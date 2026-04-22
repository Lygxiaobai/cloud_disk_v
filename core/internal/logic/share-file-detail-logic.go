package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

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
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	share, err := getShareBasic(l.ctx, sess, req.Identity, false, false)
	if err != nil {
		return nil, err
	}
	if share.ExpiredTime > 0 && share.CreatedAt.Add(time.Duration(share.ExpiredTime)*time.Second).Before(time.Now()) {
		return nil, errors.New(l.ctx, "share has expired", nil, map[string]interface{}{
			"share_identity": req.Identity,
		})
	}

	go l.svcCtx.ShareCache.IncrDailyClick(context.Background(), req.Identity)
	if _, err := l.svcCtx.Engine.Exec(
		"UPDATE share_basic SET click_num = click_num + 1, updated_at = NOW() WHERE identity = ? AND deleted_at IS NULL",
		req.Identity,
	); err != nil {
		return nil, errors.New(l.ctx, "update share click count failed", err, map[string]interface{}{
			"share_identity": req.Identity,
		})
	}

	resp, err := l.svcCtx.ShareCache.LoadShareDetail(l.ctx, req.Identity, func() (*types.ShareFileDetailResponse, error) {
		var row struct {
			RepositoryIdentity string `xorm:"repository_identity"`
			Name               string `xorm:"name"`
			Ext                string `xorm:"ext"`
			Size               int64  `xorm:"size"`
			Path               string `xorm:"path"`
			ObjectKey          string `xorm:"object_key"`
		}

		has, queryErr := l.svcCtx.Engine.Table("share_basic").
			Select("share_basic.repository_identity, user_repository.name, user_repository.ext, repository_pool.size, repository_pool.path, repository_pool.object_key").
			Join("LEFT", "repository_pool", "share_basic.repository_identity = repository_pool.identity").
			Join("LEFT", "user_repository", "share_basic.user_repository_identity = user_repository.identity").
			Where("share_basic.identity = ? AND share_basic.deleted_at IS NULL", req.Identity).
			Get(&row)
		if queryErr != nil {
			return nil, errors.New(l.ctx, "query share file detail failed", queryErr, map[string]interface{}{
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
			objectKey = l.svcCtx.OSS.GuessObjectKey(row.Path)
		}
		if objectKey != "" {
			if signedURL, signErr := l.svcCtx.OSS.SignGetObjectURL(l.ctx, objectKey, l.svcCtx.OSS.PreviewExpires(), row.Name); signErr == nil {
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
	})
	if err != nil {
		return nil, err
	}
	resp.AllowDownload = share.AllowDownload

	expectedCode := strings.TrimSpace(share.AccessCode)
	if expectedCode != "" && !strings.EqualFold(expectedCode, strings.TrimSpace(req.AccessCode)) {
		resp.NeedCode = true
	}

	return resp, nil
}
