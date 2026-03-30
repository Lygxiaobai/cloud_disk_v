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
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "启动分片上传事务失败", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback multipart upload failed: %v", rbErr)
		}
	}()

	rp := models.RepositoryPool{}
	has, err := sess.SQL("SELECT * FROM repository_pool WHERE hash = ? LIMIT 1 FOR UPDATE", req.Hash).Get(&rp)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件哈希失败", err, map[string]interface{}{
			"hash": req.Hash,
		})
	}

	resp = &types.FileUploadMultipartResponse{Identity: rp.Identity}
	if has {
		if err := sess.Commit(); err != nil {
			return nil, errors.New(l.ctx, "提交分片上传命中事务失败", err, nil)
		}
		committed = true
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
	_, err = sess.Insert(&rp)
	if err != nil {
		return nil, errors.New(l.ctx, "插入文件记录失败", err, map[string]interface{}{
			"file_name": req.Name,
			"file_hash": req.Hash,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "提交分片上传事务失败", err, nil)
	}
	committed = true

	return &types.FileUploadMultipartResponse{Identity: rp.Identity}, nil
}
