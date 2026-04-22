package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud_disk/core/internal/helper"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		return appErrors.ErrorResponse(ctx, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware("secret")
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, float64(20001), body["code"])
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		return appErrors.ErrorResponse(ctx, err)
	})

	token, err := helper.GenerateToken(1, "u-1", "tester", "user", "secret", 3600)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware("secret")
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "u-1", r.Header.Get("UserIdentity"))
		w.WriteHeader(http.StatusOK)
	})
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
