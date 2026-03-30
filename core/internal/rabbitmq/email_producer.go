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

// EmailProducer 邮件生产者（持有独立的 Channel）
type EmailProducer struct {
	channel   *amqp.Channel
	queueName string
}

// NewEmailProducer 创建邮件生产者，会从连接上创建一个独立的 Channel
func NewEmailProducer(mq *RabbitMQ, queueName string) (*EmailProducer, error) {
	ch, err := mq.NewChannel()
	if err != nil {
		return nil, fmt.Errorf("创建邮件生产者 Channel 失败: %w", err)
	}
	return &EmailProducer{
		channel:   ch,
		queueName: queueName,
	}, nil
}

// SendEmailTask 发送邮件任务到队列
func (p *EmailProducer) SendEmailTask(email string, code string) error {
	message := EmailMessage{
		Email: email,
		Code:  code,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		ctx,
		"",          // exchange: 使用默认交换机
		p.queueName, // routing key: 队列名称
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息到队列失败: %w", err)
	}

	return nil
}

// Close 关闭生产者的 Channel
func (p *EmailProducer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
}
