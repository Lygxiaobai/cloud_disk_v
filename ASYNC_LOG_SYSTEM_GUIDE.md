# RabbitMQ 异步日志系统 - 使用指南

## 📋 系统架构

```
API 接口 → LogError() → 日志生产者 → fanout 交换机 → 本地日志队列 → 本地消费者 → logs/error.log
                                                    → ES 日志队列   → ES 消费者   → Elasticsearch
```

### 核心特性

1. **fanout 交换机**：一条日志消息广播到多个队列
2. **两个独立消费者**：
   - 本地文件消费者：写入 `logs/error.log`
   - ES 消费者：写入 Elasticsearch（当前为模拟模式）
3. **并行处理**：两个消费者同时处理，互不影响
4. **故障隔离**：一个消费者失败不影响另一个
5. **降级机制**：MQ 发送失败时自动降级到本地文件

---

## 🚀 快速开始

### 1. 启动 RabbitMQ

```bash
docker start rabbitmq
```

### 2. 启动 Elasticsearch（可选）

如果需要将日志写入 ES，需要先启动 Elasticsearch：

```bash
# 使用启动脚本
bash start_elasticsearch.sh

# 或手动启动
docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  elasticsearch:8.11.0
```

验证 ES 是否启动成功：
```bash
curl http://localhost:9200
```

### 3. 编译所有程序

```bash
cd D:/Go_Project/my_cloud_disk/core

# 编译本地日志消费者
go build -o bin/log_worker_local.exe cmd/log_worker_local/main.go

# 编译 ES 日志消费者
go build -o bin/log_worker_es.exe cmd/log_worker_es/main.go

# 编译 ES 日志查询工具
go build -o bin/query_es_logs.exe cmd/query_es_logs/main.go

# 编译主服务
go build -o bin/core.exe core.go
```

### 4. 启动服务

**方式 1：使用启动脚本（推荐）**
```bash
bash start_async_log_system.sh
```

**方式 2：手动启动**
```bash
# 终端 1：启动本地日志消费者
./bin/log_worker_local.exe -f etc/core-api.yaml

# 终端 2：启动 ES 日志消费者
./bin/log_worker_es.exe -f etc/core-api.yaml

# 终端 3：启动主服务
./bin/core.exe -f etc/core-api.yaml
```

### 4. 测试日志系统

```bash
bash test_async_log_system.sh
```

### 5. 查询 ES 日志

```bash
# 查看最新 10 条日志
./bin/query_es_logs.exe -n 10

# 只查看 ERROR 级别的日志
./bin/query_es_logs.exe -level ERROR -n 20

# 按 trace_id 查询
./bin/query_es_logs.exe -trace abc123

# 查看帮助
./bin/query_es_logs.exe -h
```

---

## 📊 监控和验证

### 1. RabbitMQ 管理界面

访问：http://localhost:15672
- 用户名：guest
- 密码：guest

查看：
- **Exchanges** → `log_exchange`（fanout 类型）
- **Queues** → `local_log_queue` 和 `es_log_queue`
- 确认两个队列都有消费者连接

### 2. 查看队列状态（命令行）

```bash
# 查看交换机
curl -u guest:guest http://localhost:15672/api/exchanges/%2F/log_exchange | jq

# 查看本地日志队列
curl -u guest:guest http://localhost:15672/api/queues/%2F/local_log_queue | jq

# 查看 ES 日志队列
curl -u guest:guest http://localhost:15672/api/queues/%2F/es_log_queue | jq
```

### 3. 查看本地日志文件

```bash
# 查看最新日志
tail -f D:/Go_Project/my_cloud_disk/core/logs/error.log

# 格式化查看
tail -10 D:/Go_Project/my_cloud_disk/core/logs/error.log | jq '.'
```

### 4. 查看 Elasticsearch 日志

```bash
# 使用查询工具
cd D:/Go_Project/my_cloud_disk/core
./bin/query_es_logs.exe -n 10

# 或直接查询 ES API
curl -X GET "http://localhost:9200/logs-*/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10,
  "sort": [{"timestamp": "desc"}],
  "query": {"match_all": {}}
}
'

# 查看索引列表
curl http://localhost:9200/_cat/indices/logs-*?v

# 查看索引模板
curl http://localhost:9200/_index_template/logs-template?pretty
```

---

## 🔧 配置说明

### core-api.yaml

```yaml
RabbitMQ:
  URL: amqp://guest:guest@localhost:5672/
  EmailQueue: email_queue
  LogExchange: log_exchange        # 日志交换机（fanout）
  LocalLogQueue: local_log_queue   # 本地日志队列
  ESLogQueue: es_log_queue         # ES 日志队列

Elasticsearch:
  Addresses:
    - http://localhost:9200
  Username: ""                     # 如果 ES 启用了认证
  Password: ""
  IndexPrefix: logs                # 索引前缀（logs-2026-03-18）
```

---

## 📝 日志消息格式

```json
{
  "timestamp": "2026-03-18 15:30:45",
  "level": "ERROR",
  "trace_id": "abc123",
  "user_id": "user_001",
  "method": "POST",
  "path": "/api/login",
  "message": "登录失败: 密码错误",
  "stack_trace": "goroutine 1 [running]...",
  "extra": {
    "ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0"
  }
}
```

---

## 🎯 使用示例

### 在代码中记录日志

```go
import "cloud_disk/core/internal/logger"

// 记录错误日志
logger.LogError(ctx, "用户登录失败", err, map[string]interface{}{
    "email": "user@example.com",
    "ip": "192.168.1.100",
})

// 记录致命错误
logger.LogFatal(ctx, "数据库连接失败", err, nil)

// 记录 panic
logger.LogPanic(ctx, panicValue, map[string]interface{}{
    "request_id": "req_123",
})
```

### 日志流程

1. **调用 LogError()**
2. **生产者发送到 fanout 交换机**（5-10ms）
3. **交换机广播到两个队列**
4. **本地消费者写入文件**（并行）
5. **ES 消费者写入 ES**（并行）

---

## 🔍 故障排查

### Q1: 消费者启动失败？

**检查：**
```bash
# 1. RabbitMQ 是否运行
docker ps | grep rabbitmq

# 2. 配置文件是否正确
cat etc/core-api.yaml | grep RabbitMQ -A 5

# 3. 端口是否被占用
netstat -an | grep 5672
```

### Q2: 日志没有写入本地文件？

**检查：**
```bash
# 1. 本地消费者是否运行
ps aux | grep log_worker_local

# 2. 队列是否有消费者
curl -u guest:guest http://localhost:15672/api/queues/%2F/local_log_queue | jq '.consumers'

# 3. 查看消费者日志
# 应该看到 "✓ 日志处理成功" 的输出
```

### Q3: 日志没有发送到 MQ？

**检查：**
```bash
# 1. 主服务是否初始化异步日志
# 启动日志应该显示：
# "✓ 异步日志系统初始化成功（RabbitMQ fanout 模式）"

# 2. 查看交换机消息统计
curl -u guest:guest http://localhost:15672/api/exchanges/%2F/log_exchange | jq '.message_stats'

# 3. 触发一个错误，查看是否有消息发布
```

### Q4: ES 日志消费者连接失败？

**检查：**
```bash
# 1. Elasticsearch 是否运行
curl http://localhost:9200

# 2. 检查 ES 容器状态
docker ps | grep elasticsearch

# 3. 查看 ES 日志
docker logs elasticsearch

# 4. 如果 ES 未启动，运行启动脚本
bash start_elasticsearch.sh
```

### Q5: 如何查看 ES 中的日志？

**方法 1：使用查询工具**
```bash
./bin/query_es_logs.exe -n 10
```

**方法 2：直接查询 ES API**
```bash
curl -X GET "http://localhost:9200/logs-*/_search?pretty&size=10&sort=timestamp:desc"
```

### Q6: MQ 发送失败，如何降级？

系统会自动降级到本地文件：
```go
err := globalLogger.logProducer.SendLogMessage(logMsg)
if err != nil {
    // 自动降级到本地文件
    log.Printf("发送日志到 MQ 失败，降级到本地文件: %v", err)
    writeToLocalFile(errorLog)
}
```

---

## 📈 性能优势

### 对比同步日志

| 指标 | 同步日志 | 异步日志（RabbitMQ） |
|------|---------|---------------------|
| 日志写入时间 | 阻塞 5-10ms | 非阻塞 1-2ms |
| ES 写入时间 | 阻塞 50-100ms | 异步处理 |
| 接口响应影响 | 直接影响 | 不影响 |
| 故障隔离 | 日志失败影响接口 | 完全隔离 |
| 可扩展性 | 受限 | 可水平扩展 |

### 并行处理优势

```
同步模式：
API → 写本地文件(5ms) → 写ES(50ms) = 55ms

异步模式：
API → 发送MQ(1ms) → 立即返回
                  ↓
            本地写入(5ms) ← 并行
            ES写入(50ms)  ← 并行
```

---

## 🚀 扩展建议

### 1. ES 索引生命周期管理

随着日志增长，可以配置索引生命周期策略（ILM）：

```bash
# 创建 ILM 策略（保留 7 天）
curl -X PUT "http://localhost:9200/_ilm/policy/logs-policy?pretty" -H 'Content-Type: application/json' -d'
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_age": "1d",
            "max_size": "50gb"
          }
        }
      },
      "delete": {
        "min_age": "7d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
'
```

### 2. ES 批量写入优化

修改 `cmd/log_worker_es/main.go`，实现批量写入：

```go
// 使用缓冲区收集日志，每 100 条或 5 秒批量写入
var logBuffer []rabbitmq.LogMessage
var bufferMutex sync.Mutex

func batchWriteHandler(logMsg *rabbitmq.LogMessage) error {
    bufferMutex.Lock()
    logBuffer = append(logBuffer, *logMsg)
    shouldFlush := len(logBuffer) >= 100
    bufferMutex.Unlock()

    if shouldFlush {
        return flushLogs()
    }
    return nil
}
```

### 3. 实现真实的 ES 写入

当前已实现完整的 ES 写入功能，包括：
- ✅ 自动创建索引模板
- ✅ 按日期创建索引（logs-2026-03-18）
- ✅ 错误重试机制（最多 3 次）
- ✅ 日志查询工具

### 4. 添加更多消费者

可以启动多个消费者实例来提高吞吐量：

```bash
# 启动 3 个本地日志消费者
./bin/log_worker_local.exe -f etc/core-api.yaml &
./bin/log_worker_local.exe -f etc/core-api.yaml &
./bin/log_worker_local.exe -f etc/core-api.yaml &

# 启动 2 个 ES 消费者
./bin/log_worker_es.exe -f etc/core-api.yaml &
./bin/log_worker_es.exe -f etc/core-api.yaml &
```

### 3. 添加其他日志目标

可以添加更多队列和消费者：
- 发送到 Kafka
- 发送到 Sentry
- 发送到钉钉/企业微信告警

---

## 📦 文件清单

```
core/
├── internal/
│   ├── rabbitmq/
│   │   ├── rabbitmq.go           # RabbitMQ 连接管理（新增交换机支持）
│   │   ├── log_producer.go       # 日志生产者
│   │   └── log_consumer.go       # 日志消费者
│   ├── logger/
│   │   └── simple_logger.go      # 日志记录器（支持异步）
│   ├── config/
│   │   └── config.go             # 配置（新增日志队列配置）
│   └── svc/
│       └── service-context.go    # 服务上下文（初始化异步日志）
├── cmd/
│   ├── log_worker_local/
│   │   └── main.go               # 本地日志消费者
│   ├── log_worker_es/
│   │   └── main.go               # ES 日志消费者
│   └── query_es_logs/
│       └── main.go               # ES 日志查询工具
├── etc/
│   └── core-api.yaml             # 配置文件
└── bin/
    ├── log_worker_local.exe      # 编译后的本地消费者
    ├── log_worker_es.exe         # 编译后的 ES 消费者
    ├── query_es_logs.exe         # 编译后的查询工具
    └── core.exe                  # 编译后的主服务
```

---

## ✅ 总结

### 实现的功能

1. ✅ fanout 交换机广播日志消息
2. ✅ 两个独立消费者（本地文件 + ES）
3. ✅ 并行处理，互不影响
4. ✅ 故障隔离和降级机制
5. ✅ 完整的监控和测试工具
6. ✅ Elasticsearch 索引模板自动创建
7. ✅ 日志查询工具（支持过滤和搜索）
8. ✅ ES 写入重试机制

### 技术亮点

1. **解耦**：日志写入与业务逻辑完全分离
2. **并行**：本地和 ES 同时写入，速度快
3. **可靠**：消息持久化，手动确认
4. **可扩展**：可随时添加更多消费者
5. **降级**：MQ 失败自动降级到本地文件

---

**文档版本：** v1.0
**创建时间：** 2026-03-18
