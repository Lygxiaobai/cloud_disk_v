package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileFavoriteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileFavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileFavoriteLogic {
	return &UserFileFavoriteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileFavoriteLogic) UserFileFavorite(req *types.UserFileFavoriteRequest, userIdentity string) (*types.UserFileFavoriteResponse, error) {
	affected, err := l.svcCtx.Engine.Where("identity = ? AND user_identity = ? AND deleted_at IS NULL", req.Identity, userIdentity).
		Cols("is_favorite").
		Update(&models.UserRepository{IsFavorite: req.IsFavorite})
	if err != nil {
		return nil, errors.New(l.ctx, "update favorite failed", err, map[string]interface{}{
			"identity": req.Identity,
		})
	}
	if affected == 0 {
		return nil, errors.New(l.ctx, "file does not exist", nil, map[string]interface{}{
			"identity": req.Identity,
		})
	}
	return &types.UserFileFavoriteResponse{}, nil
}
