package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Store.DB.Close() })
	return s
}

func put(t *testing.T, s *Server, kind, id string, value any) {
	t.Helper()
	if err := s.Store.Put(kind, id, value); err != nil {
		t.Fatal(err)
	}
}

func rawSample(t *testing.T, s *Server, id, monitor string, at int64, status string) {
	t.Helper()
	r := ct.Result{ID: id, MonitorID: monitor, NodeID: "node", ScheduledAt: at, Status: status, Stage: "response"}
	data, _ := json.Marshal(r)
	if _, err := s.Store.DB.Exec(`INSERT INTO runs VALUES(?,?,?,?,?,?,?,?)`, id, monitor, "node", at, at, status, nil, string(data)); err != nil {
		t.Fatal(err)
	}
}

func TestUptimeUsesArchivedSamplesAndPreservesGaps(t *testing.T) {
	s := testServer(t)
	from := Now()/900*900 - 3600
	m := ct.Monitor{ID: "m", Interval: 60, Enabled: true, Version: 1, UpdatedAt: from, NodeIDs: []string{"node"}}
	rawSample(t, s, "ok", m.ID, from, "success")
	rawSample(t, s, "bad", m.ID, from+900, "failure")
	rawSample(t, s, "error", m.ID, from+2700, "error")
	if _, err := s.Store.DB.Exec(`INSERT INTO buckets VALUES(?,?,?,?,?,?,?,?,?)`, m.ID, "node", from+300, 300, 5, 0, 0, 10, 5); err != nil {
		t.Fatal(err)
	}
	v, err := s.uptimeSeries(m, []ct.Monitor{m}, from, from+3599, 900)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Cells) != 4 || v.Cells[0].Success != 6 || v.Cells[0].Status != "incomplete" || v.Cells[1].Status != "down" || v.Cells[2].Status != "unknown" || v.Cells[3].Status != "unknown" {
		t.Fatalf("incorrect archive/gap semantics: %+v", v)
	}
	if v.Cells[2].Availability != nil || *v.Cells[2].Coverage != 0 {
		t.Fatal("missing samples became healthy")
	}
	if v.Availability == nil || math.Abs(*v.Availability-600.0/7) > 0.00001 || *v.Coverage >= 100 {
		t.Fatalf("invalid summary: %+v", v)
	}
}

func TestExpectedSamplesRespectPauseAndMembership(t *testing.T) {
	m := ct.Monitor{Interval: 60, Enabled: true, UpdatedAt: 600, Version: 1, NodeIDs: []string{"a", "b"}}
	paused := m
	paused.Version = 2
	paused.UpdatedAt = 1200
	paused.Enabled = false
	if n := expectedVersionRuns([]ct.Monitor{paused, m}, "", 600, 1799); n != 20 {
		t.Fatalf("paused interval counted: %d", n)
	}
}

func TestPublicStatusRedactionAndCustomResolution(t *testing.T) {
	s := testServer(t)
	put(t, s, "settings", "instance", model.Settings{Name: "Fixture", PublicStatus: true})
	m := ct.Monitor{ID: "public", Name: "Public API", Public: true, Enabled: true, Interval: 60, Check: ct.Check{Target: "https://private.internal/?secret=x"}, NodeIDs: []string{"private-node"}}
	put(t, s, "monitor", m.ID, m)
	m.ID = "secret-monitor"
	m.Name = "Secret database"
	m.Public = false
	put(t, s, "monitor", m.ID, m)
	router := s.Router(t.TempDir())
	for _, minutes := range []int{15, 60, 35} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status?hours=24&bucket_minutes=%d", minutes), nil))
		if w.Code != 200 {
			t.Fatalf("custom interval rejected: %d %s", w.Code, w.Body.String())
		}
		for _, secret := range []string{"private.internal", "secret-monitor", "Secret database", "private-node"} {
			if strings.Contains(w.Body.String(), secret) {
				t.Fatalf("public leak: %s", secret)
			}
		}
		var response struct {
			Monitors []struct {
				Uptime uptimeSeries `json:"uptime"`
			} `json:"monitors"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if len(response.Monitors) != 1 || response.Monitors[0].Uptime.Bucket != int64(minutes*60) {
			t.Fatal("resolution ignored")
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/status?hours=720&bucket_minutes=5", nil))
	if w.Code != 400 {
		t.Fatal("unbounded response accepted")
	}
}

func authenticated(t *testing.T, s *Server, role string) string {
	t.Helper()
	token := ID()
	u := model.User{ID: ID(), Username: role, Role: role}
	put(t, s, "user", u.ID, u)
	put(t, s, "session", Hash(token), model.Session{ID: Hash(token), UserID: u.ID, Expires: Now() + 3600})
	return token
}

func request(router *gin.Engine, token, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, r)
	return w
}

func TestDiagnosticBatchesAndLateCancelledResults(t *testing.T) {
	s := testServer(t)
	router := s.Router(t.TempDir())
	token := authenticated(t, s, "admin")
	for _, id := range []string{"a", "b"} {
		put(t, s, "node", id, model.Node{ID: id, Enabled: true, Info: ct.Poll{Capabilities: []string{"udp"}}})
	}
	body := `{"check":{"type":"udp","target":"example.com","port":9000,"udp_payload":"ping"},"node_ids":["a","b"]}`
	first := request(router, token, "POST", "/api/v1/diagnostics", body)
	if first.Code != 201 {
		t.Fatal(first.Body.String())
	}
	var jobs []ct.Job
	json.Unmarshal(first.Body.Bytes(), &jobs)
	if len(jobs) != 2 || jobs[0].BatchID == "" || jobs[0].BatchID != jobs[1].BatchID {
		t.Fatal("batch not retained")
	}
	request(router, token, "POST", "/api/v1/diagnostics", body)
	filtered := request(router, token, "GET", "/api/v1/diagnostics?batch_id="+jobs[0].BatchID, "")
	var rows []json.RawMessage
	json.Unmarshal(filtered.Body.Bytes(), &rows)
	if len(rows) != 2 {
		t.Fatalf("mixed diagnostic runs: %s", filtered.Body.String())
	}
	viewer := authenticated(t, s, "viewer")
	if w := request(router, viewer, "POST", "/api/v1/diagnostics", body); w.Code != 403 {
		t.Fatal("viewer can execute")
	}
	request(router, token, "POST", "/api/v1/diagnostics/"+jobs[0].ID+"/cancel", `{}`)
	if !s.acceptDiagnostic(ct.Result{ID: jobs[0].ID, JobID: jobs[0].ID, NodeID: "a", Status: "success"}) {
		t.Fatal("late result was not acknowledged")
	}
	var job ct.Job
	s.Store.Get("job", jobs[0].ID, &job)
	if job.Status != "cancelled" {
		t.Fatal("late upload undid cancellation")
	}
}

func TestHistoryPaginationAndTimeSelection(t *testing.T) {
	s := testServer(t)
	router := s.Router(t.TempDir())
	token := authenticated(t, s, "viewer")
	put(t, s, "monitor", "m", ct.Monitor{ID: "m", Interval: 60, NodeIDs: []string{"node"}})
	for i := 0; i < 125; i++ {
		rawSample(t, s, fmt.Sprint(i), "m", Now()-int64(i*60), "failure")
	}
	w := request(router, token, "GET", "/api/v1/monitors/m/history?status=failure&stage=response&limit=50&offset=50", "")
	var v struct {
		Total   int         `json:"total"`
		Results []ct.Result `json:"results"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &v) != nil || v.Total != 125 || len(v.Results) != 50 {
		t.Fatalf("pagination failed: %s", w.Body.String())
	}
}
