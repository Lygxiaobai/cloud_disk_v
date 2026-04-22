package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestRateLimitMiddlewareReturnsStandardBody(t *testing.T) {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		return appErrors.ErrorResponse(ctx, err)
	})

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()

	middleware := NewRateLimitMiddleware(rdb, IPKey("/login"), 1, time.Minute)
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(rec1, req)
	require.Equal(t, http.StatusOK, rec1.Code)

	handler(rec2, req)
	require.Equal(t, http.StatusTooManyRequests, rec2.Code)
	require.NotEmpty(t, rec2.Header().Get("Retry-After"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &body))
	require.Equal(t, float64(20005), body["code"])
}
