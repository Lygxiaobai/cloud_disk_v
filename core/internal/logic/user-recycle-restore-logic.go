package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserRecycleRestoreLogic 负责回收站恢复。
type UserRecycleRestoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRecycleRestoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRecycleRestoreLogic {
	return &UserRecycleRestoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserRecycleRestore 的关键规则：
// 1. 可以批量恢复
// 2. 恢复时要把子树一起恢复
// 3. 不能在父目录还在回收站时先恢复子文件
func (l *UserRecycleRestoreLogic) UserRecycleRestore(req *types.UserRecycleRestoreRequest, userIdentity string) (*types.UserRecycleRestoreResponse, error) {
	if len(req.Identities) == 0 {
		return &types.UserRecycleRestoreResponse{}, nil
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start recycle restore transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback recycle restore failed: %v", err)
		}
	}()

	requestedSet := make(map[string]struct{}, len(req.Identities))
	for _, identity := range req.Identities {
		requestedSet[identity] = struct{}{}
		item, err := getUserRepository(l.ctx, sess, userIdentity, identity, true, true)
		if err != nil {
			return nil, err
		}

		// 如果父目录还没恢复，就不能直接恢复子文件，
		// 否则目录结构会变得不完整。
		if item.ParentId != 0 {
			parent, err := getUserRepositoryByIDRaw(l.ctx, sess, userIdentity, item.ParentId, true)
			if err != nil {
				return nil, err
			}
			if !parent.DeletedAt.IsZero() {
				if _, ok := requestedSet[parent.Identity]; !ok {
					return nil, errors.New(l.ctx, "parent folder is still in recycle bin", nil, map[string]interface{}{
						"identity": identity,
					})
				}
			}
		}

		// 恢复后目录里也不能出现同名冲突。
		if err := ensureNameAvailable(l.ctx, sess, userIdentity, item.ParentId, item.Name, item.Identity); err != nil {
			return nil, err
		}
	}

	identities, err := collectSubtreeIdentities(l.ctx, sess, userIdentity, req.Identities, true, true)
	if err != nil {
		return nil, err
	}
	if err := restoreFiles(l.ctx, sess, userIdentity, identities); err != nil {
		return nil, err
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit recycle restore failed", err, nil)
	}
	committed = true
	return &types.UserRecycleRestoreResponse{}, nil
}
