package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// LogHandler 日志处理函数类型
// 参数：logMsg（日志消息）
// 返回：error（处理失败时返回错误）
type LogHandler func(logMsg *LogMessage) error

// LogConsumer 日志消费者
type LogConsumer struct {
	rabbitmq  *RabbitMQ
	queueName string
	handler   LogHandler // 日志处理函数
}

// NewLogConsumer 创建日志消费者
func NewLogConsumer(rabbitmq *RabbitMQ, queueName string, handler LogHandler) *LogConsumer {
	return &LogConsumer{
		rabbitmq:  rabbitmq,
		queueName: queueName,
		handler:   handler,
	}
}

// Start 启动消费者（阻塞运行）
func (c *LogConsumer) Start() error {
	// 1. 获取消息通道
	msgs, err := c.rabbitmq.GetChannel().Consume(
		c.queueName, // 队列名称
		"",          // consumer: 消费者标识（空字符串表示自动生成）
		false,       // autoAck: 手动确认
		false,       // exclusive: 不独占
		false,       // noLocal
		false,       // noWait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("注册日志消费者失败: %w", err)
	}

	log.Printf("✓ 日志消费者启动成功，监听队列: %s", c.queueName)

	// 2. 持续监听消息
	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			c.handleMessage(msg)
		}
	}()

	log.Println("等待日志消息...")
	<-forever // 阻塞

	return nil
}

// handleMessage 处理单条消息
func (c *LogConsumer) handleMessage(msg amqp.Delivery) {
	// 1. 解析消息
	var logMsg LogMessage
	err := json.Unmarshal(msg.Body, &logMsg)
	if err != nil {
		log.Printf("✗ 解析日志消息失败: %v, 消息内容: %s", err, string(msg.Body))
		msg.Nack(false, false) // 拒绝消息，不重新入队
		return
	}

	log.Printf("收到日志消息: level=%s, trace_id=%s, message=%s",
		logMsg.Level, logMsg.TraceID, logMsg.Message)

	// 2. 调用日志处理函数
	err = c.handler(&logMsg)
	if err != nil {
		log.Printf("✗ 处理日志失败: %v", err)
		msg.Nack(false, true) // 拒绝消息并重新入队
		return
	}

	log.Printf("✓ 日志处理成功: %s", logMsg.TraceID)

	// 3. 确认消息
	msg.Ack(false)
}
