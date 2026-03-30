package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// LogHandler 日志处理函数类型
type LogHandler func(logMsg *LogMessage) error

// LogConsumer 日志消费者（Start 时创建独立的 Channel）
type LogConsumer struct {
	rabbitmq  *RabbitMQ
	queueName string
	handler   LogHandler
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
	ch, err := c.rabbitmq.NewChannel()
	if err != nil {
		return fmt.Errorf("创建消费者 Channel 失败: %w", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		c.queueName, // 队列名称
		"",          // consumer
		false,       // autoAck: 手动确认
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("注册日志消费者失败: %w", err)
	}

	log.Printf("日志消费者启动成功，监听队列: %s", c.queueName)
	log.Println("等待日志消息...")

	for msg := range msgs {
		c.handleMessage(msg)
	}

	log.Println("日志消费者已退出（Channel 已关闭）")
	return nil
}

// handleMessage 处理单条消息
func (c *LogConsumer) handleMessage(msg amqp.Delivery) {
	var logMsg LogMessage
	err := json.Unmarshal(msg.Body, &logMsg)
	if err != nil {
		log.Printf("解析日志消息失败: %v, 消息内容: %s", err, string(msg.Body))
		msg.Nack(false, false)
		return
	}

	log.Printf("收到日志消息: level=%s, trace_id=%s, message=%s",
		logMsg.Level, logMsg.TraceID, logMsg.Message)

	err = c.handler(&logMsg)
	if err != nil {
		log.Printf("处理日志失败: %v", err)
		msg.Nack(false, true)
		return
	}

	log.Printf("日志处理成功: %s", logMsg.TraceID)
	msg.Ack(false)
}
