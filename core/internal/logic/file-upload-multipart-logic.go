// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
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
		return nil, err
	}

	resp = &types.FileUploadMultipartResponse{
		Identity: rp.Identity,
	}
	if has {
		return resp, nil
	}

	filePath, err := helper.FileUploadMultipart(req.Name, fileBuf)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &types.FileUploadMultipartResponse{
		Identity: rp.Identity,
	}, nil
}
