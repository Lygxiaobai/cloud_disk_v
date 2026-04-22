package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileBatchFavoriteLogic 负责批量收藏 / 取消收藏。
type UserFileBatchFavoriteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileBatchFavoriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileBatchFavoriteLogic {
	return &UserFileBatchFavoriteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileBatchFavoriteLogic) UserFileBatchFavorite(req *types.UserFileBatchFavoriteRequest, userIdentity string) (*types.UserFileBatchFavoriteResponse, error) {
	if len(req.Identities) == 0 {
		return &types.UserFileBatchFavoriteResponse{}, nil
	}

	args := make([]interface{}, 0, len(req.Identities)+2)
	args = append(args, req.IsFavorite, userIdentity)
	for _, identity := range req.Identities {
		args = append(args, identity)
	}

	sql := fmt.Sprintf(
		"UPDATE user_repository SET is_favorite = ? WHERE user_identity = ? AND identity IN (%s) AND deleted_at IS NULL",
		placeholders(len(req.Identities)),
	)
	execArgs := append([]interface{}{sql}, args...)
	if _, err := l.svcCtx.Engine.Exec(execArgs...); err != nil {
		return nil, errors.New(l.ctx, "batch update favorite failed", err, map[string]interface{}{
			"count": len(req.Identities),
		})
	}

	return &types.UserFileBatchFavoriteResponse{}, nil
}
