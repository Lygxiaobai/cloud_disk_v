// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadLogic {
	return &FileUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileUploadLogic) FileUpload(req *types.FileUploadRequest) (resp *types.FileUploadResponse, err error) {
	//将上传的文件信息存入数据库
	rp := models.RepositoryPool{
		Identity: helper.UUID(),
		Name:     req.Name,
		Ext:      req.Ext,
		Size:     req.Size,
		Path:     req.Path,
		Hash:     req.Hash,
	}
	_, err = l.svcCtx.Engine.InsertOne(&rp)
	if err != nil {
		// 记录错误日志
		return nil, errors.New(l.ctx, "文件上传失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_size": req.Size,
			"file_hash": req.Hash,
		})
	}
	return &types.FileUploadResponse{
		Identity: rp.Identity,
		Ext:      rp.Ext,
		Name:     rp.Name,
	}, nil
}
