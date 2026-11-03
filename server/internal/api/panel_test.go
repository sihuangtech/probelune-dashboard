package api

import (
	"encoding/json"
	"testing"

	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
)

func TestPanelExecutorOwnsChecksAndDiagnostics(t *testing.T) {
	s := testServer(t)
	router := s.Router(t.TempDir())
	token := authenticated(t, s, "admin")
	put(t, s, "node", "panel", model.Node{ID: "panel", Role: "panel", Enabled: true, Info: ct.Poll{Capabilities: []string{"tcp"}}})
	put(t, s, "node", "remote", model.Node{ID: "remote", Enabled: true, Info: ct.Poll{Capabilities: []string{"tcp"}}})
	w := request(router, token, "POST", "/api/v1/monitors", `{"name":"Port","check":{"type":"tcp","target":"example.com","port":443},"node_ids":["remote"],"interval":60,"fail_threshold":3,"recover_threshold":2,"enabled":true}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var monitor ct.Monitor
	if err := json.Unmarshal(w.Body.Bytes(), &monitor); err != nil {
		t.Fatal(err)
	}
	if len(monitor.NodeIDs) != 1 || monitor.NodeIDs[0] != "panel" {
		t.Fatalf("monitor assigned to remote: %+v", monitor.NodeIDs)
	}
	w = request(router, token, "POST", "/api/v1/diagnostics", `{"check":{"type":"tcp","target":"example.com","port":443},"node_ids":["remote"]}`)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var jobs []ct.Job
	if err := json.Unmarshal(w.Body.Bytes(), &jobs); err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].NodeID != "panel" {
		t.Fatalf("diagnostic assigned to remote: %+v", jobs)
	}
}

func TestPanelRoleCannotBeDuplicatedOrForged(t *testing.T) {
	s := testServer(t)
	router := s.Router(t.TempDir())
	token := authenticated(t, s, "admin")
	for _, tc := range []struct {
		role   string
		status int
	}{{"invalid", 400}, {"panel", 201}, {"panel", 409}} {
		w := request(router, token, "POST", "/api/v1/nodes", `{"name":"Panel","role":"`+tc.role+`","policy":{"allow_cidrs":[]}}`)
		if w.Code != tc.status {
			t.Fatalf("role %s: status %d, want %d: %s", tc.role, w.Code, tc.status, w.Body.String())
		}
	}
	operator := authenticated(t, s, "operator")
	if w := request(router, operator, "POST", "/api/v1/nodes", `{"name":"Panel","role":"panel"}`); w.Code != 403 {
		t.Fatal("operator can create a panel executor")
	}
}
