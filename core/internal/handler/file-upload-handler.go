package handler

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"net/http"
	"path"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FileUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if maxSize := svcCtx.Config.Upload.MaxSize; maxSize > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		}

		var req types.FileUploadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		ext := path.Ext(fileHeader.Filename)
		if err := helper.ValidateUploadExt(ext, svcCtx.Config.Upload.BlockedExtensions); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		fileHash, err := helper.HashAndReset(file, svcCtx.Config.Upload.MaxSize)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		rp := new(models.RepositoryPool)
		has, err := svcCtx.Engine.Where("hash = ? AND size = ?", fileHash, fileHeader.Size).Get(rp)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if has {
			httpx.OkJson(w, &types.FileUploadResponse{
				Identity: rp.Identity,
				Ext:      rp.Ext,
				Name:     rp.Name,
			})
			return
		}

		fileOSSPath, err := helper.FileUpload(fileHeader.Filename, file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		req.Name = fileHeader.Filename
		req.Ext = ext
		req.Size = fileHeader.Size
		req.Path = fileOSSPath
		req.Hash = fileHash

		l := logic.NewFileUploadLogic(r.Context(), svcCtx)
		resp, err := l.FileUpload(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
