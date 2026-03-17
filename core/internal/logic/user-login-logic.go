// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

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
	// 从 context 中获取 TraceID（由中间件设置）
	traceID, _ := l.ctx.Value("trace_id").(string)

	// 构建上下文信息
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "path", "/user/login")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//1.从数据库查询当前用户
	var user = new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("name = ? and password = ?", req.Name, helper.MD5(req.Password)).Get(user)
	if err != nil {
		logger.LogError(ctx, "数据库查询失败", err, map[string]interface{}{
			"username": req.Name,
		})
		return nil, err
	}
	//没有查到用户
	if !has {
		err = errors.New("用户名或密码错误")
		logger.LogError(ctx, "用户登录失败", err, map[string]interface{}{
			"username": req.Name,
			"reason":   "用户名或密码错误",
		})
		return nil, err
	}
	//生成Token
	token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.TokenExpireTime)
	if err != nil {
		logger.LogError(ctx, "生成Token失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
		return nil, err
	}
	//生成refreshToken
	refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.RefreshTokenExpireTime)
	if err != nil {
		logger.LogError(ctx, "生成RefreshToken失败", err, map[string]interface{}{
			"user_id":   user.Id,
			"user_name": user.Name,
		})
		return nil, err
	}
	resp = new(types.LoginResponse)
	resp.Token = token
	resp.RefreshToken = refreshToken
	return resp, nil
}
