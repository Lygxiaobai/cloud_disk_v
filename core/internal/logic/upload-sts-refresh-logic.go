package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UploadStsRefreshLogic 负责上传过程中的 STS 续签。
// 前端上传大文件时，如果 STS 快过期，就会调用这里换一组新的临时凭证。
type UploadStsRefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadStsRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadStsRefreshLogic {
	return &UploadStsRefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadStsRefreshLogic) UploadStsRefresh(req *types.UploadStsRefreshRequest, userIdentity string) (*types.UploadStsRefreshResponse, error) {
	session := new(models.UploadSession)
	has, err := l.svcCtx.Engine.Where("identity = ? AND user_identity = ?", req.SessionIdentity, userIdentity).Get(session)
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
		return nil, errors.New(l.ctx, "upload session already completed", nil, map[string]interface{}{
			"session_identity": req.SessionIdentity,
		})
	}

	// 续签时 objectKey 不变，只是更换一组新的临时凭证。
	sts, err := l.svcCtx.OSS.IssueUploadSTS(l.ctx, session.Identity, session.ObjectKey)
	if err != nil {
		return nil, errors.New(l.ctx, "issue upload sts failed", err, map[string]interface{}{
			"session_identity": req.SessionIdentity,
		})
	}

	return &types.UploadStsRefreshResponse{
		SessionIdentity: session.Identity,
		ObjectKey:       session.ObjectKey,
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
