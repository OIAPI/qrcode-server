package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log/slog"

	"qrcode-server/config" // 替换为实际模块名
)

// Start 启动HTTP服务（接收Gin引擎和日志实例）
func Start(handler http.Handler, log *slog.Logger) {
	// 从配置获取服务参数
	cfg := config.Get()
	srv := &http.Server{
		Addr:         cfg.Server.GetAddr(),       // 服务端口
		Handler:      handler,                     // Gin路由引擎
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 1. 异步启动服务（避免阻塞信号监听）
	go func() {
		log.Info("服务器启动", "addr", srv.Addr)
		// 启动失败且非"服务关闭"错误时，退出程序
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("服务器启动失败", "error", err)
			os.Exit(1)
		}
	}()

	// 2. 监听退出信号（Ctrl+C / kill命令）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞等待信号
	log.Info("服务器关闭...")

	// 3. 优雅关闭服务（5秒超时，确保正在处理的请求完成）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("服务器关闭失败", "error", err)
		os.Exit(1)
	}

	log.Info("服务器关闭成功")
}
