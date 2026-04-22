// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	appErrors "cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserDetailRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if req.Identity == "" {
			token := r.Header.Get("Authorization")
			if token == "" {
				httpx.ErrorCtx(r.Context(), w, appErrors.New(r.Context(), "identity or authorization is required", nil, nil))
				return
			}

			uc, err := helper.AnalyzeToken(token, svcCtx.Config.JWT.AccessSecret)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			req.Identity = uc.Identity
		}

		l := logic.NewUserDetailLogic(r.Context(), svcCtx)
		resp, err := l.UserDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
