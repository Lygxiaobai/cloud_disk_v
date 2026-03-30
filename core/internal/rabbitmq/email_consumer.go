package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailHandler 邮件处理函数类型
type EmailHandler func(email string, code string) error

// EmailConsumer 邮件消费者（Start 时创建独立的 Channel）
type EmailConsumer struct {
	rabbitmq  *RabbitMQ
	queueName string
	handler   EmailHandler
}

// NewEmailConsumer 创建邮件消费者
func NewEmailConsumer(rabbitmq *RabbitMQ, queueName string, handler EmailHandler) *EmailConsumer {
	return &EmailConsumer{
		rabbitmq:  rabbitmq,
		queueName: queueName,
		handler:   handler,
	}
}

// Start 启动消费者（阻塞运行）
func (c *EmailConsumer) Start() error {
	// 创建独立的 Channel
	ch, err := c.rabbitmq.NewChannel()
	if err != nil {
		return fmt.Errorf("创建消费者 Channel 失败: %w", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		c.queueName, // 队列名称
		"",          // consumer: 自动生成
		false,       // autoAck: 手动确认
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	log.Printf("邮件消费者启动成功，监听队列: %s", c.queueName)
	log.Println("等待邮件任务...")

	// 直接阻塞在 range 上，Channel 关闭时 range 自动退出
	for msg := range msgs {
		c.handleMessage(msg)
	}

	log.Println("邮件消费者已退出（Channel 已关闭）")
	return nil
}

// handleMessage 处理单条消息
func (c *EmailConsumer) handleMessage(msg amqp.Delivery) {
	var emailMsg EmailMessage
	err := json.Unmarshal(msg.Body, &emailMsg)
	if err != nil {
		log.Printf("解析消息失败: %v, 消息内容: %s", err, string(msg.Body))
		msg.Nack(false, false)
		return
	}

	log.Printf("收到邮件任务: email=%s, code=%s", emailMsg.Email, emailMsg.Code)

	err = c.handler(emailMsg.Email, emailMsg.Code)
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		msg.Nack(false, true)
		return
	}

	log.Printf("邮件发送成功: %s", emailMsg.Email)
	msg.Ack(false)
}
