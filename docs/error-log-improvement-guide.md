# 错误日志系统 - 后续改进建议

## 一、改进路线图

```
简化版 (当前)
    ↓
第一阶段：日志轮转
    ↓
第二阶段：异步写入
    ↓
第三阶段：消息队列
    ↓
第四阶段：Elasticsearch
    ↓
第五阶段：监控告警
```

---

## 二、第一阶段：日志轮转

### 问题
简化版的日志文件会无限增长，最终占满磁盘。

### 解决方案
使用 `lumberjack` 库实现日志轮转。

### 实现步骤

**1. 安装依赖**
```bash
go get gopkg.in/natefinch/lumberjack.v2
```

**2. 修改 simple_logger.go**
```go
import (
    "gopkg.in/natefinch/lumberjack.v2"
)

func InitSimpleLogger(logFilePath string) error {
    // 使用 lumberjack 替代 os.File
    logWriter := &lumberjack.Logger{
        Filename:   logFilePath,
        MaxSize:    100,  // MB
        MaxBackups: 10,   // 保留旧文件数量
        MaxAge:     30,   // 天
        Compress:   true, // 压缩旧文件
    }

    globalLogger = &SimpleLogger{
        logFile: logWriter,
    }

    return nil
}
```

### 效果
- 单个日志文件最大 100MB
- 自动轮转，保留最近 10 个文件
- 超过 30 天的日志自动删除
- 旧日志自动压缩，节省空间

### 学习重点
- 日志轮转的概念
- lumberjack 库的使用
- 文件大小和时间管理

---

## 三、第二阶段：异步写入

### 问题
简化版是同步写入，每次记录日志都会阻塞业务代码，影响性能。

### 解决方案
使用 channel 实现异步写入。

### 实现步骤

**1. 添加 channel 缓冲区**
```go
type SimpleLogger struct {
    logFile  *os.File
    logChan  chan *ErrorLog  // 日志通道
    stopChan chan struct{}   // 停止信号
    wg       sync.WaitGroup  // 等待组
}

func InitSimpleLogger(logFilePath string) error {
    // ... 打开文件 ...

    globalLogger = &SimpleLogger{
        logFile:  file,
        logChan:  make(chan *ErrorLog, 1000), // 缓冲 1000 条
        stopChan: make(chan struct{}),
    }

    // 启动后台写入协程
    globalLogger.wg.Add(1)
    go globalLogger.writeLoop()

    return nil
}
```

**2. 实现后台写入**
```go
func (l *SimpleLogger) writeLoop() {
    defer l.wg.Done()

    for {
        select {
        case log := <-l.logChan:
            l.writeToFile(log)
        case <-l.stopChan:
            // 处理剩余日志
            for len(l.logChan) > 0 {
                log := <-l.logChan
                l.writeToFile(log)
            }
            return
        }
    }
}
```

**3. 修改 LogError 等方法**
```go
func LogError(ctx context.Context, message string, err error, extra map[string]interface{}) {
    errorLog := &ErrorLog{
        // ... 构建日志 ...
    }

    // 异步发送到 channel
    select {
    case globalLogger.logChan <- errorLog:
        // 发送成功
    default:
        // channel 满了，直接输出到控制台
        log.Printf("日志 channel 已满，丢弃日志: %s", message)
    }
}
```

**4. 优雅关闭**
```go
func Close() error {
    if globalLogger != nil {
        close(globalLogger.stopChan)
        globalLogger.wg.Wait()
        return globalLogger.logFile.Close()
    }
    return nil
}
```

### 效果
- 日志记录不阻塞业务代码
- 性能提升 10-100 倍
- 支持高并发场景

### 学习重点
- Go channel 的使用
- 异步编程模式
- 优雅关闭（graceful shutdown）
- sync.WaitGroup 的使用

---

## 四、第三阶段：消息队列（RabbitMQ）

### 问题
- 单机日志处理能力有限
- 日志写入失败会丢失
- 无法支持多个消费者（文件、ES、告警等）

### 解决方案
引入 RabbitMQ 作为消息队列，解耦日志生产和消费。

### 架构设计

```
业务代码
    ↓
Logger.LogError()
    ↓
RabbitMQ Producer → Exchange (Fanout)
                        ↓
            ┌───────────┴───────────┐
            ↓                       ↓
       File Queue              ES Queue
            ↓                       ↓
       File Consumer          ES Consumer
            ↓                       ↓
       logs/error.log        Elasticsearch
```

### 实现步骤

**1. 安装 RabbitMQ**
```bash
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

**2. 安装依赖**
```bash
go get github.com/rabbitmq/amqp091-go
```

**3. 创建 Producer**
```go
type Producer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func NewProducer(url string) (*Producer, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }

    channel, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    return &Producer{conn: conn, channel: channel}, nil
}

func (p *Producer) Publish(message interface{}) error {
    body, _ := json.Marshal(message)
    return p.channel.Publish(
        "error_log_exchange", // exchange
        "",                   // routing key
        false,                // mandatory
        false,                // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}
```

**4. 创建 Consumer**
```go
type Consumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (c *Consumer) Start() {
    msgs, _ := c.channel.Consume(
        "error_log_queue", // queue
        "",                // consumer
        true,              // auto-ack
        false,             // exclusive
        false,             // no-local
        false,             // no-wait
        nil,               // args
    )

    for msg := range msgs {
        var log ErrorLog
        json.Unmarshal(msg.Body, &log)
        // 处理日志...
    }
}
```

### 效果
- 日志不会丢失（持久化到 RabbitMQ）
- 支持多个消费者
- 生产者和消费者解耦
- 可以独立扩展

### 学习重点
- RabbitMQ 基本概念（Exchange、Queue、Binding）
- AMQP 协议
- 消息队列的使用场景
- 生产者-消费者模式

---

## 五、第四阶段：Elasticsearch

### 问题
- 日志文件难以搜索
- 无法快速定位问题
- 不支持复杂查询

### 解决方案
将日志写入 Elasticsearch，支持全文搜索和聚合分析。

### 实现步骤

**1. 安装 Elasticsearch**
```bash
docker run -d --name elasticsearch \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  elasticsearch:8.11.0
```

**2. 安装依赖**
```bash
go get github.com/elastic/go-elasticsearch/v8
```

**3. 创建 ES Client**
```go
type ESClient struct {
    client *elasticsearch.Client
    index  string
}

func NewESClient(addresses []string, index string) (*ESClient, error) {
    cfg := elasticsearch.Config{
        Addresses: addresses,
    }

    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        return nil, err
    }

    return &ESClient{client: client, index: index}, nil
}

func (c *ESClient) IndexLog(log *ErrorLog) error {
    body, _ := json.Marshal(log)

    req := esapi.IndexRequest{
        Index: c.index,
        Body:  bytes.NewReader(body),
    }

    res, err := req.Do(context.Background(), c.client)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    return nil
}
```

**4. 批量优化**
```go
type ESClient struct {
    // ... 其他字段 ...
    buffer    []*ErrorLog
    batchSize int
    mu        sync.Mutex
}

func (c *ESClient) AddLog(log *ErrorLog) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.buffer = append(c.buffer, log)

    if len(c.buffer) >= c.batchSize {
        go c.flush()
    }
}

func (c *ESClient) flush() {
    // 批量写入 ES
    // 使用 Bulk API
}
```

### 效果
- 支持全文搜索
- 支持复杂查询（按时间、用户、错误类型等）
- 支持聚合分析（错误统计、趋势分析）
- 可视化展示（配合 Kibana）

### 学习重点
- Elasticsearch 基本概念（Index、Document、Mapping）
- 全文搜索原理
- 批量写入优化
- RESTful API 的使用

---

## 六、第五阶段：监控告警

### 问题
- 错误发生后无法及时发现
- 需要人工查看日志
- 无法预警潜在问题

### 解决方案
添加监控指标和告警机制。

### 实现步骤

**1. 添加指标收集**
```go
type Metrics struct {
    ErrorCount  int64
    FatalCount  int64
    PanicCount  int64
    LastError   time.Time
}

var globalMetrics Metrics

func LogError(...) {
    // ... 记录日志 ...

    // 更新指标
    atomic.AddInt64(&globalMetrics.ErrorCount, 1)
    globalMetrics.LastError = time.Now()
}
```

**2. 暴露 Prometheus 指标**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    errorCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "error_log_total",
            Help: "Total number of error logs",
        },
        []string{"level"},
    )
)

func init() {
    prometheus.MustRegister(errorCounter)
}

func LogError(...) {
    // ... 记录日志 ...

    errorCounter.WithLabelValues("ERROR").Inc()
}
```

**3. 配置告警规则**
```yaml
# prometheus.yml
groups:
  - name: error_log_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(error_log_total{level="ERROR"}[5m]) > 10
        for: 5m
        annotations:
          summary: "错误日志过多"
          description: "最近 5 分钟错误日志超过 10 条/秒"
```

**4. 集成告警通知**
```go
func sendAlert(message string) {
    // 发送到钉钉、企业微信、邮件等
}

func LogFatal(...) {
    // ... 记录日志 ...

    // 立即告警
    go sendAlert(fmt.Sprintf("Fatal 错误: %s", message))
}
```

### 效果
- 实时监控错误日志
- 自动告警通知
- 可视化仪表盘
- 趋势分析和预警

### 学习重点
- Prometheus 监控系统
- 指标收集和暴露
- 告警规则配置
- 通知渠道集成

---

## 七、高级特性

### 7.1 自动重连

**问题**：RabbitMQ 或 ES 连接断开后无法恢复

**解决方案**：
```go
func (p *Producer) monitorConnection() {
    for {
        notifyClose := make(chan *amqp.Error)
        p.conn.NotifyClose(notifyClose)

        err := <-notifyClose
        if err != nil {
            log.Printf("连接断开: %v", err)
            p.reconnect()
        }
    }
}

func (p *Producer) reconnect() {
    backoff := time.Second
    for {
        if err := p.connect(); err == nil {
            log.Printf("重连成功")
            return
        }
        time.Sleep(backoff)
        backoff *= 2
        if backoff > 30*time.Second {
            backoff = 30 * time.Second
        }
    }
}
```

### 7.2 死信队列

**问题**：消息处理失败后丢失

**解决方案**：
```go
// 声明死信队列
channel.QueueDeclare(
    "error_log_dlq", // 死信队列
    true,            // durable
    false,           // auto-delete
    false,           // exclusive
    false,           // no-wait
    nil,             // args
)

// 主队列绑定死信交换机
channel.QueueDeclare(
    "error_log_queue",
    true,
    false,
    false,
    false,
    amqp.Table{
        "x-dead-letter-exchange": "error_log_dlx",
    },
)
```

### 7.3 链路追踪

**问题**：分布式系统中难以追踪请求

**解决方案**：
```go
import "go.opentelemetry.io/otel"

func LogError(ctx context.Context, ...) {
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()

    errorLog := &ErrorLog{
        TraceID: traceID,
        // ... 其他字段 ...
    }
}
```

### 7.4 日志采样

**问题**：高频错误导致日志爆炸

**解决方案**：
```go
type Sampler struct {
    rate     float64
    counters sync.Map
}

func (s *Sampler) ShouldLog(key string) bool {
    count, _ := s.counters.LoadOrStore(key, int64(0))
    newCount := atomic.AddInt64(count.(*int64), 1)

    // 每 N 条记录一次
    return newCount%int64(1/s.rate) == 0
}

func LogError(...) {
    if !sampler.ShouldLog(message) {
        return
    }
    // ... 记录日志 ...
}
```

---

## 八、性能优化

### 8.1 批量写入

**优化前**：每条日志单独写入
```go
func (c *ESClient) IndexLog(log *ErrorLog) error {
    // 单条写入
}
```

**优化后**：批量写入
```go
func (c *ESClient) BulkIndex(logs []*ErrorLog) error {
    // 批量写入，减少网络开销
}
```

**效果**：性能提升 10-50 倍

### 8.2 连接池

**优化前**：每次请求创建新连接
```go
conn, _ := amqp.Dial(url)
defer conn.Close()
```

**优化后**：使用连接池
```go
type ConnectionPool struct {
    conns chan *amqp.Connection
}

func (p *ConnectionPool) Get() *amqp.Connection {
    return <-p.conns
}

func (p *ConnectionPool) Put(conn *amqp.Connection) {
    p.conns <- conn
}
```

### 8.3 压缩

**优化前**：原始 JSON 传输
```go
body, _ := json.Marshal(log)
```

**优化后**：压缩后传输
```go
import "compress/gzip"

var buf bytes.Buffer
gw := gzip.NewWriter(&buf)
json.NewEncoder(gw).Encode(log)
gw.Close()
body := buf.Bytes()
```

**效果**：减少 70% 网络传输

---

## 九、生产环境最佳实践

### 9.1 安全加固

1. **修改默认密码**
```bash
# RabbitMQ
docker exec rabbitmq rabbitmqctl change_password guest new_password

# Elasticsearch
curl -X POST "localhost:9200/_security/user/elastic/_password" \
  -H 'Content-Type: application/json' \
  -d '{"password":"new_password"}'
```

2. **启用 TLS**
```go
cfg := elasticsearch.Config{
    Addresses: []string{"https://localhost:9200"},
    CACert:    caCert,
}
```

3. **限制访问**
```yaml
# docker-compose.yml
services:
  rabbitmq:
    networks:
      - internal
```

### 9.2 高可用部署

1. **RabbitMQ 集群**
```bash
docker run -d --name rabbitmq1 \
  -e RABBITMQ_ERLANG_COOKIE='secret' \
  rabbitmq:3-management

docker run -d --name rabbitmq2 \
  -e RABBITMQ_ERLANG_COOKIE='secret' \
  --link rabbitmq1 \
  rabbitmq:3-management

docker exec rabbitmq2 rabbitmqctl stop_app
docker exec rabbitmq2 rabbitmqctl join_cluster rabbit@rabbitmq1
docker exec rabbitmq2 rabbitmqctl start_app
```

2. **Elasticsearch 集群**
```yaml
# docker-compose.yml
services:
  es01:
    image: elasticsearch:8.11.0
    environment:
      - cluster.name=es-cluster
      - node.name=es01
      - discovery.seed_hosts=es02,es03

  es02:
    image: elasticsearch:8.11.0
    environment:
      - cluster.name=es-cluster
      - node.name=es02
      - discovery.seed_hosts=es01,es03

  es03:
    image: elasticsearch:8.11.0
    environment:
      - cluster.name=es-cluster
      - node.name=es03
      - discovery.seed_hosts=es01,es02
```

### 9.3 资源限制

1. **队列长度限制**
```go
channel.QueueDeclare(
    "error_log_queue",
    true,
    false,
    false,
    false,
    amqp.Table{
        "x-max-length": 100000, // 最多 10 万条
    },
)
```

2. **ES 索引生命周期**
```json
PUT _ilm/policy/error_log_policy
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_size": "50GB",
            "max_age": "7d"
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

3. **磁盘空间监控**
```go
func checkDiskSpace() {
    var stat syscall.Statfs_t
    syscall.Statfs("/var/log", &stat)

    available := stat.Bavail * uint64(stat.Bsize)
    if available < 1*1024*1024*1024 { // 小于 1GB
        log.Printf("磁盘空间不足: %d MB", available/1024/1024)
    }
}
```

---

## 十、学习资源

### 书籍
- 《Go 语言编程》- 基础语法
- 《Go 并发编程实战》- channel、goroutine
- 《RabbitMQ 实战》- 消息队列
- 《Elasticsearch 权威指南》- 搜索引擎

### 在线资源
- Go 官方文档: https://go.dev/doc/
- RabbitMQ 教程: https://www.rabbitmq.com/tutorials
- Elasticsearch 文档: https://www.elastic.co/guide/
- Prometheus 文档: https://prometheus.io/docs/

### 实践项目
1. 实现日志轮转（第一阶段）
2. 实现异步写入（第二阶段）
3. 集成 RabbitMQ（第三阶段）
4. 集成 Elasticsearch（第四阶段）
5. 添加监控告警（第五阶段）

---

## 十一、总结

### 改进优先级

**必须做**（生产环境必备）：
1. ✅ 日志轮转 - 防止磁盘爆满
2. ✅ 异步写入 - 提升性能
3. ✅ 错误处理 - 防止日志系统崩溃

**应该做**（提升可靠性）：
4. ✅ 消息队列 - 解耦和持久化
5. ✅ 自动重连 - 提高可用性
6. ✅ 监控告警 - 及时发现问题

**可以做**（锦上添花）：
7. ⭕ Elasticsearch - 日志搜索和分析
8. ⭕ 链路追踪 - 分布式追踪
9. ⭕ 日志采样 - 减少日志量

### 学习路径

```
第 1 周：熟悉简化版，理解基本概念
    ↓
第 2 周：实现日志轮转和异步写入
    ↓
第 3-4 周：学习 RabbitMQ，实现消息队列
    ↓
第 5-6 周：学习 Elasticsearch，实现日志搜索
    ↓
第 7-8 周：添加监控告警，完善生产环境部署
```

### 最终架构

```
业务代码
    ↓
Logger (异步)
    ↓
RabbitMQ Cluster (高可用)
    ↓
┌─────────┬─────────┬─────────┐
│  File   │   ES    │ Alert   │
│Consumer │Consumer │Consumer │
└─────────┴─────────┴─────────┘
    ↓         ↓         ↓
  本地文件   ES集群   告警系统
                ↓
              Kibana (可视化)
```

---

**文档版本**: v1.0
**更新日期**: 2026-03-17
**适用版本**: 简化版 v1.0 及以上
