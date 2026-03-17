// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
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
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "path", "/file/upload/multipart")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	rp := models.RepositoryPool{}
	has, err := l.svcCtx.Engine.Where("hash =?", req.Hash).Get(&rp)
	if err != nil {
		logger.LogError(ctx, "查询文件哈希失败", err, map[string]interface{}{
			"hash": req.Hash,
		})
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
		logger.LogError(ctx, "分片上传失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_size": req.Size,
		})
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
		logger.LogError(ctx, "插入文件记录失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_hash": req.Hash,
		})
		return nil, err
	}

	return &types.FileUploadMultipartResponse{
		Identity: rp.Identity,
	}, nil
}
