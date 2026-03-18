// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// Context 已经在中间件中构建完成，直接使用 l.ctx

	// 1. 从数据库查询当前用户
	var user = new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("name = ? and password = ?", req.Name, helper.MD5(req.Password)).Get(user)
	if err != nil {
		return nil, errors.New(l.ctx, "数据库查询失败", err, map[string]interface{}{
			"username": req.Name,
		})
	}

	// 2. 检查用户是否存在
	if !has {
		return nil, errors.New(l.ctx, "用户登录失败", nil, map[string]interface{}{
			"username": req.Name,
			"reason":   "用户名或密码错误",
		})
	}

	// 3. 生成 Token
	token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.TokenExpireTime)
	if err != nil {
		return nil, errors.New(l.ctx, "生成Token失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
	}

	// 4. 生成 RefreshToken
	refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.RefreshTokenExpireTime)
	if err != nil {
		return nil, errors.New(l.ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
	}

	resp = new(types.LoginResponse)
	resp.Token = token
	resp.RefreshToken = refreshToken
	return resp, nil
}
