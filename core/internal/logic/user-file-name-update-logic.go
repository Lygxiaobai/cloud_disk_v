// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"

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
	//当当前目录存在与修改名字相同的文件时，不允许修改
	count, err := l.svcCtx.Engine.Where("name=? AND parent_id=(select parent_id from user_repository where user_identity = ? and identity =?)", req.Name, userIdentity, req.Identity).Count(&models.UserRepository{})
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件名失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
		})
	}
	if count > 0 {
		return nil, errors.New(l.ctx, "重命名文件失败", nil, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
			"reason":        "文件名已存在",
		})
	}

	up := models.UserRepository{
		Name: req.Name,
	}
	_, err = l.svcCtx.Engine.Where("identity =? AND user_identity =?", req.Identity, userIdentity).Update(&up)
	if err != nil {
		return nil, errors.New(l.ctx, "更新文件名失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
		})
	}

	return
}
