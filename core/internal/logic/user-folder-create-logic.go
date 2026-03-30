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

type UserFolderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFolderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderCreateLogic {
	return &UserFolderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFolderCreateLogic) UserFolderCreate(req *types.UserFolderCreateRequest, userIdentity string) (*types.UserFolderCreateResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start create folder transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback folder create failed: %v", rbErr)
		}
	}()

	if err := ensureNameAvailable(l.ctx, sess, userIdentity, req.ParentId, req.Name, ""); err != nil {
		return nil, err
	}

	up := models.UserRepository{
		Identity:     helper.UUID(),
		Name:         req.Name,
		ParentId:     req.ParentId,
		UserIdentity: userIdentity,
		IsDir:        1,
		IsFavorite:   0,
	}
	if _, err := sess.Insert(&up); err != nil {
		return nil, errors.New(l.ctx, "insert folder failed", err, map[string]interface{}{
			"folder_name": req.Name,
			"parent_id":   req.ParentId,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit create folder failed", err, nil)
	}
	committed = true

	return &types.UserFolderCreateResponse{Identity: up.Identity}, nil
}
