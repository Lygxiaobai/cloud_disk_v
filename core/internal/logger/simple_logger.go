package logger

import (
	"cloud_disk/core/internal/rabbitmq"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// SimpleLogger 简化版错误日志记录器
type SimpleLogger struct {
	mu          sync.Mutex
	logFile     *os.File
	logProducer *rabbitmq.LogProducer
	useAsync    bool
}

// ErrorLog 错误日志结构（V4 完整版）
type ErrorLog struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	TraceID    string                 `json:"trace_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Message    string                 `json:"message"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// 全局唯一 logger，通过 atomic.Pointer 保证并发读写安全
var globalLogger atomic.Pointer[SimpleLogger]

// InitSimpleLogger 初始化日志记录器（同步模式，写入本地文件）
func InitSimpleLogger(logFilePath string) error {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	globalLogger.Store(&SimpleLogger{
		logFile:  file,
		useAsync: false,
	})

	log.Printf("日志记录器初始化成功（同步模式）: %s", logFilePath)
	return nil
}

// InitAsyncLogger 初始化异步日志记录器（使用 RabbitMQ）
func InitAsyncLogger(logFilePath string, logProducer *rabbitmq.LogProducer) error {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	globalLogger.Store(&SimpleLogger{
		logFile:     file,
		logProducer: logProducer,
		useAsync:    true,
	})

	log.Printf("日志记录器初始化成功（异步模式 - RabbitMQ）: %s", logFilePath)
	return nil
}

// LogError 记录错误日志
func LogError(ctx context.Context, message string, err error, extra map[string]interface{}) {
	logger := globalLogger.Load()
	if logger == nil {
		log.Printf("日志记录器未初始化: %s - %v", message, err)
		return
	}

	errorLog := &ErrorLog{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      "ERROR",
		TraceID:    getStringFromContext(ctx, "trace_id"),
		UserID:     getStringFromContext(ctx, "user_id"),
		Method:     getStringFromContext(ctx, "method"),
		Path:       getStringFromContext(ctx, "path"),
		Message:    fmt.Sprintf("%s: %v", message, err),
		StackTrace: string(debug.Stack()),
		Extra:      extra,
	}

	if logger.useAsync && logger.logProducer != nil {
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

		sendErr := logger.logProducer.SendLogMessage(logMsg)
		if sendErr != nil {
			log.Printf("发送日志到 MQ 失败，降级到本地文件: %v", sendErr)
			logger.writeToLocalFile(errorLog)
		}
	} else {
		logger.writeToLocalFile(errorLog)
	}

	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

func (s *SimpleLogger) writeToLocalFile(errorLog *ErrorLog) {
	if s == nil || s.logFile == nil {
		return
	}

	jsonData, jsonErr := json.Marshal(errorLog)
	if jsonErr != nil {
		log.Printf("序列化日志失败: %v", jsonErr)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.logFile.Write(append(jsonData, '\n'))
}

// LogFatal 记录致命错误日志
func LogFatal(ctx context.Context, message string, err error, extra map[string]interface{}) {
	logger := globalLogger.Load()
	if logger == nil {
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

	if logger.useAsync && logger.logProducer != nil {
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
		if sendErr := logger.logProducer.SendLogMessage(logMsg); sendErr != nil {
			log.Printf("发送致命日志到 MQ 失败，降级到本地文件: %v", sendErr)
			logger.writeToLocalFile(errorLog)
		}
	} else {
		logger.writeToLocalFile(errorLog)
	}

	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

// LogPanic 记录 panic 日志
func LogPanic(ctx context.Context, panicValue interface{}, extra map[string]interface{}) {
	logger := globalLogger.Load()
	if logger == nil {
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

	if logger.useAsync && logger.logProducer != nil {
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
		if sendErr := logger.logProducer.SendLogMessage(logMsg); sendErr != nil {
			log.Printf("发送 panic 日志到 MQ 失败，降级到本地文件: %v", sendErr)
			logger.writeToLocalFile(errorLog)
		}
	} else {
		logger.writeToLocalFile(errorLog)
	}

	log.Printf("[%s] TraceID=%s - %s", errorLog.Level, errorLog.TraceID, errorLog.Message)
}

// getStringFromContext 从 context 中获取字符串值
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
	logger := globalLogger.Load()
	if logger == nil || logger.logFile == nil {
		return nil
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	return logger.logFile.Close()
}
