package main

import (
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/rabbitmq"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/core-api.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	log.Println("========================================")
	log.Println("本地日志写入工作进程启动中...")
	log.Println("========================================")

	// 2. 初始化 RabbitMQ
	mq, err := rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("初始化 RabbitMQ 失败: %v", err)
	}
	defer mq.Close()

	// 3. 声明交换机（fanout 类型）
	err = mq.DeclareExchange(c.RabbitMQ.LogExchange, "fanout")
	if err != nil {
		log.Fatalf("声明交换机失败: %v", err)
	}

	// 4. 声明本地日志队列
	err = mq.DeclareQueue(c.RabbitMQ.LocalLogQueue)
	if err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}

	// 5. 绑定队列到交换机
	err = mq.BindQueueToExchange(c.RabbitMQ.LocalLogQueue, c.RabbitMQ.LogExchange, "")
	if err != nil {
		log.Fatalf("绑定队列失败: %v", err)
	}

	// 6. 创建日志文件
	logFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}
	defer logFile.Close()

	// 7. 定义本地文件写入处理函数
	localLogHandler := func(logMsg *rabbitmq.LogMessage) error {
		log.Printf("开始写入本地日志: trace_id=%s, level=%s", logMsg.TraceID, logMsg.Level)

		// 序列化为 JSON
		jsonData, err := json.Marshal(logMsg)
		if err != nil {
			return fmt.Errorf("序列化日志失败: %w", err)
		}

		// 写入文件（每行一条 JSON）
		_, err = logFile.Write(append(jsonData, '\n'))
		if err != nil {
			return fmt.Errorf("写入日志文件失败: %w", err)
		}

		log.Printf("本地日志写入成功: %s", logMsg.TraceID)
		return nil
	}

	// 8. 创建消费者
	consumer := rabbitmq.NewLogConsumer(mq, c.RabbitMQ.LocalLogQueue, localLogHandler)

	// 9. 启动消费者（阻塞运行）
	log.Println("✓ 本地日志消费者已启动，监听队列:", c.RabbitMQ.LocalLogQueue)
	err = consumer.Start()
	if err != nil {
		log.Fatalf("消费者启动失败: %v", err)
	}
}
