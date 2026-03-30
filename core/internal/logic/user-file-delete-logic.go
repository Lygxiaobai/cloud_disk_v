package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileDeleteLogic 负责单文件删除。
// 内部仍然复用子树删除逻辑，因此删除文件夹时会把全部子节点一起送进回收站。
type UserFileDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileDeleteLogic {
	return &UserFileDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileDeleteLogic) UserFileDelete(req *types.UserFileDeleteRequest, userIdentity string) (*types.UserFileDeleteResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start delete transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback delete failed: %v", err)
		}
	}()

	identities, err := collectSubtreeIdentities(l.ctx, sess, userIdentity, []string{req.Identity}, false, true)
	if err != nil {
		return nil, err
	}
	if err := softDeleteFiles(l.ctx, sess, userIdentity, identities); err != nil {
		return nil, err
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit delete failed", err, nil)
	}
	committed = true
	return &types.UserFileDeleteResponse{}, nil
}
