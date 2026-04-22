// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"cloud_disk/core/internal/errors"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFolderChildrenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFolderChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderChildrenLogic {
	return &UserFolderChildrenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFolderChildrenLogic) UserFolderChildren(req *types.UserFolderChildrenRequest, userIdentity string) (resp *types.UserFolderChildrenResponse, err error) {
	// 能进来当前这个处理函数的都是文件夹
	//本函数 就是1.找当前文件夹的同级文件夹 2.判断当前文件夹是否有子文件夹 然后保存信息到UserFolderNode中
	list := make([]*types.UserFolderNode, 0)
	sql := `
			SELECT
				ur.id,
				ur.identity,
				ur.parent_id,
				ur.name,
				CASE WHEN EXISTS (
					SELECT 1
					FROM user_repository child
					WHERE child.parent_id = ur.id
					  AND child.user_identity = ur.user_identity
					  AND child.is_dir = 1
					  AND child.deleted_at IS NULL
				) THEN 1 ELSE 0 END AS has_children
			FROM user_repository ur
			WHERE ur.user_identity = ?
			  AND ur.parent_id = ?
			  AND ur.is_dir = 1
			  AND ur.deleted_at IS NULL
			ORDER BY ur.id ASC
			`

	err = l.svcCtx.Engine.SQL(sql, userIdentity, req.Id).Find(&list)
	if err != nil {
		return nil, errors.New(l.ctx, "查询文件夹子项失败", err, map[string]interface{}{
			"parent_id": req.Id,
		})
	}
	resp = &types.UserFolderChildrenResponse{}
	resp.List = list
	return
}
