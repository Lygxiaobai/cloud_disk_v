package svc

import (
	"cloud_disk/core/internal/cache"
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/middleware"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/rabbitmq"
	"cloud_disk/core/internal/storage"
	"cloud_disk/core/internal/task"
	"log"

	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"xorm.io/xorm"
)

// ServiceContext 是 go-zero 项目里的依赖注入容器。
// 所有 handler / logic 都通过它拿到数据库、Redis、OSS、RabbitMQ、中间件等基础能力。
type ServiceContext struct {
	Config        config.Config
	Engine        *xorm.Engine
	RDB           *redis.Client
	OSS           *storage.OSSService
	ShareCache    *cache.ShareCache
	HotShareTask  *task.HotShareTask
	RabbitMQ      *rabbitmq.RabbitMQ
	EmailProducer *rabbitmq.EmailProducer
	Auth          rest.Middleware
	Casbin        rest.Middleware
	ErrorRecovery rest.Middleware
}

// NewServiceContext 在服务启动时统一初始化所有基础依赖。
func NewServiceContext(c config.Config) *ServiceContext {
	adapter := fileadapter.NewAdapter(c.Casbin.PolicyPath)
	enforcer, err := casbin.NewSyncedEnforcer(c.Casbin.ModelPath, adapter)
	if err != nil {
		panic(err)
	}

	engine := models.Init(c.Mysql.DataSource)
	rdb := models.InitRedis(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	ossService := storage.NewOSSService(c.OSS)

	shareCache := cache.NewShareCache(rdb)
	hotShareTask := task.NewHotShareTask(engine, shareCache)
	hotShareTask.Start()

	var (
		mq            *rabbitmq.RabbitMQ
		emailProducer *rabbitmq.EmailProducer
	)

	// RabbitMQ 在本地联调阶段允许缺席，
	// 这样主链路（上传 / 预览 / 列表）不会被邮件和异步日志阻塞。
	mq, err = rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Printf("RabbitMQ unavailable, continue without async email/log pipeline: %v", err)
	} else {
		if err := mq.DeclareQueue(c.RabbitMQ.EmailQueue); err != nil {
			log.Printf("declare email queue failed, skip email worker: %v", err)
		} else {
			ep, epErr := rabbitmq.NewEmailProducer(mq, c.RabbitMQ.EmailQueue)
			if epErr != nil {
				log.Printf("create email producer failed, skip email: %v", epErr)
			} else {
				emailProducer = ep
				startEmailConsumer(mq, c.RabbitMQ.EmailQueue)
			}
		}
		initAsyncLogSystem(mq, c)
	}

	return &ServiceContext{
		Config:        c,
		Engine:        engine,
		RDB:           rdb,
		OSS:           ossService,
		ShareCache:    shareCache,
		HotShareTask:  hotShareTask,
		RabbitMQ:      mq,
		EmailProducer: emailProducer,
		Auth:          middleware.NewAuthMiddleware().Handle,
		Casbin:        middleware.NewCasbinMiddleware(enforcer).Handle,
		ErrorRecovery: middleware.NewErrorRecoveryMiddleware().Handle,
	}
}

// startEmailConsumer 在本地 goroutine 中启动邮件消费者。
// 这样注册发验证码时只负责投递任务，不阻塞当前 HTTP 请求。
func startEmailConsumer(mq *rabbitmq.RabbitMQ, queueName string) {
	emailHandler := func(email string, code string) error {
		log.Printf("start sending email, email=%s", email)
		if err := helper.MailCodeSend(email, code); err != nil {
			log.Printf("send email failed: %v", err)
			return err
		}
		log.Printf("send email success: %s", email)
		return nil
	}

	consumer := rabbitmq.NewEmailConsumer(mq, queueName, emailHandler)
	go func() {
		log.Printf("email consumer started, queue=%s", queueName)
		if err := consumer.Start(); err != nil {
			log.Fatalf("start email consumer failed: %v", err)
		}
	}()
}

// initAsyncLogSystem 初始化基于 RabbitMQ 的异步日志链路。
func initAsyncLogSystem(mq *rabbitmq.RabbitMQ, c config.Config) {
	if err := mq.DeclareExchange(c.RabbitMQ.LogExchange, "fanout"); err != nil {
		log.Fatalf("declare log exchange failed: %v", err)
	}
	if err := mq.DeclareQueue(c.RabbitMQ.LocalLogQueue); err != nil {
		log.Fatalf("declare local log queue failed: %v", err)
	}
	if err := mq.DeclareQueue(c.RabbitMQ.ESLogQueue); err != nil {
		log.Fatalf("declare ES log queue failed: %v", err)
	}
	if err := mq.BindQueueToExchange(c.RabbitMQ.LocalLogQueue, c.RabbitMQ.LogExchange, ""); err != nil {
		log.Fatalf("bind local log queue failed: %v", err)
	}
	if err := mq.BindQueueToExchange(c.RabbitMQ.ESLogQueue, c.RabbitMQ.LogExchange, ""); err != nil {
		log.Fatalf("bind ES log queue failed: %v", err)
	}

	logProducer, err := rabbitmq.NewLogProducer(mq, c.RabbitMQ.LogExchange)
	if err != nil {
		log.Fatalf("create log producer failed: %v", err)
	}
	if err := logger.InitAsyncLogger("logs/error.log", logProducer); err != nil {
		log.Fatalf("initialize async logger failed: %v", err)
	}

	log.Println("async log system initialized")
}
