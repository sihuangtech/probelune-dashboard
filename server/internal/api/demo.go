package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"github.com/sihuangtech/probelune-dashboard/server/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const demoTokenPrefix = "demo-"

type demoSeed struct {
	SeedVersion string `json:"seed_version"`
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"credentials"`
	Settings struct {
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	} `json:"settings"`
	Nodes       []demoNodeSeed       `json:"nodes"`
	Monitors    []ct.Monitor         `json:"monitors"`
	History     demoHistorySeed      `json:"history"`
	Incidents   []demoIncidentSeed   `json:"incidents"`
	Diagnostics []demoDiagnosticSeed `json:"diagnostics"`
}

type demoNodeSeed struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	Network      string   `json:"network"`
	Capabilities []string `json:"capabilities"`
}

type demoHistorySeed struct {
	Hours           int                 `json:"hours"`
	IntervalSeconds int                 `json:"interval_seconds"`
	BaseLatencyMS   int                 `json:"base_latency_ms"`
	JitterMS        int                 `json:"jitter_ms"`
	Failures        []demoFailureSeries `json:"failures"`
}

type demoFailureSeries struct {
	MonitorID   string `json:"monitor_id"`
	NodeID      string `json:"node_id"`
	LastSamples int    `json:"last_samples"`
}

type demoIncidentSeed struct {
	ID               string `json:"id"`
	MonitorID        string `json:"monitor_id"`
	NodeID           string `json:"node_id"`
	Name             string `json:"name"`
	OpenedMinutesAgo int    `json:"opened_minutes_ago"`
	Message          string `json:"message"`
}

type demoDiagnosticSeed struct {
	ID         string  `json:"id"`
	BatchID    string  `json:"batch_id"`
	MonitorID  string  `json:"monitor_id"`
	NodeID     string  `json:"node_id"`
	Type       string  `json:"type"`
	Target     string  `json:"target"`
	Port       int     `json:"port"`
	Timeout    int     `json:"timeout"`
	Status     string  `json:"status"`
	LatencyMS  float64 `json:"latency_ms"`
	MinutesAgo int     `json:"minutes_ago"`
}

// NewDemo opens an isolated database and loads its fixture from demo_seed.json.
// The JSON is copied to the demo data directory so operators can edit it without
// changing Go code; increment seed_version to rebuild only the demo database.
func NewDemo(dir, seedPath string) (*Server, error) {
	s, err := New(dir)
	if err != nil {
		return nil, err
	}
	s.DemoMode = true
	s.DemoTokenPrefix = demoTokenPrefix
	s.SetupToken = ""
	_ = os.Remove(filepath.Join(dir, "setup-token"))
	seed, err := loadDemoSeed(dir, seedPath)
	if err == nil {
		s.DemoUsername = seed.Credentials.Username
		err = seedDemo(s, seed)
	}
	if err != nil {
		s.Store.DB.Close()
		return nil, err
	}
	return s, nil
}

func loadDemoSeed(dir, seedPath string) (demoSeed, error) {
	path := filepath.Join(dir, "seed.json")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		raw, err = os.ReadFile(seedPath)
		if err != nil {
			return demoSeed{}, fmt.Errorf("演示数据文件不可读 %s: %w", seedPath, err)
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			return demoSeed{}, err
		}
	} else if err != nil {
		return demoSeed{}, err
	}
	var seed demoSeed
	if err = json.Unmarshal(raw, &seed); err != nil {
		return demoSeed{}, fmt.Errorf("演示 seed.json 格式无效: %w", err)
	}
	if seed.SeedVersion == "" || seed.Credentials.Username == "" || seed.Credentials.Password == "" || len(seed.Nodes) == 0 || len(seed.Monitors) == 0 || seed.History.Hours < 1 || seed.History.IntervalSeconds < 15 {
		return demoSeed{}, fmt.Errorf("演示 seed.json 缺少版本、登录信息、节点、监测或有效历史配置")
	}
	return seed, nil
}

func seedDemo(s *Server, seed demoSeed) error {
	var previous string
	err := s.Store.DB.QueryRow(`SELECT value FROM meta WHERE key='demo_seed_version'`).Scan(&previous)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if previous != seed.SeedVersion {
		if _, err = s.Store.DB.Exec(`DELETE FROM documents; DELETE FROM runs; DELETE FROM buckets; DELETE FROM meta WHERE key='demo_seed_version'`); err != nil {
			return err
		}
	}

	now := Now()
	var cfg model.Settings
	if err = s.Store.Get("settings", "instance", &cfg); err != nil {
		return err
	}
	cfg.Name, cfg.PublicURL, cfg.Timezone = seed.Settings.Name, "", seed.Settings.Timezone
	if cfg.Name == "" {
		cfg.Name = "Probelune Demo"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Shanghai"
	}
	if err = s.Store.Put("settings", "instance", cfg); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(seed.Credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	demoUser := model.User{ID: "demo-user", Username: seed.Credentials.Username, Password: string(hash), Role: "viewer"}
	if err = s.Store.Put("user", demoUser.ID, demoUser); err != nil {
		return err
	}
	for _, n := range seed.Nodes {
		node := model.Node{ID: n.ID, Role: "node", Name: n.Name, Region: n.Region, Network: n.Network, Enabled: true, CreatedAt: now - 30*86400, LastSeen: now, Info: ct.Poll{Protocol: ct.Protocol, Version: ct.Version, Platform: "demo", Capabilities: n.Capabilities, Time: now}}
		if err = s.Store.Put("node", node.ID, node); err != nil {
			return err
		}
	}
	updatedAt := now - int64(seed.History.Hours+24)*3600
	for i := range seed.Monitors {
		m := seed.Monitors[i]
		m.Version = max(m.Version, 1)
		m.UpdatedAt = updatedAt
		m.ChannelIDs = []string{}
		m.Tags = append(m.Tags, "demo")
		if err = s.Store.PutMany(store.Document{Kind: "monitor", ID: m.ID, Value: m}, store.Document{Kind: "monitor_version", ID: fmt.Sprintf("%s:%d", m.ID, m.Version), Value: m}); err != nil {
			return err
		}
		for _, nodeID := range m.NodeIDs {
			status := demoCurrentStatus(seed.History.Failures, m.ID, nodeID)
			result := demoResult(m, nodeID, now, status, 20)
			state := model.State{ID: m.ID + ":" + nodeID, MonitorID: m.ID, NodeID: nodeID, Version: m.Version, Status: status, LastAt: now, Result: result}
			if status == "failed" {
				state.Failures = m.FailThreshold
			} else {
				state.Successes = m.RecoverThreshold
			}
			if err = s.Store.Put("state", state.ID, state); err != nil {
				return err
			}
		}
	}
	if err = seedDemoRuns(s, seed); err != nil {
		return err
	}
	for _, incident := range seed.Incidents {
		v := model.Incident{ID: incident.ID, MonitorID: incident.MonitorID, NodeID: incident.NodeID, Name: incident.Name, OpenedAt: now - int64(incident.OpenedMinutesAgo)*60, Affected: []string{incident.NodeID}, Events: []model.Event{{At: now - int64(incident.OpenedMinutesAgo)*60, Kind: "opened", Message: incident.Message, User: seed.Credentials.Username}}}
		if err = s.Store.Put("incident", v.ID, v); err != nil {
			return err
		}
	}
	for _, sample := range seed.Diagnostics {
		created := now - int64(sample.MinutesAgo)*60
		job := ct.Job{ID: sample.ID, BatchID: sample.BatchID, MonitorID: sample.MonitorID, NodeID: sample.NodeID, Check: ct.Check{Type: sample.Type, Target: sample.Target, Port: sample.Port, Timeout: sample.Timeout}, Status: "done", CreatedAt: created, Deadline: created + 120}
		latency := sample.LatencyMS
		result := ct.Result{ID: sample.ID, MonitorID: sample.MonitorID, JobID: sample.ID, NodeID: sample.NodeID, Version: 1, ScheduledAt: created, StartedAt: created, ReceivedAt: created + 2, Status: sample.Status, Stage: "response", Latency: &latency, Data: map[string]any{"synthetic": true}}
		if err = s.Store.Put("job", job.ID, job); err != nil {
			return err
		}
		if err = s.Store.Put("diagnostic_result", job.ID, result); err != nil {
			return err
		}
	}
	_, err = s.Store.DB.Exec(`INSERT INTO meta(key,value) VALUES('demo_seed_version',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, seed.SeedVersion)
	return err
}

func demoCurrentStatus(failures []demoFailureSeries, monitorID, nodeID string) string {
	for _, series := range failures {
		if series.MonitorID == monitorID && series.NodeID == nodeID && series.LastSamples > 0 {
			return "failed"
		}
	}
	return "healthy"
}

func demoResult(m ct.Monitor, nodeID string, at int64, status string, latencyMS float64) ct.Result {
	var latency *float64
	if status == "success" || status == "healthy" {
		latency = &latencyMS
	}
	if status == "healthy" {
		status = "success"
	}
	return ct.Result{ID: m.ID + ":" + nodeID + ":latest", MonitorID: m.ID, NodeID: nodeID, Version: m.Version, ScheduledAt: at, StartedAt: at, ReceivedAt: at, Status: status, Stage: "response", Latency: latency, Data: map[string]any{"synthetic": true}}
}

func seedDemoRuns(s *Server, seed demoSeed) error {
	step := int64(seed.History.IntervalSeconds)
	end := Now() / step * step
	start := end - int64(seed.History.Hours)*3600
	for _, m := range seed.Monitors {
		for ni, nodeID := range m.NodeIDs {
			var failedSamples int
			for _, series := range seed.History.Failures {
				if series.MonitorID == m.ID && series.NodeID == nodeID {
					failedSamples = series.LastSamples
				}
			}
			for at, index := start, 0; at <= end; at, index = at+step, index+1 {
				status := "success"
				if failedSamples > 0 && at > end-int64(failedSamples)*step {
					status = "failure"
				}
				latency := float64(seed.History.BaseLatencyMS + (index+ni)%max(seed.History.JitterMS, 1))
				result := demoResult(m, nodeID, at, status, latency)
				result.ID = fmt.Sprintf("%s:%s:%d", m.ID, nodeID, at)
				result.StartedAt, result.ReceivedAt = at+1, at+2
				raw, err := json.Marshal(result)
				if err != nil {
					return err
				}
				var latencyArg any = latency
				if status == "failure" {
					latencyArg = nil
				}
				if _, err = s.Store.DB.Exec(`INSERT OR REPLACE INTO runs VALUES(?,?,?,?,?,?,?,?)`, result.ID, m.ID, nodeID, at, result.ReceivedAt, status, latencyArg, string(raw)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Server) TouchDemo() {
	if !s.DemoMode {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := Now()
	if now-s.demoTouched < 15 {
		return
	}
	s.demoTouched = now
	for _, n := range all[model.Node](s, "node") {
		n.LastSeen, n.Info.Time = now, now
		_ = s.Store.Put("node", n.ID, n)
	}
	for _, m := range all[ct.Monitor](s, "monitor") {
		for _, nodeID := range m.NodeIDs {
			var state model.State
			if s.Store.Get("state", m.ID+":"+nodeID, &state) == nil {
				state.LastAt = now
				state.Result.ScheduledAt, state.Result.StartedAt, state.Result.ReceivedAt = now, now, now
				_ = s.Store.Put("state", state.ID, state)
			}
		}
	}
}

// RouterWithDemo dispatches demo credentials and session tokens to a separate
// API and database while keeping the public site and mobile origin unchanged.
func (s *Server) RouterWithDemo(web string, demo *Server) http.Handler {
	production := s.Router(web)
	if demo == nil {
		return production
	}
	demo.DemoAvailable = true
	demoRouter := demo.Router(web)
	s.DemoAvailable = true
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/demo/login" || isDemoLogin(r, demo.DemoUsername) || hasDemoSession(r, demo.DemoTokenPrefix) {
			demo.TouchDemo()
			demoRouter.ServeHTTP(w, r)
			return
		}
		production.ServeHTTP(w, r)
	})
}

func isDemoLogin(r *http.Request, username string) bool {
	if r.Method != http.MethodPost || r.URL.Path != "/api/v1/auth/login" {
		return false
	}
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	r.Body = io.NopCloser(bytes.NewReader(raw))
	var in struct {
		Username string `json:"username"`
	}
	return json.Unmarshal(raw, &in) == nil && in.Username == username
}

func hasDemoSession(r *http.Request, prefix string) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		if cookie, err := r.Cookie("probelune_session"); err == nil {
			token = cookie.Value
		}
	}
	return prefix != "" && strings.HasPrefix(token, prefix)
}
