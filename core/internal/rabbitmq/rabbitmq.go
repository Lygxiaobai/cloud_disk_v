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

// DeclareExchange 声明交换机
// exchangeName: 交换机名称
// exchangeType: 交换机类型（fanout, direct, topic, headers）
func (r *RabbitMQ) DeclareExchange(exchangeName string, exchangeType string) error {
	err := r.channel.ExchangeDeclare(
		exchangeName, // 交换机名称
		exchangeType, // 交换机类型
		true,         // durable: 持久化
		false,        // autoDelete: 不自动删除
		false,        // internal: 不是内部交换机
		false,        // noWait: 等待服务器确认
		nil,          // arguments: 额外参数
	)
	if err != nil {
		return fmt.Errorf("声明交换机 %s 失败: %w", exchangeName, err)
	}

	log.Printf("✓ 交换机声明成功: %s (类型: %s)", exchangeName, exchangeType)
	return nil
}

// BindQueueToExchange 绑定队列到交换机
// queueName: 队列名称
// exchangeName: 交换机名称
// routingKey: 路由键（fanout 类型可以为空）
func (r *RabbitMQ) BindQueueToExchange(queueName string, exchangeName string, routingKey string) error {
	err := r.channel.QueueBind(
		queueName,    // 队列名称
		routingKey,   // routing key
		exchangeName, // 交换机名称
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("绑定队列 %s 到交换机 %s 失败: %w", queueName, exchangeName, err)
	}

	log.Printf("✓ 队列绑定成功: %s -> %s", queueName, exchangeName)
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
