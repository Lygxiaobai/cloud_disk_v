# TraceID 复用实现

## 问题

之前每个 Logic 都重新生成 TraceID，导致同一个请求的不同日志有不同的 TraceID，无法关联。

## 解决方案

### 架构设计

```
HTTP 请求
    ↓
ErrorRecoveryMiddleware
    ↓ 生成 TraceID
    ↓ 存入 Context
    ↓
Handler
    ↓
Logic (从 Context 获取 TraceID)
    ↓
Logger (使用相同的 TraceID)
```

### 实现步骤

#### 1. 中间件生成并存储 TraceID

```go
// error-recovery-middleware.go
func (m *ErrorRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 生成或获取 TraceID
        traceID := r.Header.Get("X-Trace-Id")
        if traceID == "" {
            traceID = helper.UUID()
        }

        // 存入 Context
        ctx := context.WithValue(r.Context(), "trace_id", traceID)
        r = r.WithContext(ctx)

        next(w, r)
    }
}
```

#### 2. Logic 从 Context 获取 TraceID

```go
// user-login-logic.go
func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) error {
    // 从 context 中获取 TraceID
    traceID, _ := l.ctx.Value("trace_id").(string)

    // 构建上下文
    ctx := context.WithValue(l.ctx, "trace_id", traceID)
    ctx = context.WithValue(ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/user/login")

    // 记录日志
    logger.LogError(ctx, "登录失败", err, extra)
}
```

## 效果对比

### 之前（每个 Logic 生成新的 TraceID）

```json
// 用户登录请求
{"trace_id": "abc123", "message": "数据库查询失败"}
{"trace_id": "def456", "message": "生成Token失败"}
{"trace_id": "ghi789", "message": "登录失败"}
```

❌ 无法关联同一个请求的日志

### 现在（复用中间件的 TraceID）

```json
// 用户登录请求
{"trace_id": "abc123", "message": "数据库查询失败"}
{"trace_id": "abc123", "message": "生成Token失败"}
{"trace_id": "abc123", "message": "登录失败"}
```

✅ 所有日志都有相同的 TraceID，可以轻松关联

## 使用示例

### 查找同一个请求的所有日志

```bash
# 查找 TraceID 为 abc123 的所有日志
cat ./logs/error.log | jq 'select(.trace_id=="abc123")'
```

输出：
```json
{"timestamp":"2026-03-17 20:00:00","trace_id":"abc123","message":"数据库查询失败"}
{"timestamp":"2026-03-17 20:00:01","trace_id":"abc123","message":"生成Token失败"}
{"timestamp":"2026-03-17 20:00:02","trace_id":"abc123","message":"登录失败"}
```

### 追踪请求链路

```
TraceID: abc123

20:00:00 - 用户登录请求
20:00:00 - 数据库查询用户信息
20:00:01 - 生成 Token
20:00:01 - 生成 RefreshToken
20:00:02 - 返回登录成功
```

## 标准模板

在所有 Logic 中使用这个模板：

```go
func (l *YourLogic) YourMethod(req *types.Request) error {
    // 1. 从 context 中获取 TraceID（由中间件设置）
    traceID, _ := l.ctx.Value("trace_id").(string)

    // 2. 构建上下文信息
    ctx := context.WithValue(l.ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/your/path")
    ctx = context.WithValue(ctx, "trace_id", traceID)

    // 3. 业务逻辑 + 日志记录
    if err := doSomething(); err != nil {
        logger.LogError(ctx, "操作失败", err, map[string]interface{}{
            "param": req.Param,
        })
        return err
    }

    return nil
}
```

## 优势

### 1. 请求链路追踪

同一个请求的所有操作都有相同的 TraceID：

```
用户登录 (abc123)
  ↓ 查询数据库 (abc123)
  ↓ 生成 Token (abc123)
  ↓ 写入 Redis (abc123)
  ↓ 返回结果 (abc123)
```

### 2. 快速问题定位

用户报告问题时，通过 TraceID 快速找到所有相关日志：

```bash
# 用户说："我刚才登录失败了"
# 找到失败的 TraceID: abc123

# 查看整个请求链路
cat ./logs/error.log | jq 'select(.trace_id=="abc123")'

# 立即看到：
# - 在哪一步失败的
# - 失败的原因
# - 相关的参数
```

### 3. 性能分析

统计每个请求的耗时：

```bash
# 找出同一个 TraceID 的第一条和最后一条日志
# 计算时间差，得到请求总耗时
```

### 4. 分布式追踪

如果调用其他服务，传递 TraceID：

```go
// 调用其他服务时传递 TraceID
req, _ := http.NewRequest("POST", "http://other-service/api", body)
req.Header.Set("X-Trace-Id", traceID)
resp, _ := client.Do(req)
```

这样整个分布式系统的请求链路都可以追踪。

## 注意事项

### 1. TraceID 为空的情况

如果 Logic 中获取不到 TraceID（比如不是通过 HTTP 请求调用的），可以生成一个新的：

```go
traceID, _ := l.ctx.Value("trace_id").(string)
if traceID == "" {
    traceID = helper.UUID() // 生成新的
}
```

### 2. 异步任务

如果有异步任务（goroutine），需要传递 TraceID：

```go
traceID, _ := l.ctx.Value("trace_id").(string)

go func() {
    ctx := context.WithValue(context.Background(), "trace_id", traceID)
    logger.LogError(ctx, "异步任务失败", err, extra)
}()
```

### 3. 定时任务

定时任务没有 HTTP 请求，需要生成新的 TraceID：

```go
func cronJob() {
    traceID := helper.UUID()
    ctx := context.WithValue(context.Background(), "trace_id", traceID)
    logger.LogError(ctx, "定时任务失败", err, extra)
}
```

## 总结

✅ 中间件生成一次 TraceID
✅ 所有 Logic 复用相同的 TraceID
✅ 同一个请求的所有日志都可以关联
✅ 方便追踪和排查问题

现在整个请求链路都可以通过 TraceID 追踪了！
