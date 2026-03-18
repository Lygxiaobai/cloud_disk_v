package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// LogMessage 日志消息结构（与 logger.ErrorLog 保持一致）
type LogMessage struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	TraceID    string                 `json:"trace_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Message    string                 `json:"message"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// LogProducer 日志生产者
type LogProducer struct {
	rabbitmq     *RabbitMQ
	exchangeName string
}

// NewLogProducer 创建日志生产者
func NewLogProducer(rabbitmq *RabbitMQ, exchangeName string) *LogProducer {
	return &LogProducer{
		rabbitmq:     rabbitmq,
		exchangeName: exchangeName,
	}
}

// SendLogMessage 发送日志消息到交换机（fanout 广播）
func (p *LogProducer) SendLogMessage(log *LogMessage) error {
	// 1. 序列化日志消息
	body, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("序列化日志消息失败: %w", err)
	}

	// 2. 发布到 fanout 交换机（routing key 为空）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.rabbitmq.GetChannel().PublishWithContext(
		ctx,
		p.exchangeName, // 交换机名称
		"",             // routing key（fanout 不需要）
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 持久化消息
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("发送日志消息到 MQ 失败: %w", err)
	}

	return nil
}
