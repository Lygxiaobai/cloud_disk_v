// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
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
		return nil, errors.InvalidParam(l.ctx, "refresh token不能为空", nil, map[string]interface{}{
			"reason": "empty_refresh_token",
		})
	}

	jwtCfg := l.svcCtx.Config.JWT
	uc, err := helper.AnalyzeToken(req.RefreshToken, jwtCfg.RefreshSecret)
	if err != nil {
		return nil, errors.AuthFailed(l.ctx, "refresh token无效或已过期", err, nil)
	}
	if uc == nil {
		return nil, errors.AuthFailed(l.ctx, "refresh token无效或已过期", nil, map[string]interface{}{
			"reason": "nil_claims",
		})
	}

	token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role,
		jwtCfg.AccessSecret, jwtCfg.AccessExpire)
	if err != nil {
		return nil, errors.Internal(l.ctx, "生成Token失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
	}
	refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role,
		jwtCfg.RefreshSecret, jwtCfg.RefreshExpire)
	if err != nil {
		return nil, errors.Internal(l.ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
	}

	resp = &types.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
