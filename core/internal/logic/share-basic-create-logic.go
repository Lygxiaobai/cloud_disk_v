// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"
	"strings"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareBasicCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareBasicCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareBasicCreateLogic {
	return &ShareBasicCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareBasicCreateLogic) ShareBasicCreate(req *types.ShareBasicCreateRequest, userIdentity string) (resp *types.ShareBasicCreateResponse, err error) {
	// 根据用户传过来的 Identity 查找到该记录。
	upData := &models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity = ? AND user_identity = ? AND deleted_at IS NULL", req.UserRepositoryIdentity, userIdentity).Get(upData)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件失败", err, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
		})

	}
	if !has {
		return nil, errors.New(l.ctx, "创建分享失败", nil, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
			"reason":              "文件不存在",
		})
	}

	// 这里把分享高级配置一并落库：
	// - access_code: 公共链接访问口令
	// - allow_download: 是否允许二次保存 / 下载式使用
	accessCode := strings.ToUpper(strings.TrimSpace(req.AccessCode))
	allowDownload := 1
	if req.AllowDownload != nil {
		if *req.AllowDownload == 0 {
			allowDownload = 0
		} else {
			allowDownload = 1
		}
	}

	// 向 share_basic 表中插入数据。
	var sb = models.ShareBasic{
		Identity:               helper.UUID(),
		UserIdentity:           userIdentity,
		UserRepositoryIdentity: upData.Identity,
		RepositoryIdentity:     upData.RepositoryIdentity,
		ExpiredTime:            req.ExpiredTime,
		ClickNum:               define.DefaultClickNum,
		AccessCode:             accessCode,
		AllowDownload:          allowDownload,
	}
	_, err = l.svcCtx.Engine.Insert(&sb)
	if err != nil {
		return nil, errors.New(l.ctx, "插入分享记录失败", err, map[string]interface{}{
			"repository_identity": req.UserRepositoryIdentity,
			"expired_time":        req.ExpiredTime,
			"allow_download":      allowDownload,
		})
	}
	resp = &types.ShareBasicCreateResponse{
		Identity:      sb.Identity,
		AccessCodeSet: accessCode != "",
	}
	return
}
