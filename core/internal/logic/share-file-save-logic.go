package logic

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareFileSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareFileSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareFileSaveLogic {
	return &ShareFileSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareFileSaveLogic) ShareFileSave(req *types.ShareFileSaveRequest, userIdentity string) (resp *types.ShareFileSaveResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "path", "/share/file/save")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//从公共池获取文件信息
	var rpData = &models.RepositoryPool{}
	has, err := l.svcCtx.Engine.Where("identity = ?", req.RepositoryIdentity).Get(rpData)
	if err != nil {
		logger.LogError(ctx, "查询文件失败", err, map[string]interface{}{
			"repository_identity": req.RepositoryIdentity,
		})
		return nil, err
	}
	if !has {
		err = errors.New("文件不存在")
		logger.LogError(ctx, "保存分享文件失败", err, map[string]interface{}{
			"repository_identity": req.RepositoryIdentity,
			"reason":              "文件不存在",
		})
		return nil, err
	}
	//存储到当前用户中
	upData := models.UserRepository{
		Identity:           helper.UUID(),
		UserIdentity:       userIdentity,
		ParentId:           req.ParentId,
		RepositoryIdentity: req.RepositoryIdentity,
		Name:               rpData.Name,
		Ext:                rpData.Ext,
		IsDir:              0,
	}
	_, err = l.svcCtx.Engine.Insert(upData)
	if err != nil {
		logger.LogError(ctx, "插入用户文件失败", err, map[string]interface{}{
			"repository_identity": req.RepositoryIdentity,
			"parent_id":           req.ParentId,
		})
		return nil, err
	}
	return
}
