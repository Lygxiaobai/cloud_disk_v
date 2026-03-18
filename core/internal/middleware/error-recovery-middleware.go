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

		// 构建完整的 Context（包含所有日志需要的信息）
		ctx := context.WithValue(r.Context(), "trace_id", traceID)
		ctx = context.WithValue(ctx, "method", r.Method)
		ctx = context.WithValue(ctx, "path", r.URL.Path)
		ctx = context.WithValue(ctx, "user_id", r.Header.Get("UserId"))
		r = r.WithContext(ctx)

		defer func() {
			if err := recover(); err != nil {
				// 直接使用 r.Context()，不需要重新构建
				logger.LogPanic(r.Context(), err, map[string]interface{}{
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
