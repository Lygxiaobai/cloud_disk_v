package logger

import (
	"cloud_disk/core/internal/rabbitmq"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

// SimpleLogger 简化版错误日志记录器
type SimpleLogger struct {
	logFile     *os.File              // 日志文件句柄（保留用于降级）
	logProducer *rabbitmq.LogProducer // RabbitMQ 日志生产者
	useAsync    bool                  // 是否使用异步模式
}

// ErrorLog 错误日志结构（V4 完整版）
type ErrorLog struct {
	Timestamp  string                 `json:"timestamp"`             // 时间戳
	Level      string                 `json:"level"`                 // 日志级别
	TraceID    string                 `json:"trace_id,omitempty"`    // 追踪ID
	UserID     string                 `json:"user_id,omitempty"`     // 用户ID
	Method     string                 `json:"method,omitempty"`      // HTTP方法
	Path       string                 `json:"path,omitempty"`        // 请求路径
	Message    string                 `json:"message"`               // 错误消息
	StackTrace string                 `json:"stack_trace,omitempty"` // 堆栈信息（V4 新增）
	Extra      map[string]interface{} `json:"extra,omitempty"`       // 额外字段（V4 新增）
}

// 全局变量，保存唯一的 logger 实例
var globalLogger *SimpleLogger

// InitSimpleLogger 初始化日志记录器（同步模式，写入本地文件）
func InitSimpleLogger(logFilePath string) error {
	// 1. 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 2. 打开日志文件（追加模式）
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 3. 保存到全局变量（同步模式）
	globalLogger = &SimpleLogger{
		logFile:  file,
		useAsync: false,
	}

	log.Printf("日志记录器初始化成功（同步模式）: %s", logFilePath)
	return nil
}

// InitAsyncLogger 初始化异步日志记录器（使用 RabbitMQ）
func InitAsyncLogger(logFilePath string, logProducer *rabbitmq.LogProducer) error {
	// 1. 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 2. 打开日志文件（用于降级）
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 3. 保存到全局变量（异步模式）
	globalLogger = &SimpleLogger{
		logFile:     file,
		logProducer: logProducer,
		useAsync:    true,
	}

	log.Printf("日志记录器初始化成功（异步模式 - RabbitMQ）: %s", logFilePath)
	return nil
}

// LogError 记录错误日志（V4 - 完整版，支持异步）
func LogError(ctx context.Context, message string, err error, extra map[string]interface{}) {
	// 检查是否初始化
	if globalLogger == nil {
		log.Printf("日志记录器未初始化: %s - %v", message, err)
		return
	}

	// 1. 构建日志结构体
	errorLog := &ErrorLog{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"), // 当前时间
		Level:      "ERROR",                                  // 日志级别
		TraceID:    getStringFromContext(ctx, "trace_id"),    // 从 context 提取 TraceID
		UserID:     getStringFromContext(ctx, "user_id"),     // 从 context 提取 UserID
		Method:     getStringFromContext(ctx, "method"),      // 从 context 提取 Method
		Path:       getStringFromContext(ctx, "path"),        // 从 context 提取 Path
		Message:    fmt.Sprintf("%s: %v", message, err),      // 拼接消息和错误
		StackTrace: string(debug.Stack()),                    // 获取堆栈信息（V4 新增）
		Extra:      extra,                                    // 额外字段（V4 新增）
	}

	// 2. 判断使用异步还是同步模式
	if globalLogger.useAsync && globalLogger.logProducer != nil {
		// 异步模式：发送到 RabbitMQ
		logMsg := &rabbitmq.LogMessage{
			Timestamp:  errorLog.Timestamp,
			Level:      errorLog.Level,
			TraceID:    errorLog.TraceID,
			UserID:     errorLog.UserID,
			Method:     errorLog.Method,
			Path:       errorLog.Path,
			Message:    errorLog.Message,
			StackTrace: errorLog.StackTrace,
			Extra:      errorLog.Extra,
		}

		err := globalLogger.logProducer.SendLogMessage(logMsg)
		if err != nil {
			// MQ 发送失败，降级到本地文件
			log.Printf("发送日志到 MQ 失败，降级到本地文件: %v", err)
			writeToLocalFile(errorLog)
		}
	} else {
		// 同步模式：直接写入本地文件
		writeToLocalFile(errorLog)
	}

	// 3. 同时输出到控制台
	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

// writeToLocalFile 写入本地文件（内部函数）
func writeToLocalFile(errorLog *ErrorLog) {
	if globalLogger.logFile == nil {
		return
	}

	jsonData, jsonErr := json.Marshal(errorLog)
	if jsonErr != nil {
		log.Printf("序列化日志失败: %v", jsonErr)
		return
	}

	globalLogger.logFile.Write(append(jsonData, '\n'))
}

// LogFatal 记录致命错误日志（V4 新增，支持异步）
func LogFatal(ctx context.Context, message string, err error, extra map[string]interface{}) {
	if globalLogger == nil {
		log.Printf("日志记录器未初始化: %s - %v", message, err)
		return
	}

	errorLog := &ErrorLog{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      "FATAL",
		TraceID:    getStringFromContext(ctx, "trace_id"),
		UserID:     getStringFromContext(ctx, "user_id"),
		Method:     getStringFromContext(ctx, "method"),
		Path:       getStringFromContext(ctx, "path"),
		Message:    fmt.Sprintf("%s: %v", message, err),
		StackTrace: string(debug.Stack()),
		Extra:      extra,
	}

	if globalLogger.useAsync && globalLogger.logProducer != nil {
		logMsg := &rabbitmq.LogMessage{
			Timestamp:  errorLog.Timestamp,
			Level:      errorLog.Level,
			TraceID:    errorLog.TraceID,
			UserID:     errorLog.UserID,
			Method:     errorLog.Method,
			Path:       errorLog.Path,
			Message:    errorLog.Message,
			StackTrace: errorLog.StackTrace,
			Extra:      errorLog.Extra,
		}
		globalLogger.logProducer.SendLogMessage(logMsg)
	} else {
		writeToLocalFile(errorLog)
	}

	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

// LogPanic 记录 panic 日志（V4 新增，支持异步）
func LogPanic(ctx context.Context, panicValue interface{}, extra map[string]interface{}) {
	if globalLogger == nil {
		log.Printf("日志记录器未初始化: panic=%v", panicValue)
		return
	}

	errorLog := &ErrorLog{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      "PANIC",
		TraceID:    getStringFromContext(ctx, "trace_id"),
		UserID:     getStringFromContext(ctx, "user_id"),
		Method:     getStringFromContext(ctx, "method"),
		Path:       getStringFromContext(ctx, "path"),
		Message:    fmt.Sprintf("Panic: %v", panicValue),
		StackTrace: string(debug.Stack()),
		Extra:      extra,
	}

	if globalLogger.useAsync && globalLogger.logProducer != nil {
		logMsg := &rabbitmq.LogMessage{
			Timestamp:  errorLog.Timestamp,
			Level:      errorLog.Level,
			TraceID:    errorLog.TraceID,
			UserID:     errorLog.UserID,
			Method:     errorLog.Method,
			Path:       errorLog.Path,
			Message:    errorLog.Message,
			StackTrace: errorLog.StackTrace,
			Extra:      errorLog.Extra,
		}
		globalLogger.logProducer.SendLogMessage(logMsg)
	} else {
		writeToLocalFile(errorLog)
	}

	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

// getStringFromContext 从 context 中获取字符串值（辅助函数）
func getStringFromContext(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}

// Close 关闭日志记录器
func Close() error {
	if globalLogger != nil && globalLogger.logFile != nil {
		return globalLogger.logFile.Close()
	}
	return nil
}
