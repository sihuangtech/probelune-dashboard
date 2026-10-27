package main

import (
	"context"
	"flag"
	"github.com/sihuangtech/probelune-dashboard/server/internal/api"
	"github.com/sihuangtech/probelune-dashboard/server/internal/backup"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	address := flag.String("listen", "127.0.0.1:8080", "HTTP 监听地址")
	data := flag.String("data", "./data", "数据与加密密钥目录")
	web := flag.String("web", "../web/dist", "网页静态资源目录")
	backupPath := flag.String("backup", "", "在线备份到一个不存在的目录，然后退出")
	flag.Parse()
	if *backupPath != "" {
		if err := backup.Run(*data, *backupPath); err != nil {
			slog.Error("备份失败", "error", err)
			os.Exit(1)
		}
		slog.Info("备份完成", "directory", *backupPath)
		return
	}
	s, err := api.New(*data)
	if err != nil {
		slog.Error("启动失败", "error", err)
		os.Exit(1)
	}
	defer s.Store.DB.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go s.RunWorkers(ctx)
	httpServer := &http.Server{Addr: *address, Handler: s.Router(*web), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdown)
	}()
	slog.Info("巡汐面板启动", "listen", *address, "data", *data)
	if s.SetupToken != "" {
		slog.Info("首次初始化：在受控终端读取数据目录的 setup-token 文件")
	}
	if err = httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP 服务失败", "error", err)
		os.Exit(1)
	}
}
