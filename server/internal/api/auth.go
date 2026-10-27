package api

import (
	"crypto/subtle"
	"github.com/gin-gonic/gin"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type throttle struct {
	Count int
	At    int64
}

var loginMu sync.Mutex
var attempts = map[string]throttle{}

func rateLimit(c *gin.Context) {
	loginMu.Lock()
	defer loginMu.Unlock()
	now := Now()
	for k, v := range attempts {
		if now-v.At > 60 {
			delete(attempts, k)
		}
	}
	key := c.ClientIP()
	v := attempts[key]
	if v.Count >= 15 {
		fail(c, 429, "请求过于频繁，请稍后重试")
		return
	}
	v.Count++
	if v.At == 0 {
		v.At = now
	}
	attempts[key] = v
}
func (s *Server) authRoutes(r *gin.Engine) {
	a := r.Group("/api/v1/auth")
	a.GET("/status", func(c *gin.Context) {
		s.mu.Lock()
		defer s.mu.Unlock()
		c.JSON(200, gin.H{"initialized": len(all[model.User](s, "user")) > 0})
	})
	a.POST("/setup", rateLimit, s.setup)
	a.POST("/login", rateLimit, s.login)
	a.POST("/pair/exchange", rateLimit, s.exchangePair)
	a.GET("/me", s.authenticate, func(c *gin.Context) { c.JSON(200, c.MustGet("user")) })
	a.POST("/logout", s.authenticate, func(c *gin.Context) {
		s.Store.Delete("session", c.GetString("session_id"))
		s.cookie(c, "", -1)
		c.Status(204)
	})
	a.GET("/sessions", s.authenticate, func(c *gin.Context) {
		out := []model.Session{}
		for _, v := range all[model.Session](s, "session") {
			if v.UserID == c.GetString("user_id") && v.Expires > Now() {
				out = append(out, v)
			}
		}
		c.JSON(200, out)
	})
	a.DELETE("/sessions/:id", s.authenticate, func(c *gin.Context) {
		var v model.Session
		if s.Store.Get("session", c.Param("id"), &v) != nil || v.UserID != c.GetString("user_id") {
			fail(c, 404, "会话不存在")
			return
		}
		s.Store.Delete("session", v.ID)
		c.Status(204)
	})
	a.POST("/pair", s.authenticate, func(c *gin.Context) {
		code := ID()
		p := model.Pair{UserID: c.GetString("user_id"), Expires: Now() + 120}
		if s.save(c, "pair", Hash(code), p) {
			c.JSON(201, gin.H{"code": code, "expires": p.Expires})
		}
	})
}
func (s *Server) cookie(c *gin.Context, token string, age int) {
	c.SetSameSite(3)
	c.SetCookie("probelune_session", token, age, "/", "", c.Request.TLS != nil || strings.HasPrefix(os.Getenv("PROBELUNE_PUBLIC_URL"), "https://"), true)
}
func (s *Server) issue(c *gin.Context, u model.User) {
	token := ID()
	age := 8 * 3600
	if c.FullPath() == "/api/v1/auth/pair/exchange" || strings.HasPrefix(c.GetHeader("User-Agent"), "Probelune Mobile/") {
		age = 30 * 86400
	}
	session := model.Session{ID: Hash(token), UserID: u.ID, Expires: Now() + int64(age), Device: c.GetHeader("User-Agent")}
	if len(session.Device) > 200 {
		session.Device = session.Device[:200]
	}
	if !s.save(c, "session", session.ID, session) {
		return
	}
	s.cookie(c, token, age)
	u.Password = ""
	var cfg model.Settings
	s.Store.Get("settings", "instance", &cfg)
	c.JSON(200, gin.H{"token": token, "expires": session.Expires, "user": u, "instance_id": cfg.ID})
}
func (s *Server) setup(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(all[model.User](s, "user")) > 0 {
		fail(c, 409, "实例已初始化")
		return
	}
	if s.SetupToken == "" || subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Setup-Token")), []byte(s.SetupToken)) != 1 {
		fail(c, 403, "请提供服务器数据目录内的 setup-token")
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !bind(c, &input) {
		return
	}
	if len(input.Username) < 3 || len(input.Username) > 64 || len(input.Password) < 12 || len(input.Password) > 72 {
		fail(c, 400, "用户名 3—64 字符，密码 12—72 字节")
		return
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if e != nil {
		fail(c, 400, "密码不合法")
		return
	}
	u := model.User{ID: ID(), Username: input.Username, Password: string(hash), Role: "admin"}
	if !s.save(c, "user", u.ID, u) {
		return
	}
	os.Remove(filepath.Join(s.Dir, "setup-token"))
	s.SetupToken = ""
	s.issue(c, u)
}
func (s *Server) login(c *gin.Context) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !bind(c, &in) {
		return
	}
	for _, u := range all[model.User](s, "user") {
		if u.Username == in.Username && !u.Disabled && bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)) == nil {
			s.issue(c, u)
			return
		}
	}
	time.Sleep(150 * time.Millisecond)
	fail(c, 401, "用户名或密码错误")
}
func (s *Server) authenticate(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token == "" {
		token, _ = c.Cookie("probelune_session")
	}
	var session model.Session
	var u model.User
	if token == "" || s.Store.Get("session", Hash(token), &session) != nil || session.Expires < Now() || s.Store.Get("user", session.UserID, &u) != nil || u.Disabled {
		fail(c, 401, "请重新登录")
		return
	}
	u.Password = ""
	c.Set("user", u)
	c.Set("user_id", u.ID)
	c.Set("session_id", session.ID)
	c.Set("role", u.Role)
}
func (s *Server) role(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !contains(roles, c.GetString("role")) {
			fail(c, 403, "当前角色无权执行此操作")
		}
	}
}
func (s *Server) exchangePair(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var in struct {
		Code string `json:"code"`
	}
	if !bind(c, &in) {
		return
	}
	var p model.Pair
	var u model.User
	if s.Store.Get("pair", Hash(in.Code), &p) != nil || p.Expires < Now() || s.Store.Get("user", p.UserID, &u) != nil || u.Disabled {
		fail(c, 401, "配对码已失效")
		return
	}
	s.Store.Delete("pair", Hash(in.Code))
	s.issue(c, u)
}
