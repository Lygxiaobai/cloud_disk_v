package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ 连接管理器（全局共享）。
// amqp.Connection 是线程安全的，可以在多个 goroutine 之间共享；
// 但 amqp.Channel 不是线程安全的，因此每个 producer/consumer 必须拥有自己的 Channel。
type RabbitMQ struct {
	conn *amqp.Connection // RabbitMQ 连接（线程安全）
	URL  string           // 连接地址
}

// NewRabbitMQ 创建 RabbitMQ 实例
// url: RabbitMQ 连接地址，格式：amqp://用户名:密码@地址:端口/
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	rabbitmq := &RabbitMQ{
		URL: url,
	}

	err := rabbitmq.connect()
	if err != nil {
		return nil, err
	}

	return rabbitmq, nil
}

// connect 建立 RabbitMQ 连接（仅连接，不创建 Channel）
func (r *RabbitMQ) connect() error {
	var err error
	r.conn, err = amqp.Dial(r.URL)
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}

	log.Println("RabbitMQ 连接成功")
	return nil
}

// NewChannel 创建一个新的独立 Channel。
// 每个 producer/consumer 应该在初始化时调用此方法获取自己的 Channel。
func (r *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建信道失败: %w", err)
	}
	return ch, nil
}

// DeclareQueue 声明队列（如果队列不存在则创建）
func (r *RabbitMQ) DeclareQueue(queueName string) error {
	ch, err := r.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName, // 队列名称
		true,      // durable: 持久化
		false,     // autoDelete: 不自动删除
		false,     // exclusive: 不独占
		false,     // noWait: 等待服务器确认
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列 %s 失败: %w", queueName, err)
	}

	log.Printf("队列声明成功: %s", queueName)
	return nil
}

// DeclareExchange 声明交换机
func (r *RabbitMQ) DeclareExchange(exchangeName string, exchangeType string) error {
	ch, err := r.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // 交换机名称
		exchangeType, // 交换机类型
		true,         // durable: 持久化
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("声明交换机 %s 失败: %w", exchangeName, err)
	}

	log.Printf("交换机声明成功: %s (类型: %s)", exchangeName, exchangeType)
	return nil
}

// BindQueueToExchange 绑定队列到交换机
func (r *RabbitMQ) BindQueueToExchange(queueName string, exchangeName string, routingKey string) error {
	ch, err := r.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.QueueBind(
		queueName,    // 队列名称
		routingKey,   // routing key
		exchangeName, // 交换机名称
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("绑定队列 %s 到交换机 %s 失败: %w", queueName, exchangeName, err)
	}

	log.Printf("队列绑定成功: %s -> %s", queueName, exchangeName)
	return nil
}

// Close 关闭连接
func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	log.Println("RabbitMQ 连接已关闭")
}
