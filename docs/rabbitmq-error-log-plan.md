# RabbitMQ 错误日志异步写入方案

## 一、方案概述

使用 RabbitMQ 的发布订阅（Fanout Exchange）模式实现错误日志的异步处理，将错误日志同时写入 Elasticsearch 和本地日志文件。

### 核心优势
- **解耦**：日志生产者与消费者完全解耦
- **异步**：不阻塞主业务流程
- **可扩展**：可随时添加新的日志消费者
- **可靠性**：RabbitMQ 保证消息不丢失

## 二、架构设计

```
┌─────────────┐
│  业务代码    │
│ (Producer)  │
└──────┬──────┘
       │ 发布错误日志
       ▼
┌─────────────────────┐
│  RabbitMQ Exchange  │
│   (Fanout 模式)     │
└──────┬──────┬───────┘
       │      │
       │      └──────────────┐
       │                     │
       ▼                     ▼
┌─────────────┐      ┌─────────────┐
│  Queue 1    │      │  Queue 2    │
│ (ES队列)    │      │ (File队列)  │
└──────┬──────┘      └──────┬──────┘
       │                     │
       ▼                     ▼
┌─────────────┐      ┌─────────────┐
│ Consumer 1  │      │ Consumer 2  │
│ 写入ES      │      │ 写入本地文件│
└─────────────┘      └─────────────┘
```

## 三、目录结构规划

```
core/
├── internal/
│   ├── config/
│   │   └── config.go              # 添加 RabbitMQ 和 ES 配置
│   ├── mq/
│   │   ├── producer.go            # RabbitMQ 生产者
│   │   ├── consumer_es.go         # ES 消费者
│   │   ├── consumer_file.go       # 文件消费者
│   │   └── message.go             # 消息结构定义
│   ├── logger/
│   │   ├── error_logger.go        # 错误日志封装
│   │   └── es_client.go           # Elasticsearch 客户端
│   └── middleware/
│       └── error-recovery-middleware.go  # 全局错误捕获中间件
└── etc/
    └── core.yaml                  # 配置文件
```

## 四、实现步骤

### 4.1 添加依赖包

需要在 `go.mod` 中添加以下依赖：
```go
github.com/rabbitmq/amqp091-go v1.10.0
github.com/elastic/go-elasticsearch/v8 v8.15.0
```

### 4.2 配置文件修改

在 `core/etc/core.yaml` 中添加：
```yaml
RabbitMQ:
  Host: localhost
  Port: 5672
  Username: guest
  Password: guest
  Exchange: error_log_exchange
  ExchangeType: fanout
  QueueES: error_log_queue_es
  QueueFile: error_log_queue_file

Elasticsearch:
  Addresses:
    - http://localhost:9200
  Username: elastic
  Password: ""
  Index: error_logs

ErrorLog:
  FilePath: ./logs/error.log
  MaxSize: 100  # MB
  MaxBackups: 10
  MaxAge: 30    # days
```

### 4.3 核心文件实现

#### 1. 消息结构 (`internal/mq/message.go`)
定义统一的错误日志消息格式：
```go
type ErrorLogMessage struct {
    Timestamp   time.Time
    Level       string
    Service     string
    TraceID     string
    UserID      string
    Method      string
    Path        string
    ErrorMsg    string
    StackTrace  string
    Extra       map[string]interface{}
}
```

#### 2. 生产者 (`internal/mq/producer.go`)
- 初始化 RabbitMQ 连接
- 创建 Fanout Exchange
- 提供发布错误日志的方法
- 实现连接断开重连机制

#### 3. ES 消费者 (`internal/mq/consumer_es.go`)
- 订阅 ES 队列
- 批量写入 Elasticsearch
- 错误重试机制

#### 4. 文件消费者 (`internal/mq/consumer_file.go`)
- 订阅文件队列
- 使用 lumberjack 实现日志轮转
- 格式化日志输出

#### 5. 错误日志封装 (`internal/logger/error_logger.go`)
- 提供统一的错误日志记录接口
- 自动收集上下文信息（TraceID、UserID 等）
- 调用 MQ 生产者发布消息

#### 6. 全局错误中间件 (`internal/middleware/error-recovery-middleware.go`)
- 捕获 panic
- 自动记录错误日志
- 返回友好的错误响应

### 4.4 集成到项目

1. **修改 `internal/config/config.go`**
   添加 RabbitMQ、Elasticsearch、ErrorLog 配置结构

2. **修改 `internal/svc/service-context.go`**
   初始化 MQ 生产者和消费者

3. **修改 `internal/handler/routes.go`**
   注册全局错误恢复中间件

4. **在业务代码中使用**
   ```go
   // 记录错误日志
   logger.Error(ctx, err, "文件上传失败", map[string]interface{}{
       "file_name": fileName,
       "file_size": fileSize,
   })
   ```

## 五、关键技术点

### 5.1 RabbitMQ Fanout Exchange
- 不需要 routing key
- 消息广播到所有绑定的队列
- 适合发布订阅场景

### 5.2 消息持久化
- Exchange 持久化：`Durable: true`
- Queue 持久化：`Durable: true`
- Message 持久化：`DeliveryMode: amqp.Persistent`

### 5.3 消费者确认机制
- 手动 ACK：`AutoAck: false`
- 处理成功后才确认
- 失败时 NACK 并重新入队

### 5.4 批量写入 ES
- 使用 Bulk API
- 每 100 条或 5 秒批量提交
- 提高写入性能

### 5.5 日志轮转
- 使用 `gopkg.in/natefinch/lumberjack.v2`
- 按大小和时间自动轮转
- 自动压缩旧日志

## 六、错误处理

### 6.1 MQ 连接断开
- 实现自动重连机制
- 使用指数退避策略
- 记录重连日志

### 6.2 ES 写入失败
- 重试 3 次
- 失败后降级到本地文件
- 记录失败日志

### 6.3 消息堆积
- 设置队列最大长度
- 监控队列深度
- 告警机制

## 七、测试方案

### 7.1 单元测试
- 测试消息序列化/反序列化
- 测试生产者发送消息
- 测试消费者处理消息

### 7.2 集成测试
- 启动 RabbitMQ、ES
- 发送测试错误日志
- 验证 ES 和文件中的数据

### 7.3 压力测试
- 模拟高并发错误日志
- 验证消息不丢失
- 监控性能指标

## 八、监控指标

### 8.1 RabbitMQ 监控
- 消息发布速率
- 消息消费速率
- 队列深度
- 连接状态

### 8.2 ES 监控
- 写入成功率
- 写入延迟
- 索引大小

### 8.3 应用监控
- 错误日志数量
- 日志级别分布
- 响应时间

## 九、部署说明

### 9.1 环境准备
```bash
# 启动 RabbitMQ
docker run -d --name rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management

# 启动 Elasticsearch
docker run -d --name elasticsearch \
  -p 9200:9200 -p 9300:9300 \
  -e "discovery.type=single-node" \
  elasticsearch:8.15.0
```

### 9.2 配置修改
- 修改 `core/etc/core.yaml` 中的连接信息
- 根据环境调整队列和索引名称

### 9.3 启动顺序
1. 启动 RabbitMQ 和 ES
2. 启动消费者（自动创建 Exchange 和 Queue）
3. 启动主应用

## 十、优化建议

### 10.1 性能优化
- ES 批量写入
- 消息压缩
- 连接池复用

### 10.2 可靠性优化
- 消息持久化
- 消费者确认
- 死信队列

### 10.3 可观测性
- 添加 Prometheus 指标
- 集成链路追踪
- 日志聚合分析

## 十一、实施时间估算

- 配置和依赖安装：0.5 小时
- MQ 生产者和消费者：2 小时
- ES 客户端和文件写入：1.5 小时
- 错误日志封装和中间件：1 小时
- 集成和测试：2 小时
- 文档和优化：1 小时

**总计：约 8 小时**

## 十二、风险评估

### 12.1 技术风险
- **RabbitMQ 单点故障**：建议使用集群模式
- **ES 写入压力**：设置合理的批量大小
- **消息丢失**：启用持久化和确认机制

### 12.2 业务风险
- **日志量过大**：设置日志采样率
- **敏感信息泄露**：脱敏处理
- **存储成本**：定期清理旧日志

---

## 附录：相关命令

```bash
# 安装依赖
go get github.com/rabbitmq/amqp091-go
go get github.com/elastic/go-elasticsearch/v8
go get gopkg.in/natefinch/lumberjack.v2

# 查看 RabbitMQ 队列
rabbitmqctl list_queues

# 查看 ES 索引
curl -X GET "localhost:9200/_cat/indices?v"

# 查看错误日志
curl -X GET "localhost:9200/error_logs/_search?pretty"
```
