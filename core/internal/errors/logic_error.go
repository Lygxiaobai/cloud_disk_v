package errors

import (
	"cloud_disk/core/internal/logger"
	"context"
	"fmt"
)

// LogicError 业务逻辑错误（自动记录日志）
type LogicError struct {
	message string
	err     error
	level   string
	ctx     context.Context
}

func (e *LogicError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

// ==================== ERROR 级别 ====================

// New 创建 ERROR 级别错误（自动记录日志）
// 用于普通业务错误：用户不存在、参数错误、数据库查询失败等
//
// 参数：
//   - ctx: 上下文（包含 trace_id, method, path, user_id）
//   - message: 错误消息
//   - err: 原始错误（可以为 nil）
//   - extra: 额外信息（可以为 nil）
//
// 示例：
//
//	errors.New(ctx, "数据库查询失败", err, map[string]interface{}{"username": "test"})
//	errors.New(ctx, "用户不存在", nil, map[string]interface{}{"username": "test"})
//	errors.New(ctx, "参数错误", nil, nil)
//	errors.New(ctx, fmt.Sprintf("用户 %s 不存在", name), nil, nil)
func New(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	// 自动记录 ERROR 日志
	logger.LogError(ctx, message, err, extra)

	return &LogicError{
		message: message,
		err:     err,
		level:   "ERROR",
		ctx:     ctx,
	}
}

// ==================== FATAL 级别 ====================

// Fatal 创建 FATAL 级别错误（自动记录日志）
// 用于致命错误：数据库连接失败、Redis连接失败、配置加载失败等
//
// 参数：
//   - ctx: 上下文（包含 trace_id, method, path, user_id）
//   - message: 错误消息
//   - err: 原始错误（可以为 nil）
//   - extra: 额外信息（可以为 nil）
//
// 示例：
//
//	errors.Fatal(ctx, "数据库连接失败", err, nil)
//	errors.Fatal(ctx, "Redis连接失败", err, nil)
//	errors.Fatal(ctx, fmt.Sprintf("配置文件 %s 加载失败", path), err, nil)
func Fatal(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	// 自动记录 FATAL 日志
	logger.LogFatal(ctx, message, err, extra)

	return &LogicError{
		message: message,
		err:     err,
		level:   "FATAL",
		ctx:     ctx,
	}
}
