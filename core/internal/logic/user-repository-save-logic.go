package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserRepositySaveLogic 主要用于“把已有资源保存到当前用户网盘”。
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

func (l *UserRepositySaveLogic) UserRepositorySave(req *types.UserRepositySaveRequest, userIdentity string) (*types.UserRepositySaveResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start save file transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback save file failed: %v", rbErr)
		}
	}()

	if err := ensureNameAvailable(l.ctx, sess, userIdentity, req.ParentId, req.Name, ""); err != nil {
		return nil, err
	}

	up := &models.UserRepository{
		Identity:           helper.UUID(),
		UserIdentity:       userIdentity,
		ParentId:           req.ParentId,
		RepositoryIdentity: req.ReposityIdentity,
		Ext:                req.Ext,
		Name:               req.Name,
		IsDir:              0,
		IsFavorite:         0,
	}
	if _, err := sess.Insert(up); err != nil {
		return nil, errors.New(l.ctx, "save file to user space failed", err, map[string]interface{}{
			"file_name":           req.Name,
			"repository_identity": req.ReposityIdentity,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit save file failed", err, nil)
	}
	committed = true

	return &types.UserRepositySaveResponse{}, nil
}
