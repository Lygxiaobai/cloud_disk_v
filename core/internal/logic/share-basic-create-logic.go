// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareBasicCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareBasicCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareBasicCreateLogic {
	return &ShareBasicCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareBasicCreateLogic) ShareBasicCreate(req *types.ShareBasicCreateRequest, userIdentity string) (resp *types.ShareBasicCreateResponse, err error) {
	// 从 context 中获取 TraceID（由中间件设置）
	traceID, _ := l.ctx.Value("trace_id").(string)

	// 构建上下文信息
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "path", "/share/basic/create")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//根据用户传过来的Identity查找到该记录
	upData := &models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity = ?", req.UserRepositoryIdentity).Get(upData)
	if err != nil {
		logger.LogError(ctx, "查询文件失败", err, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
		})
		return nil, err

	}
	if !has {
		err = errors.New("不存在该文件！")
		logger.LogError(ctx, "创建分享失败", err, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
			"reason":              "文件不存在",
		})
		return nil, err
	}

	//向share_basic表中插入数据
	var sb = models.ShareBasic{
		Identity:               helper.UUID(),
		UserIdentity:           userIdentity,
		UserRepositoryIdentity: upData.Identity,
		RepositoryIdentity:     upData.RepositoryIdentity,
		ExpiredTime:            req.ExpiredTime,
		ClickNum:               define.DefaultClickNum,
	}
	_, err = l.svcCtx.Engine.Insert(&sb)
	if err != nil {
		logger.LogError(ctx, "插入分享记录失败", err, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
			"expired_time":        req.ExpiredTime,
		})
		return nil, err
	}
	resp = &types.ShareBasicCreateResponse{
		Identity: sb.Identity,
	}
	return
}
