package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type ErrorRecoveryMiddleware struct{}

func NewErrorRecoveryMiddleware() *ErrorRecoveryMiddleware {
	return &ErrorRecoveryMiddleware{}
}

func (m *ErrorRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = helper.UUID()
		}

		r.Header.Set("X-Trace-Id", traceID)

		ctx := context.WithValue(r.Context(), "trace_id", traceID)
		ctx = context.WithValue(ctx, "method", r.Method)
		ctx = context.WithValue(ctx, "path", r.URL.Path)
		ctx = context.WithValue(ctx, "user_id", r.Header.Get("UserId"))
		r = r.WithContext(ctx)

		defer func() {
			if err := recover(); err != nil {
				logger.LogPanic(r.Context(), err, map[string]interface{}{
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.UserAgent(),
				})

				httpx.ErrorCtx(r.Context(), w, appErrors.Internal(r.Context(), "internal server error", nil, map[string]interface{}{
					"panic": err,
				}))
			}
		}()

		next(w, r)
	}
}
