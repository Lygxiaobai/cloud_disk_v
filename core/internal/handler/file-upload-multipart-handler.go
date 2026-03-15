// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"path"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FileUploadMultipartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileUploadMultipartRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		fileBuf, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		req.Name = header.Filename
		req.Ext = path.Ext(header.Filename)
		req.Size = header.Size
		req.Hash = fmt.Sprintf("%x", md5.Sum(fileBuf))

		l := logic.NewFileUploadMultipartLogic(r.Context(), svcCtx)
		resp, err := l.FileUploadMultipart(&req, fileBuf)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
