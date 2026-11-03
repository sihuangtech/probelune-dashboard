package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"sort"
	"strconv"
)

type uptimeCell struct {
	At           int64    `json:"at"`
	End          int64    `json:"end"`
	Status       string   `json:"status"`
	Success      int64    `json:"success"`
	Failed       int64    `json:"failed"`
	Errors       int64    `json:"errors"`
	Expected     int64    `json:"expected"`
	Availability *float64 `json:"availability"`
	Coverage     *float64 `json:"coverage"`
}

type uptimeSeries struct {
	From         int64        `json:"from"`
	Until        int64        `json:"until"`
	Bucket       int64        `json:"bucket_seconds"`
	Cells        []uptimeCell `json:"cells"`
	Availability *float64     `json:"availability"`
	Coverage     *float64     `json:"coverage"`
}

func uptimeRange(c *gin.Context) (int64, int64, int64, error) {
	hours, e1 := strconv.Atoi(c.DefaultQuery("hours", "24"))
	minutes, e2 := strconv.Atoi(c.DefaultQuery("bucket_minutes", "15"))
	if e1 != nil || e2 != nil || hours < 1 || hours > 720 || minutes < 5 || minutes > 1440 || minutes%5 != 0 || (hours*60+minutes-1)/minutes > 2016 {
		return 0, 0, 0, fmt.Errorf("时间范围为 1—720 小时；色块粒度为 5—1440 分钟（5 的倍数），最多 2016 个色块")
	}
	width := int64(minutes * 60)
	until := Now()
	from := (until-int64(hours)*3600)/width*width + width
	return from, until, width, nil
}

func percentage(n, d int64) *float64 {
	if d <= 0 {
		return nil
	}
	v := min(float64(n)*100/float64(d), 100)
	return &v
}

// Raw and archived five-minute samples are disjoint. Aggregate in SQL before
// serializing so the public endpoint never returns targets, nodes or raw data.
func (s *Server) uptimeSeries(m ct.Monitor, versions []ct.Monitor, from, until, width int64) (uptimeSeries, error) {
	out := uptimeSeries{From: from, Until: until, Bucket: width, Cells: []uptimeCell{}}
	rows, err := s.Store.DB.Query(`SELECT at,SUM(ok),SUM(bad),SUM(errs) FROM (
 SELECT scheduled/?*? AS at,SUM(status='success') AS ok,SUM(status='failure') AS bad,SUM(status NOT IN ('success','failure')) AS errs FROM runs WHERE monitor=? AND scheduled>=? AND scheduled<=? GROUP BY at
 UNION ALL SELECT bucket/?*? AS at,success,failed,errors FROM buckets WHERE monitor=? AND width=300 AND bucket>=? AND bucket<=?
 ) GROUP BY at ORDER BY at`, width, width, m.ID, from, until, width, width, m.ID, from, until)
	if err != nil {
		return out, err
	}
	samples := map[int64]uptimeCell{}
	for rows.Next() {
		var v uptimeCell
		if err = rows.Scan(&v.At, &v.Success, &v.Failed, &v.Errors); err != nil {
			rows.Close()
			return out, err
		}
		samples[v.At] = v
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	var success, failed, expected int64
	for at := from; at <= until; at += width {
		v := samples[at]
		v.At, v.End = at, at+width
		v.Expected = expectedVersionRuns(versions, "", at, min(at+width-1, until-int64(m.Check.Timeout)))
		v.Availability = percentage(v.Success, v.Success+v.Failed)
		v.Coverage = percentage(v.Success+v.Failed, v.Expected)
		v.Status = "unknown"
		switch {
		case v.Failed > 0 && v.Success == 0:
			v.Status = "down"
		case v.Failed > 0:
			v.Status = "partial"
		case v.Success > 0 && v.Errors == 0 && v.Success >= v.Expected:
			v.Status = "healthy"
		case v.Success > 0:
			v.Status = "incomplete"
		}
		success += v.Success
		failed += v.Failed
		expected += v.Expected
		out.Cells = append(out.Cells, v)
	}
	out.Availability = percentage(success, success+failed)
	out.Coverage = percentage(success+failed, expected)
	return out, nil
}

func monitorVersions(s *Server, id string) []ct.Monitor {
	out := []ct.Monitor{}
	for _, m := range all[ct.Monitor](s, "monitor_version") {
		if m.ID == id {
			out = append(out, m)
		}
	}
	return out
}

func (s *Server) uptime(c *gin.Context) {
	var m ct.Monitor
	if s.Store.Get("monitor", c.Param("id"), &m) != nil {
		fail(c, 404, "监测不存在")
		return
	}
	from, until, width, err := uptimeRange(c)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	series, err := s.uptimeSeries(m, monitorVersions(s, m.ID), from, until, width)
	if err != nil {
		fail(c, 500, "可用性历史查询失败")
		return
	}
	c.JSON(200, series)
}

func (s *Server) publicStatus(c *gin.Context) {
	var cfg model.Settings
	s.Store.Get("settings", "instance", &cfg)
	if !cfg.PublicStatus {
		fail(c, 404, "状态页未启用")
		return
	}
	from, until, width, err := uptimeRange(c)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	versions := map[string][]ct.Monitor{}
	for _, m := range all[ct.Monitor](s, "monitor_version") {
		versions[m.ID] = append(versions[m.ID], m)
	}
	list := []gin.H{}
	public := map[string]bool{}
	for _, m := range all[ct.Monitor](s, "monitor") {
		if !m.Public || m.Archived {
			continue
		}
		public[m.ID] = true
		series, err := s.uptimeSeries(m, versions[m.ID], from, until, width)
		if err != nil {
			fail(c, 500, "可用性历史查询失败")
			return
		}
		v := s.monitorView(m)
		list = append(list, gin.H{"id": m.ID, "name": m.Name, "status": v["status"], "coverage_complete": v["coverage_complete"], "maintenance": v["maintenance"], "interval": m.Interval, "uptime": series})
	}
	sort.Slice(list, func(i, j int) bool { return list[i]["name"].(string) < list[j]["name"].(string) })
	incidents := []gin.H{}
	for _, v := range all[model.Incident](s, "incident") {
		if public[v.MonitorID] && v.OpenedAt <= until && (v.RecoveredAt == 0 || v.RecoveredAt >= from) {
			incidents = append(incidents, gin.H{"monitor_id": v.MonitorID, "opened_at": v.OpenedAt, "recovered_at": v.RecoveredAt, "suppressed": v.Suppressed})
		}
	}
	c.JSON(200, gin.H{"name": cfg.Name, "monitors": list, "incidents": incidents, "updated_at": until, "timezone": cfg.Timezone})
}
