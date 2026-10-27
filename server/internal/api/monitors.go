package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"github.com/sihuangtech/probelune-dashboard/server/internal/store"
	"sort"
)

func cleanMonitor(m ct.Monitor) gin.H {
	has := len(m.Check.Headers) > 0
	m.Check.Headers = nil
	raw, _ := json.Marshal(m)
	out := gin.H{}
	json.Unmarshal(raw, &out)
	out["has_headers"] = has
	return out
}
func (s *Server) saveMonitor(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var m ct.Monitor
	if !bind(c, &m) {
		return
	}
	var old ct.Monitor
	id := c.Param("id")
	if id != "" {
		if s.Store.Get("monitor", id, &old) != nil {
			fail(c, 404, "监测不存在")
			return
		}
		if m.Version != old.Version {
			fail(c, 409, "配置已被修改，请刷新")
			return
		}
		m.ID = id
		m.Version = old.Version + 1
		if m.Check.Headers == nil {
			m.Check.Headers = old.Check.Headers
		}
	} else {
		m.ID = ID()
		m.Version = 1
	}
	if err := m.Check.Validate(); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if !contains([]string{"http", "tcp", "ping", "dns"}, m.Check.Type) {
		fail(c, 400, "持续监测仅支持 HTTP、TCP、ICMP、DNS")
		return
	}
	if m.Name == "" || len(m.Name) > 100 || len(m.NodeIDs) == 0 || len(m.NodeIDs) > 50 || m.Interval < 15 || m.Interval > 86400 || m.Interval <= m.Check.Timeout || m.FailThreshold < 1 || m.FailThreshold > 20 || m.RecoverThreshold < 1 || m.RecoverThreshold > 20 {
		fail(c, 400, "检查名称、节点、间隔和确认阈值；间隔必须大于截止时间")
		return
	}
	if m.MinFailedNodes == 0 {
		m.MinFailedNodes = 1
	}
	if m.MinFailedNodes > len(m.NodeIDs) || m.MinFailedNodes < 1 {
		fail(c, 400, "故障节点阈值不合法")
		return
	}
	seen := map[string]bool{}
	for _, id := range m.NodeIDs {
		var n model.Node
		if seen[id] || s.Store.Get("node", id, &n) != nil || n.Revoked {
			fail(c, 400, "包含重复或不可用节点")
			return
		}
		seen[id] = true
	}
	for _, id := range m.ChannelIDs {
		var channel model.Channel
		if s.Store.Get("channel", id, &channel) != nil {
			fail(c, 400, "通知渠道不存在")
			return
		}
	}
	m.UpdatedAt = Now()
	if err := s.Store.PutMany(store.Document{Kind: "monitor", ID: m.ID, Value: m}, store.Document{Kind: "monitor_version", ID: fmt.Sprintf("%s:%d", m.ID, m.Version), Value: m}); err != nil {
		fail(c, 500, "监测配置保存失败")
		return
	}
	s.audit(c, "monitor.save", m.ID)
	c.JSON(200, cleanMonitor(m))
}
func (s *Server) monitorView(m ct.Monitor) gin.H {
	out := cleanMonitor(m)
	states := []model.State{}
	normal, failed, known := 0, 0, 0
	now := Now()
	maintenanceEnd := s.maintenanceEnded(m.ID)
	for _, id := range m.NodeIDs {
		var state model.State
		var node model.Node
		if s.Store.Get("state", m.ID+":"+id, &state) != nil {
			state = model.State{ID: m.ID + ":" + id, MonitorID: m.ID, NodeID: id, Status: "unknown"}
		}
		if s.Store.Get("node", id, &node) != nil || node.Revoked || !node.Enabled || now-node.LastSeen > 60 || now-state.LastAt > int64(2*m.Interval+m.Check.Timeout) || state.Version != m.Version || state.LastAt < maintenanceEnd {
			state.Status = "unknown"
		}
		if state.Status == "healthy" {
			normal++
			known++
		}
		if state.Status == "failed" {
			failed++
			known++
		}
		states = append(states, state)
	}
	status := "unknown"
	if normal == len(m.NodeIDs) {
		status = "healthy"
	} else if failed == len(m.NodeIDs) {
		status = "down"
	} else if failed > 0 {
		status = "partial"
	} else if known > 0 {
		status = "pending"
	}
	if m.Archived {
		status = "archived"
	} else if !m.Enabled {
		status = "paused"
	}
	out["status"] = status
	out["states"] = states
	out["coverage_complete"] = known == len(m.NodeIDs)
	out["maintenance"] = s.maintained(m.ID)
	return out
}
func (s *Server) monitors(c *gin.Context) {
	out := []gin.H{}
	for _, m := range all[ct.Monitor](s, "monitor") {
		out = append(out, s.monitorView(m))
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]["name"]) < fmt.Sprint(out[j]["name"]) })
	c.JSON(200, out)
}
func (s *Server) overview(c *gin.Context) {
	out := []gin.H{}
	summary := map[string]int{"healthy": 0, "partial": 0, "down": 0, "unknown": 0, "paused": 0, "pending": 0}
	for _, m := range all[ct.Monitor](s, "monitor") {
		if m.Archived {
			continue
		}
		view := s.monitorView(m)
		summary[view["status"].(string)]++
		out = append(out, view)
	}
	online := 0
	nodes := []model.Node{}
	for _, n := range all[model.Node](s, "node") {
		nodes = append(nodes, publicNode(n))
		if !n.Revoked && n.Enabled && Now()-n.LastSeen <= 60 {
			online++
		}
	}
	incidents := []model.Incident{}
	for _, v := range all[model.Incident](s, "incident") {
		if v.RecoveredAt == 0 {
			incidents = append(incidents, v)
		}
	}
	var cfg model.Settings
	s.Store.Get("settings", "instance", &cfg)
	c.JSON(200, gin.H{"instance_id": cfg.ID, "monitors": out, "summary": summary, "nodes": nodes, "online_nodes": online, "incidents": incidents, "time": Now()})
}
func (s *Server) publicStatus(c *gin.Context) {
	var cfg model.Settings
	s.Store.Get("settings", "instance", &cfg)
	if !cfg.PublicStatus {
		fail(c, 404, "状态页未启用")
		return
	}
	list := []gin.H{}
	for _, m := range all[ct.Monitor](s, "monitor") {
		if m.Public && !m.Archived {
			v := s.monitorView(m)
			list = append(list, gin.H{"id": m.ID, "name": m.Name, "status": v["status"], "coverage_complete": v["coverage_complete"], "maintenance": v["maintenance"]})
		}
	}
	c.JSON(200, gin.H{"name": cfg.Name, "monitors": list, "updated_at": Now()})
}
