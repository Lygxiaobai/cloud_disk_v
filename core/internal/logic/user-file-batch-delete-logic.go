package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileBatchDeleteLogic 负责批量删除。
// 删除并不是立即物理删除，而是统一进入回收站。
type UserFileBatchDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileBatchDeleteLogic {
	return &UserFileBatchDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileBatchDeleteLogic) UserFileBatchDelete(req *types.UserFileBatchDeleteRequest, userIdentity string) (*types.UserFileBatchDeleteResponse, error) {
	if len(req.Identities) == 0 {
		return &types.UserFileBatchDeleteResponse{}, nil
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start batch delete transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback batch delete failed: %v", err)
		}
	}()

	// 批量删除要带上文件夹子树，不能只删顶层节点。
	identities, err := collectSubtreeIdentities(l.ctx, sess, userIdentity, req.Identities, false, true)
	if err != nil {
		return nil, err
	}
	if err := softDeleteFiles(l.ctx, sess, userIdentity, identities); err != nil {
		return nil, err
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit batch delete failed", err, nil)
	}
	committed = true
	return &types.UserFileBatchDeleteResponse{}, nil
}
