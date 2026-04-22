package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// UploadInitLogic 负责“上传初始化”这一步。
// 这一层的核心职责是判断当前文件是否可以秒传，如果不能秒传，则为前端创建上传会话并发放 STS。
type UploadInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadInitLogic {
	return &UploadInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadInit 上传初始化流程：
// 1. 校验参数
// 2. 解析目标目录
// 3. 检查同目录是否重名
// 4. 查询资源池是否命中秒传
// 5. 若未命中，则创建 upload_session 并签发 STS
//
// 新增能力：
// 6. 当 target_file_identity 存在时，将当前上传视为“上传新版本”
func (l *UploadInitLogic) UploadInit(req *types.UploadInitRequest, userIdentity string) (*types.UploadInitResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Hash = strings.TrimSpace(req.Hash)
	req.Ext = strings.TrimSpace(req.Ext)
	req.TargetFileIdentity = strings.TrimSpace(req.TargetFileIdentity)
	if req.Name == "" || req.Hash == "" || req.Size <= 0 {
		return nil, errors.New(l.ctx, "invalid upload init params", nil, map[string]interface{}{
			"name": req.Name,
			"size": req.Size,
		})
	}
	if req.Ext == "" {
		req.Ext = path.Ext(req.Name)
	}
	// 大小上限
	if maxSize := l.svcCtx.Config.Upload.MaxSize; maxSize > 0 && req.Size > maxSize {
		return nil, errors.New(l.ctx, "文件过大", nil, map[string]interface{}{
			"size": req.Size,
			"max":  maxSize,
		})
	}
	// 扩展名黑名单
	if err := helper.ValidateUploadExt(req.Ext, l.svcCtx.Config.Upload.BlockedExtensions); err != nil {
		return nil, errors.New(l.ctx, err.Error(), nil, map[string]interface{}{
			"ext": req.Ext,
		})
	}

	// init 阶段里目录校验、秒传判断、会话创建必须放在一个事务中，
	// 否则并发场景下容易出现重名或重复会话问题。
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, errors.New(l.ctx, "start upload transaction failed", err, nil)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if err := sess.Rollback(); err != nil {
			logx.WithContext(l.ctx).Errorf("rollback upload init failed: %v", err)
		}
	}()

	var (
		parentID int64
		err      error
	)
	if req.TargetFileIdentity != "" {
		// 上传新版本时，目录和逻辑文件名都以目标文件为准，避免前端传入错误目录。
		targetFile, err := getUserRepository(l.ctx, sess, userIdentity, req.TargetFileIdentity, false, true)
		if err != nil {
			return nil, err
		}
		if targetFile.IsDir == 1 {
			return nil, errors.New(l.ctx, "folder does not support version replace", nil, map[string]interface{}{
				"file_identity": req.TargetFileIdentity,
			})
		}
		parentID = int64(targetFile.ParentId)
		req.Name = targetFile.Name
	} else {
		parentID, err = resolveParentID(l.ctx, sess, userIdentity, req.ParentId, req.ParentIdentity, true)
		if err != nil {
			return nil, err
		}
	}

	// 上传新版本时复用已有逻辑文件，不再做同目录重名检查。
	if req.TargetFileIdentity == "" {
		if err := ensureNameAvailable(l.ctx, sess, userIdentity, parentID, req.Name, ""); err != nil {
			return nil, err
		}
	}

	// repository_pool 代表“物理文件资源池”，按 hash + size 做全局去重。
	repository := new(models.RepositoryPool)
	has, err := sess.Where("hash = ? AND size = ?", req.Hash, req.Size).Get(repository)
	if err != nil {
		return nil, errors.New(l.ctx, "query repository pool failed", err, map[string]interface{}{
			"hash": req.Hash,
			"size": req.Size,
		})
	}

	if has {
		if req.TargetFileIdentity != "" {
			// 版本替换命中秒传时，直接把逻辑文件切到已有物理版本。
			file, err := replaceUserFileRepository(l.ctx, sess, userIdentity, req.TargetFileIdentity, repository)
			if err != nil {
				return nil, err
			}
			if err := sess.Commit(); err != nil {
				return nil, errors.New(l.ctx, "commit instant version replace failed", err, nil)
			}
			committed = true
			return &types.UploadInitResponse{
				InstantHit:         true,
				FileIdentity:       file.Identity,
				RepositoryIdentity: repository.Identity,
			}, nil
		}

		// 秒传命中:
		// 说明物理文件已经存在，不需要再上传 OSS，
		// 只给当前用户新增一条逻辑文件记录即可。
		file := &models.UserRepository{
			Identity:           helper.UUID(),
			UserIdentity:       userIdentity,
			ParentId:           parentID,
			RepositoryIdentity: repository.Identity,
			Name:               req.Name,
			Ext:                req.Ext,
			IsDir:              0,
			IsFavorite:         0,
		}
		if _, err := sess.Insert(file); err != nil {
			return nil, errors.New(l.ctx, "save instant upload file failed", err, map[string]interface{}{
				"name": req.Name,
			})
		}
		if err := sess.Commit(); err != nil {
			return nil, errors.New(l.ctx, "commit instant upload failed", err, nil)
		}
		committed = true
		return &types.UploadInitResponse{
			InstantHit:         true,
			FileIdentity:       file.Identity,
			RepositoryIdentity: repository.Identity,
		}, nil
	}

	// 未命中秒传时，先由后端生成 objectKey。
	// 前端永远不能自己决定 OSS 的写入路径。
	objectKey := l.svcCtx.OSS.BuildObjectKey(userIdentity, req.Name)

	// upload_session 用于承接这次“尚未完成的上传”：
	// 前端的暂停、继续、续签、完成确认都基于这个会话来做。
	session := &models.UploadSession{
		Identity:           helper.UUID(),
		UserIdentity:       userIdentity,
		ParentId:           parentID,
		TargetFileIdentity: req.TargetFileIdentity,
		Name:               req.Name,
		Ext:                req.Ext,
		Hash:               req.Hash,
		Size:               req.Size,
		ObjectKey:          objectKey,
		Status:             "pending",
	}
	if _, err := sess.Insert(session); err != nil {
		return nil, errors.New(l.ctx, "create upload session failed", err, map[string]interface{}{
			"name": req.Name,
		})
	}

	// 发给前端的 STS 只允许写这次会话对应的 objectKey。
	sts, err := l.svcCtx.OSS.IssueUploadSTS(l.ctx, session.Identity, objectKey)
	if err != nil {
		return nil, errors.New(l.ctx, "issue upload sts failed", err, map[string]interface{}{
			"session_identity": session.Identity,
		})
	}

	if err := sess.Commit(); err != nil {
		return nil, errors.New(l.ctx, "commit upload session failed", err, nil)
	}
	committed = true

	return &types.UploadInitResponse{
		InstantHit:      false,
		SessionIdentity: session.Identity,
		ObjectKey:       objectKey,
		OssBucket:       l.svcCtx.OSS.Bucket(),
		OssRegion:       l.svcCtx.OSS.BrowserRegion(),
		OssEndpoint:     l.svcCtx.OSS.Endpoint(),
		Sts: &types.UploadSTS{
			AccessKeyId:     sts.AccessKeyID,
			AccessKeySecret: sts.AccessKeySecret,
			SecurityToken:   sts.SecurityToken,
			Expiration:      sts.Expiration,
		},
	}, nil
}
