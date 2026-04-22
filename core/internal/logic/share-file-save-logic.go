package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareFileSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareFileSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareFileSaveLogic {
	return &ShareFileSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareFileSaveLogic) ShareFileSave(req *types.ShareFileSaveRequest, userIdentity string) (resp *types.ShareFileSaveResponse, err error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start share save transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := sess.Rollback(); rbErr != nil {
			logx.WithContext(l.ctx).Errorf("rollback share save failed: %v", rbErr)
		}
	}()

	// 新版优先通过 share_identity + access_code 校验分享访问权限，
	// 这样“保存到我的网盘”就不会绕过分享口令与下载开关。
	repositoryIdentity := req.RepositoryIdentity
	if req.ShareIdentity != "" {
		share, err := validateShareAccess(l.ctx, sess, req.ShareIdentity, req.AccessCode, true)
		if err != nil {
			return nil, err
		}
		if share.AllowDownload == 0 {
			return nil, errors.New(l.ctx, "share does not allow saving", nil, map[string]interface{}{
				"share_identity": req.ShareIdentity,
			})
		}
		repositoryIdentity = share.RepositoryIdentity
	}

	var rpData = &models.RepositoryPool{}
	has, err := sess.Where("identity = ?", repositoryIdentity).Get(rpData)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件失败", err, map[string]interface{}{
			"repository_identity": repositoryIdentity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "保存分享文件失败", nil, map[string]interface{}{
			"repository_identity": repositoryIdentity,
			"reason":              "文件不存在",
		})
	}

	if err := ensureNameAvailable(l.ctx, sess, userIdentity, req.ParentId, rpData.Name, ""); err != nil {
		return nil, err
	}

	upData := models.UserRepository{
		Identity:           helper.UUID(),
		UserIdentity:       userIdentity,
		ParentId:           req.ParentId,
		RepositoryIdentity: repositoryIdentity,
		Name:               rpData.Name,
		Ext:                rpData.Ext,
		IsDir:              0,
	}
	_, err = sess.Insert(&upData)
	if err != nil {
		return nil, errors.New(l.ctx, "插入用户文件失败", err, map[string]interface{}{
			"repository_identity": repositoryIdentity,
			"parent_id":           req.ParentId,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit share save failed", err, nil)
	}
	committed = true
	return &types.ShareFileSaveResponse{
		Identity: upData.Identity,
	}, nil
}
