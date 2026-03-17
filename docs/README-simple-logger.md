# 错误日志系统 - 简化版说明

## 📋 概述

我已经为你创建了一个**简化版错误日志系统**，专为新手设计，代码简洁易懂。

## 📁 文件清单

### 核心代码
```
core/internal/logger/
├── simple_logger.go              # 简化版日志记录器（~150行）
└── error_logger.go               # 复杂版日志记录器（保留）

core/internal/middleware/
├── simple-error-middleware.go    # 简化版中间件（~30行）
└── error-recovery-middleware.go  # 复杂版中间件（保留）

core/
├── test_simple_logger.go         # 完整测试程序
└── test_simple_demo.go           # 快速演示
```

### 文档
```
docs/
├── simple-error-log-guide.md         # 简化版使用指南
├── error-log-improvement-guide.md    # 后续改进建议
├── error-log-test-report.md          # 复杂版测试报告（参考）
└── rabbitmq-*.md                     # 复杂版文档（参考）
```

## 🎯 简化版 vs 复杂版

| 对比项 | 简化版 | 复杂版 |
|--------|--------|--------|
| **代码行数** | ~180 行 | ~970 行 |
| **外部依赖** | 0 个 | 3 个 (RabbitMQ, ES, lumberjack) |
| **学习难度** | ⭐ 新手友好 | ⭐⭐⭐⭐ 需要分布式知识 |
| **功能** | 本地文件日志 | 文件 + ES + MQ |
| **适用场景** | 单机应用、学习 | 生产环境、分布式系统 |

## 🚀 快速开始

### 1. 运行演示
```bash
cd D:/Go_Project/my_cloud_disk/core
go run test_simple_demo.go
```

### 2. 查看日志
```bash
cat ./logs/error.log
```

### 3. 在代码中使用

```go
package main

import (
    "cloud_disk/core/internal/logger"
    "context"
    "errors"
)

func main() {
    // 初始化
    logger.InitSimpleLogger("./logs/error.log")
    defer logger.Close()

    // 使用
    ctx := context.Background()
    ctx = context.WithValue(ctx, "trace_id", "trace-123")
    ctx = context.WithValue(ctx, "user_id", "user-456")

    err := errors.New("connection timeout")
    logger.LogError(ctx, "操作失败", err, map[string]interface{}{
        "file_name": "test.pdf",
    })
}
```

## 📖 核心 API

### InitSimpleLogger
初始化日志记录器
```go
func InitSimpleLogger(logFilePath string) error
```

### LogError
记录 Error 级别日志
```go
func LogError(ctx context.Context, message string, err error, extra map[string]interface{})
```

### LogFatal
记录 Fatal 级别日志（严重错误）
```go
func LogFatal(ctx context.Context, message string, err error, extra map[string]interface{})
```

### LogPanic
记录 Panic 级别日志（程序崩溃）
```go
func LogPanic(ctx context.Context, panicValue interface{}, extra map[string]interface{})
```

### Close
关闭日志记录器
```go
func Close() error
```

## 📝 日志格式

每条日志以 JSON 格式保存：

```json
{
  "timestamp": "2026-03-17 10:30:45",
  "level": "ERROR",
  "trace_id": "trace-123",
  "user_id": "user-456",
  "method": "POST",
  "path": "/file/upload",
  "message": "操作失败: connection timeout",
  "stack_trace": "goroutine 1 [running]:\n...",
  "extra": {
    "file_name": "test.pdf"
  }
}
```

## 🎓 学习路径

### 第 1 周：理解简化版
1. 阅读 `simple_logger.go`（~150行）
2. 运行 `test_simple_demo.go`
3. 在自己的项目中集成

### 第 2-4 周：逐步改进
参考 `error-log-improvement-guide.md`：
1. 添加日志轮转（防止文件过大）
2. 实现异步写入（提升性能）
3. 学习 RabbitMQ（消息队列）
4. 学习 Elasticsearch（日志搜索）

### 第 5-8 周：生产环境
1. 添加监控告警
2. 高可用部署
3. 性能优化

## 🔧 后续改进

详见 `docs/error-log-improvement-guide.md`，包括：

1. **日志轮转** - 使用 lumberjack 自动管理日志文件
2. **异步写入** - 使用 channel 提升性能
3. **消息队列** - 使用 RabbitMQ 解耦
4. **Elasticsearch** - 支持日志搜索和分析
5. **监控告警** - 集成 Prometheus

## ⚠️ 注意事项

### 简化版的限制
1. ❌ 日志文件会无限增长（需要手动清理）
2. ❌ 同步写入，高并发下可能影响性能
3. ❌ 不支持日志搜索
4. ❌ 单机存储，不适合分布式系统

### 何时升级到复杂版？
- 日志量大（每天 > 10GB）
- 需要搜索和分析日志
- 分布式系统
- 高可用要求

## 📚 相关文档

- **使用指南**: `docs/simple-error-log-guide.md`
- **改进建议**: `docs/error-log-improvement-guide.md`
- **复杂版参考**: `docs/error-log-test-report.md`

## 💡 常见问题

### Q: 日志文件在哪里？
A: 默认在 `./logs/error.log`

### Q: 如何查看日志？
A:
```bash
# 查看全部
cat ./logs/error.log

# 实时监控
tail -f ./logs/error.log

# 格式化查看（需要 jq）
cat ./logs/error.log | jq .
```

### Q: 日志文件会无限增长吗？
A: 简化版不会自动轮转。参考改进指南第一阶段添加日志轮转功能。

### Q: 性能如何？
A: 简化版是同步写入，适合中小型应用。如需高性能，参考改进指南第二阶段。

## 🎉 总结

简化版特点：
- ✅ 代码简洁（~180行）
- ✅ 零依赖
- ✅ 易于理解
- ✅ 快速上手

适合：
- 🎓 学习 Go 语言
- 🔰 新手入门
- 🏠 单机应用
- 🧪 快速原型

---

**版本**: v1.0
**日期**: 2026-03-17
**作者**: Claude
**状态**: ✅ 已测试
