// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileNameUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileNameUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileNameUpdateLogic {
	return &UserFileNameUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileNameUpdateLogic) UserFileNameUpdate(req *types.UserFileNameUpdateRequest, userIdentity string) (resp *types.UserFileNameUpdateResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "PUT")
	ctx = context.WithValue(ctx, "path", "/user/file/name/update")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//当当前目录存在与修改名字相同的文件时，不允许修改
	count, err := l.svcCtx.Engine.Where("name=? AND parent_id=(select parent_id from user_repository where user_identity = ? and identity =?)", req.Name, userIdentity, req.Identity).Count(&models.UserRepository{})
	if err != nil {
		logger.LogError(ctx, "查询文件名失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
		})
		return nil, err
	}
	if count > 0 {
		err = errors.New("该文件名已经存在")
		logger.LogError(ctx, "重命名文件失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
			"reason":        "文件名已存在",
		})
		return nil, err
	}

	up := models.UserRepository{
		Name: req.Name,
	}
	_, err = l.svcCtx.Engine.Where("identity =? AND user_identity =?", req.Identity, userIdentity).Update(&up)
	if err != nil {
		logger.LogError(ctx, "更新文件名失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
		})
		return nil, err
	}

	return
}
