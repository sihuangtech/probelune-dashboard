package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"sort"
	"strconv"
)

func (s *Server) history(c *gin.Context) {
	id := c.Param("id")
	var m ct.Monitor
	if s.Store.Get("monitor", id, &m) != nil {
		fail(c, 404, "监测不存在")
		return
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours < 1 || hours > 8760 {
		hours = 24
	}
	width := int64(300)
	if hours > 90*24 {
		width = 3600
	}
	from := (Now() - int64(hours)*3600) / width * width
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "300"))
	if limit < 1 || limit > 1000 {
		limit = 300
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	node := c.Query("node_id")
	rows, err := s.Store.DB.Query(`SELECT payload FROM runs WHERE monitor=? AND scheduled>=? AND (?='' OR node=?) ORDER BY scheduled DESC LIMIT ? OFFSET ?`, id, from, node, node, limit, offset)
	if err != nil {
		fail(c, 500, "历史查询失败")
		return
	}
	results := []ct.Result{}
	for rows.Next() {
		var raw string
		var r ct.Result
		if rows.Scan(&raw) != nil {
			continue
		}
		if json.Unmarshal([]byte(raw), &r) == nil {
			results = append(results, r)
		}
	}
	rows.Close()
	buckets, err := s.historyBuckets(id, node, from, width)
	if err != nil {
		fail(c, 500, "聚合查询失败")
		return
	}
	summaries := []gin.H{}
	nodes := m.NodeIDs
	if node != "" {
		nodes = []string{node}
	}
	for _, n := range nodes {
		var success, failed, errors, latencyCount int64
		var latencySum float64
		var avg *float64
		for _, b := range buckets {
			if b.NodeID != n {
				continue
			}
			success += b.Success
			failed += b.Failed
			errors += b.Errors
			latencySum += b.LatencySum
			latencyCount += b.LatencyCount
		}
		if latencyCount > 0 {
			value := latencySum / float64(latencyCount)
			avg = &value
		}
		var availability any
		if success+failed > 0 {
			availability = float64(success) * 100 / float64(success+failed)
		}
		// 按启用的配置版本计算应执行数，断连或租约到期不会缩小分母。
		expected := s.expectedRuns(id, n, from, Now()-int64(m.Check.Timeout))
		var coverage any
		if expected > 0 {
			value := float64(success+failed) * 100 / float64(expected)
			if value > 100 {
				value = 100
			}
			coverage = value
		}
		summaries = append(summaries, gin.H{"node_id": n, "success": success, "failed": failed, "errors": errors, "availability": availability, "coverage": coverage, "expected": expected, "latency_ms": avg})
	}
	c.JSON(200, gin.H{"results": results, "buckets": buckets, "bucket_seconds": width, "summary": summaries, "limit": limit, "offset": offset, "from": from, "coverage_basis": "Expected cycles from enabled configuration versions; offline nodes count as missing. Maintenance included; paused periods excluded. Range rounded to bucket boundaries."})
}

func (s *Server) expectedRuns(monitor, node string, from, to int64) int64 {
	var count int64
	versions := []ct.Monitor{}
	for _, m := range all[ct.Monitor](s, "monitor_version") {
		if m.ID == monitor {
			versions = append(versions, m)
		}
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].Version < versions[j].Version })
	for i, m := range versions {
		if !m.Enabled || m.Archived || !contains(m.NodeIDs, node) {
			continue
		}
		start, end := m.UpdatedAt, to
		if i+1 < len(versions) {
			end = min(end, versions[i+1].UpdatedAt-1)
		}
		if start < from {
			start = from
		}
		if end > to {
			end = to
		}
		if end < start {
			continue
		}
		step := int64(m.Interval)
		first := (start + step - 1) / step * step
		if end >= first {
			count += (end-first)/step + 1
		}
	}
	return count
}
