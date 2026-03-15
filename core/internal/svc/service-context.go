// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/middleware"
	"cloud_disk/core/internal/models"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"xorm.io/xorm"
)

type ServiceContext struct {
	Config config.Config
	Engine *xorm.Engine
	RDB    *redis.Client
	Auth   rest.Middleware
	Casbin rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	adapter := fileadapter.NewAdapter(c.Casbin.PolicyPath)
	enforcer, err := casbin.NewEnforcer(c.Casbin.ModelPath, adapter)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		Engine: models.Init(c.Mysql.DataSource),
		RDB:    models.InitRedis(c.Redis.Addr),
		Auth:   middleware.NewAuthMiddleware().Handle,
		Casbin: middleware.NewCasbinMiddleware(enforcer).Handle,
	}
}
