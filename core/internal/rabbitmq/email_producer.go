package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailMessage 邮件消息结构
type EmailMessage struct {
	Email string `json:"email"` // 收件人邮箱
	Code  string `json:"code"`  // 验证码
}

// EmailProducer 邮件生产者
type EmailProducer struct {
	rabbitmq  *RabbitMQ
	queueName string
}

// NewEmailProducer 创建邮件生产者
func NewEmailProducer(rabbitmq *RabbitMQ, queueName string) *EmailProducer {
	return &EmailProducer{
		rabbitmq:  rabbitmq,
		queueName: queueName,
	}
}

// SendEmailTask 发送邮件任务到队列
func (p *EmailProducer) SendEmailTask(email string, code string) error {
	// 1. 构造消息
	message := EmailMessage{
		Email: email,
		Code:  code,
	}

	// 2. 序列化为 JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 3. 设置超时上下文（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 4. 发布消息到队列
	err = p.rabbitmq.GetChannel().PublishWithContext(
		ctx,
		"",          // exchange: 使用默认交换机
		p.queueName, // routing key: 队列名称
		false,       // mandatory: 如果没有队列绑定，不返回错误
		false,       // immediate: 不要求立即投递
		amqp.Publishing{
			ContentType:  "application/json", // 内容类型
			Body:         body,               // 消息体
			DeliveryMode: amqp.Persistent,    // 持久化消息（重启不丢失）
			Timestamp:    time.Now(),         // 时间戳
		},
	)

	if err != nil {
		return fmt.Errorf("发送消息到队列失败: %w", err)
	}

	return nil
}
