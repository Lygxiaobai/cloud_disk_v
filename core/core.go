// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"log"

	"cloud_disk/core/internal/config"
	appErrors "cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/handler"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/core-api.yaml", "the config file")

func main() {
	flag.Parse()

	// 初始化简化版日志系统
	if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	defer logger.Close()
	log.Println("日志系统初始化成功")

	var c config.Config
	// UseEnv 让 YAML 中的 ${VAR} 占位符从环境变量读取
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	httpx.SetErrorHandlerCtx(appErrors.ErrorResponse)

	serverOpts := []rest.RunOption{}
	if len(c.CORS.AllowedOrigins) > 0 {
		serverOpts = append(serverOpts, rest.WithCors(c.CORS.AllowedOrigins...))
	}
	server := rest.MustNewServer(c.RestConf, serverOpts...)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
