// Package backup 使用 SQLite 在线备份快照，并同时保存解密必需的密钥。
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sihuangtech/probelune-dashboard/server/internal/store"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Run(source, destination string) error {
	if _, err := os.Stat(filepath.Join(source, "probelune.db")); err != nil {
		return fmt.Errorf("source database does not exist: %w", err)
	}
	abs, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	// 不覆盖既有备份；失败也保留目录，供操作者检查。
	if err = os.Mkdir(abs, 0700); err != nil {
		return err
	}
	db, err := store.Open(source)
	if err != nil {
		return err
	}
	defer db.DB.Close()
	if _, err = db.DB.Exec("VACUUM INTO ?", filepath.Join(abs, "probelune.db")); err != nil {
		return err
	}
	key, err := os.ReadFile(filepath.Join(source, "encryption.key"))
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(abs, "encryption.key"), key, 0600); err != nil {
		return err
	}
	if err = os.Chmod(filepath.Join(abs, "probelune.db"), 0600); err != nil {
		return err
	}
	files := map[string]string{}
	for _, name := range []string{"probelune.db", "encryption.key"} {
		file, e := os.Open(filepath.Join(abs, name))
		if e != nil {
			return e
		}
		hash := sha256.New()
		_, e = io.Copy(hash, file)
		file.Close()
		if e != nil {
			return e
		}
		files[name] = hex.EncodeToString(hash.Sum(nil))
	}
	manifest, _ := json.MarshalIndent(map[string]any{"format": 1, "created_at": time.Now().UTC(), "sha256": files}, "", "  ")
	return os.WriteFile(filepath.Join(abs, "manifest.json"), manifest, 0600)
}
