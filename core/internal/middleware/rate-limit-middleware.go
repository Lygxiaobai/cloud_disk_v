package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type RateLimitKeyFunc func(r *http.Request) string

type RateLimitMiddleware struct {
	rdb    *redis.Client
	keyFn  RateLimitKeyFunc
	max    int
	window time.Duration
}

func NewRateLimitMiddleware(rdb *redis.Client, keyFn RateLimitKeyFunc, max int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{rdb: rdb, keyFn: keyFn, max: max, window: window}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := m.keyFn(r)
		if key == "" || m.rdb == nil {
			next(w, r)
			return
		}

		ctx := r.Context()
		count, err := m.rdb.Incr(ctx, key).Result()
		if err != nil {
			logx.WithContext(ctx).Errorf("rate-limit incr failed: key=%s err=%v", key, err)
			next(w, r)
			return
		}
		if count == 1 {
			if err := m.rdb.Expire(ctx, key, m.window).Err(); err != nil {
				logx.WithContext(ctx).Errorf("rate-limit expire failed: key=%s err=%v", key, err)
			}
		}
		if count > int64(m.max) {
			ttl, _ := m.rdb.TTL(ctx, key).Result()
			if ttl > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))
			}
			httpx.ErrorCtx(ctx, w, appErrors.RateLimited(ctx, "too many requests", nil, map[string]interface{}{
				"retry_after_seconds": int(ttl.Seconds()),
			}))
			return
		}

		next(w, r)
	}
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func IPKey(path string) RateLimitKeyFunc {
	return func(r *http.Request) string {
		ip := ClientIP(r)
		if ip == "" {
			return ""
		}
		return "ratelimit:ip:" + path + ":" + ip
	}
}

func EmailKey(path string) RateLimitKeyFunc {
	return func(r *http.Request) string {
		if r.Body == nil {
			return ""
		}
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			return ""
		}
		r.Body = io.NopCloser(strings.NewReader(string(buf)))

		var payload struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return ""
		}
		email := strings.TrimSpace(strings.ToLower(payload.Email))
		if email == "" {
			return ""
		}
		return "ratelimit:email:" + path + ":" + email
	}
}

func ComposeKey(fns ...RateLimitKeyFunc) RateLimitKeyFunc {
	return func(r *http.Request) string {
		for _, f := range fns {
			if k := f(r); k != "" {
				return k
			}
		}
		return ""
	}
}

var _ = context.TODO
