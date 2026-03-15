// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareFileDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareFileDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareFileDetailLogic {
	return &ShareFileDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareFileDetailLogic) ShareFileDetail(req *types.ShareFileDetailRequest) (resp *types.ShareFileDetailResponse, err error) {
	//每次点击分享链接时，次数加一
	_, err = l.svcCtx.Engine.Exec("update share_basic set click_num = click_num + 1 where identity = ?", req.Identity)
	if err != nil {
		return nil, err
	}
	//1.从shareIdentity中获取导Repository信息
	resp = &types.ShareFileDetailResponse{}
	_, err = l.svcCtx.Engine.Table("share_basic").
		Select("share_basic.repository_identity, user_repository.name, user_repository.ext,repository_pool.size,repository_pool.path").
		Join("LEFT", "repository_pool", "share_basic.repository_identity = repository_pool.identity").
		Join("LEFT", "user_repository", "share_basic.user_repository_identity = user_repository.identity").
		Where("share_basic.identity = ?", req.Identity).Get(resp)

	return
}
