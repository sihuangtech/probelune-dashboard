package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

func (s *Server) adminRoutes(v *gin.RouterGroup) {
	v.GET("/settings", func(c *gin.Context) {
		var cfg model.Settings
		s.Store.Get("settings", "instance", &cfg)
		c.JSON(200, cfg)
	})
	v.PUT("/settings", s.role("admin"), s.settings)
	v.GET("/maintenance", func(c *gin.Context) { c.JSON(200, all[model.Maintenance](s, "maintenance")) })
	v.POST("/maintenance", s.role("admin", "operator"), s.maintenance)
	v.DELETE("/maintenance/:id", s.role("admin", "operator"), func(c *gin.Context) {
		s.mu.Lock()
		defer s.mu.Unlock()
		var m model.Maintenance
		if s.Store.Get("maintenance", c.Param("id"), &m) != nil {
			fail(c, 404, "维护窗口不存在")
			return
		}
		if m.Start <= Now() {
			m.End = min(m.End, Now())
			if !s.save(c, "maintenance", m.ID, m) {
				return
			}
		} else if s.Store.Delete("maintenance", m.ID) != nil {
			fail(c, 500, "删除失败")
			return
		}
		s.audit(c, "maintenance.delete", c.Param("id"))
		c.Status(204)
	})
	v.GET("/users", s.role("admin"), func(c *gin.Context) {
		items := all[model.User](s, "user")
		for i := range items {
			items[i].Password = ""
		}
		c.JSON(200, items)
	})
	v.POST("/users", s.role("admin"), s.saveUser)
	v.PUT("/users/:id", s.role("admin"), s.saveUser)
	v.GET("/channels", s.role("admin"), s.channels)
	v.POST("/channels", s.role("admin"), s.saveChannel)
	v.PUT("/channels/:id", s.role("admin"), s.saveChannel)
	v.POST("/channels/:id/test", s.role("admin"), func(c *gin.Context) {
		var ch model.Channel
		if s.Store.Get("channel", c.Param("id"), &ch) != nil {
			fail(c, 404, "渠道不存在")
			return
		}
		d := model.Delivery{ID: ID(), ChannelID: ch.ID, Kind: "test", Status: "pending", NextAt: Now(), CreatedAt: Now()}
		if s.save(c, "delivery", d.ID, d) {
			s.signal()
			c.JSON(202, d)
		}
	})
	v.GET("/deliveries", s.role("admin", "operator"), func(c *gin.Context) { c.JSON(200, all[model.Delivery](s, "delivery")) })
	v.GET("/audit", s.role("admin"), func(c *gin.Context) { c.JSON(200, all[model.Audit](s, "audit")) })
	v.POST("/devices", s.saveDevice)
	v.GET("/devices", func(c *gin.Context) {
		out := []model.Device{}
		for _, d := range all[model.Device](s, "device") {
			if d.UserID == c.GetString("user_id") {
				d.Token = ""
				out = append(out, d)
			}
		}
		c.JSON(200, out)
	})
	v.DELETE("/devices/:id", func(c *gin.Context) {
		var d model.Device
		if s.Store.Get("device", c.Param("id"), &d) != nil || d.UserID != c.GetString("user_id") {
			fail(c, 404, "设备不存在")
			return
		}
		s.Store.Delete("device", d.ID)
		c.Status(204)
	})
}
func (s *Server) settings(c *gin.Context) {
	var cfg model.Settings
	if !bind(c, &cfg) {
		return
	}
	if cfg.Name == "" || cfg.RawDays < 1 || cfg.RawDays > 90 || cfg.EventDays < 7 || cfg.EventDays > 365 {
		fail(c, 400, "原始数据保留 1—90 天，事件保留 7—365 天")
		return
	}
	if _, e := time.LoadLocation(cfg.Timezone); e != nil {
		fail(c, 400, "无效时区")
		return
	}
	if cfg.PublicURL != "" {
		u, e := url.Parse(cfg.PublicURL)
		if e != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			fail(c, 400, "实例地址无效")
			return
		}
	}
	var old model.Settings
	s.Store.Get("settings", "instance", &old)
	cfg.ID = old.ID
	if s.save(c, "settings", "instance", cfg) {
		s.audit(c, "settings.update", "instance")
		c.JSON(200, cfg)
	}
}
func (s *Server) maintenance(c *gin.Context) {
	var m model.Maintenance
	if !bind(c, &m) {
		return
	}
	if m.Name == "" || m.End <= m.Start || m.End <= Now() || m.End-m.Start > 30*86400 {
		fail(c, 400, "维护窗口最长 30 天，结束时间必须晚于开始时间")
		return
	}
	m.ID = ID()
	if s.save(c, "maintenance", m.ID, m) {
		s.audit(c, "maintenance.create", m.ID)
		c.JSON(201, m)
	}
}
func (s *Server) saveUser(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Disabled bool   `json:"disabled"`
	}
	if !bind(c, &in) {
		return
	}
	if len(in.Username) < 3 || len(in.Username) > 64 || !contains([]string{"admin", "operator", "viewer"}, in.Role) {
		fail(c, 400, "用户名或角色无效")
		return
	}
	id := c.Param("id")
	var user model.User
	if id != "" {
		if s.Store.Get("user", id, &user) != nil {
			fail(c, 404, "成员不存在")
			return
		}
	} else {
		id = ID()
	}
	for _, u := range all[model.User](s, "user") {
		if u.ID != id && u.Username == in.Username {
			fail(c, 409, "用户名已存在")
			return
		}
	}
	if id == c.GetString("user_id") && (in.Disabled || in.Role != "admin") {
		fail(c, 400, "不能禁用自己或降低当前管理员权限")
		return
	}
	if in.Password != "" {
		if len(in.Password) < 12 || len(in.Password) > 72 {
			fail(c, 400, "密码需要 12—72 字节")
			return
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if e != nil {
			fail(c, 400, "密码无效")
			return
		}
		user.Password = string(hash)
	}
	if user.Password == "" {
		fail(c, 400, "新成员必须设置密码")
		return
	}
	user.ID = id
	user.Username = in.Username
	user.Role = in.Role
	user.Disabled = in.Disabled
	if s.save(c, "user", id, user) {
		if in.Password != "" || in.Disabled {
			for _, session := range all[model.Session](s, "session") {
				if session.UserID == id && session.ID != c.GetString("session_id") {
					s.Store.Delete("session", session.ID)
				}
			}
		}
		s.audit(c, "user.save", id)
		user.Password = ""
		c.JSON(200, user)
	}
}
func (s *Server) channels(c *gin.Context) {
	out := []gin.H{}
	for _, ch := range all[model.Channel](s, "channel") {
		has := ch.Secret != "" || ch.Password != ""
		ch.Secret = ""
		ch.Password = ""
		out = append(out, gin.H{"channel": ch, "has_secret": has})
	}
	c.JSON(200, out)
}
func (s *Server) saveChannel(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ch model.Channel
	if !bind(c, &ch) {
		return
	}
	if ch.Name == "" || !contains([]string{"webhook", "smtp", "push"}, ch.Type) {
		fail(c, 400, "渠道名称或类型无效")
		return
	}
	if ch.ReminderMinutes < 0 || ch.ReminderMinutes > 1440 {
		fail(c, 400, "持续提醒间隔为 0—1440 分钟，0 为关闭")
		return
	}
	if err := ch.Policy.Validate(); err != nil {
		fail(c, 400, err.Error())
		return
	}
	ch.ID = c.Param("id")
	if ch.ID != "" {
		var old model.Channel
		if s.Store.Get("channel", ch.ID, &old) != nil {
			fail(c, 404, "渠道不存在")
			return
		}
		if ch.Secret == "" {
			ch.Secret = old.Secret
		}
		if ch.Password == "" {
			ch.Password = old.Password
		}
	} else {
		ch.ID = ID()
	}
	if ch.Type == "webhook" {
		u, e := url.Parse(ch.URL)
		if e != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			fail(c, 400, "Webhook 地址无效")
			return
		}
		if ch.Secret == "" {
			ch.Secret = ID()
		}
	}
	if ch.Type == "smtp" {
		if ch.SMTPPort < 1 || ch.SMTPPort > 65535 || ch.SMTPHost == "" || !contains([]string{"tls", "starttls"}, ch.TLSMode) {
			fail(c, 400, "SMTP 必须配置 TLS/STARTTLS、主机与端口")
			return
		}
		if _, e := mail.ParseAddress(ch.From); e != nil {
			fail(c, 400, "发件地址无效")
			return
		}
		if _, e := mail.ParseAddressList(ch.To); e != nil {
			fail(c, 400, "收件地址无效")
			return
		}
	}
	if s.save(c, "channel", ch.ID, ch) {
		s.audit(c, "channel.save", ch.ID)
		secret := ch.Secret
		ch.Secret = ""
		ch.Password = ""
		c.JSON(200, gin.H{"channel": ch, "signing_secret": secret})
	}
}
func (s *Server) saveDevice(c *gin.Context) {
	var d model.Device
	if !bind(c, &d) {
		return
	}
	if len(d.Token) < 10 || len(d.Token) > 4096 || !contains([]string{"android", "ios"}, d.Platform) {
		fail(c, 400, "设备信息无效")
		return
	}
	d.ID = Hash(c.GetString("user_id") + d.Token)
	d.UserID = c.GetString("user_id")
	d.CreatedAt = Now()
	d.Token = strings.TrimSpace(d.Token)
	if s.save(c, "device", d.ID, d) {
		d.Token = ""
		c.JSON(200, d)
	}
}
