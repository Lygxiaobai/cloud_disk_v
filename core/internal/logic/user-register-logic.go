// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logger"
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
	// 从 context 中获取 TraceID（由中间件设置）
	traceID, _ := l.ctx.Value("trace_id").(string)

	// 构建上下文信息
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "path", "/user/register")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//判断验证码是否存在且有效（key 包含邮箱）
	redisKey := "code:" + req.Email
	code, err := l.svcCtx.RDB.Get(l.ctx, redisKey).Result()
	if err != nil || code != req.Code {
		err = errors.New("验证码有误")
		logger.LogError(ctx, "验证码验证失败", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, err
	}
	//1.先判断邮箱是否已注册
	count, err := l.svcCtx.Engine.Where("email=?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		logger.LogError(ctx, "查询邮箱失败", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, err
	}
	if count > 0 {
		err = errors.New("邮箱已被注册")
		logger.LogError(ctx, "用户注册失败", err, map[string]interface{}{
			"email":  req.Email,
			"reason": "邮箱已被注册",
		})
		return nil, err
	}
	//2.判断用户名是否已注册
	count, err = l.svcCtx.Engine.Where("name=?", req.Name).Count(&models.UserBasic{})
	if err != nil {
		logger.LogError(ctx, "查询用户名失败", err, map[string]interface{}{
			"username": req.Name,
		})
		return nil, err
	}
	if count > 0 {
		err = errors.New("用户名已被注册")
		logger.LogError(ctx, "用户注册失败", err, map[string]interface{}{
			"username": req.Name,
			"reason":   "用户名已被注册",
		})
		return nil, err
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
		logger.LogError(ctx, "插入用户失败", err, map[string]interface{}{
			"username": req.Name,
			"email":    req.Email,
		})
		return nil, err
	}
	return
}
