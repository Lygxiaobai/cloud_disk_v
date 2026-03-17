package middleware

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"context"
	"net/http"
)

type ErrorRecoveryMiddleware struct {
}

func NewErrorRecoveryMiddleware() *ErrorRecoveryMiddleware {
	return &ErrorRecoveryMiddleware{}
}

func (m *ErrorRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 生成或获取 TraceID
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = helper.UUID() // 自动生成 TraceID
		}

		// 将 TraceID 设置到 Header 中
		r.Header.Set("X-Trace-Id", traceID)

		// 将 TraceID 存入 Context，供后续 Logic 使用
		ctx := context.WithValue(r.Context(), "trace_id", traceID)
		r = r.WithContext(ctx)

		defer func() {
			if err := recover(); err != nil {
				// 构建上下文信息
				panicCtx := context.WithValue(r.Context(), "method", r.Method)
				panicCtx = context.WithValue(panicCtx, "path", r.URL.Path)
				panicCtx = context.WithValue(panicCtx, "user_id", r.Header.Get("UserId"))
				panicCtx = context.WithValue(panicCtx, "trace_id", traceID)

				// 记录 panic 日志
				logger.LogPanic(panicCtx, err, map[string]interface{}{
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.UserAgent(),
				})

				// 返回 500 错误
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("服务器内部错误"))
			}
		}()

		next(w, r)
	}
}
