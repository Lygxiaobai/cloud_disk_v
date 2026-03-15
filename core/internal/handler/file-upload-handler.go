// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"crypto/md5"
	"fmt"
	"net/http"
	"path"

	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FileUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FileUploadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			return
		}
		b := make([]byte, fileHeader.Size)
		// Read可能会有bug
		if _, err := file.Read(b); err != nil {
			return
		}
		//计算文件的hash值
		fH := fmt.Sprintf("%x", md5.Sum(b))
		//1.判断文件是否已经上传到oss
		rp := new(models.RepositoryPool)
		has, err := svcCtx.Engine.Where("hash=?", fH).Get(rp)
		if err != nil {
			return
		}
		//2.1若有则直接返回
		if has {
			httpx.OkJson(w, &types.FileUploadResponse{
				Identity: rp.Identity,
				Ext:      rp.Ext,
				Name:     rp.Name,
			})
			return
		}
		//2.1.2若无则上传 在req设置值传给logic使用
		fileOSSPath, err := helper.FileUpload(r)

		req.Name = fileHeader.Filename
		req.Ext = path.Ext(fileHeader.Filename)
		req.Size = fileHeader.Size
		req.Path = fileOSSPath
		req.Hash = fH

		l := logic.NewFileUploadLogic(r.Context(), svcCtx)
		resp, err := l.FileUpload(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
