// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"
	"errors"

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
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "GET")
	ctx = context.WithValue(ctx, "path", "/user/detail")
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//1.根据用户的唯一标识查询出用户 然后返回用户信息
	user := new(models.UserBasic)
	has, err := l.svcCtx.Engine.Where("identity = ?", req.Identity).Get(user)
	if err != nil {
		logger.LogError(ctx, "查询用户详情失败", err, map[string]interface{}{
			"user_identity": req.Identity,
		})
		return nil, err
	}
	if !has {
		err = errors.New("不合法的用户")
		logger.LogError(ctx, "查询用户详情失败", err, map[string]interface{}{
			"user_identity": req.Identity,
			"reason":        "用户不存在",
		})
		return nil, err
	}
	resp = new(types.UserDetailResponse)
	resp.Name = user.Name
	resp.Email = user.Email
	return
}
