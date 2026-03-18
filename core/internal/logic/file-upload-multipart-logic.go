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

type FileUploadMultipartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadMultipartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadMultipartLogic {
	return &FileUploadMultipartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileUploadMultipartLogic) FileUploadMultipart(req *types.FileUploadMultipartRequest, fileBuf []byte) (resp *types.FileUploadMultipartResponse, err error) {
	rp := models.RepositoryPool{}
	has, err := l.svcCtx.Engine.Where("hash =?", req.Hash).Get(&rp)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件哈希失败", err, map[string]interface{}{
			"hash": req.Hash,
		})
	}

	resp = &types.FileUploadMultipartResponse{
		Identity: rp.Identity,
	}
	if has {
		return resp, nil
	}

	filePath, err := helper.FileUploadMultipart(req.Name, fileBuf)
	if err != nil {
		return nil, errors.New(l.ctx, "分片上传失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_size": req.Size,
		})
	}

	rp = models.RepositoryPool{
		Identity: helper.UUID(),
		Hash:     req.Hash,
		Name:     req.Name,
		Ext:      req.Ext,
		Size:     req.Size,
		Path:     filePath,
	}
	_, err = l.svcCtx.Engine.Insert(&rp)
	if err != nil {
		return nil, errors.New(l.ctx, "插入文件记录失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_hash": req.Hash,
		})
	}

	return &types.FileUploadMultipartResponse{
		Identity: rp.Identity,
	}, nil
}
