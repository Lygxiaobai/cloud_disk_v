package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// ShareBasicDeleteLogic 用于撤销分享链接。
// 这里走软删除，这样历史点击统计和审计字段仍然可以保留。
type ShareBasicDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareBasicDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareBasicDeleteLogic {
	return &ShareBasicDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareBasicDeleteLogic) ShareBasicDelete(req *types.ShareBasicDeleteRequest, userIdentity string) (*types.ShareBasicDeleteResponse, error) {
	if len(req.Identities) == 0 {
		return &types.ShareBasicDeleteResponse{}, nil
	}

	args := make([]interface{}, 0, len(req.Identities)+1)
	args = append(args, userIdentity)
	for _, identity := range req.Identities {
		args = append(args, identity)
	}

	sql := fmt.Sprintf(
		"UPDATE share_basic SET deleted_at = NOW(), updated_at = NOW() WHERE user_identity = ? AND identity IN (%s) AND deleted_at IS NULL",
		placeholders(len(req.Identities)),
	)
	execArgs := append([]interface{}{sql}, args...)
	if _, err := l.svcCtx.Engine.Exec(execArgs...); err != nil {
		return nil, errors.New(l.ctx, "delete share failed", err, map[string]interface{}{
			"user_identity": userIdentity,
			"count":         len(req.Identities),
		})
	}

	return &types.ShareBasicDeleteResponse{}, nil
}
