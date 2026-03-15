// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"cloud_disk/core/internal/helper"
	"errors"
	"net/http"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserFolderChildrenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserFolderChildrenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		token := r.Header.Get("Authorization")
		if token == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("identity or authorization is required"))
			return
		}
		uc, err := helper.AnalyzeToken(token)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
		identity := uc.Identity
		l := logic.NewUserFolderChildrenLogic(r.Context(), svcCtx)
		resp, err := l.UserFolderChildren(&req, identity)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
