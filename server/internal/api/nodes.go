package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	ct "github.com/sihuangtech/probelune-contracts"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"strings"
)

func (s *Server) nodeRoutes(r *gin.Engine) {
	r.POST("/api/v1/node/register", rateLimit, s.registerNode)
	n := r.Group("/api/v1/node", s.nodeAuth)
	n.POST("/poll", s.poll)
	n.POST("/results", s.receive)
}
func (s *Server) nodeAuth(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	id := c.GetHeader("X-Node-ID")
	var n model.Node
	if token == "" || s.Store.Get("node", id, &n) != nil || n.Revoked || Hash(token) != n.TokenHash {
		fail(c, 401, "节点身份已失效")
		return
	}
	c.Set("node_id", id)
}
func publicNode(n model.Node) model.Node { n.TokenHash = ""; n.RegisterHash = ""; return n }
func (s *Server) nodes(c *gin.Context) {
	out := []model.Node{}
	for _, n := range all[model.Node](s, "node") {
		out = append(out, publicNode(n))
	}
	c.JSON(200, out)
}
func (s *Server) createNode(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n model.Node
	if !bind(c, &n) {
		return
	}
	if len(n.Name) < 1 || len(n.Name) > 100 {
		fail(c, 400, "节点名称不能为空且最多 100 字符")
		return
	}
	if err := n.Policy.Validate(); err != nil {
		fail(c, 400, err.Error())
		return
	}
	code := ID()
	n = model.Node{ID: ID(), Name: n.Name, Region: n.Region, Network: n.Network, Enabled: true, Policy: n.Policy, CreatedAt: Now(), RegisterHash: Hash(code), RegisterExpires: Now() + 600}
	if s.save(c, "node", n.ID, n) {
		s.audit(c, "node.create", n.ID)
		c.JSON(201, gin.H{"node": publicNode(n), "registration_token": code, "expires": n.RegisterExpires})
	}
}
func (s *Server) updateNode(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n, in model.Node
	if s.Store.Get("node", c.Param("id"), &n) != nil {
		fail(c, 404, "节点不存在")
		return
	}
	if !bind(c, &in) {
		return
	}
	if in.Name == "" {
		fail(c, 400, "名称不能为空")
		return
	}
	if err := in.Policy.Validate(); err != nil {
		fail(c, 400, err.Error())
		return
	}
	n.Name = in.Name
	n.Region = in.Region
	n.Network = in.Network
	n.Enabled = in.Enabled
	n.Policy = in.Policy
	if s.save(c, "node", n.ID, n) {
		for _, job := range all[ct.Job](s, "job") {
			if job.NodeID == n.ID && (job.Status == "queued" || job.Status == "running") {
				job.Status = "cancelled"
				if !s.save(c, "job", job.ID, job) {
					return
				}
			}
		}
		s.audit(c, "node.update", n.ID)
		c.JSON(200, publicNode(n))
	}
}
func (s *Server) rotateNode(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n model.Node
	var in struct {
		Revoke bool `json:"revoke"`
	}
	if !bind(c, &in) {
		return
	}
	if s.Store.Get("node", c.Param("id"), &n) != nil {
		fail(c, 404, "节点不存在")
		return
	}
	n.TokenHash = ""
	n.Revoked = in.Revoke
	n.LastSeen = 0
	n.RegisterHash = ""
	n.RegisterExpires = 0
	code := ""
	if !in.Revoke {
		code = ID()
		n.RegisterHash = Hash(code)
		n.RegisterExpires = Now() + 600
	}
	if s.save(c, "node", n.ID, n) {
		s.audit(c, "node.credentials", n.ID)
		c.JSON(200, gin.H{"registration_token": code, "node": publicNode(n)})
	}
}
func (s *Server) registerNode(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var in struct {
		NodeID string `json:"node_id"`
		Token  string `json:"token"`
	}
	if !bind(c, &in) {
		return
	}
	var n model.Node
	if s.Store.Get("node", in.NodeID, &n) != nil || n.Revoked || n.RegisterHash == "" || n.RegisterExpires < Now() || n.RegisterHash != Hash(in.Token) {
		fail(c, 401, "注册凭据无效或已使用")
		return
	}
	token := ID()
	n.TokenHash = Hash(token)
	n.RegisterHash = ""
	n.RegisterExpires = 0
	if s.save(c, "node", n.ID, n) {
		c.JSON(200, gin.H{"token": token, "node_id": n.ID, "protocol": ct.Protocol})
	}
}
func (s *Server) poll(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var info ct.Poll
	if !bind(c, &info) {
		return
	}
	if info.Protocol != ct.Protocol {
		fail(c, 409, "节点协议版本不兼容，请升级")
		return
	}
	var n model.Node
	if s.Store.Get("node", c.GetString("node_id"), &n) != nil || n.Revoked || n.TokenHash != Hash(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")) {
		fail(c, 401, "节点身份已撤销")
		return
	}
	now := Now()
	n.LastSeen = now
	n.Info = info
	n.ClockSkew = info.Time - now
	if !s.save(c, "node", n.ID, n) {
		return
	}
	work := ct.Work{Assignments: []ct.Assignment{}, Jobs: []ct.Job{}, Cancelled: []string{}, ServerTime: now}
	if n.Enabled && n.ClockSkew < 30 && n.ClockSkew > -30 {
		for _, m := range all[ct.Monitor](s, "monitor") {
			if m.Check.CertificateMinDays > 0 && !contains(info.Capabilities, "certificate_expiry") {
				continue
			}
			if !m.Enabled || m.Archived || !contains(m.NodeIDs, n.ID) || !contains(info.Capabilities, m.Check.Type) {
				continue
			}
			policyJSON, _ := json.Marshal(n.Policy)
			key := Hash(ct.RunID(m.ID, n.ID, m.Version, 0) + string(policyJSON))
			var a ct.Assignment
			if s.Store.Get("assignment", key, &a) != nil {
				a = ct.Assignment{ID: key, Monitor: m, NodeID: n.ID, Start: now / int64(m.Interval) * int64(m.Interval)}
			}
			a.Until = now + 900
			a.Policy = n.Policy
			if !s.save(c, "assignment", key, a) {
				return
			}
			work.Assignments = append(work.Assignments, a)
		}
	}
	for _, j := range all[ct.Job](s, "job") {
		if j.NodeID != n.ID {
			continue
		}
		if j.Status == "cancelled" || j.Deadline < now {
			work.Cancelled = append(work.Cancelled, j.ID)
			continue
		}
		if !n.Enabled {
			work.Cancelled = append(work.Cancelled, j.ID)
			continue
		}
		if n.ClockSkew >= 30 || n.ClockSkew <= -30 {
			work.Cancelled = append(work.Cancelled, j.ID)
			continue
		}
		if j.Status == "queued" || j.Status == "running" {
			j.Status = "running"
			if !s.save(c, "job", j.ID, j) {
				return
			}
			work.Jobs = append(work.Jobs, j)
		}
	}
	c.JSON(200, work)
}
