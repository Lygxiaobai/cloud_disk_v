// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ShareBasicDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 这里支持批量撤销分享，前端会按 identity 数组提交。
		var req types.ShareBasicDeleteRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewShareBasicDeleteLogic(r.Context(), svcCtx)
		resp, err := l.ShareBasicDelete(&req, r.Header.Get("UserIdentity"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
