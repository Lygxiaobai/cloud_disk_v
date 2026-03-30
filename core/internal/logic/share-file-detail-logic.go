package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// ShareFileDetailLogic 负责公开分享页的文件详情查询。
// 这里除了查详情，还会顺手累计分享点击数。
type ShareFileDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareFileDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareFileDetailLogic {
	return &ShareFileDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareFileDetailLogic) ShareFileDetail(req *types.ShareFileDetailRequest) (*types.ShareFileDetailResponse, error) {
	// 热门分享统计仍然保留，点击详情即视为一次访问。
	go l.svcCtx.ShareCache.IncrDailyClick(context.Background(), req.Identity)
	if _, err := l.svcCtx.Engine.Exec("UPDATE share_basic SET click_num = click_num + 1, updated_at = NOW() WHERE identity = ?", req.Identity); err != nil {
		return nil, errors.New(l.ctx, "update share click count failed", err, map[string]interface{}{
			"share_identity": req.Identity,
		})
	}

	var row struct {
		RepositoryIdentity string `xorm:"repository_identity"`
		Name               string `xorm:"name"`
		Ext                string `xorm:"ext"`
		Size               int64  `xorm:"size"`
		Path               string `xorm:"path"`
		ObjectKey          string `xorm:"object_key"`
	}
	has, err := l.svcCtx.Engine.Table("share_basic").
		Select("share_basic.repository_identity, user_repository.name, user_repository.ext, repository_pool.size, repository_pool.path, repository_pool.object_key").
		Join("LEFT", "repository_pool", "share_basic.repository_identity = repository_pool.identity").
		Join("LEFT", "user_repository", "share_basic.user_repository_identity = user_repository.identity").
		Where("share_basic.identity = ?", req.Identity).
		Get(&row)
	if err != nil {
		return nil, errors.New(l.ctx, "query share file detail failed", err, map[string]interface{}{
			"share_identity": req.Identity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "share file does not exist", nil, map[string]interface{}{
			"share_identity": req.Identity,
		})
	}

	path := row.Path
	objectKey := row.ObjectKey
	if objectKey == "" {
		// 兼容旧数据：历史记录中可能只有 path，没有 object_key。
		objectKey = l.svcCtx.OSS.GuessObjectKey(row.Path)
	}
	if objectKey != "" {
		// 分享页也尽量返回签名 URL，保证分享访问时无需公开 bucket。
		if signedURL, err := l.svcCtx.OSS.SignGetObjectURL(l.ctx, objectKey, l.svcCtx.OSS.PreviewExpires(), row.Name); err == nil {
			path = signedURL
		} else {
			path = l.svcCtx.OSS.BuildObjectURL(objectKey)
		}
	}

	return &types.ShareFileDetailResponse{
		RepositoryIdentity: row.RepositoryIdentity,
		Name:               row.Name,
		Ext:                row.Ext,
		Size:               row.Size,
		Path:               path,
	}, nil
}
