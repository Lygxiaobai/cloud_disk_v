package errors

import (
	"cloud_disk/core/internal/logger"
	"context"
)

// LogicError 业务逻辑错误（自动记录日志）
type LogicError struct {
	message string
	err     error
	level   string
	ctx     context.Context
	code    int
	status  int
}

// Error 只返回用户可见的 message，内部的 err 仅写入日志（由 errors.New/Fatal 自动记录），
// 避免把数据库 SQL、内部调用栈等敏感信息通过 HTTP 响应泄露。
func (e *LogicError) Error() string {
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

	status, code := classify(message, err)

	return &LogicError{
		message: message,
		err:     err,
		level:   "ERROR",
		ctx:     ctx,
		code:    code,
		status:  status,
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

	status, code := classify(message, err)
	if status < 500 {
		status = 500
	}
	if code < 50000 {
		code = CodeInternalError
	}

	return &LogicError{
		message: message,
		err:     err,
		level:   "FATAL",
		ctx:     ctx,
		code:    code,
		status:  status,
	}
}

func (e *LogicError) StatusCode() int {
	if e == nil || e.status == 0 {
		return 500
	}
	return e.status
}

func (e *LogicError) BusinessCode() int {
	if e == nil || e.code == 0 {
		return CodeInternalError
	}
	return e.code
}
