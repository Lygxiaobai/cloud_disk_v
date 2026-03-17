# 简化版错误日志使用指南

## 一、简介

这是一个简化版的错误日志系统，专为新手设计，易于理解和使用。

### 核心功能
- ✅ 记录三种级别的错误日志：Error、Fatal、Panic
- ✅ 自动记录堆栈信息
- ✅ 支持上下文信息（TraceID、UserID、请求路径等）
- ✅ 日志以 JSON 格式保存到文件
- ✅ 同时输出到控制台

### 与复杂版本的区别
| 功能 | 简化版 | 复杂版 |
|------|--------|--------|
| 日志存储 | 本地文件 | 本地文件 + Elasticsearch |
| 消息队列 | 无 | RabbitMQ |
| 自动重连 | 无 | 有 |
| 批量写入 | 无 | 有 |
| 代码行数 | ~150 行 | ~800 行 |
| 依赖包 | 0 个 | 3 个 |

---

## 二、快速开始

### 1. 初始化日志记录器

在程序启动时初始化：

```go
package main

import (
    "cloud_disk/core/internal/logger"
    "log"
)

func main() {
    // 初始化日志记录器
    if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
        log.Fatalf("初始化日志记录器失败: %v", err)
    }
    defer logger.Close()

    // 你的业务代码...
}
```

### 2. 记录错误日志

```go
import (
    "cloud_disk/core/internal/logger"
    "context"
    "errors"
)

func uploadFile() error {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "trace_id", "trace-123")
    ctx = context.WithValue(ctx, "user_id", "user-456")
    ctx = context.WithValue(ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/file/upload")

    // 模拟错误
    err := errors.New("connection timeout")

    // 记录错误日志
    logger.LogError(ctx, "文件上传失败", err, map[string]interface{}{
        "file_name": "test.pdf",
        "file_size": 1024000,
    })

    return err
}
```

### 3. 在中间件中使用

```go
package middleware

import (
    "cloud_disk/core/internal/logger"
    "context"
    "fmt"
    "net/http"
)

func SimpleErrorRecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                ctx := context.WithValue(r.Context(), "method", r.Method)
                ctx = context.WithValue(ctx, "path", r.URL.Path)

                logger.LogPanic(ctx, err, nil)

                http.Error(w, "服务器内部错误", http.StatusInternalServerError)
            }
        }()

        next(w, r)
    }
}
```

---

## 三、API 说明

### InitSimpleLogger

初始化日志记录器。

```go
func InitSimpleLogger(logFilePath string) error
```

**参数**：
- `logFilePath`: 日志文件路径，例如 `"./logs/error.log"`

**返回**：
- `error`: 初始化失败时返回错误

**示例**：
```go
err := logger.InitSimpleLogger("./logs/error.log")
```

---

### LogError

记录 Error 级别的错误日志。

```go
func LogError(ctx context.Context, message string, err error, extra map[string]interface{})
```

**参数**：
- `ctx`: 上下文，包含 trace_id、user_id、method、path 等信息
- `message`: 错误描述
- `err`: 错误对象
- `extra`: 额外的自定义字段（可选）

**示例**：
```go
logger.LogError(ctx, "数据库查询失败", err, map[string]interface{}{
    "sql": "SELECT * FROM users",
    "retry": 3,
})
```

---

### LogFatal

记录 Fatal 级别的错误日志（严重错误）。

```go
func LogFatal(ctx context.Context, message string, err error, extra map[string]interface{})
```

**使用场景**：
- 数据库连接失败
- 关键配置缺失
- 系统资源耗尽

**示例**：
```go
logger.LogFatal(ctx, "Redis 连接失败", err, map[string]interface{}{
    "host": "localhost",
    "port": 6379,
})
```

---

### LogPanic

记录 Panic 级别的错误日志（程序崩溃）。

```go
func LogPanic(ctx context.Context, panicValue interface{}, extra map[string]interface{})
```

**使用场景**：
- 捕获 panic
- 空指针引用
- 数组越界

**示例**：
```go
defer func() {
    if r := recover(); r != nil {
        logger.LogPanic(ctx, r, map[string]interface{}{
            "action": "delete_file",
        })
    }
}()
```

---

### Close

关闭日志记录器，释放文件句柄。

```go
func Close() error
```

**示例**：
```go
defer logger.Close()
```

---

## 四、日志格式

### JSON 格式

每条日志以 JSON 格式保存，一行一条：

```json
{
  "timestamp": "2026-03-17 10:30:45",
  "level": "ERROR",
  "trace_id": "trace-123",
  "user_id": "user-456",
  "method": "POST",
  "path": "/file/upload",
  "message": "文件上传失败: connection timeout",
  "stack_trace": "goroutine 1 [running]:\n...",
  "extra": {
    "file_name": "test.pdf",
    "file_size": 1024000
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| timestamp | string | 时间戳（格式：2006-01-02 15:04:05） |
| level | string | 日志级别（ERROR/FATAL/PANIC） |
| trace_id | string | 追踪ID（用于链路追踪） |
| user_id | string | 用户ID |
| method | string | HTTP 方法（GET/POST/PUT/DELETE） |
| path | string | 请求路径 |
| message | string | 错误消息 |
| stack_trace | string | 堆栈信息 |
| extra | object | 自定义字段 |

---

## 五、测试

### 运行测试程序

```bash
cd D:/Go_Project/my_cloud_disk/core
go run test_simple_logger.go
```

### 预期输出

```
========================================
简化版错误日志测试
========================================

测试 1: 初始化日志记录器
✅ 日志记录器初始化成功

测试 2: 记录 Error 级别日志
✅ Error 日志已记录

测试 3: 记录 Fatal 级别日志
✅ Fatal 日志已记录

测试 4: 记录 Panic 级别日志
✅ Panic 日志已记录

========================================
测试完成！请查看 ./logs/error.log 文件
========================================
```

### 查看日志文件

```bash
cat ./logs/error.log
```

---

## 六、常见问题

### Q1: 日志文件在哪里？

A: 默认在 `./logs/error.log`，你可以在初始化时指定其他路径。

### Q2: 如何查看日志？

A: 使用文本编辑器打开，或使用命令：
```bash
# 查看全部
cat ./logs/error.log

# 实时监控
tail -f ./logs/error.log

# 格式化查看（需要 jq）
cat ./logs/error.log | jq .
```

### Q3: 日志文件会无限增长吗？

A: 简化版不会自动轮转。如果需要日志轮转功能，请参考"后续改进建议"。

### Q4: 如何在 go-zero 中使用？

A: 在 `main.go` 中初始化，在 handler 中使用：

```go
// main.go
func main() {
    logger.InitSimpleLogger("./logs/error.log")
    defer logger.Close()

    // 启动服务...
}

// handler
func (l *UploadLogic) Upload(req *types.UploadReq) error {
    if err := l.uploadToOSS(req); err != nil {
        logger.LogError(l.ctx, "上传失败", err, map[string]interface{}{
            "file_name": req.FileName,
        })
        return err
    }
    return nil
}
```

### Q5: 性能如何？

A: 简化版直接写文件，性能足够日常使用。如果需要高性能，请参考"后续改进建议"。

---

## 七、与复杂版本对比

### 代码复杂度

**简化版**：
- `simple_logger.go`: ~150 行
- `simple-error-middleware.go`: ~30 行
- 总计: ~180 行

**复杂版**：
- `error_logger.go`: ~170 行
- `es_client.go`: ~175 行
- `producer.go`: ~210 行
- `consumer_file.go`: ~200 行
- `consumer_es.go`: ~150 行
- `message.go`: ~30 行
- `error-recovery-middleware.go`: ~37 行
- 总计: ~970 行

### 依赖包

**简化版**：
- 无外部依赖（只使用 Go 标准库）

**复杂版**：
- github.com/rabbitmq/amqp091-go
- github.com/elastic/go-elasticsearch/v8
- gopkg.in/natefinch/lumberjack.v2

### 功能对比

| 功能 | 简化版 | 复杂版 |
|------|--------|--------|
| 本地文件日志 | ✅ | ✅ |
| Elasticsearch | ❌ | ✅ |
| RabbitMQ | ❌ | ✅ |
| 自动重连 | ❌ | ✅ |
| 批量写入 | ❌ | ✅ |
| 日志轮转 | ❌ | ✅ |
| 分布式追踪 | ✅ | ✅ |
| 堆栈信息 | ✅ | ✅ |

---

## 八、何时升级到复杂版本？

当你遇到以下情况时，可以考虑升级：

1. **日志量大**：每天产生超过 10GB 日志
2. **需要搜索**：需要快速搜索和分析日志
3. **分布式系统**：多个服务需要统一日志管理
4. **高可用要求**：需要日志不丢失的保证
5. **性能要求**：需要异步处理，不阻塞业务

---

## 九、文件清单

```
core/
├── internal/
│   ├── logger/
│   │   └── simple_logger.go          # 简化版日志记录器
│   └── middleware/
│       └── simple-error-middleware.go # 简化版中间件
├── test_simple_logger.go              # 测试程序
└── logs/
    └── error.log                      # 日志文件（自动创建）
```

---

## 十、下一步

1. 运行测试程序，熟悉基本用法
2. 在你的项目中集成简化版日志
3. 阅读"后续改进建议"文档，了解可以优化的方向
4. 当需求增长时，逐步升级到复杂版本
