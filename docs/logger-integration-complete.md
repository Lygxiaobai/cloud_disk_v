# 日志系统完善总结

## 已完成的工作

### 1. 注册错误恢复中间件 ✅

**修改文件：**
- `core/internal/svc/service-context.go` - 添加 ErrorRecovery 中间件
- `core/internal/handler/routes.go` - 在所有路由组中注册中间件

**效果：**
- 所有接口都被错误恢复中间件保护
- 任何 panic 都会被捕获并记录日志
- 程序不会因为单个请求的 bug 而崩溃

### 2. 在关键业务逻辑中添加日志记录 ✅

**已添加日志的文件：**

#### 用户相关
- `user-login-logic.go` - 用户登录
  - 数据库查询失败
  - 用户名或密码错误
  - Token 生成失败

- `user-register-logic.go` - 用户注册
  - 验证码验证失败
  - 邮箱已被注册
  - 用户名已被注册
  - 插入用户失败

#### 文件相关
- `file-upload-logic.go` - 文件上传
  - 文件上传失败

- `user-file-delete-logic.go` - 文件删除
  - 删除文件失败

#### 分享相关
- `share-basic-create-logic.go` - 创建分享
  - 查询文件失败
  - 文件不存在
  - 插入分享记录失败

### 3. 日志记录的信息

每条日志都包含：
- **时间戳** - 错误发生时间
- **日志级别** - ERROR/FATAL/PANIC
- **上下文信息**
  - method - HTTP 方法（POST/GET/DELETE）
  - path - 请求路径
  - user_identity - 用户标识
- **错误消息** - 具体的错误描述
- **堆栈信息** - 完整的调用堆栈
- **额外字段** - 业务相关的详细信息

## 日志记录示例

### 用户登录失败
```json
{
  "timestamp": "2026-03-17 19:30:45",
  "level": "ERROR",
  "method": "POST",
  "path": "/user/login",
  "message": "用户登录失败: 用户名或密码错误",
  "stack_trace": "...",
  "extra": {
    "username": "test@example.com",
    "reason": "用户名或密码错误"
  }
}
```

### 文件删除失败
```json
{
  "timestamp": "2026-03-17 19:31:20",
  "level": "ERROR",
  "method": "DELETE",
  "path": "/user/file/delete",
  "user_identity": "user-123",
  "message": "删除文件失败: database error",
  "stack_trace": "...",
  "extra": {
    "file_identity": "file-456",
    "user_identity": "user-123"
  }
}
```

### 程序 Panic
```json
{
  "timestamp": "2026-03-17 19:32:10",
  "level": "PANIC",
  "method": "POST",
  "path": "/file/upload",
  "user_id": "user-789",
  "message": "Panic: runtime error: invalid memory address",
  "stack_trace": "...",
  "extra": {
    "remote_addr": "192.168.1.100:12345",
    "user_agent": "Mozilla/5.0..."
  }
}
```

## 如何使用

### 启动服务
```bash
cd D:/Go_Project/my_cloud_disk/core
go run core.go
```

### 查看日志
```bash
# 查看全部日志
cat ./logs/error.log

# 实时监控
tail -f ./logs/error.log

# 格式化查看（需要 jq）
cat ./logs/error.log | jq .

# 查看最新 10 条
tail -10 ./logs/error.log | jq .

# 筛选 ERROR 级别
cat ./logs/error.log | jq 'select(.level=="ERROR")'

# 筛选用户登录相关
cat ./logs/error.log | jq 'select(.path=="/user/login")'

# 筛选特定用户
cat ./logs/error.log | jq 'select(.user_identity=="user-123")'
```

## 测试建议

### 1. 测试用户登录错误
```bash
curl -X POST http://localhost:8888/user/login \
  -H "Content-Type: application/json" \
  -d '{"name":"wrong","password":"wrong"}'
```

### 2. 测试用户注册错误
```bash
# 先注册一个用户
curl -X POST http://localhost:8888/user/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test","email":"test@example.com","password":"123456","code":"123456"}'

# 再次注册相同用户（会失败）
curl -X POST http://localhost:8888/user/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test","email":"test@example.com","password":"123456","code":"123456"}'
```

### 3. 测试文件删除错误
```bash
# 删除不存在的文件
curl -X DELETE "http://localhost:8888/user/file/delete?identity=not-exist" \
  -H "Authorization: your-token"
```

### 4. 查看日志
```bash
# 查看最新的错误日志
tail -5 ./logs/error.log | jq .
```

## 后续可以添加日志的地方

### 还未添加日志的 Logic 文件
- `mail-code-send-register-logic.go` - 发送验证码
- `file-upload-multipart-logic.go` - 分片上传
- `user-file-move-logic.go` - 移动文件
- `user-file-name-update-logic.go` - 重命名文件
- `user-folder-create-logic.go` - 创建文件夹
- `share-file-save-logic.go` - 保存分享文件
- `refresh-token-logic.go` - 刷新 Token

### 添加方法
参考已有的实现，在每个 Logic 的错误处理中添加：

```go
import "cloud_disk/core/internal/logger"

func (l *YourLogic) YourMethod(req *types.Request) error {
    // 构建上下文
    ctx := context.WithValue(l.ctx, "method", "POST")
    ctx = context.WithValue(ctx, "path", "/your/path")

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

## 优势

### 1. 防止程序崩溃
- 任何 panic 都会被中间件捕获
- 服务继续运行，不影响其他用户

### 2. 完整的错误追踪
- 每个错误都有详细的上下文信息
- 包含完整的堆栈信息
- 方便快速定位问题

### 3. 便于问题排查
- JSON 格式，易于解析和搜索
- 可以按时间、用户、路径等维度筛选
- 支持实时监控

### 4. 生产环境就绪
- 日志文件持久化存储
- 不影响业务性能
- 可以随时查看历史日志

## 注意事项

1. **不要记录敏感信息**
   - 密码、Token 等不要记录到日志
   - 用户隐私信息要脱敏

2. **控制日志量**
   - 不要在循环中频繁记录日志
   - 相同错误可以考虑采样记录

3. **定期清理日志**
   - 日志文件会持续增长
   - 建议定期清理或实现日志轮转

4. **及时修复 bug**
   - 看到 PANIC 日志要立即修复代码
   - ERROR 日志也要定期检查

## 下一步改进

当日志量增大时，可以考虑：
1. **日志轮转** - 使用 lumberjack 库
2. **异步写入** - 使用 channel 提升性能
3. **集成 Elasticsearch** - 支持全文搜索
4. **添加监控告警** - 错误自动通知

详见：`docs/error-log-improvement-guide.md`
