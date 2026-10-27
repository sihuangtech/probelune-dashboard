package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"strings"
)

func (s *Server) receive(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var batch []ct.Result
	if !bind(c, &batch) {
		return
	}
	if len(batch) > 100 {
		fail(c, 400, "每批最多 100 条结果")
		return
	}
	nodeID := c.GetString("node_id")
	var node model.Node
	if s.Store.Get("node", nodeID, &node) != nil || node.Revoked || node.TokenHash != Hash(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")) {
		fail(c, 401, "节点身份已撤销")
		return
	}
	accepted := []string{}
	rejected := []string{}
	for _, r := range batch {
		if !contains([]string{"success", "failure", "error", "cancelled"}, r.Status) || r.NodeID != nodeID || r.ScheduledAt > Now()+30 || r.StartedAt > Now()+30 || r.StartedAt < r.ScheduledAt-30 {
			rejected = append(rejected, r.ID)
			continue
		}
		r.ReceivedAt = Now()
		r.Error = ct.Redact(r.Error)
		encoded, sizeErr := json.Marshal(r)
		if sizeErr != nil || len(encoded) > 65536 {
			rejected = append(rejected, r.ID)
			continue
		}
		if r.JobID != "" {
			if s.acceptDiagnostic(r) {
				accepted = append(accepted, r.ID)
			} else {
				rejected = append(rejected, r.ID)
			}
			continue
		}
		var a ct.Assignment
		if s.Store.Get("assignment", r.AssignmentID, &a) != nil || a.NodeID != nodeID || a.Monitor.ID != r.MonitorID || a.Monitor.Version != r.Version || r.ScheduledAt < a.Start || r.ScheduledAt > a.Until || r.ScheduledAt%int64(a.Monitor.Interval) != 0 || r.ID != ct.RunID(r.MonitorID, nodeID, r.Version, r.ScheduledAt) {
			rejected = append(rejected, r.ID)
			continue
		}
		r.Late = Now()-r.ScheduledAt > int64(2*a.Monitor.Interval+a.Monitor.Check.Timeout)
		var current ct.Monitor
		currentPolicy, _ := json.Marshal(node.Policy)
		leasePolicy, _ := json.Marshal(a.Policy)
		if s.Store.Get("monitor", r.MonitorID, &current) != nil || current.Version != r.Version || string(currentPolicy) != string(leasePolicy) {
			r.Late = true
		}
		raw, e := json.Marshal(r)
		if e != nil || len(raw) > 65536 {
			rejected = append(rejected, r.ID)
			continue
		}
		_, e = s.Store.DB.Exec(`INSERT OR IGNORE INTO runs VALUES(?,?,?,?,?,?,?,?)`, r.ID, r.MonitorID, nodeID, r.ScheduledAt, r.ReceivedAt, r.Status, r.Latency, string(raw))
		if e != nil {
			fail(c, 500, "结果保存失败，请重试")
			return
		}
		if !r.Late {
			e = s.updateState(a.Monitor, r)
			if e != nil {
				fail(c, 500, "状态保存失败，请重试")
				return
			}
		}
		accepted = append(accepted, r.ID)
	}
	s.signal()
	c.JSON(200, gin.H{"accepted": accepted, "rejected": rejected})
}

func (s *Server) acceptDiagnostic(r ct.Result) bool {
	var j ct.Job
	if s.Store.Get("job", r.JobID, &j) != nil || j.NodeID != r.NodeID || r.ID != j.ID {
		return false
	}
	if j.Status == "cancelled" || j.Status == "done" {
		return true
	}
	if Now() > j.Deadline+60 {
		return false
	}
	if e := s.Store.Put("diagnostic_result", j.ID, r); e != nil {
		return false
	}
	j.Status = "done"
	return s.Store.Put("job", j.ID, j) == nil
}

func (s *Server) updateState(m ct.Monitor, r ct.Result) error {
	key := m.ID + ":" + r.NodeID
	var state model.State
	s.Store.Get("state", key, &state)
	if state.LastAt >= r.ScheduledAt {
		return nil
	}
	end := s.maintenanceEnded(m.ID)
	if state.Version != r.Version || r.ScheduledAt-state.LastAt > int64(m.Interval+m.Check.Timeout) || (state.LastAt < end && r.ScheduledAt >= end) {
		state = model.State{Status: "unknown"}
	}
	state.ID = key
	state.MonitorID = m.ID
	state.NodeID = r.NodeID
	state.Version = r.Version
	state.LastAt = r.ScheduledAt
	state.Result = r
	switch r.Status {
	case "success":
		state.Successes++
		state.Failures = 0
		if state.Successes >= m.RecoverThreshold {
			state.Status = "healthy"
		} else if state.Status != "failed" {
			state.Status = "pending"
		}
	case "failure":
		state.Failures++
		state.Successes = 0
		if state.Failures >= m.FailThreshold {
			state.Status = "failed"
		} else if state.Status != "healthy" {
			state.Status = "pending"
		}
	default:
		state.Failures = 0
		state.Successes = 0
		state.Status = "unknown"
	}
	return s.Store.Put("state", key, state)
}
