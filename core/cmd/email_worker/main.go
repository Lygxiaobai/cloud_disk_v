package main

import (
	"cloud_disk/core/internal/config"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/rabbitmq"
	"flag"
	"log"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/core-api.yaml", "配置文件路径")

// 全局邮件配置（在 main 中初始化，供 emailHandler 使用）
var mailCfg helper.MailConfig

func main() {
	flag.Parse()

	// 1. 加载配置（启用环境变量替换）
	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	mailCfg = helper.MailConfig{
		From:       c.Mail.From,
		Host:       c.Mail.Host,
		Username:   c.Mail.Username,
		Password:   c.Mail.Password,
		ServerName: c.Mail.ServerName,
	}

	log.Println("==========================================")
	log.Println("邮件消费者服务启动")
	log.Println("==========================================")

	// 2. 连接 RabbitMQ
	mq, err := rabbitmq.NewRabbitMQ(c.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
	}
	defer mq.Close()

	// 3. 声明邮件队列
	err = mq.DeclareQueue(c.RabbitMQ.EmailQueue)
	if err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}

	// 4. 创建邮件消费者
	consumer := rabbitmq.NewEmailConsumer(mq, c.RabbitMQ.EmailQueue, emailHandler)

	// 5. 启动消费者（阻塞运行）
	log.Printf("开始监听队列: %s", c.RabbitMQ.EmailQueue)
	err = consumer.Start()
	if err != nil {
		log.Fatalf("启动消费者失败: %v", err)
	}
}

// emailHandler 邮件发送处理函数
// 这个函数会被消费者调用，用于实际发送邮件
func emailHandler(email string, code string) error {
	log.Printf("开始发送邮件: email=%s, code=%s", email, code)

	err := helper.MailCodeSend(email, code, mailCfg)
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		return err
	}

	log.Printf("邮件发送成功: %s", email)
	return nil
}
