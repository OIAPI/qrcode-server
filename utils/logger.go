package utils

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat-go/file-rotatelogs" // 日志切割库
	"qrcode-server/config"                  // 你的配置模块名
)

var (
	globalLogger *slog.Logger
	initOnce     sync.Once
)

// InitLogger 初始化全局日志（支持控制台输出+文件大小切割）
func InitLogger() {
	initOnce.Do(func() {
		cfg := config.Get()
		logLevel := parseLogLevel(cfg.Log.Level)

		// 1. 准备输出目标：默认包含控制台
		var writers []io.Writer
		writers = append(writers, os.Stdout)

		// 2. 处理日志文件（含大小限制+切割）
		if cfg.Log.Path != "" {
			// 2.1 解析日志路径（分离目录和文件名前缀）
			logDir := filepath.Dir(cfg.Log.Path)   // 日志目录（如 "logs"）
			logName := filepath.Base(cfg.Log.Path) // 日志文件名前缀（如 "qrcode.log"）
			// 确保日志目录存在（不存在则创建）
			if err := os.MkdirAll(logDir, 0755); err != nil {
				slog.Error("create log directory failed", "dir", logDir, "error", err)
			}

			// 2.2 配置日志切割（按大小限制）
			// - 单文件最大 size：cfg.Log.MaxSize（MB 转 Byte）
			// - 切割后文件名格式：qrcode.log.20240920123456（前缀+时间戳）
			rotator, err := rotatelogs.New(
				// 切割后的日志文件名模板（路径+前缀+时间戳）
				filepath.Join(logDir, logName+".%Y%m%d%H%M%S"),
				rotatelogs.WithMaxAge(7*24*time.Hour), // 日志保留时间（可选，如7天）
				rotatelogs.WithRotationSize(int64(cfg.Log.MaxSize) * 1024 * 1024), // 单文件最大大小（MB→Byte）
			)
			if err != nil {
				slog.Error("init log rotator failed", "path", cfg.Log.Path, "error", err)
			} else {
				// 将切割 writer 加入输出目标
				writers = append(writers, rotator)
			}
		}

		// 3. 合并所有输出目标（控制台+切割文件）
		multiWriter := io.MultiWriter(writers...)
		// 4. 创建 slog 处理器并初始化全局日志
		handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: logLevel,
		})
		globalLogger = slog.New(handler)
		slog.SetDefault(globalLogger)
	})
}

// GetLogger 对外提供全局日志实例
func GetLogger() *slog.Logger {
	if globalLogger == nil {
		InitLogger()
	}
	return globalLogger
}

// parseLogLevel 转换配置的日志级别为 slog.Level
func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
