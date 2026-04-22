package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileVersionRestoreLogic restores one historical version as the current version.
// The operation runs in a transaction so version snapshotting and repository switching stay atomic.
type UserFileVersionRestoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileVersionRestoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileVersionRestoreLogic {
	return &UserFileVersionRestoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileVersionRestoreLogic) UserFileVersionRestore(req *types.UserFileVersionRestoreRequest, userIdentity string) (*types.UserFileVersionRestoreResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start version restore transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback version restore failed: %v", rbErr)
		}
	}()

	file, repository, err := restoreUserFileVersion(l.ctx, sess, userIdentity, req.FileIdentity, req.VersionIdentity)
	if err != nil {
		return nil, err
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit version restore failed", err, nil)
	}
	committed = true

	return &types.UserFileVersionRestoreResponse{
		FileIdentity:       file.Identity,
		RepositoryIdentity: repository.Identity,
	}, nil
}
