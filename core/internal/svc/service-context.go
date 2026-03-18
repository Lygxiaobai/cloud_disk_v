// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"cloud_disk/core/internal/cache"
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/middleware"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/rabbitmq"
	"cloud_disk/core/internal/task"
	"log"

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
	RabbitMQ      *rabbitmq.RabbitMQ
	EmailProducer *rabbitmq.EmailProducer
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

	// 初始化 RabbitMQ
	mq, err := rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("初始化 RabbitMQ 失败: %v", err)
	}

	// 声明邮件队列
	err = mq.DeclareQueue(c.RabbitMQ.EmailQueue)
	if err != nil {
		log.Fatalf("声明邮件队列失败: %v", err)
	}

	// 创建邮件生产者
	emailProducer := rabbitmq.NewEmailProducer(mq, c.RabbitMQ.EmailQueue)

	// 启动邮件消费者（在后台 Goroutine 中运行）
	startEmailConsumer(mq, c.RabbitMQ.EmailQueue)

	return &ServiceContext{
		Config:        c,
		Engine:        engine,
		RDB:           rdb,
		ShareCache:    shareCache,
		HotShareTask:  hotShareTask,
		RabbitMQ:      mq,
		EmailProducer: emailProducer,
		Auth:          middleware.NewAuthMiddleware().Handle,
		Casbin:        middleware.NewCasbinMiddleware(enforcer).Handle,
		ErrorRecovery: middleware.NewErrorRecoveryMiddleware().Handle,
	}
}

// startEmailConsumer 启动邮件消费者（在后台运行）
func startEmailConsumer(mq *rabbitmq.RabbitMQ, queueName string) {
	// 邮件发送处理函数
	emailHandler := func(email string, code string) error {
		log.Printf("开始发送邮件: email=%s, code=%s", email, code)
		err := helper.MailCodeSend(email, code)
		if err != nil {
			log.Printf("发送邮件失败: %v", err)
			return err
		}
		log.Printf("邮件发送成功: %s", email)
		return nil
	}

	// 创建消费者
	consumer := rabbitmq.NewEmailConsumer(mq, queueName, emailHandler)

	// 在后台 Goroutine 中启动消费者
	go func() {
		log.Printf("✓ 邮件消费者已在后台启动，监听队列: %s", queueName)
		err := consumer.Start()
		if err != nil {
			log.Fatalf("邮件消费者启动失败: %v", err)
		}
	}()
}
