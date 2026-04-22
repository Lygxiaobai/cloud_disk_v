package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UploadCompleteLogic 负责“上传完成确认”这一步。
type UploadCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCompleteLogic {
	return &UploadCompleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadComplete 处理流程：
// 1. 锁住 upload_session，确认当前用户拥有这次上传会话
// 2. 通过 HeadObject 校验 OSS 上对象已经存在且大小正确
// 3. 锁住 repository_pool 的 hash+size 记录，防止并发重复入库
// 4. 普通上传则写入 user_repository
// 5. 版本替换则更新已有逻辑文件的 repository_identity
// 6. 把 upload_session 状态改成 completed
func (l *UploadCompleteLogic) UploadComplete(req *types.UploadCompleteRequest, userIdentity string) (*types.UploadCompleteResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start upload complete transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback upload complete failed: %v", err)
		}
	}()

	session := new(models.UploadSession)
	has, err := sess.SQL(
		"SELECT * FROM upload_session WHERE identity = ? AND user_identity = ? LIMIT 1 FOR UPDATE",
		req.SessionIdentity,
		userIdentity,
	).Get(session)
	if err != nil {
		return nil, errors.New(l.ctx, "query upload session failed", err, map[string]interface{}{
			"session_identity": req.SessionIdentity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "upload session does not exist", nil, map[string]interface{}{
			"session_identity": req.SessionIdentity,
		})
	}

	if session.Status == "completed" {
		existing := new(models.UserRepository)
		lookupIdentity := session.TargetFileIdentity
		if lookupIdentity == "" {
			has, err = sess.Where(
				"user_identity = ? AND parent_id = ? AND repository_identity = ? AND name = ? AND deleted_at IS NULL",
				userIdentity,
				session.ParentId,
				session.RepositoryIdentity,
				session.Name,
			).Desc("id").Get(existing)
		} else {
			has, err = sess.Where("identity = ? AND user_identity = ? AND deleted_at IS NULL", lookupIdentity, userIdentity).Get(existing)
		}
		if err == nil && has {
			if err := sess.Commit(); err != nil {
				return nil, errors.New(l.ctx, "commit upload complete idempotent failed", err, nil)
			}
			committed = true
			return &types.UploadCompleteResponse{
				FileIdentity:       existing.Identity,
				RepositoryIdentity: existing.RepositoryIdentity,
			}, nil
		}
	}

	head, err := l.svcCtx.OSS.HeadObject(l.ctx, session.ObjectKey)
	if err != nil {
		return nil, errors.New(l.ctx, "verify uploaded object failed", err, map[string]interface{}{
			"object_key": session.ObjectKey,
		})
	}
	if head.ContentLength != session.Size {
		return nil, errors.New(l.ctx, "uploaded object size mismatch", nil, map[string]interface{}{
			"expected": session.Size,
			"actual":   head.ContentLength,
		})
	}

	if session.TargetFileIdentity == "" {
		if err := ensureNameAvailable(l.ctx, sess, userIdentity, session.ParentId, session.Name, ""); err != nil {
			return nil, err
		}
	}

	repository := new(models.RepositoryPool)
	has, err = sess.SQL(
		"SELECT * FROM repository_pool WHERE hash = ? AND size = ? LIMIT 1 FOR UPDATE",
		session.Hash,
		session.Size,
	).Get(repository)
	if err != nil {
		return nil, errors.New(l.ctx, "query repository pool failed", err, map[string]interface{}{
			"hash": session.Hash,
		})
	}

	createdRepository := false
	if !has {
		repository = &models.RepositoryPool{
			Identity:  helper.UUID(),
			Hash:      session.Hash,
			Name:      session.Name,
			Ext:       session.Ext,
			Size:      session.Size,
			Path:      l.svcCtx.OSS.BuildObjectURL(session.ObjectKey),
			ObjectKey: session.ObjectKey,
		}
		if _, err := sess.Insert(repository); err != nil {
			return nil, errors.New(l.ctx, "save repository pool failed", err, map[string]interface{}{
				"object_key": session.ObjectKey,
			})
		}
		createdRepository = true
	}

	var file *models.UserRepository
	if session.TargetFileIdentity != "" {
		// 上传新版本时，不新建逻辑文件，只替换已有逻辑文件背后的 repository。
		file, err = replaceUserFileRepository(l.ctx, sess, userIdentity, session.TargetFileIdentity, repository)
		if err != nil {
			return nil, err
		}
	} else {
		file = &models.UserRepository{
			Identity:           helper.UUID(),
			UserIdentity:       userIdentity,
			ParentId:           session.ParentId,
			RepositoryIdentity: repository.Identity,
			Name:               session.Name,
			Ext:                session.Ext,
			IsDir:              0,
			IsFavorite:         0,
		}
		if _, err := sess.Insert(file); err != nil {
			return nil, errors.New(l.ctx, "save uploaded file failed", err, map[string]interface{}{
				"name": session.Name,
			})
		}
	}

	if _, err := sess.Where("id = ?", session.Id).Cols("status", "repository_identity").Update(&models.UploadSession{
		Status:             "completed",
		RepositoryIdentity: repository.Identity,
	}); err != nil {
		return nil, errors.New(l.ctx, "update upload session failed", err, map[string]interface{}{
			"session_identity": session.Identity,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit upload complete failed", err, nil)
	}
	committed = true

	// 如果因为并发命中了已有资源池，则把本次多余的 OSS 对象清掉。
	if !createdRepository && repository.ObjectKey != "" && repository.ObjectKey != session.ObjectKey {
		_ = l.svcCtx.OSS.DeleteObject(l.ctx, session.ObjectKey)
	}

	return &types.UploadCompleteResponse{
		FileIdentity:       file.Identity,
		RepositoryIdentity: repository.Identity,
	}, nil
}
