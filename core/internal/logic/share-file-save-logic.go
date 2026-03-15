// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/helper"
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
	//从公共池获取文件信息
	var rpData = &models.RepositoryPool{}
	has, err := l.svcCtx.Engine.Where("identity = ?", req.RepositoryIdentity).Get(rpData)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("错误操作")
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
		return nil, err
	}
	return
}
