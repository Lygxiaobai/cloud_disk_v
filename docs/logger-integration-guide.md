# 日志系统集成完成

## 已完成的工作

### 1. 初始化日志系统
在 `core/core.go` 的 main 函数中添加了日志系统初始化：
```go
if err := logger.InitSimpleLogger("./logs/error.log"); err != nil {
    log.Fatalf("初始化日志系统失败: %v", err)
}
defer logger.Close()
```

### 2. 创建错误恢复中间件
新建 `core/internal/middleware/error-recovery-middleware.go`，用于捕获 panic：
- 自动捕获所有 panic
- 记录完整的堆栈信息
- 返回友好的错误提示

### 3. 在业务逻辑中使用
在 `file-upload-logic.go` 中添加了错误日志记录示例：
```go
logger.LogError(ctx, "文件上传失败", err, map[string]interface{}{
    "file_name": req.Name,
    "file_size": req.Size,
    "file_hash": req.Hash,
})
```

## 如何使用

### 在 Logic 层记录错误

```go
import "cloud_disk/core/internal/logger"

func (l *YourLogic) YourMethod(req *types.Request) error {
    // 构建上下文
    ctx := context.WithValue(l.ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/your/path")
    ctx = context.WithValue(ctx, "user_id", "user-123")

    // 业务逻辑
    if err := doSomething(); err != nil {
        // 记录错误日志
        logger.LogError(ctx, "操作失败", err, map[string]interface{}{
            "param1": req.Param1,
            "param2": req.Param2,
        })
        return err
    }

    return nil
}
```

### 使用错误恢复中间件

在 `routes.go` 中注册中间件（需要手动添加）：
```go
server.Use(middleware.NewErrorRecoveryMiddleware().Handle)
```

### 三种日志级别的使用场景

**ERROR - 一般错误**
```go
logger.LogError(ctx, "文件上传失败", err, extra)
```
- 文件上传失败
- 数据库查询失败
- 网络请求超时
- 业务逻辑错误

**FATAL - 致命错误**
```go
logger.LogFatal(ctx, "数据库连接失败", err, extra)
```
- 数据库连接失败
- Redis 连接失败
- 关键配置缺失
- 系统资源耗尽

**PANIC - 程序崩溃**
```go
logger.LogPanic(ctx, panicValue, extra)
```
- 空指针引用
- 数组越界
- 类型断言失败
- 在 recover() 中使用

## 日志文件位置

- 日志文件：`./logs/error.log`
- 每条日志一行 JSON 格式

## 查看日志

```bash
# 查看全部日志
cat ./logs/error.log

# 实时监控
tail -f ./logs/error.log

# 格式化查看（需要 jq）
cat ./logs/error.log | jq .

# 查看最后 10 条
tail -10 ./logs/error.log | jq .

# 筛选 ERROR 级别
cat ./logs/error.log | jq 'select(.level=="ERROR")'

# 筛选特定用户
cat ./logs/error.log | jq 'select(.user_id=="user-123")'
```

## 测试

运行集成测试：
```bash
cd core
go run test_logger_integration.go
```

## 下一步建议

1. **在更多 Logic 中添加日志记录**
   - user-login-logic.go
   - user-register-logic.go
   - file-upload-multipart-logic.go

2. **注册错误恢复中间件**
   - 在 routes.go 中全局注册

3. **添加 TraceID 生成**
   - 在请求入口生成唯一 TraceID
   - 方便追踪整个请求链路

4. **考虑日志轮转**
   - 当日志量大时，实现第一阶段改进
   - 使用 lumberjack 库

## 注意事项

- 不要在循环中频繁记录日志
- 敏感信息（密码、token）不要记录到日志
- extra 字段不要放太大的数据
- 日志文件会持续增长，需要定期清理或实现轮转
