// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"cloud_disk/core/internal/helper"
	"net/http"
	"strconv"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//1.从Header中回去用户token信息
		token := r.Header.Get("Authorization")
		//2.没有信息 返回未授权
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		//3.解析token
		uc, err := helper.AnalyzeToken(token)
		//4.解析失败
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		//4.解析成功 在Header中设置用户的Identity，Name,Id
		r.Header.Set("UserId", strconv.Itoa(uc.ID))
		r.Header.Set("UserIdentity", uc.Identity)
		r.Header.Set("UserName", uc.Name)
		r.Header.Set("UserRole", uc.Role)

		next(w, r)
	}
}
