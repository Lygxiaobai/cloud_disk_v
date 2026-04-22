package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	accessSecret string
}

func NewAuthMiddleware(accessSecret string) *AuthMiddleware {
	return &AuthMiddleware{accessSecret: accessSecret}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			httpx.ErrorCtx(r.Context(), w, appErrors.Unauthorized(r.Context(), "identity or authorization is required", nil, nil))
			return
		}

		uc, err := helper.AnalyzeToken(token, m.accessSecret)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("auth token parse failed: %v", err)
			httpx.ErrorCtx(r.Context(), w, appErrors.AuthFailed(r.Context(), "authentication failed", err, nil))
			return
		}

		r.Header.Set("UserId", strconv.Itoa(uc.ID))
		r.Header.Set("UserIdentity", uc.Identity)
		r.Header.Set("UserName", uc.Name)
		r.Header.Set("UserRole", uc.Role)

		next(w, r)
	}
}
