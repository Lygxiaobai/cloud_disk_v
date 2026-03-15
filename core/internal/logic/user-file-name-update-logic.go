// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
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
	//当当前目录存在与修改名字相同的文件时，不允许修改
	count, err := l.svcCtx.Engine.Where("name=? AND parent_id=(select parent_id from user_repository where user_identity = ? and identity =?)", req.Name, userIdentity, req.Identity).Count(&models.UserRepository{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该文件名已经存在")
	}

	up := models.UserRepository{
		Name: req.Name,
	}
	_, err = l.svcCtx.Engine.Where("identity =? AND user_identity =?", req.Identity, userIdentity).Update(&up)
	if err != nil {
		return nil, err
	}

	return
}
