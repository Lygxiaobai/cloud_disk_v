// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	if req.RefreshToken == "" {
		return nil, errors.New(l.ctx, "刷新Token失败", nil, map[string]interface{}{
			"reason": "refresh_token为空",
		})
	}

	// 这里传的是 refreshToken
	uc, err := helper.AnalyzeToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New(l.ctx, "解析RefreshToken失败", err, nil)
	}
	if uc == nil {
		return nil, errors.New(l.ctx, "刷新Token失败", nil, map[string]interface{}{
			"reason": "无效的refresh_token",
		})
	}

	// 根据 refreshToken 生成新的一组 token
	token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.TokenExpireTime)
	if err != nil {
		return nil, errors.New(l.ctx, "生成Token失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
	}
	refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.RefreshTokenExpireTime)
	if err != nil {
		return nil, errors.New(l.ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
	}
	resp = &types.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return
}
