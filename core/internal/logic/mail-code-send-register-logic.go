package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"time"

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

func (l *MailCodeSendRegisterLogic) MailCodeSendRegister(req *types.MailCodeRequest) (*types.MailCodeResponse, error) {
	mailCfg := l.svcCtx.Config.Mail

	count, err := l.svcCtx.Engine.Where("email = ?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		return nil, errors.Internal(l.ctx, "注册验证码校验失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}
	if count > 0 {
		return nil, errors.Conflict(l.ctx, "邮箱已被注册", nil, map[string]interface{}{
			"email": req.Email,
		})
	}

	code := helper.RandCode(mailCfg.CodeLen)
	redisKey := "code:" + req.Email
	if err := l.svcCtx.RDB.Set(l.ctx, redisKey, code, time.Second*time.Duration(mailCfg.CodeExpire)).Err(); err != nil {
		return nil, errors.Internal(l.ctx, "写入验证码失败", err, map[string]interface{}{
			"email": req.Email,
		})
	}

	if l.svcCtx.EmailProducer != nil {
		if err := l.svcCtx.EmailProducer.SendEmailTask(req.Email, code); err != nil {
			return nil, errors.Internal(l.ctx, "发送邮件任务失败", err, map[string]interface{}{
				"email": req.Email,
			})
		}
	} else {
		hc := helper.MailConfig{
			From:       mailCfg.From,
			Host:       mailCfg.Host,
			Username:   mailCfg.Username,
			Password:   mailCfg.Password,
			ServerName: mailCfg.ServerName,
		}
		if err := helper.MailCodeSend(req.Email, code, hc); err != nil {
			return nil, errors.Internal(l.ctx, "发送验证码邮件失败", err, map[string]interface{}{
				"email": req.Email,
			})
		}
	}

	return &types.MailCodeResponse{Message: "验证码已发送到您的邮箱"}, nil
}
