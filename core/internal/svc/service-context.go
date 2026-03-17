// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"cloud_disk/core/internal/cache"
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/middleware"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/task"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"xorm.io/xorm"
)

type ServiceContext struct {
	Config        config.Config
	Engine        *xorm.Engine
	RDB           *redis.Client
	ShareCache    *cache.ShareCache
	HotShareTask  *task.HotShareTask
	Auth          rest.Middleware
	Casbin        rest.Middleware
	ErrorRecovery rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	adapter := fileadapter.NewAdapter(c.Casbin.PolicyPath)
	enforcer, err := casbin.NewEnforcer(c.Casbin.ModelPath, adapter)
	if err != nil {
		panic(err)
	}

	// 初始化数据库和 Redis
	engine := models.Init(c.Mysql.DataSource)
	rdb := models.InitRedis(c.Redis.Addr)

	// 初始化分享缓存
	shareCache := cache.NewShareCache(rdb)

	// 初始化热门分享统计任务
	hotShareTask := task.NewHotShareTask(engine, shareCache)
	hotShareTask.Start() // 启动定时任务

	return &ServiceContext{
		Config:        c,
		Engine:        engine,
		RDB:           rdb,
		ShareCache:    shareCache,
		HotShareTask:  hotShareTask,
		Auth:          middleware.NewAuthMiddleware().Handle,
		Casbin:        middleware.NewCasbinMiddleware(enforcer).Handle,
		ErrorRecovery: middleware.NewErrorRecoveryMiddleware().Handle,
	}
}
