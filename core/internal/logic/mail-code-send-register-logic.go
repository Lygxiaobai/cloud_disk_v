package logic

import (
	"cloud_disk/core/internal/define"
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
	count, err := l.svcCtx.Engine.Where("email = ?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		return nil, errors.New(l.ctx, "mail verification failed", err, map[string]interface{}{
			"email": req.Email,
		})
	}
	if count > 0 {
		return nil, errors.New(l.ctx, "email already registered", nil, map[string]interface{}{
			"email": req.Email,
		})
	}

	code := helper.RandCode()
	redisKey := "code:" + req.Email
	if err := l.svcCtx.RDB.Set(l.ctx, redisKey, code, time.Second*time.Duration(define.CodeExpireTime)).Err(); err != nil {
		return nil, errors.New(l.ctx, "write verification code to redis failed", err, map[string]interface{}{
			"email": req.Email,
		})
	}

	if l.svcCtx.EmailProducer != nil {
		if err := l.svcCtx.EmailProducer.SendEmailTask(req.Email, code); err != nil {
			return nil, errors.New(l.ctx, "send email task failed", err, map[string]interface{}{
				"email": req.Email,
			})
		}
	} else if err := helper.MailCodeSend(req.Email, code); err != nil {
		return nil, errors.New(l.ctx, "send email failed", err, map[string]interface{}{
			"email": req.Email,
		})
	}

	return &types.MailCodeResponse{Code: code}, nil
}
