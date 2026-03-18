// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDetailLogic {
	return &UserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDetailLogic) UserDetail(req *types.UserDetailRequest) (resp *types.UserDetailResponse, err error) {
	//1.根据用户的唯一标识查询出用户 然后返回用户信息
	user := new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("identity = ?", req.Identity).Get(user)
	if err != nil {
		return nil, errors.New(l.ctx, "查询用户详情失败", err, map[string]interface{}{
			"user_identity": req.Identity,
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "查询用户详情失败", nil, map[string]interface{}{
			"user_identity": req.Identity,
			"reason":        "用户不存在",
		})
	}
	resp = new(types.UserDetailResponse)
	resp.Name = user.Name
	resp.Email = user.Email
	return
}
