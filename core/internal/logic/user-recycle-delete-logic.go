package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserRecycleDeleteLogic 负责“彻底删除”。
// 这一步会真正删除数据库记录，不可恢复。
type UserRecycleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRecycleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRecycleDeleteLogic {
	return &UserRecycleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRecycleDeleteLogic) UserRecycleDelete(req *types.UserRecycleDeleteRequest, userIdentity string) (*types.UserRecycleDeleteResponse, error) {
	if len(req.Identities) == 0 {
		return &types.UserRecycleDeleteResponse{}, nil
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start recycle delete transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback recycle delete failed: %v", err)
		}
	}()

	// 彻底删除前先确认这些文件的确已经在回收站里。
	for _, identity := range req.Identities {
		item, err := getUserRepository(l.ctx, sess, userIdentity, identity, true, true)
		if err != nil {
			return nil, err
		}
		if item.DeletedAt.IsZero() {
			return nil, errors.New(l.ctx, "file is not in recycle bin", nil, map[string]interface{}{
				"identity": identity,
			})
		}
	}

	identities, err := collectSubtreeIdentities(l.ctx, sess, userIdentity, req.Identities, true, true)
	if err != nil {
		return nil, err
	}
	if err := hardDeleteFiles(l.ctx, sess, userIdentity, identities); err != nil {
		return nil, err
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit recycle delete failed", err, nil)
	}
	committed = true

	// 最近文件中如果还保留着这些 identity，也需要同步清掉。
	if len(identities) > 0 && l.svcCtx.RDB != nil {
		members := make([]interface{}, 0, len(identities))
		for _, identity := range identities {
			members = append(members, identity)
		}
		l.svcCtx.RDB.ZRem(l.ctx, recentRedisKey(userIdentity), members...)
	}

	return &types.UserRecycleDeleteResponse{}, nil
}
