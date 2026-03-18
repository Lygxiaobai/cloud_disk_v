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
	"time"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MailCodeSendRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMailCodeSendRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MailCodeSendRegisterLogic {
	return &MailCodeSendRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MailCodeSendRegisterLogic) MailCodeSendRegister(req *types.MailCodeRequest) (resp *types.MailCodeResponse, err error) {
	//设置traceId
	traceId := l.ctx.Value("trace_id").(string)
	//构建上下文信息
	ctx := context.WithValue(l.ctx, "method", "POST")
	ctx = context.WithValue(ctx, "trace_id", traceId)
	ctx = context.WithValue(ctx, "path", "/mail/code/send/register")

	//1.检验邮箱是否注册
	count, err := l.svcCtx.Engine.Where("email=?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		//写入文件系统
		err := errors.New("邮箱有误")
		logger.LogError(ctx, "邮箱验证失败", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, err
	}
	if count > 0 {
		//写入文件系统
		err := errors.New("该邮箱已被注册")
		logger.LogError(ctx, "该邮箱已被注册", err, map[string]interface{}{
			"email": req.Email,
		})
		return nil, err
	}
	//2.未注册 发送验证码
	//2.1生成随机验证码
	code := helper.RandCode()
	//3.存储到redis（key 包含邮箱，避免多用户冲突）
	redisKey := "code:" + req.Email
	err = l.svcCtx.RDB.Set(l.ctx, redisKey, code, time.Second*time.Duration(define.CodeExpireTime)).Err()
	if err != nil {
		//写入文件系统
		err := errors.New("验证码写入redis出现错误")
		logger.LogError(ctx, "验证码写入redis出现错误", err, map[string]interface{}{
			"code": code,
		})
		return nil, err
	}

	// 4. 发送邮件任务到 RabbitMQ（异步）
	err = l.svcCtx.EmailProducer.SendEmailTask(req.Email, code)
	if err != nil {
		// 发送到队列失败，记录错误
		logger.LogError(ctx, "发送邮件任务到队列失败", err, map[string]interface{}{
			"email": req.Email,
			"code":  code,
		})
		return nil, errors.New("发送验证码失败，请稍后重试")
	}

	// 5. 立即返回响应（不等待邮件发送）
	return &types.MailCodeResponse{
		code,
	}, nil
}
