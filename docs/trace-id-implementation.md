# TraceID 实现说明

## 问题

之前的实现中，TraceID 从 HTTP Header 中读取 `X-Trace-Id`，但没有任何地方设置这个值，导致 TraceID 一直为空。

## 解决方案

### 方案 1：中间件自动生成（已实现）

在 `error-recovery-middleware.go` 中：

```go
func (m *ErrorRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 获取或生成 TraceID
        traceID := r.Header.Get("X-Trace-Id")
        if traceID == "" {
            traceID = helper.UUID() // 自动生成
        }

        // 2. 设置到 Header 中
        r.Header.Set("X-Trace-Id", traceID)

        // 3. 后续使用
        defer func() {
            if err := recover(); err != nil {
                ctx = context.WithValue(ctx, "trace_id", traceID)
                logger.LogPanic(ctx, err, extra)
            }
        }()

        next(w, r)
    }
}
```

### 方案 2：在 Logic 层生成

在每个 Logic 方法中：

```go
func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) error {
    // 生成 TraceID
    traceID := helper.UUID()
    ctx := context.WithValue(l.ctx, "trace_id", traceID)

    // 使用 ctx 记录日志
    logger.LogError(ctx, "登录失败", err, extra)
}
```

## TraceID 的作用

### 1. 链路追踪

在分布式系统中，一个请求可能经过多个服务：

```
用户请求 → 网关 → 用户服务 → 数据库
         ↓
      文件服务 → OSS
```

使用相同的 TraceID，可以追踪整个请求链路。

### 2. 日志关联

同一个请求的所有日志都有相同的 TraceID：

```json
// 请求开始
{"trace_id": "abc123", "message": "用户登录请求"}

// 数据库查询
{"trace_id": "abc123", "message": "查询用户信息"}

// 生成 Token
{"trace_id": "abc123", "message": "生成 Token"}

// 请求结束
{"trace_id": "abc123", "message": "登录成功"}
```

通过 TraceID 可以快速找到同一个请求的所有日志。

### 3. 问题排查

当用户报告问题时，可以通过 TraceID 快速定位：

```bash
# 查找特定请求的所有日志
cat ./logs/error.log | jq 'select(.trace_id=="abc123")'
```

## 使用方式

### 客户端传递 TraceID（可选）

如果客户端已经有 TraceID（比如前端生成），可以通过 Header 传递：

```bash
curl -X POST http://localhost:8888/user/login \
  -H "X-Trace-Id: frontend-generated-id-123" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","password":"123456"}'
```

### 服务端自动生成

如果客户端没有传递 TraceID，中间件会自动生成：

```bash
curl -X POST http://localhost:8888/user/login \
  -H "Content-Type: application/json" \
  -d '{"name":"test","password":"123456"}'

# 服务端自动生成 TraceID: "550e8400-e29b-41d4-a716-446655440000"
```

### 在 Logic 中使用

```go
func (l *YourLogic) YourMethod(req *types.Request) error {
    // 生成 TraceID
    traceID := helper.UUID()
    ctx := context.WithValue(l.ctx, "trace_id", traceID)
    ctx = context.WithValue(ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/your/path")

    // 记录日志
    if err := doSomething(); err != nil {
        logger.LogError(ctx, "操作失败", err, map[string]interface{}{
            "param": req.Param,
        })
        return err
    }

    return nil
}
```

## 查看 TraceID

### 查看日志
```bash
# 查看所有日志的 TraceID
cat ./logs/error.log | jq '.trace_id'

# 查找特定 TraceID 的所有日志
cat ./logs/error.log | jq 'select(.trace_id=="abc123")'

# 统计每个 TraceID 的日志数量
cat ./logs/error.log | jq -r '.trace_id' | sort | uniq -c
```

### 日志示例

```json
{
  "timestamp": "2026-03-17 20:00:00",
  "level": "ERROR",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/user/login",
  "message": "用户登录失败: 用户名或密码错误",
  "extra": {
    "username": "test@example.com"
  }
}
```

## 最佳实践

### 1. 在入口处生成

在请求的最开始（中间件）生成 TraceID，然后传递给后续所有操作。

### 2. 保持一致

同一个请求的所有日志都应该使用相同的 TraceID。

### 3. 跨服务传递

如果调用其他服务，需要通过 Header 传递 TraceID：

```go
req, _ := http.NewRequest("POST", "http://other-service/api", body)
req.Header.Set("X-Trace-Id", traceID)
resp, _ := client.Do(req)
```

### 4. 返回给客户端

可以在响应 Header 中返回 TraceID，方便客户端追踪：

```go
w.Header().Set("X-Trace-Id", traceID)
```

## 进阶：集成 OpenTelemetry

如果需要更完善的链路追踪，可以集成 OpenTelemetry：

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (l *Logic) Method(req *Request) error {
    // 从 context 中获取 span
    span := trace.SpanFromContext(l.ctx)
    traceID := span.SpanContext().TraceID().String()

    // 使用 traceID
    ctx := context.WithValue(l.ctx, "trace_id", traceID)
    logger.LogError(ctx, "error", err, extra)
}
```

## 总结

- ✅ 中间件自动生成 TraceID
- ✅ 支持客户端传递 TraceID
- ✅ 所有日志都包含 TraceID
- ✅ 方便追踪和排查问题

现在每个请求都有唯一的 TraceID，可以轻松追踪整个请求链路！
