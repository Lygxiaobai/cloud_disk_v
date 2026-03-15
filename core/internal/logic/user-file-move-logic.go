// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
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
	//将当前文件的parentId改到要移动的文件夹的id下即可
	//获取要移动文件夹的id
	var parentFolder = models.UserRepository{}
	has, err := l.svcCtx.Engine.Where("identity =?", req.ParentIdentity).Get(&parentFolder)
	if err != nil {
		return nil, err

	}
	if parentFolder.IsDir == 0 {
		return nil, errors.New("error")
	}
	if !has {
		return nil, errors.New("不存在的文件夹")
	}
	var cup = models.UserRepository{
		ParentId: int64(parentFolder.Id),
	}
	//更改当前要移动文件的parentId
	one, err := l.svcCtx.Engine.Where("identity =?", req.Identity).Update(&cup)
	if err != nil {
		return nil, err
	}
	if one == 0 {
		return nil, errors.New("不合法的操作")
	}
	return
}
