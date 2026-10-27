// Package store 管理 SQLite 持久化。含凭据的配置统一用 AES-GCM 加密。
package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type Store struct {
	DB     *sql.DB
	cipher cipher.AEAD
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	keyPath := filepath.Join(dir, "encryption.key")
	key, err := os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		if _, dbErr := os.Stat(filepath.Join(dir, "probelune.db")); dbErr == nil {
			return nil, errors.New("existing database is missing encryption.key; restore the original key")
		}
		key = make([]byte, 32)
		if _, err = rand.Read(key); err == nil {
			err = os.WriteFile(keyPath, key, 0600)
		}
	}
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "probelune.db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db, aead}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON;
 CREATE TABLE IF NOT EXISTS documents(kind TEXT NOT NULL,id TEXT NOT NULL,payload BLOB NOT NULL,PRIMARY KEY(kind,id));
 CREATE TABLE IF NOT EXISTS runs(id TEXT PRIMARY KEY,monitor TEXT NOT NULL,node TEXT NOT NULL,scheduled INTEGER NOT NULL,received INTEGER NOT NULL,status TEXT NOT NULL,latency REAL,payload TEXT NOT NULL);
 CREATE INDEX IF NOT EXISTS runs_history ON runs(monitor,node,scheduled DESC);
 CREATE TABLE IF NOT EXISTS buckets(monitor TEXT,node TEXT,bucket INTEGER,width INTEGER,success INTEGER,failed INTEGER,errors INTEGER,latency_sum REAL,latency_count INTEGER,PRIMARY KEY(monitor,node,bucket,width));
 CREATE TABLE IF NOT EXISTS meta(key TEXT PRIMARY KEY,value TEXT);
 INSERT OR IGNORE INTO meta VALUES('schema_version','1');`)
	if err != nil {
		db.Close()
		return nil, err
	}
	var version string
	if err = db.QueryRow("SELECT value FROM meta WHERE key='schema_version'").Scan(&version); err != nil || version != "1" {
		db.Close()
		return nil, errors.New("unsupported database schema; use a compatible release")
	}
	return s, nil
}

func (s *Store) Put(kind, id string, value any) error {
	return s.PutMany(Document{kind, id, value})
}

type Document struct {
	Kind, ID string
	Value    any
}

// 配置及其历史版本必须原子落盘，避免覆盖率计算与当前配置不一致。
func (s *Store) PutMany(items ...Document) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		encrypted, err := s.encode(item.Kind, item.ID, item.Value)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(`INSERT INTO documents VALUES(?,?,?) ON CONFLICT(kind,id) DO UPDATE SET payload=excluded.payload`, item.Kind, item.ID, encrypted); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) encode(kind, id string, value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, s.cipher.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.cipher.Seal(nonce, nonce, raw, []byte(kind+":"+id)), nil
}

func (s *Store) Get(kind, id string, value any) error {
	var raw []byte
	if err := s.DB.QueryRow(`SELECT payload FROM documents WHERE kind=? AND id=?`, kind, id).Scan(&raw); err != nil {
		return err
	}
	return s.decode(kind, id, raw, value)
}

func (s *Store) decode(kind, id string, raw []byte, value any) error {
	n := s.cipher.NonceSize()
	if len(raw) < n {
		return errors.New("invalid encrypted record")
	}
	plain, err := s.cipher.Open(nil, raw[:n], raw[n:], []byte(kind+":"+id))
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, value)
}

func List[T any](s *Store, kind string) ([]T, error) {
	rows, err := s.DB.Query(`SELECT id,payload FROM documents WHERE kind=? ORDER BY id`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		var id string
		var raw []byte
		if err = rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		var v T
		if err = s.decode(kind, id, raw, &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) Delete(kind, id string) error {
	_, err := s.DB.Exec(`DELETE FROM documents WHERE kind=? AND id=?`, kind, id)
	return err
}
