package api

import (
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"sort"
)

func (s *Server) createJobs(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var in struct {
		Check     ct.Check `json:"check"`
		NodeIDs   []string `json:"node_ids"`
		MonitorID string   `json:"monitor_id"`
	}
	if !bind(c, &in) {
		return
	}
	if in.MonitorID != "" {
		var m ct.Monitor
		if s.Store.Get("monitor", in.MonitorID, &m) != nil {
			fail(c, 404, "监测不存在")
			return
		}
		in.Check = m.Check
		if len(in.NodeIDs) == 0 {
			in.NodeIDs = m.NodeIDs
		}
	}
	if err := in.Check.Validate(); err != nil {
		fail(c, 400, err.Error())
		return
	}
	if len(in.NodeIDs) == 0 || len(in.NodeIDs) > 20 {
		fail(c, 400, "请选择 1—20 个节点")
		return
	}
	jobs := []ct.Job{}
	seen := map[string]bool{}
	for _, id := range in.NodeIDs {
		var n model.Node
		if seen[id] || s.Store.Get("node", id, &n) != nil || !n.Enabled || n.Revoked {
			fail(c, 400, "包含无效节点")
			return
		}
		seen[id] = true
		if in.Check.CertificateMinDays > 0 && !contains(n.Info.Capabilities, "certificate_expiry") {
			fail(c, 400, "节点需要升级以支持证书有效期断言")
			return
		}
		if !contains(n.Info.Capabilities, in.Check.Type) {
			fail(c, 400, "节点不支持此检测类型")
			return
		}
		jobs = append(jobs, ct.Job{ID: ID(), NodeID: id, Check: in.Check, Policy: n.Policy, Status: "queued", CreatedAt: Now(), Deadline: Now() + 120})
	}
	for _, j := range jobs {
		if !s.save(c, "job", j.ID, j) {
			return
		}
		s.audit(c, "diagnostic.create", j.ID)
	}
	for i := range jobs {
		jobs[i].Check.Headers = nil
	}
	c.JSON(201, jobs)
}
func (s *Server) jobs(c *gin.Context) {
	jobs := all[ct.Job](s, "job")
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].CreatedAt > jobs[j].CreatedAt })
	if len(jobs) > 200 {
		jobs = jobs[:200]
	}
	out := []gin.H{}
	for _, j := range jobs {
		j.Check.Headers = nil
		if j.Deadline < Now() && (j.Status == "queued" || j.Status == "running") {
			j.Status = "expired"
		}
		var result ct.Result
		s.Store.Get("diagnostic_result", j.ID, &result)
		out = append(out, gin.H{"job": j, "result": result})
	}
	c.JSON(200, out)
}
func (s *Server) cancelJob(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var j ct.Job
	if s.Store.Get("job", c.Param("id"), &j) != nil {
		fail(c, 404, "任务不存在")
		return
	}
	if j.Status == "queued" || j.Status == "running" {
		j.Status = "cancelled"
		if !s.save(c, "job", j.ID, j) {
			return
		}
		s.audit(c, "diagnostic.cancel", j.ID)
	}
	c.JSON(200, gin.H{"status": j.Status})
}
