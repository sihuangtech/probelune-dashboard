package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sihuangtech/probelune-dashboard/server/internal/model"
	"github.com/sihuangtech/probelune-dashboard/server/internal/store"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Store      *store.Store
	Dir        string
	SetupToken string
	mu         sync.Mutex
	notify     chan struct{}
}

func ID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Hash(v string) string { b := sha256.Sum256([]byte(v)); return hex.EncodeToString(b[:]) }
func Now() int64           { return time.Now().Unix() }
func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
func fail(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
func bind(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		fail(c, 400, "请求数据格式不正确")
		return false
	}
	return true
}
func (s *Server) save(c *gin.Context, kind, id string, v any) bool {
	if err := s.Store.Put(kind, id, v); err != nil {
		slog.Error("save failed", "kind", kind, "error", err)
		fail(c, 500, "保存失败")
		return false
	}
	return true
}
func all[T any](s *Server, kind string) []T {
	v, err := store.List[T](s.Store, kind)
	if err != nil {
		panic(err)
	}
	return v
}
func (s *Server) audit(c *gin.Context, action, id string) {
	a := model.Audit{ID: ID(), At: Now(), User: c.GetString("user_id"), Action: action, Resource: id}
	if err := s.Store.Put("audit", a.ID, a); err != nil {
		slog.Error("audit", "error", err)
	}
}
func (s *Server) signal() {
	select {
	case s.notify <- struct{}{}:
	default:
	}
}

func New(dir string) (*Server, error) {
	db, err := store.Open(dir)
	if err != nil {
		return nil, err
	}
	s := &Server{Store: db, Dir: dir, notify: make(chan struct{}, 1)}
	var settings model.Settings
	if err = db.Get("settings", "instance", &settings); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			db.DB.Close()
			return nil, err
		}
		settings = model.Settings{Name: "巡汐 Probelune", Timezone: "Asia/Shanghai", RawDays: 7, EventDays: 180}
		if err = db.Put("settings", "instance", settings); err != nil {
			return nil, err
		}
	}
	if settings.ID == "" {
		settings.ID = ID()
		if err = db.Put("settings", "instance", settings); err != nil {
			return nil, err
		}
	}
	if len(all[model.User](s, "user")) == 0 {
		path := filepath.Join(dir, "setup-token")
		raw, e := os.ReadFile(path)
		if errors.Is(e, os.ErrNotExist) {
			raw = []byte(ID())
			e = os.WriteFile(path, raw, 0600)
		}
		if e != nil {
			return nil, e
		}
		s.SetupToken = strings.TrimSpace(string(raw))
	}
	return s, nil
}

func (s *Server) Router(web string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
		}
		origin := c.Request.Header.Get("Origin")
		if origin != "" && c.Request.Method != "GET" && c.GetHeader("Authorization") == "" {
			scheme := "https"
			if c.Request.TLS == nil {
				scheme = "http"
			}
			allowed := os.Getenv("PROBELUNE_PUBLIC_URL")
			if origin != scheme+"://"+c.Request.Host && origin != allowed {
				fail(c, 403, "请求来源不允许")
				return
			}
		}
		c.Next()
	})
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "version": "0.1.0"}) })
	s.authRoutes(r)
	s.nodeRoutes(r)
	v := r.Group("/api/v1", s.authenticate)
	v.GET("/overview", s.overview)
	v.GET("/monitors", s.monitors)
	v.GET("/monitors/:id/history", s.history)
	v.POST("/monitors", s.role("admin", "operator"), s.saveMonitor)
	v.PUT("/monitors/:id", s.role("admin", "operator"), s.saveMonitor)
	v.GET("/nodes", s.nodes)
	v.POST("/nodes", s.role("admin"), s.createNode)
	v.PUT("/nodes/:id", s.role("admin"), s.updateNode)
	v.POST("/nodes/:id/credentials", s.role("admin"), s.rotateNode)
	v.GET("/diagnostics", s.jobs)
	v.POST("/diagnostics", s.role("admin", "operator"), s.createJobs)
	v.POST("/diagnostics/:id/cancel", s.role("admin", "operator"), s.cancelJob)
	v.GET("/incidents", s.incidents)
	v.POST("/incidents/:id/actions", s.role("admin", "operator"), s.incidentAction)
	s.adminRoutes(v)
	r.GET("/api/v1/status", s.publicStatus)
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" || strings.HasPrefix(c.Request.URL.Path, "/api/") {
			fail(c, 404, "未找到资源")
			return
		}
		relative := filepath.Clean("/" + c.Request.URL.Path)
		file := filepath.Join(web, relative)
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			c.File(file)
			return
		}
		if _, err := os.Stat(filepath.Join(web, "index.html")); err != nil {
			fail(c, 503, "网页尚未构建，请先执行 npm run build")
			return
		}
		c.File(filepath.Join(web, "index.html"))
	})
	return r
}
