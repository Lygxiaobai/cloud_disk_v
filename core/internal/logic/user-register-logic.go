// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRegisterLogic {
	return &UserRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRegisterLogic) UserRegister(req *types.UserRegisterRequest) (resp *types.UserRegisterResponse, err error) {
	//判断验证码是否存在且有效
	code, err := l.svcCtx.RDB.Get(l.ctx, "code").Result()
	if err != nil || code != req.Code {
		err = errors.New("验证码有误")
		return nil, err
	}
	//1.先判断邮箱是否已注册
	count, err := l.svcCtx.Engine.Where("email=?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("邮箱已被注册")
	}
	//2.判断用户名是否已注册
	count, err = l.svcCtx.Engine.Where("name=?", req.Name).Count(&models.UserBasic{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已被注册")
	}
	//3.新增一条用户
	user := models.UserBasic{
		Identity: helper.UUID(),
		Name:     req.Name,
		Email:    req.Email,
		Password: helper.MD5(req.Password),
	}
	_, err = l.svcCtx.Engine.InsertOne(&user)
	if err != nil {
		return nil, err
	}
	return
}
