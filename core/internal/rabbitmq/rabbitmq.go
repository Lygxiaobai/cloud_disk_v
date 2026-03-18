package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ 连接管理器（全局共享）
type RabbitMQ struct {
	conn    *amqp.Connection // RabbitMQ 连接
	channel *amqp.Channel    // RabbitMQ 信道
	URL     string           // 连接地址
}

// NewRabbitMQ 创建 RabbitMQ 实例
// url: RabbitMQ 连接地址，格式：amqp://用户名:密码@地址:端口/
// 例如：amqp://guest:guest@localhost:5672/
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	rabbitmq := &RabbitMQ{
		URL: url,
	}

	// 建立连接
	err := rabbitmq.connect()
	if err != nil {
		return nil, err
	}

	return rabbitmq, nil
}

// connect 建立 RabbitMQ 连接
func (r *RabbitMQ) connect() error {
	var err error

	// 1. 创建连接
	r.conn, err = amqp.Dial(r.URL)
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	// 2. 创建信道
	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("创建信道失败: %w", err)
	}

	log.Println("✓ RabbitMQ 连接成功")
	return nil
}

// DeclareQueue 声明队列（如果队列不存在则创建）
// queueName: 队列名称
func (r *RabbitMQ) DeclareQueue(queueName string) error {
	_, err := r.channel.QueueDeclare(
		queueName, // 队列名称
		true,      // durable: 持久化，RabbitMQ 重启后队列不会丢失
		false,     // autoDelete: 不自动删除
		false,     // exclusive: 不独占（允许多个消费者）
		false,     // noWait: 等待服务器确认
		nil,       // arguments: 额外参数
	)
	if err != nil {
		return fmt.Errorf("声明队列 %s 失败: %w", queueName, err)
	}

	log.Printf("✓ 队列声明成功: %s", queueName)
	return nil
}

// GetChannel 获取信道（供生产者和消费者使用）
func (r *RabbitMQ) GetChannel() *amqp.Channel {
	return r.channel
}

// Close 关闭连接
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	log.Println("RabbitMQ 连接已关闭")
}
