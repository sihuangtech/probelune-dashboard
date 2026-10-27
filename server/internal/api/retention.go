package api

import (
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"log/slog"
)

func (s *Server) cleanup() {
	var cfg model.Settings
	if s.Store.Get("settings", "instance", &cfg) != nil {
		return
	}
	cutoff := Now() - int64(cfg.RawDays)*86400
	tx, err := s.Store.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	// 聚合与删除在同一事务中执行，重启后不会重复累加旧样本。
	for _, width := range []int{300, 3600} {
		_, err = tx.Exec(`INSERT INTO buckets SELECT monitor,node,scheduled/?*?, ?,SUM(status='success'),SUM(status='failure'),SUM(status NOT IN ('success','failure')),COALESCE(SUM(latency),0),COUNT(latency) FROM runs WHERE scheduled<? GROUP BY monitor,node,scheduled/? ON CONFLICT(monitor,node,bucket,width) DO UPDATE SET success=success+excluded.success,failed=failed+excluded.failed,errors=errors+excluded.errors,latency_sum=latency_sum+excluded.latency_sum,latency_count=latency_count+excluded.latency_count`, width, width, width, cutoff, width)
		if err != nil {
			slog.Error("历史聚合失败", "error", err)
			return
		}
	}
	if _, err = tx.Exec(`DELETE FROM runs WHERE scheduled<?`, cutoff); err != nil {
		return
	}
	if _, err = tx.Exec(`DELETE FROM buckets WHERE (width=300 AND bucket<?) OR bucket<?`, Now()-90*86400, Now()-365*86400); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	for _, v := range all[model.Session](s, "session") {
		if v.Expires < Now() {
			s.Store.Delete("session", v.ID)
		}
	}
	for _, v := range all[ct.Job](s, "job") {
		if v.CreatedAt < Now()-7*86400 {
			s.Store.Delete("job", v.ID)
			s.Store.Delete("diagnostic_result", v.ID)
		}
	}
	for _, v := range all[model.Incident](s, "incident") {
		if v.RecoveredAt > 0 && v.RecoveredAt < Now()-int64(cfg.EventDays)*86400 {
			s.Store.Delete("incident", v.ID)
		}
	}
	for _, v := range all[model.Audit](s, "audit") {
		if v.At < Now()-int64(cfg.EventDays)*86400 {
			s.Store.Delete("audit", v.ID)
		}
	}
	for _, v := range all[model.Delivery](s, "delivery") {
		if v.CreatedAt < Now()-int64(cfg.EventDays)*86400 {
			s.Store.Delete("delivery", v.ID)
		}
	}
}
