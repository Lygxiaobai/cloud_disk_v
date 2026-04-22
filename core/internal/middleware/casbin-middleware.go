package middleware

import (
	appErrors "cloud_disk/core/internal/errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type CasbinMiddleware struct {
	enforcer *casbin.SyncedEnforcer
}

func NewCasbinMiddleware(enforcer *casbin.SyncedEnforcer) *CasbinMiddleware {
	return &CasbinMiddleware{enforcer: enforcer}
}

func (m *CasbinMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("UserRole")
		if role == "" {
			httpx.ErrorCtx(r.Context(), w, appErrors.ForbiddenError(r.Context(), "permission denied", nil, nil))
			return
		}

		ok, err := m.enforcer.Enforce(role, r.URL.Path, r.Method)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("casbin enforce failed: %v", err)
			httpx.ErrorCtx(r.Context(), w, appErrors.Internal(r.Context(), "internal server error", err, nil))
			return
		}
		if !ok {
			httpx.ErrorCtx(r.Context(), w, appErrors.ForbiddenError(r.Context(), "permission denied", nil, nil))
			return
		}

		next(w, r)
	}
}
