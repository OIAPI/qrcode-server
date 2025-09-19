package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 项目总配置结构体
type Config struct {
	Server ServerConfig `yaml:"server"` // 服务配置（拆分 Host/Port）
	Log    LogConfig    `yaml:"log"`
	QRCode QRCodeConfig `yaml:"qrcode"`
}

// ServerConfig 服务配置（拆分 Host 和 Port，独立配置）
type ServerConfig struct {
	Host         string `yaml:"host"`          // 绑定地址
	Port         string `yaml:"port"`          // 端口（如 "8080"）
	ReadTimeout  int    `yaml:"read_timeout"`  // 读取超时（秒）
	WriteTimeout int    `yaml:"write_timeout"` // 写入超时（秒）
}

// GetAddr 拼接 Host:Port 完整地址（给外部模块调用，无需手动拼接）
func (s *ServerConfig) GetAddr() string {
	// 处理 Host 为空的情况（默认用 "0.0.0.0"，兼容仅配置 Port 的场景）
	host := s.Host
	if strings.TrimSpace(host) == "" {
		host = "0.0.0.0"
	}
	return host + ":" + s.Port
}

// 以下 LogConfig、QRCodeConfig 结构体保持不变...
type LogConfig struct {
	Level     string `yaml:"level"`
	Path      string `yaml:"path"`
	MaxSize   int    `yaml:"max_size"`
	MaxAge    int    `yaml:"max_age"`
	MaxBackup int    `yaml:"max_backup"`
}

type QRCodeConfig struct {
	DefaultSize  int      `yaml:"default_size"`
	DefaultLevel string   `yaml:"default_level"`
	SupportTypes []string `yaml:"support_types"`
}

// InitConfig 初始化配置（更新 Server 默认值）
func InitConfig(configPath string) error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return err
	}

	// 设置默认值（重点更新 Server 的 Host/Port 默认值）
	setDefaultConfig(&cfg)
	// 自动创建日志目录
	if err := createRequiredDirs(&cfg); err != nil {
		return err
	}

	globalConfig = &cfg
	return nil
}

// setDefaultConfig 配置默认值（更新 Server 部分）
func setDefaultConfig(cfg *Config) {
	// Server 配置默认值：Host 空默认 "0.0.0.0"，Port 空默认 "8080"
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080" // 端口默认 8080
	}
	// Host 不设默认值，空时在 GetAddr() 中自动补 "0.0.0.0"
	if cfg.Server.ReadTimeout <= 0 {
		cfg.Server.ReadTimeout = 5
	}
	if cfg.Server.WriteTimeout <= 0 {
		cfg.Server.WriteTimeout = 10
	}

	// 以下 Log、QRCode 默认值保持不变...
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "logs/qrcode.log"
	}
	if cfg.Log.MaxSize <= 0 {
		cfg.Log.MaxSize = 10
	}
	if cfg.Log.MaxAge <= 0 {
		cfg.Log.MaxAge = 7
	}
	if cfg.Log.MaxBackup <= 0 {
		cfg.Log.MaxBackup = 10
	}

	if cfg.QRCode.DefaultSize <= 0 {
		cfg.QRCode.DefaultSize = 300
	}
	if cfg.QRCode.DefaultLevel == "" {
		cfg.QRCode.DefaultLevel = "M"
	}
	if len(cfg.QRCode.SupportTypes) == 0 {
		cfg.QRCode.SupportTypes = []string{"png", "jpeg"}
	}
}

// createRequiredDirs 自动创建目录（保持不变）
func createRequiredDirs(cfg *Config) error {
	logDir := filepath.Dir(cfg.Log.Path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	return nil
}

// Get 对外提供全局配置（保持不变）
func Get() *Config {
	if globalConfig == nil {
		panic("配置未初始化")
	}
	return globalConfig
}

var globalConfig *Config
