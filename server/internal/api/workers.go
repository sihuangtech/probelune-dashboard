package api

import (
	"context"
	"fmt"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"github.com/sihuangtech/probelune-dashboard/server/internal/notify"
	"log/slog"
	"strings"
	"time"
)

func (s *Server) RunWorkers(ctx context.Context) {
	go s.deliveryLoop(ctx)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	lastCleanup := int64(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.notify:
		}
		s.mu.Lock()
		s.reconcile()
		s.mu.Unlock()
		if Now()-lastCleanup >= 3600 {
			s.cleanup()
			lastCleanup = Now()
		}
	}
}
func (s *Server) enqueue(v model.Incident, kind string, ids []string) {
	if ids == nil {
		for _, ch := range all[model.Channel](s, "channel") {
			if ch.Enabled {
				ids = append(ids, ch.ID)
			}
		}
	}
	for _, id := range ids {
		key := v.ID + ":" + kind + ":" + id
		var existing model.Delivery
		if s.Store.Get("delivery", key, &existing) == nil {
			continue
		}
		d := model.Delivery{ID: key, IncidentID: v.ID, ChannelID: id, Kind: kind, Status: "pending", NextAt: Now(), CreatedAt: Now()}
		if err := s.Store.Put("delivery", key, d); err != nil {
			slog.Error("通知入队失败", "error", err)
		}
	}
}
func (s *Server) deliveryLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for _, d := range all[model.Delivery](s, "delivery") {
			if d.Status != "pending" || d.NextAt > Now() {
				continue
			}
			s.deliver(ctx, d)
		}
	}
}
func (s *Server) deliver(ctx context.Context, d model.Delivery) {
	var ch model.Channel
	var event model.Incident
	var cfg model.Settings
	if s.Store.Get("channel", d.ChannelID, &ch) != nil || !ch.Enabled {
		d.Status = "skipped"
		d.Error = "渠道已停用"
		s.Store.Put("delivery", d.ID, d)
		return
	}
	s.Store.Get("incident", d.IncidentID, &event)
	s.Store.Get("settings", "instance", &cfg)
	if d.Kind != "test" && (s.maintained(event.MonitorID) || event.SilencedUntil > Now()) {
		d.Status = "suppressed"
		s.Store.Put("delivery", d.ID, d)
		return
	}
	m := notify.Message{InstanceID: cfg.ID, ID: event.ID, Instance: cfg.Name, Kind: d.Kind, Name: event.Name, At: Now(), InstanceURL: strings.TrimRight(cfg.PublicURL, "/"), Affected: event.Affected}
	m.URL = m.InstanceURL + "/incidents?id=" + event.ID
	call, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var err error
	if ch.Type == "push" {
		devices := all[model.Device](s, "device")
		if len(devices) == 0 {
			err = fmt.Errorf("尚未注册推送设备")
		}
		for _, device := range devices {
			var u model.User
			if s.Store.Get("user", device.UserID, &u) != nil || u.Disabled {
				continue
			}
			if e := notify.Push(call, device.Token, m); e != nil {
				err = e
			}
		}
	} else {
		err = notify.Send(call, ch, m)
	}
	d.Attempts++
	d.Error = ""
	if err == nil {
		d.Status = "sent"
	} else {
		d.Error = err.Error()
		if d.Attempts >= 3 {
			d.Status = "failed"
		} else {
			d.NextAt = Now() + int64(5*(1<<d.Attempts))
		}
	}
	if e := s.Store.Put("delivery", d.ID, d); e != nil {
		slog.Error("通知状态保存失败", "error", e)
	}
}
