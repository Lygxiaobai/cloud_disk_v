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

type UserFileMoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileMoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileMoveLogic {
	return &UserFileMoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFileMoveLogic) UserFileMove(req *types.UserFileMoveRequest, userIdentity string) (resp *types.UserFileMoveResponse, err error) {
	// 从 context 中获取 TraceID
	traceID, _ := l.ctx.Value("trace_id").(string)
	ctx := context.WithValue(l.ctx, "method", "PUT")
	ctx = context.WithValue(ctx, "path", "/user/file/move")
	ctx = context.WithValue(ctx, "user_identity", userIdentity)
	ctx = context.WithValue(ctx, "trace_id", traceID)

	//将当前文件的parentId改到要移动的文件夹的id下即可
	//获取要移动文件夹的id
	var parentFolder = models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity =?", req.ParentIdentity).Get(&parentFolder)
	if err != nil {
		logger.LogError(ctx, "查询目标文件夹失败", err, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
		})
		return nil, err

	}
	if parentFolder.IsDir == 0 {
		err = errors.New("目标不是文件夹")
		logger.LogError(ctx, "移动文件失败", err, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
			"reason":          "目标不是文件夹",
		})
		return nil, err
	}
	if !has {
		err = errors.New("不存在的文件夹")
		logger.LogError(ctx, "移动文件失败", err, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
			"reason":          "文件夹不存在",
		})
		return nil, err
	}
	var cup = models.UserRepository{
		ParentId: int64(parentFolder.Id),
	}
	//更改当前要移动文件的parentId
	one, err := l.svcCtx.Engine.Where("identity =?", req.Identity).Update(&cup)
	if err != nil {
		logger.LogError(ctx, "更新文件位置失败", err, map[string]interface{}{
			"file_identity":   req.Identity,
			"parent_identity": req.ParentIdentity,
		})
		return nil, err
	}
	if one == 0 {
		err = errors.New("不合法的操作")
		logger.LogError(ctx, "移动文件失败", err, map[string]interface{}{
			"file_identity": req.Identity,
			"reason":        "文件不存在或无权限",
		})
		return nil, err
	}
	return
}
