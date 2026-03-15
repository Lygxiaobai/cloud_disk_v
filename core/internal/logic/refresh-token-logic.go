// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
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
	if req.RefreshToken == "" {
		return nil, errors.New("refresh_token is required")
	}

	// 这里传的是 refreshToken
	uc, err := helper.AnalyzeToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if uc == nil {
		return nil, errors.New("invalid refresh token")
	}

	// 根据 refreshToken 生成新的一组 token
	token, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.TokenExpireTime)
	if err != nil {
		return nil, err
	}
	refreshToken, err := helper.GenerateToken(uc.ID, uc.Identity, uc.Name, uc.Role, define.RefreshTokenExpireTime)
	if err != nil {
		return nil, err
	}
	resp = &types.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return
}
