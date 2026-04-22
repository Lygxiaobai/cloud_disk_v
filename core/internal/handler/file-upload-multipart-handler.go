package handler

import (
	"net/http"
	"path"

	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FileUploadMultipartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if maxSize := svcCtx.Config.Upload.MaxSize; maxSize > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		}

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

		ext := path.Ext(header.Filename)
		if err := helper.ValidateUploadExt(ext, svcCtx.Config.Upload.BlockedExtensions); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		fileHash, err := helper.HashAndReset(file, svcCtx.Config.Upload.MaxSize)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		req.Name = header.Filename
		req.Ext = ext
		req.Size = header.Size
		req.Hash = fileHash

		l := logic.NewFileUploadMultipartLogic(r.Context(), svcCtx)
		resp, err := l.FileUploadMultipart(&req, file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
