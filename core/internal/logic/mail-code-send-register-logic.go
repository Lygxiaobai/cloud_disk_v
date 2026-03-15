// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/helper"
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
	//1.检验邮箱是否注册
	count, err := l.svcCtx.Engine.Where("email=?", req.Email).Count(&models.UserBasic{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该邮箱已被注册")
	}
	//2.未注册 发送验证码
	//2.1生成随机验证码
	code := helper.RandCode()
	//3.存储到redis
	err = l.svcCtx.RDB.Set(l.ctx, "code", code, time.Second*time.Duration(define.CodeExpireTime)).Err()
	if err != nil {
		return nil, err
	}
	err = helper.MailCodeSend(req.Email, code)
	return &types.MailCodeResponse{
		code,
	}, err
}
