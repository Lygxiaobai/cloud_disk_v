package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileVersionListLogic 返回某个逻辑文件的版本历史。
// 列表第一项始终是当前版本，后续才是历史快照。
type UserFileVersionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileVersionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileVersionListLogic {
	return &UserFileVersionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileVersionListLogic) UserFileVersionList(req *types.UserFileVersionListRequest, userIdentity string) (*types.UserFileVersionListResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	file, err := getUserRepository(l.ctx, sess, userIdentity, req.Identity, false, false)
	if err != nil {
		return nil, err
	}
	if file.IsDir == 1 {
		return nil, errors.New(l.ctx, "folder does not support version history", nil, map[string]interface{}{
			"file_identity": req.Identity,
		})
	}

	currentRepository := new(models.RepositoryPool)
	has, err := sess.Where("identity = ?", file.RepositoryIdentity).Get(currentRepository)
	if err != nil {
		return nil, errors.New(l.ctx, "query current repository failed", err, map[string]interface{}{
			"file_identity": req.Identity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "current repository does not exist", nil, map[string]interface{}{
			"file_identity": req.Identity,
		})
	}

	list := []*types.UserFileVersionItem{
		{
			Identity:           "current",
			FileIdentity:       file.Identity,
			RepositoryIdentity: currentRepository.Identity,
			Name:               file.Name,
			Ext:                file.Ext,
			Size:               currentRepository.Size,
			Hash:               currentRepository.Hash,
			Action:             "current",
			IsCurrent:          1,
			CreatedAt:          file.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	history := make([]*types.UserFileVersionItem, 0)
	historySQL := `
SELECT
  v.identity,
  v.file_identity,
  v.repository_identity,
  v.name,
  v.ext,
  v.size,
  v.hash,
  v.action,
  0 AS is_current,
  DATE_FORMAT(v.created_at, '%Y-%m-%d %H:%i:%s') AS created_at
FROM user_file_version v
WHERE v.user_identity = ?
  AND v.file_identity = ?
ORDER BY v.created_at DESC, v.id DESC`
	if err := sess.SQL(historySQL, userIdentity, file.Identity).Find(&history); err != nil {
		return nil, errors.New(l.ctx, "query file version history failed", err, map[string]interface{}{
			"file_identity": file.Identity,
		})
	}

	list = append(list, history...)
	return &types.UserFileVersionListResponse{List: list}, nil
}
