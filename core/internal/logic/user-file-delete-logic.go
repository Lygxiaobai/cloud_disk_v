// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/logger"
	"cloud_disk/core/internal/models"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFileDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileDeleteLogic {
	return &UserFileDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileDeleteLogic) UserFileDelete(req *types.UserFileDeleteRequest, userIdentity string) (resp *types.UserFileDeleteResponse, err error) {
	// 从 context 中获取 TraceID（由中间件设置）
	traceID, _ := l.ctx.Value("trace_id").(string)

	// 构建上下文信息
	ctx := context.WithValue(l.ctx, "method", "DELETE")
	ctx = context.WithValue(ctx, "path", "/user/file/delete")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//逻辑删除文件
	_, err = l.svcCtx.Engine.Where("identity = ? AND user_identity=?", req.Identity, userIdentity).Delete(&models.UserRepository{})
	if err != nil {
		logger.LogError(ctx, "删除文件失败", err, map[string]interface{}{
			"file_identity": req.Identity,
			"user_identity": userIdentity,
		})
		return nil, err
	}

	return
}
