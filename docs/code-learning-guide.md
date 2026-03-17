# RabbitMQ 错误日志系统 - 代码学习指南（新手版）

> 本文档详细解释所有新增代码，帮助新手理解每一行代码的作用

## 📚 目录
1. [系统架构](#系统架构)
2. [核心概念](#核心概念)
3. [代码详解](#代码详解)

---

## 系统架构

### 整体流程图
```
业务代码发生错误
    ↓
logger.Error() 记录错误
    ↓
MQ Producer 发送消息
    ↓
RabbitMQ Exchange (fanout广播)
    ↓
┌──────────┴──────────┐
↓                     ↓
ES Queue         File Queue
↓                     ↓
ES Consumer      File Consumer
↓                     ↓
Elasticsearch    logs/error.log
```

### 为什么这样设计？

**传统方式的问题**：
- 直接写文件会阻塞业务代码（慢）
- 日志难以查询和分析
- 无法实时监控

**新方案的优势**：
- **异步**：发送到 MQ 后立即返回，不等待写入
- **双存储**：ES 用于查询，文件用于备份
- **解耦**：日志处理独立于业务逻辑

---

## 核心概念

### RabbitMQ 基础

1. **Producer（生产者）**：发送消息的一方
2. **Exchange（交换机）**：接收消息并分发到队列
3. **Queue（队列）**：存储消息
4. **Consumer（消费者）**：处理消息

### Fanout 模式
- 广播模式：一条消息发送到所有绑定的队列
- 本项目：一条错误日志同时发送到 ES 队列和文件队列

---

## 代码详解

### 1. 消息结构 (mq/message.go)

```go
package mq

import (
	"encoding/json"
	"time"
)

// ErrorLogMessage 错误日志消息结构
// 定义了一条错误日志包含哪些信息
type ErrorLogMessage struct {
	// 时间戳：错误发生的时间
	Timestamp  time.Time              `json:"timestamp"`

	// 日志级别：ERROR/FATAL/PANIC
	Level      string                 `json:"level"`

	// 服务名称：标识哪个服务产生的错误
	Service    string                 `json:"service"`

	// 追踪ID：用于追踪一个请求的完整链路
	TraceID    string                 `json:"trace_id"`

	// 用户ID：标识哪个用户的操作
	UserID     string                 `json:"user_id"`

	// HTTP 方法：GET/POST/PUT/DELETE
	Method     string                 `json:"method"`

	// 请求路径：如 /file/upload
	Path       string                 `json:"path"`

	// 错误信息：具体的错误描述
	ErrorMsg   string                 `json:"error_msg"`

	// 堆栈跟踪：错误发生时的代码调用链
	StackTrace string                 `json:"stack_trace"`

	// 额外信息：自定义字段（如文件名、大小等）
	// omitempty: 如果为空则不序列化
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// ToJSON 将消息转换为 JSON 字节数组
// 为什么？RabbitMQ 传输的是字节数组
func (m *ErrorLogMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 字节数组解析消息
// 为什么？消费者收到字节数组后需要还原成结构体
func FromJSON(data []byte) (*ErrorLogMessage, error) {
	var msg ErrorLogMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
```

**关键点**：
- `json:"timestamp"` 是结构体标签，指定 JSON 字段名
- `map[string]interface{}` 表示键是字符串，值可以是任意类型
- `*ErrorLogMessage` 返回指针，避免复制大对象

---

### 2. 生产者 (mq/producer.go)

```go
package mq

import (
	"fmt"
	"log"
	"sync"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer RabbitMQ 生产者
type Producer struct {
	conn         *amqp.Connection  // RabbitMQ 连接
	channel      *amqp.Channel     // 通道（类似数据库连接）
	exchange     string            // 交换机名称
	exchangeType string            // 交换机类型
	url          string            // 连接地址
	mu           sync.RWMutex      // 读写锁（并发安全）
	closed       bool              // 是否已关闭
	reconnecting bool              // 是否正在重连
}

// NewProducer 创建生产者
func NewProducer(host string, port int, username, password, exchange, exchangeType string) (*Producer, error) {
	// 构建连接 URL: amqp://guest:guest@localhost:5672/
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

	producer := &Producer{
		exchange:     exchange,
		exchangeType: exchangeType,
		url:          url,
		closed:       false,
		reconnecting: false,
	}

	// 建立连接
	if err := producer.connect(); err != nil {
		return nil, err
	}

	// 启动连接监控（后台运行）
	// go 关键字：在新的 goroutine 中运行，不阻塞主流程
	go producer.monitorConnection()

	return producer, nil
}

// connect 建立连接
func (p *Producer) connect() error {
	// 加锁：防止多个 goroutine 同时连接
	p.mu.Lock()
	defer p.mu.Unlock()  // defer: 函数返回前执行

	var err error

	// 1. 连接到 RabbitMQ
	p.conn, err = amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 2. 创建通道
	// 为什么需要 Channel？一个连接可以有多个通道，提高并发
	p.channel, err = p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 3. 声明交换机
	err = p.channel.ExchangeDeclare(
		p.exchange,     // name: 交换机名称
		p.exchangeType, // type: fanout（广播）
		true,           // durable: 持久化（重启不丢失）
		false,          // auto-deleted: 不自动删除
		false,          // internal: 不是内部交换机
		false,          // no-wait: 等待服务器确认
		nil,            // arguments: 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("RabbitMQ Producer connected successfully")
	return nil
}

// monitorConnection 监控连接状态
// 作用：连接断开时自动重连
func (p *Producer) monitorConnection() {
	for {
		if p.closed {
			return
		}

		p.mu.RLock()
		conn := p.conn
		p.mu.RUnlock()

		if conn == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// 监听连接关闭事件
		notifyClose := make(chan *amqp.Error)
		conn.NotifyClose(notifyClose)

		// 阻塞等待连接关闭
		err := <-notifyClose
		if err != nil && !p.closed {
			log.Printf("RabbitMQ connection closed: %v, attempting to reconnect...", err)
			p.reconnect()
		}
	}
}

// reconnect 重新连接
// 使用指数退避：1秒 -> 2秒 -> 4秒 -> ... -> 最多30秒
func (p *Producer) reconnect() {
	p.mu.Lock()
	if p.reconnecting {
		p.mu.Unlock()
		return
	}
	p.reconnecting = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.reconnecting = false
		p.mu.Unlock()
	}()

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for !p.closed {
		log.Printf("Attempting to reconnect to RabbitMQ...")
		if err := p.connect(); err != nil {
			log.Printf("Failed to reconnect: %v, retrying in %v", err, backoff)
			time.Sleep(backoff)
			backoff *= 2  // 等待时间翻倍
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		} else {
			log.Printf("Successfully reconnected to RabbitMQ")
			return
		}
	}
}

// Publish 发布消息（核心方法）
func (p *Producer) Publish(message interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	// 类型断言：确保是 ErrorLogMessage 类型
	var msg *ErrorLogMessage
	switch v := message.(type) {
	case *ErrorLogMessage:
		msg = v
	default:
		return fmt.Errorf("unsupported message type")
	}

	// 转换为 JSON
	body, err := msg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	err = p.channel.Publish(
		p.exchange, // exchange
		"",         // routing key（fanout 不需要）
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,  // 持久化
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	log.Printf("RabbitMQ Producer closed")
	return nil
}
```

**关键概念**：
- **sync.RWMutex**：读写锁，多个 goroutine 可以同时读，写时独占
- **defer**：延迟执行，常用于释放资源
- **goroutine**：Go 的轻量级线程，用 `go` 启动
- **channel**：Go 的通信机制，用于 goroutine 间传递数据

---

### 3. ES 消费者 (mq/consumer_es.go)

```go
package mq

import (
	"fmt"
	"log"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ESWriter ES 写入接口
// 为什么用接口？解耦，方便测试和替换实现
type ESWriter interface {
	AddLog(doc map[string]interface{}) error
}

// ConsumerES Elasticsearch 消费者
type ConsumerES struct {
	conn      *amqp.Connection  // RabbitMQ 连接
	channel   *amqp.Channel     // 通道
	queue     string            // 队列名称
	esWriter  ESWriter          // ES 写入器
	url       string            // 连接地址
	closed    bool              // 是否已关闭
	closeChan chan struct{}     // 关闭信号通道
}

// NewConsumerES 创建 ES 消费者
func NewConsumerES(host string, port int, username, password, exchange, exchangeType, queue string, esWriter ESWriter) (*ConsumerES, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

	consumer := &ConsumerES{
		queue:     queue,
		esWriter:  esWriter,
		url:       url,
		closed:    false,
		closeChan: make(chan struct{}),
	}

	if err := consumer.connect(exchange, exchangeType); err != nil {
		return nil, err
	}

	return consumer, nil
}

// connect 建立连接
func (c *ConsumerES) connect(exchange, exchangeType string) error {
	var err error

	// 1. 连接 RabbitMQ
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 2. 创建通道
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 3. 声明交换机（与生产者保持一致）
	err = c.channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 4. 声明队列
	_, err = c.channel.QueueDeclare(
		c.queue, // name
		true,    // durable: 持久化
		false,   // delete when unused
		false,   // exclusive: 不独占
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 5. 绑定队列到交换机
	err = c.channel.QueueBind(
		c.queue,  // queue name
		"",       // routing key（fanout 不需要）
		exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("ES Consumer connected and bound to queue: %s", c.queue)
	return nil
}

// Start 开始消费
func (c *ConsumerES) Start() error {
	// 注册消费者
	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer: 消费者标签（空表示自动生成）
		false,   // auto-ack: 手动确认（重要！）
		false,   // exclusive: 不独占
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("ES Consumer started, waiting for messages...")

	// 启动消息处理 goroutine
	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				// ok 为 false 表示通道已关闭
				if !ok {
					log.Printf("ES Consumer channel closed")
					return
				}
				c.handleMessage(msg)
			case <-c.closeChan:
				return
			}
		}
	}()

	return nil
}

// handleMessage 处理消息
func (c *ConsumerES) handleMessage(msg amqp.Delivery) {
	// 1. 解析消息
	errorLog, err := FromJSON(msg.Body)
	if err != nil {
		log.Printf("Failed to parse message: %v", err)
		msg.Nack(false, false)  // 拒绝消息，不重新入队
		return
	}

	// 2. 转换为 ES 文档格式
	doc := map[string]interface{}{
		"timestamp":   errorLog.Timestamp,
		"level":       errorLog.Level,
		"service":     errorLog.Service,
		"trace_id":    errorLog.TraceID,
		"user_id":     errorLog.UserID,
		"method":      errorLog.Method,
		"path":        errorLog.Path,
		"error_msg":   errorLog.ErrorMsg,
		"stack_trace": errorLog.StackTrace,
		"extra":       errorLog.Extra,
	}

	// 3. 写入 ES
	if err := c.esWriter.AddLog(doc); err != nil {
		log.Printf("Failed to add log to ES: %v", err)
		// 重试逻辑
		if c.retryWithBackoff(msg, 3) {
			msg.Ack(false)  // 确认消息
		} else {
			msg.Nack(false, true)  // 重新入队
		}
		return
	}

	// 4. 确认消息
	// 为什么要确认？告诉 RabbitMQ 消息已处理，可以删除
	msg.Ack(false)
}

// retryWithBackoff 重试机制
// maxRetries: 最大重试次数
func (c *ConsumerES) retryWithBackoff(msg amqp.Delivery, maxRetries int) bool {
	backoff := time.Second
	for i := 0; i < maxRetries; i++ {
		time.Sleep(backoff)

		errorLog, err := FromJSON(msg.Body)
		if err != nil {
			return false
		}

		doc := map[string]interface{}{
			"timestamp":   errorLog.Timestamp,
			"level":       errorLog.Level,
			"service":     errorLog.Service,
			"trace_id":    errorLog.TraceID,
			"user_id":     errorLog.UserID,
			"method":      errorLog.Method,
			"path":        errorLog.Path,
			"error_msg":   errorLog.ErrorMsg,
			"stack_trace": errorLog.StackTrace,
			"extra":       errorLog.Extra,
		}

		if err := c.esWriter.AddLog(doc); err == nil {
			log.Printf("Retry %d succeeded", i+1)
			return true
		}

		backoff *= 2  // 指数退避
	}
	log.Printf("All retries failed")
	return false
}

// Close 关闭消费者
func (c *ConsumerES) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.closeChan)

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}

	log.Printf("ES Consumer closed")
	return nil
}
```

**关键点**：
- **auto-ack: false**：手动确认，确保消息不丢失
- **msg.Ack()**：确认消息已处理
- **msg.Nack()**：拒绝消息，可选择是否重新入队
- **select**：Go 的多路复用，同时监听多个 channel

---

### 4. 文件消费者 (mq/consumer_file.go)

```go
package mq

import (
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ConsumerFile 文件消费者
type ConsumerFile struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     string
	logger    *lumberjack.Logger  // 日志轮转器
	url       string
	closed    bool
	closeChan chan struct{}
}

// LogConfig 日志配置
type LogConfig struct {
	FilePath   string  // 日志文件路径
	MaxSize    int     // 单个文件最大大小（MB）
	MaxBackups int     // 保留的旧文件数量
	MaxAge     int     // 保留天数
}

// NewConsumerFile 创建文件消费者
func NewConsumerFile(host string, port int, username, password, exchange, exchangeType, queue string, logConfig LogConfig) (*ConsumerFile, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

	// 创建日志轮转器
	// lumberjack 会自动管理日志文件的轮转
	logger := &lumberjack.Logger{
		Filename:   logConfig.FilePath,    // 日志文件路径
		MaxSize:    logConfig.MaxSize,     // 100MB
		MaxBackups: logConfig.MaxBackups,  // 保留 10 个旧文件
		MaxAge:     logConfig.MaxAge,      // 保留 30 天
		Compress:   true,                  // 压缩旧文件
	}

	consumer := &ConsumerFile{
		queue:     queue,
		logger:    logger,
		url:       url,
		closed:    false,
		closeChan: make(chan struct{}),
	}

	if err := consumer.connect(exchange, exchangeType); err != nil {
		return nil, err
	}

	return consumer, nil
}

// connect 建立连接（与 ES 消费者类似）
func (c *ConsumerFile) connect(exchange, exchangeType string) error {
	var err error
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = c.channel.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	_, err = c.channel.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.channel.QueueBind(
		c.queue,
		"",
		exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("File Consumer connected and bound to queue: %s", c.queue)
	return nil
}

// Start 开始消费
func (c *ConsumerFile) Start() error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,  // 手动确认
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("File Consumer started, waiting for messages...")

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("File Consumer channel closed")
					return
				}
				c.handleMessage(msg)
			case <-c.closeChan:
				return
			}
		}
	}()

	return nil
}

// handleMessage 处理消息
func (c *ConsumerFile) handleMessage(msg amqp.Delivery) {
	// 1. 解析消息
	errorLog, err := FromJSON(msg.Body)
	if err != nil {
		log.Printf("Failed to parse message: %v", err)
		msg.Nack(false, false)
		return
	}

	// 2. 格式化日志
	// 格式：[时间] [级别] [服务] TraceID=xxx UserID=xxx GET /path - 错误信息
	logLine := fmt.Sprintf("[%s] [%s] [%s] TraceID=%s UserID=%s %s %s - %s\n",
		errorLog.Timestamp.Format("2006-01-02 15:04:05"),
		errorLog.Level,
		errorLog.Service,
		errorLog.TraceID,
		errorLog.UserID,
		errorLog.Method,
		errorLog.Path,
		errorLog.ErrorMsg,
	)

	// 如果有堆栈跟踪，追加到日志
	if errorLog.StackTrace != "" {
		logLine += fmt.Sprintf("StackTrace:\n%s\n", errorLog.StackTrace)
	}

	// 3. 写入文件
	// lumberjack 会自动处理文件轮转
	if _, err := c.logger.Write([]byte(logLine)); err != nil {
		log.Printf("Failed to write log to file: %v", err)
		msg.Nack(false, true)  // 重新入队
		return
	}

	// 4. 确认消息
	msg.Ack(false)
}

// Close 关闭消费者
func (c *ConsumerFile) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.closeChan)

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	if c.logger != nil {
		c.logger.Close()
	}

	log.Printf("File Consumer closed")
	return nil
}
```

**关键点**：
- **lumberjack**：自动日志轮转库
- **日志格式**：人类可读的格式，方便直接查看
- **轮转策略**：按大小、数量、时间三个维度管理

---

### 5. 错误日志记录器 (logger/error_logger.go)

```go
package logger

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

// ErrorLogMessage 错误日志消息结构
// 为什么重复定义？避免循环依赖（logger 不依赖 mq）
type ErrorLogMessage struct {
	Timestamp  time.Time
	Level      string
	Service    string
	TraceID    string
	UserID     string
	Method     string
	Path       string
	ErrorMsg   string
	StackTrace string
	Extra      map[string]interface{}
}

// MQPublisher MQ 发布接口
// 为什么用接口？解耦，logger 不需要知道 MQ 的具体实现
type MQPublisher interface {
	Publish(msg interface{}) error
}

// 全局生产者
var globalProducer MQPublisher

// InitErrorLogger 初始化错误日志记录器
// 在应用启动时调用，设置全局生产者
func InitErrorLogger(producer MQPublisher) {
	globalProducer = producer
}

// Error 记录错误日志
// 参数：
//   - ctx: 上下文，包含 trace_id、user_id 等信息
//   - err: 错误对象
//   - message: 错误描述
//   - extra: 额外信息（如文件名、大小等）
func Error(ctx context.Context, err error, message string, extra map[string]interface{}) {
	if globalProducer == nil {
		fmt.Printf("Error logger not initialized: %v\n", err)
		return
	}

	// 从上下文中提取信息
	traceID := getTraceID(ctx)
	userID := getUserID(ctx)
	method := getMethod(ctx)
	path := getPath(ctx)

	// 构建错误日志消息
	msg := &ErrorLogMessage{
		Timestamp:  time.Now(),
		Level:      "ERROR",
		Service:    "core-api",
		TraceID:    traceID,
		UserID:     userID,
		Method:     method,
		Path:       path,
		ErrorMsg:   fmt.Sprintf("%s: %v", message, err),
		StackTrace: string(debug.Stack()),  // 获取堆栈跟踪
		Extra:      extra,
	}

	// 异步发送到 MQ
	if err := globalProducer.Publish(msg); err != nil {
		fmt.Printf("Failed to publish error log: %v\n", err)
	}
}

// Fatal 记录致命错误日志
func Fatal(ctx context.Context, err error, message string, extra map[string]interface{}) {
	if globalProducer == nil {
		fmt.Printf("Error logger not initialized: %v\n", err)
		return
	}

	traceID := getTraceID(ctx)
	userID := getUserID(ctx)
	method := getMethod(ctx)
	path := getPath(ctx)

	msg := &ErrorLogMessage{
		Timestamp:  time.Now(),
		Level:      "FATAL",
		Service:    "core-api",
		TraceID:    traceID,
		UserID:     userID,
		Method:     method,
		Path:       path,
		ErrorMsg:   fmt.Sprintf("%s: %v", message, err),
		StackTrace: string(debug.Stack()),
		Extra:      extra,
	}

	if err := globalProducer.Publish(msg); err != nil {
		fmt.Printf("Failed to publish fatal log: %v\n", err)
	}
}

// Panic 记录 panic 日志
func Panic(ctx context.Context, panicValue interface{}, extra map[string]interface{}) {
	if globalProducer == nil {
		fmt.Printf("Error logger not initialized: panic=%v\n", panicValue)
		return
	}

	traceID := getTraceID(ctx)
	userID := getUserID(ctx)
	method := getMethod(ctx)
	path := getPath(ctx)

	msg := &ErrorLogMessage{
		Timestamp:  time.Now(),
		Level:      "PANIC",
		Service:    "core-api",
		TraceID:    traceID,
		UserID:     userID,
		Method:     method,
		Path:       path,
		ErrorMsg:   fmt.Sprintf("Panic: %v", panicValue),
		StackTrace: string(debug.Stack()),
		Extra:      extra,
	}

	if err := globalProducer.Publish(msg); err != nil {
		fmt.Printf("Failed to publish panic log: %v\n", err)
	}
}

// 辅助函数：从上下文中提取信息
// 为什么用 context？Go 的标准做法，用于传递请求级别的数据

func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// 类型断言：将 interface{} 转换为 string
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

func getUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getMethod(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if method, ok := ctx.Value("method").(string); ok {
		return method
	}
	return ""
}

func getPath(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if path, ok := ctx.Value("path").(string); ok {
		return path
	}
	return ""
}
```

**关键点**：
- **context.Context**：Go 的标准做法，传递请求级别的数据
- **debug.Stack()**：获取当前的堆栈跟踪
- **类型断言**：`value.(string)` 将 interface{} 转换为具体类型
- **全局变量**：`globalProducer` 在整个应用中共享

---

### 6. Elasticsearch 客户端 (logger/es_client.go)

```go
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ESClient Elasticsearch 客户端
type ESClient struct {
	client     *elasticsearch.Client
	index      string                      // 索引名称
	batchSize  int                         // 批量大小
	flushTime  time.Duration               // 刷新时间
	buffer     []*map[string]interface{}   // 缓冲区
	mu         sync.Mutex                  // 互斥锁
	stopCh     chan struct{}               // 停止信号
	wg         sync.WaitGroup              // 等待组
}

// NewESClient 创建 ES 客户端
func NewESClient(addresses []string, username, password, index string) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ES: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES returned error: %s", res.String())
	}

	esClient := &ESClient{
		client:    client,
		index:     index,
		batchSize: 100,              // 攒够 100 条再写入
		flushTime: 5 * time.Second,  // 或者每 5 秒写入一次
		buffer:    make([]*map[string]interface{}, 0, 100),
		stopCh:    make(chan struct{}),
	}

	// 启动定时刷新
	esClient.wg.Add(1)
	go esClient.autoFlush()

	log.Printf("Elasticsearch client connected successfully")
	return esClient, nil
}

// AddLog 添加日志到缓冲区
// 为什么用缓冲区？批量写入比单条写入快很多
func (c *ESClient) AddLog(doc map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer = append(c.buffer, &doc)

	// 如果达到批量大小，立即刷新
	if len(c.buffer) >= c.batchSize {
		go c.flush()
	}

	return nil
}

// autoFlush 自动刷新
// 作用：即使没攒够 100 条，也要定期写入
func (c *ESClient) autoFlush() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.flush()
		case <-c.stopCh:
			c.flush()  // 最后刷新一次
			return
		}
	}
}

// flush 批量写入 ES
func (c *ESClient) flush() {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return
	}

	// 复制缓冲区并清空
	docs := make([]*map[string]interface{}, len(c.buffer))
	copy(docs, c.buffer)
	c.buffer = c.buffer[:0]
	c.mu.Unlock()

	// 批量写入
	if err := c.bulkIndex(docs); err != nil {
		log.Printf("Failed to bulk index to ES: %v", err)
	}
}

// bulkIndex 批量索引
func (c *ESClient) bulkIndex(docs []*map[string]interface{}) error {
	if len(docs) == 0 {
		return nil
	}

	// 构建批量请求
	// ES 的批量格式：
	// {"index":{"_index":"error_logs"}}
	// {"timestamp":"2026-03-15",...}
	// {"index":{"_index":"error_logs"}}
	// {"timestamp":"2026-03-15",...}
	var buf bytes.Buffer
	for _, doc := range docs {
		// 索引元数据
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": c.index,
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			log.Printf("Failed to encode meta: %v", err)
			continue
		}

		// 文档内容
		if err := json.NewEncoder(&buf).Encode(*doc); err != nil {
			log.Printf("Failed to encode document: %v", err)
			continue
		}
	}

	// 执行批量请求
	req := esapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request returned error: %s", res.String())
	}

	log.Printf("Successfully indexed %d documents to ES", len(docs))
	return nil
}

// Close 关闭客户端
func (c *ESClient) Close() error {
	close(c.stopCh)
	c.wg.Wait()  // 等待 autoFlush goroutine 结束
	log.Printf("Elasticsearch client closed")
	return nil
}
```

**关键点**：
- **批量写入**：攒够 100 条或 5 秒后写入，提高性能
- **sync.WaitGroup**：等待 goroutine 结束
- **ticker**：定时器，每隔一段时间触发一次
- **ES Bulk API**：批量操作的标准格式

---

### 7. 错误恢复中间件 (middleware/error-recovery-middleware.go)

```go
package middleware

import (
	"cloud_disk/core/internal/logger"
	"context"
	"fmt"
	"net/http"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ErrorRecoveryMiddleware 全局错误恢复中间件
// 作用：捕获所有 panic，防止程序崩溃
func ErrorRecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// defer + recover 是 Go 捕获 panic 的标准做法
		defer func() {
			// recover() 捕获 panic
			if err := recover(); err != nil {
				// 构建上下文
				ctx := context.WithValue(r.Context(), "method", r.Method)
				ctx = context.WithValue(ctx, "path", r.URL.Path)
				ctx = context.WithValue(ctx, "trace_id", r.Header.Get("X-Trace-ID"))

				// 记录 panic 日志
				logger.Panic(ctx, err, map[string]interface{}{
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.UserAgent(),
					"headers":     r.Header,
				})

				// 返回友好的错误响应
				httpx.Error(w, fmt.Errorf("internal server error"))
			}
		}()

		// 执行下一个处理器
		next(w, r)
	}
}
```

**关键点**：
- **defer + recover**：Go 捕获 panic 的标准做法
- **中间件模式**：在请求处理前后执行额外逻辑
- **context.WithValue**：向上下文中添加键值对

**为什么需要这个中间件？**
- 防止 panic 导致整个程序崩溃
- 自动记录 panic 日志，方便排查问题
- 返回友好的错误信息给用户

---

### 8. 配置文件 (core/etc/core-api.yaml)

```yaml
# 新增的 RabbitMQ 配置
RabbitMQ:
  Host: localhost          # RabbitMQ 服务器地址
  Port: 5672               # RabbitMQ 端口
  Username: guest          # 用户名
  Password: guest          # 密码
  Exchange: error_log_exchange      # 交换机名称
  ExchangeType: fanout     # 交换机类型（广播）
  QueueES: error_log_queue_es       # ES 队列名称
  QueueFile: error_log_queue_file   # 文件队列名称

# 新增的 Elasticsearch 配置
Elasticsearch:
  Addresses:
    - http://localhost:9200  # ES 服务器地址（可以配置多个）
  Username: elastic          # ES 用户名
  Password: ""               # ES 密码（空表示不需要认证）
  Index: error_logs          # 索引名称

# 新增的错误日志配置
ErrorLog:
  FilePath: ./logs/error.log  # 日志文件路径
  MaxSize: 100                # 单个文件最大 100MB
  MaxBackups: 10              # 保留 10 个旧文件
  MaxAge: 30                  # 保留 30 天
```

**配置说明**：
- **RabbitMQ**：消息队列配置
- **Elasticsearch**：搜索引擎配置（可选）
- **ErrorLog**：本地文件日志配置

---

### 9. 配置结构体 (core/internal/config/config.go)

```go
// 新增的配置结构体
type Config struct {
	// ... 原有配置 ...

	// RabbitMQ 配置
	RabbitMQ struct {
		Host         string
		Port         int
		Username     string
		Password     string
		Exchange     string
		ExchangeType string
		QueueES      string
		QueueFile    string
	}

	// Elasticsearch 配置
	Elasticsearch struct {
		Addresses []string  // 支持多个地址
		Username  string
		Password  string
		Index     string
	}

	// 错误日志配置
	ErrorLog struct {
		FilePath   string
		MaxSize    int
		MaxBackups int
		MaxAge     int
	}
}
```

**关键点**：
- 结构体字段名首字母大写（公开）
- 字段名与 YAML 配置文件对应

---

### 10. 服务上下文 (core/internal/svc/service-context.go)

```go
type ServiceContext struct {
	// ... 原有字段 ...

	// 新增字段
	MQProducer     *mq.Producer        // MQ 生产者
	ESConsumer     *mq.ConsumerES      // ES 消费者
	FileConsumer   *mq.ConsumerFile    // 文件消费者
	ErrorRecovery  rest.Middleware     // 错误恢复中间件
}

func NewServiceContext(c config.Config) *ServiceContext {
	// ... 原有代码 ...

	// 1. 初始化 RabbitMQ 生产者
	producer, err := mq.NewProducer(
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Exchange,
		c.RabbitMQ.ExchangeType,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize MQ Producer: %v", err)
	}

	// 2. 初始化错误日志记录器
	logger.InitErrorLogger(producer)

	// 3. 初始化 Elasticsearch 客户端
	esClient, err := logger.NewESClient(
		c.Elasticsearch.Addresses,
		c.Elasticsearch.Username,
		c.Elasticsearch.Password,
		c.Elasticsearch.Index,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize ES Client: %v", err)
	}

	// 4. 初始化 ES 消费者
	var esConsumer *mq.ConsumerES
	if esClient != nil {
		esConsumer, err = mq.NewConsumerES(
			c.RabbitMQ.Host,
			c.RabbitMQ.Port,
			c.RabbitMQ.Username,
			c.RabbitMQ.Password,
			c.RabbitMQ.Exchange,
			c.RabbitMQ.ExchangeType,
			c.RabbitMQ.QueueES,
			esClient,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize ES Consumer: %v", err)
		} else {
			// 启动消费者
			if err := esConsumer.Start(); err != nil {
				log.Printf("Warning: Failed to start ES Consumer: %v", err)
			}
		}
	}

	// 5. 初始化文件消费者
	fileConsumer, err := mq.NewConsumerFile(
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Exchange,
		c.RabbitMQ.ExchangeType,
		c.RabbitMQ.QueueFile,
		mq.LogConfig{
			FilePath:   c.ErrorLog.FilePath,
			MaxSize:    c.ErrorLog.MaxSize,
			MaxBackups: c.ErrorLog.MaxBackups,
			MaxAge:     c.ErrorLog.MaxAge,
		},
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize File Consumer: %v", err)
	} else {
		// 启动消费者
		if err := fileConsumer.Start(); err != nil {
			log.Printf("Warning: Failed to start File Consumer: %v", err)
		}
	}

	return &ServiceContext{
		// ... 原有字段 ...
		MQProducer:    producer,
		ESConsumer:    esConsumer,
		FileConsumer:  fileConsumer,
		ErrorRecovery: middleware.ErrorRecoveryMiddleware,
	}
}
```

**初始化顺序**：
1. 创建 MQ 生产者
2. 初始化日志记录器（设置全局生产者）
3. 创建 ES 客户端
4. 创建 ES 消费者并启动
5. 创建文件消费者并启动

**为什么用 Warning？**
- 日志系统失败不应该阻止应用启动
- 只记录警告，应用继续运行

---

### 11. 路由注册 (core/internal/handler/routes.go)

```go
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 注册全局错误恢复中间件
	// 这个中间件会应用到所有路由
	server.Use(serverCtx.ErrorRecovery)

	// ... 原有路由注册 ...
}
```

**关键点**：
- `server.Use()` 注册全局中间件
- 中间件会在所有路由处理前执行

---

## 工作流程详解

### 完整流程示例

假设用户上传文件失败，看看日志是如何记录的：

```go
// 1. 业务代码中发生错误
func (l *FileUploadLogic) FileUpload(req *types.FileUploadRequest) error {
	file, err := l.uploadToOSS(req.File)
	if err != nil {
		// 2. 调用 logger.Error 记录错误
		logger.Error(l.ctx, err, "文件上传到OSS失败", map[string]interface{}{
			"file_name": req.FileName,
			"file_size": req.FileSize,
			"user_id":   l.userId,
		})
		return err
	}
	return nil
}
```

**流程分解**：

1. **logger.Error() 被调用**
   - 从 context 中提取 trace_id、user_id 等信息
   - 获取堆栈跟踪
   - 构建 ErrorLogMessage 对象

2. **发送到 MQ Producer**
   - 将 ErrorLogMessage 转换为 JSON
   - 调用 producer.Publish() 发送消息
   - 消息发送到 RabbitMQ Exchange

3. **Exchange 广播消息**
   - fanout 模式：将消息复制到所有绑定的队列
   - 消息同时进入 ES Queue 和 File Queue

4. **ES Consumer 处理**
   - 从 ES Queue 中取出消息
   - 解析 JSON 为 ErrorLogMessage
   - 转换为 ES 文档格式
   - 添加到 ES 客户端的缓冲区
   - 攒够 100 条或 5 秒后批量写入 ES
   - 确认消息（Ack）

5. **File Consumer 处理**
   - 从 File Queue 中取出消息
   - 解析 JSON 为 ErrorLogMessage
   - 格式化为人类可读的日志行
   - 写入本地文件（lumberjack 自动轮转）
   - 确认消息（Ack）

6. **查询日志**
   - 本地文件：`tail -f logs/error.log`
   - Elasticsearch：`curl http://localhost:9200/error_logs/_search`

---

## 使用示例

### 示例 1：记录普通错误

```go
package logic

import (
	"cloud_disk/core/internal/logger"
	"context"
	"fmt"
)

func (l *FileUploadLogic) FileUpload(req *types.FileUploadRequest) error {
	// 业务逻辑
	err := l.doSomething()
	if err != nil {
		// 记录错误日志
		logger.Error(l.ctx, err, "操作失败", map[string]interface{}{
			"operation": "file_upload",
			"file_name": req.FileName,
		})
		return err
	}
	return nil
}
```

### 示例 2：记录致命错误

```go
func (l *DatabaseLogic) Connect() error {
	err := l.db.Connect()
	if err != nil {
		// 记录致命错误
		logger.Fatal(l.ctx, err, "数据库连接失败", map[string]interface{}{
			"db_host": "localhost",
			"db_port": 3306,
		})
		return err
	}
	return nil
}
```

### 示例 3：Panic 自动捕获

```go
// 不需要手动调用，中间件会自动捕获
func (l *SomeLogic) DoSomething() {
	// 如果这里发生 panic
	panic("something went wrong")
	// ErrorRecoveryMiddleware 会自动捕获并记录
}
```

---

## 常见问题解答

### Q1：为什么要用消息队列？

**A**：
- **异步处理**：不阻塞业务代码
- **削峰填谷**：高并发时消息暂存在队列中
- **解耦**：日志处理和业务逻辑分离
- **可靠性**：消息持久化，不会丢失

### Q2：为什么要双存储（ES + 文件）？

**A**：
- **ES**：强大的查询能力，适合分析和监控
- **文件**：简单可靠，作为备份
- **互补**：ES 挂了还有文件，文件满了还有 ES

### Q3：消息会丢失吗？

**A**：不会，因为：
- Exchange 和 Queue 都是持久化的
- 消息设置为 Persistent
- 消费者使用手动确认（Ack）
- 只有成功写入后才确认

### Q4：如何保证消息不重复？

**A**：
- 本系统允许少量重复（幂等性）
- 如果需要严格去重，可以：
  - 使用 trace_id + timestamp 作为唯一键
  - 在 ES 中使用文档 ID 去重

### Q5：性能如何？

**A**：
- 发送消息：微秒级（异步）
- ES 批量写入：100 条/次，高效
- 文件写入：lumberjack 自动缓冲

### Q6：如何监控系统？

**A**：
- RabbitMQ 管理界面：http://localhost:15672
- 查看队列深度、消费速率
- 查看 ES 索引大小
- 查看本地日志文件大小

---

## 学习建议

### 新手学习路径

1. **理解概念**（1-2 天）
   - RabbitMQ 基础：Producer、Exchange、Queue、Consumer
   - Elasticsearch 基础：Index、Document、Query
   - Go 并发：goroutine、channel、mutex

2. **阅读代码**（2-3 天）
   - 从 message.go 开始，理解数据结构
   - 阅读 producer.go，理解如何发送消息
   - 阅读 consumer_*.go，理解如何处理消息
   - 阅读 logger/*.go，理解如何使用

3. **实践操作**（3-5 天）
   - 启动 RabbitMQ 和 ES
   - 运行项目，触发错误
   - 查看日志文件和 ES 中的数据
   - 修改配置，观察效果

4. **深入理解**（持续）
   - 研究 RabbitMQ 的其他模式（direct、topic）
   - 学习 ES 的查询语法
   - 优化性能参数

### 推荐资源

- **RabbitMQ 官方教程**：https://www.rabbitmq.com/getstarted.html
- **Elasticsearch 官方文档**：https://www.elastic.co/guide/
- **Go 并发编程**：《Go 并发编程实战》

---

## 总结

这个错误日志系统的核心思想是：

1. **异步化**：业务代码不等待日志写入
2. **解耦**：日志处理独立于业务逻辑
3. **可靠性**：消息持久化 + 手动确认
4. **高性能**：批量写入 + 缓冲
5. **易用性**：简单的 API（logger.Error）

通过这个系统，你可以：
- 快速定位问题（堆栈跟踪）
- 分析错误趋势（ES 查询）
- 监控系统健康（RabbitMQ 管理界面）

希望这份文档能帮助你理解和掌握这个系统！

