package api

import (
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"sort"
	"strings"
)

func (s *Server) maintained(id string) bool {
	now := Now()
	for _, w := range all[model.Maintenance](s, "maintenance") {
		if w.Start <= now && w.End > now && (len(w.MonitorIDs) == 0 || contains(w.MonitorIDs, id)) {
			return true
		}
	}
	return false
}

// 维护结束后重新累计确认阈值，不能沿用维护期内的失败/成功计数。
func (s *Server) maintenanceEnded(id string) int64 {
	var end int64
	for _, w := range all[model.Maintenance](s, "maintenance") {
		if w.End <= Now() && w.End > end && (len(w.MonitorIDs) == 0 || contains(w.MonitorIDs, id)) {
			end = w.End
		}
	}
	return end
}
func (s *Server) incidents(c *gin.Context) {
	items := all[model.Incident](s, "incident")
	sort.Slice(items, func(i, j int) bool { return items[i].OpenedAt > items[j].OpenedAt })
	if len(items) > 500 {
		items = items[:500]
	}
	c.JSON(200, items)
}
func (s *Server) incidentAction(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var v model.Incident
	if s.Store.Get("incident", c.Param("id"), &v) != nil {
		fail(c, 404, "事件不存在")
		return
	}
	var in struct {
		Action  string `json:"action"`
		Minutes int    `json:"minutes"`
		Note    string `json:"note"`
	}
	if !bind(c, &in) {
		return
	}
	switch in.Action {
	case "ack":
		v.AckBy = c.GetString("user_id")
	case "silence":
		if in.Minutes < 1 || in.Minutes > 1440 {
			fail(c, 400, "静默范围为 1—1440 分钟")
			return
		}
		v.SilencedUntil = Now() + int64(in.Minutes)*60
	case "note":
		if strings.TrimSpace(in.Note) == "" || len(in.Note) > 2000 {
			fail(c, 400, "备注不能为空且最多 2000 字符")
			return
		}
	default:
		fail(c, 400, "不支持的操作")
		return
	}
	v.Events = append(v.Events, model.Event{At: Now(), Kind: in.Action, Message: in.Note, User: c.GetString("user_id")})
	if s.save(c, "incident", v.ID, v) {
		s.audit(c, "incident."+in.Action, v.ID)
		c.JSON(200, v)
	}
}

func (s *Server) reconcile() {
	incidents := all[model.Incident](s, "incident")
	for _, m := range all[ct.Monitor](s, "monitor") {
		if !m.Enabled || m.Archived {
			continue
		}
		view := s.monitorView(m)
		states := view["states"].([]model.State)
		failed := []string{}
		allHealthy := true
		for _, v := range states {
			if v.Status == "failed" {
				failed = append(failed, v.NodeID)
			}
			if v.Status != "healthy" {
				allHealthy = false
			}
		}
		var current *model.Incident
		for i := range incidents {
			if incidents[i].MonitorID == m.ID && incidents[i].RecoveredAt == 0 {
				current = &incidents[i]
				break
			}
		}
		if current == nil && len(failed) >= m.MinFailedNodes {
			v := model.Incident{ID: ID(), MonitorID: m.ID, Name: m.Name, OpenedAt: Now(), Affected: failed, Suppressed: s.maintained(m.ID), Events: []model.Event{{At: Now(), Kind: "opened", Message: "节点达到连续失败确认阈值"}}}
			if s.Store.Put("incident", v.ID, v) == nil && !v.Suppressed {
				s.enqueue(v, "opened", m.ChannelIDs)
			}
		} else if current != nil {
			if len(failed) >= m.MinFailedNodes {
				s.remind(*current, m.ChannelIDs)
			}
			if allHealthy {
				current.RecoveredAt = Now()
				current.Events = append(current.Events, model.Event{At: Now(), Kind: "recovered", Message: "所有已配置节点均已确认恢复"})
				if s.Store.Put("incident", current.ID, current) == nil && !current.Suppressed && !s.maintained(m.ID) && current.SilencedUntil < Now() {
					s.enqueue(*current, "recovered", m.ChannelIDs)
				}
			} else {
				sort.Strings(failed)
				sort.Strings(current.Affected)
				changed := strings.Join(failed, ",") != strings.Join(current.Affected, ",")
				if changed {
					current.Affected = failed
					current.Events = append(current.Events, model.Event{At: Now(), Kind: "coverage", Message: "失败节点范围变化；缺测不代表恢复"})
				}
				if current.Suppressed && !s.maintained(m.ID) && len(failed) >= m.MinFailedNodes {
					current.Suppressed = false
					s.enqueue(*current, "opened", m.ChannelIDs)
					changed = true
				}
				if changed {
					s.Store.Put("incident", current.ID, current)
				}
			}
		}
	}
	for _, n := range all[model.Node](s, "node") {
		if !n.Enabled || n.Revoked || n.LastSeen == 0 {
			continue
		}
		var active *model.Incident
		for i := range incidents {
			if incidents[i].NodeID == n.ID && incidents[i].RecoveredAt == 0 {
				active = &incidents[i]
				break
			}
		}
		if Now()-n.LastSeen > 60 && active == nil {
			v := model.Incident{ID: ID(), NodeID: n.ID, Name: n.Name, OpenedAt: Now(), Events: []model.Event{{At: Now(), Kind: "offline", Message: "节点离线，监测覆盖不足"}}}
			if s.Store.Put("incident", v.ID, v) == nil {
				s.enqueue(v, "node_offline", nil)
			}
		} else if Now()-n.LastSeen <= 60 && active != nil {
			active.RecoveredAt = Now()
			active.Events = append(active.Events, model.Event{At: Now(), Kind: "online", Message: "节点重新连接"})
			if s.Store.Put("incident", active.ID, active) == nil {
				s.enqueue(*active, "node_online", nil)
			}
		}
	}
}
