package handler

import (
	"net/http"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserRecycleRestoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserRecycleRestoreRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserRecycleRestoreLogic(r.Context(), svcCtx)
		resp, err := l.UserRecycleRestore(&req, r.Header.Get("UserIdentity"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
