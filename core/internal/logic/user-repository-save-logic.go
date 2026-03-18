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

type UserRepositySaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRepositorySaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRepositySaveLogic {
	return &UserRepositySaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRepositySaveLogic) UserRepositorySave(req *types.UserRepositySaveRequest, userIdentity string) (resp *types.UserRepositySaveResponse, err error) {
	up := &models.UserRepository{
		Identity:           helper.UUID(),
		UserIdentity:       userIdentity,
		ParentId:           req.ParentId,
		RepositoryIdentity: req.ReposityIdentity,
		Ext:                req.Ext,
		Name:               req.Name,
		IsDir:              0,
	}
	_, err = l.svcCtx.Engine.Insert(up)
	if err != nil {
		return nil, errors.New(l.ctx, "保存文件到用户空间失败", err, map[string]interface{}{
			"file_name":           req.Name,
			"repository_identity": req.ReposityIdentity,
		})
	}

	return
}
