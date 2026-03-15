// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
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
	//1.从数据库查询当前用户
	var user = new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("name = ? and password = ?", req.Name, helper.MD5(req.Password)).Get(user)
	if err != nil {
		return nil, err
	}
	//没有查到用户
	if !has {
		err = errors.New("用户名或密码错误")
		return nil, err
	}
	//生成Token
	token, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.TokenExpireTime)
	if err != nil {
		return nil, err
	}
	//生成refreshToken
	refreshToken, err := helper.GenerateToken(user.Id, user.Identity, user.Name, user.Role, define.RefreshTokenExpireTime)
	resp = new(types.LoginResponse)
	resp.Token = token
	resp.RefreshToken = refreshToken
	return resp, nil
}
