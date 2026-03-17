// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"context"
	"errors"

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
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "PUT")
	ctx = context.WithValue(ctx, "path", "/refresh/token")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	if req.RefreshToken == "" {
		err = errors.New("refresh_token is required")
		logger.LogError(ctx, "刷新Token失败", err, map[string]interface{}{
			"reason": "refresh_token为空",
		})
		return nil, err
	}

	// 这里传的是 refreshToken
	uc, err := helper.AnalyzeToken(req.RefreshToken)
	if err != nil {
		logger.LogError(ctx, "解析RefreshToken失败", err, nil)
		return nil, err
	}
	if uc == nil {
		err = errors.New("invalid refresh token")
		logger.LogError(ctx, "刷新Token失败", err, map[string]interface{}{
			"reason": "无效的refresh_token",
		})
		return nil, err
	}

	// 根据 refreshToken 生成新的一组 token
	token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.TokenExpireTime)
	if err != nil {
		logger.LogError(ctx, "生成Token失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
		return nil, err
	}
	refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.RefreshTokenExpireTime)
	if err != nil {
		logger.LogError(ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id": uc.ID,
		})
		return nil, err
	}
	resp = &types.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return
}
