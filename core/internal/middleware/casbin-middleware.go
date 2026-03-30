package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
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
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden: role missing"))
			return
		}

		ok, err := m.enforcer.Enforce(role, r.URL.Path, r.Method)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
			return
		}

		next(w, r)
	}
}
