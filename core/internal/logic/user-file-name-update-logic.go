package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileNameUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileNameUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileNameUpdateLogic {
	return &UserFileNameUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileNameUpdateLogic) UserFileNameUpdate(req *types.UserFileNameUpdateRequest, userIdentity string) (resp *types.UserFileNameUpdateResponse, err error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start rename transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback rename failed: %v", rbErr)
		}
	}()

	file := &models.UserRepository{}
	has, err := sess.SQL("SELECT * FROM user_repository WHERE identity = ? AND user_identity = ? AND deleted_at IS NULL LIMIT 1 FOR UPDATE", req.Identity, userIdentity).Get(file)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件失败", err, map[string]interface{}{
			"file_identity": req.Identity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "重命名文件失败", nil, map[string]interface{}{
			"file_identity": req.Identity,
			"reason":        "文件不存在",
		})
	}

	if err := ensureNameAvailable(l.ctx, sess, userIdentity, int64(file.ParentId), req.Name, file.Identity); err != nil {
		return nil, err
	}

	up := models.UserRepository{Name: req.Name}
	_, err = sess.Where("identity = ? AND user_identity = ?", req.Identity, userIdentity).Update(&up)
	if err != nil {
		return nil, errors.New(l.ctx, "更新文件名失败", err, map[string]interface{}{
			"new_name":      req.Name,
			"file_identity": req.Identity,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit rename failed", err, nil)
	}
	committed = true
	return
}
