package logic

import (
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserFileMoveLogic 是单文件移动接口的适配层。
// 它内部直接复用批量移动逻辑，这样单个和批量的校验规则能保持一致。
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

func (l *UserFileMoveLogic) UserFileMove(req *types.UserFileMoveRequest, userIdentity string) (*types.UserFileMoveResponse, error) {
	_, err := NewUserFileBatchMoveLogic(l.ctx, l.svcCtx).UserFileBatchMove(&types.UserFileBatchMoveRequest{
		Identities:     []string{req.Identity},
		ParentIdentity: req.ParentIdentity,
	}, userIdentity)
	if err != nil {
		return nil, err
	}
	return &types.UserFileMoveResponse{}, nil
}
