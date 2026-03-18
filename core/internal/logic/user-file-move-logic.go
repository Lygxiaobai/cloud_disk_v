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
	//将当前文件的parentId改到要移动的文件夹的id下即可
	//获取要移动文件夹的id
	var parentFolder = models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity =?", req.ParentIdentity).Get(&parentFolder)
	if err != nil {
		return nil, errors.New(l.ctx, "查询目标文件夹失败", err, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
		})

	}
	if parentFolder.IsDir == 0 {
		return nil, errors.New(l.ctx, "移动文件失败", nil, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
			"reason":          "目标不是文件夹",
		})
	}
	if !has {
		return nil, errors.New(l.ctx, "移动文件失败", nil, map[string]interface{}{
			"parent_identity": req.ParentIdentity,
			"reason":          "文件夹不存在",
		})
	}
	var cup = models.UserRepository{
		ParentId: int64(parentFolder.Id),
	}
	//更改当前要移动文件的parentId
	one, err := l.svcCtx.Engine.Where("identity =?", req.Identity).Update(&cup)
	if err != nil {
		return nil, errors.New(l.ctx, "更新文件位置失败", err, map[string]interface{}{
			"file_identity":   req.Identity,
			"parent_identity": req.ParentIdentity,
		})
	}
	if one == 0 {
		return nil, errors.New(l.ctx, "移动文件失败", nil, map[string]interface{}{
			"file_identity": req.Identity,
			"reason":        "文件不存在或无权限",
		})
	}
	return
}
