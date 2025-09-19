package main

import (
	"flag"
	"log/slog"
	"os"

	"qrcode-server/config" // 配置模块
	"qrcode-server/router" // 路由模块
	"qrcode-server/server" // 服务器模块
	"qrcode-server/utils"  // 日志模块
)

func main() {
	// 第一步：解析命令行参数（指定配置文件路径）
	// -config：参数名；./config.yaml：默认值；指定配置文件路径（如 -config ./conf/prod.yaml）：参数说明
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yaml", "specify config file path (e.g. -config ./conf/prod.yaml)")
	flag.Parse() // 必须调用，解析命令行参数

	// 第二步：加载配置（使用命令行指定的路径，自动创建日志目录）
	slog.Info("加载配置文件", "path", configPath)
	if err := config.InitConfig(configPath); err != nil {
		slog.Error("无法加载配置", "error", err, "path", configPath)
		os.Exit(1) // 配置加载失败，退出程序
	}

	// 第三步：初始化日志（使用配置中的日志参数：级别、路径、大小限制等）
	utils.InitLogger()
	log := utils.GetLogger()
	log.Info("Logger成功初始化")

	// 第四步：初始化路由（注入日志实例）
	r := router.InitRouter(log)
	log.Info("路由成功初始化")

	// 第五步：启动 HTTP 服务（使用配置中的端口）
	server.Start(r, log)
}
