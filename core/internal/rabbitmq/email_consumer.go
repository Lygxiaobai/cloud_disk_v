package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailHandler 邮件处理函数类型
// 参数：email（收件人邮箱）, code（验证码）
// 返回：error（发送失败时返回错误）
type EmailHandler func(email string, code string) error

// EmailConsumer 邮件消费者
type EmailConsumer struct {
	rabbitmq  *RabbitMQ
	queueName string
	handler   EmailHandler // 邮件发送处理函数
}

// NewEmailConsumer 创建邮件消费者
// rabbitmq: RabbitMQ 连接
// queueName: 队列名称
// handler: 邮件发送处理函数（实际发送邮件的逻辑）
func NewEmailConsumer(rabbitmq *RabbitMQ, queueName string, handler EmailHandler) *EmailConsumer {
	return &EmailConsumer{
		rabbitmq:  rabbitmq,
		queueName: queueName,
		handler:   handler,
	}
}

// Start 启动消费者（阻塞运行）
func (c *EmailConsumer) Start() error {
	// 1. 获取消息通道
	msgs, err := c.rabbitmq.GetChannel().Consume(
		c.queueName, // 队列名称
		"",          // consumer: 消费者标识（空字符串表示自动生成）
		false,       // autoAck: 手动确认（false = 需要手动 Ack）
		false,       // exclusive: 不独占
		false,       // noLocal: 不接收同一连接发布的消息
		false,       // noWait: 等待服务器确认
		nil,         // args: 额外参数
	)
	if err != nil {
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	log.Printf("✓ 邮件消费者启动成功，监听队列: %s", c.queueName)

	// 2. 持续监听消息
	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			// 处理消息
			c.handleMessage(msg)
		}
	}()

	log.Println("等待邮件任务...")
	<-forever // 阻塞，保持消费者运行

	return nil
}

// handleMessage 处理单条消息
func (c *EmailConsumer) handleMessage(msg amqp.Delivery) {
	// 1. 解析消息
	var emailMsg EmailMessage
	err := json.Unmarshal(msg.Body, &emailMsg)
	if err != nil {
		log.Printf("✗ 解析消息失败: %v, 消息内容: %s", err, string(msg.Body))
		// 解析失败，拒绝消息（不重新入队）
		msg.Nack(false, false)
		return
	}

	log.Printf("收到邮件任务: email=%s, code=%s", emailMsg.Email, emailMsg.Code)

	// 2. 调用邮件发送处理函数
	err = c.handler(emailMsg.Email, emailMsg.Code)
	if err != nil {
		log.Printf("✗ 发送邮件失败: %v", err)
		// 发送失败，拒绝消息并重新入队（让其他消费者重试）
		msg.Nack(false, true)
		return
	}

	log.Printf("✓ 邮件发送成功: %s", emailMsg.Email)

	// 3. 确认消息（从队列中删除）
	msg.Ack(false)
}
