// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"context"

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
	//判断验证码是否存在且有效（key 包含邮箱）
	redisKey := "code:" + req.Email
	code, err := l.svcCtx.RDB.Get(l.ctx, redisKey).Result()
	if err != nil || code != req.Code {
		return nil, errors.New(l.ctx, "验证码验证失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}
	//1.先判断邮箱是否已注册
	count, err := l.svcCtx.Engine.Where("email=?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		return nil, errors.New(l.ctx, "查询邮箱失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}
	if count > 0 {
		return nil, errors.New(l.ctx, "用户注册失败", nil, map[string]interface{}{
			"email":  req.Email,
			"reason": "邮箱已被注册",
		})
	}
	//2.判断用户名是否已注册
	count, err = l.svcCtx.Engine.Where("name=?", req.Name).Count(&models.UserBasic{})
	if err != nil {
		return nil, errors.New(l.ctx, "查询用户名失败", err, map[string]interface{}{
			"username": req.Name,
		})
	}
	if count > 0 {
		return nil, errors.New(l.ctx, "用户注册失败", nil, map[string]interface{}{
			"username": req.Name,
			"reason":   "用户名已被注册",
		})
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
		return nil, errors.New(l.ctx, "插入用户失败", err, map[string]interface{}{
			"username": req.Name,
			"email":    req.Email,
		})
	}
	return
}
