package api

import (
	"fmt"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"log/slog"
)

// 仅提醒当前仍确认的异常；稳定时间槽避免重启后回放历史提醒。
func (s *Server) remind(v model.Incident, ids []string) {
	if v.Suppressed || v.SilencedUntil > Now() || s.maintained(v.MonitorID) {
		return
	}
	for _, id := range ids {
		var ch model.Channel
		if s.Store.Get("channel", id, &ch) != nil || !ch.Enabled || ch.ReminderMinutes == 0 {
			continue
		}
		slot := (Now() - v.OpenedAt) / int64(ch.ReminderMinutes*60)
		if slot < 1 {
			continue
		}
		key := fmt.Sprintf("%s:reminder:%s:%d", v.ID, id, slot)
		var existing model.Delivery
		if s.Store.Get("delivery", key, &existing) == nil {
			continue
		}
		d := model.Delivery{ID: key, IncidentID: v.ID, ChannelID: id, Kind: "reminder", Status: "pending", NextAt: Now(), CreatedAt: Now()}
		if err := s.Store.Put("delivery", key, d); err != nil {
			slog.Error("提醒入队失败", "error", err)
		}
	}
}
