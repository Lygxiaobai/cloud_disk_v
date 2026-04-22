package logic

import (
	"cloud_disk/core/internal/errors"
	"context"

	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserFolderPathLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFolderPathLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFolderPathLogic {
	return &UserFolderPathLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserFolderPathLogic) UserFolderPath(req *types.UserFolderPathRequest, userIdentity string) (*types.UserFolderPathResponse, error) {
	root := &types.FolderPathItem{
		Id:       0,
		Identity: "",
		Name:     "全部文件",
	}
	if req.Identity == "" {
		return &types.UserFolderPathResponse{List: []*types.FolderPathItem{root}}, nil
	}

	sql := `
WITH RECURSIVE folder_path AS (
	SELECT id, identity, parent_id, name, 0 AS depth
	FROM user_repository
	WHERE identity = ? AND user_identity = ? AND is_dir = 1 AND deleted_at IS NULL
	UNION ALL
	SELECT parent.id, parent.identity, parent.parent_id, parent.name, folder_path.depth + 1
	FROM user_repository parent
	INNER JOIN folder_path ON folder_path.parent_id = parent.id
	WHERE parent.user_identity = ? AND parent.is_dir = 1 AND parent.deleted_at IS NULL
)
SELECT id, identity, name
FROM folder_path
ORDER BY depth DESC`

	rows := make([]*types.FolderPathItem, 0)
	if err := l.svcCtx.Engine.SQL(sql, req.Identity, userIdentity, userIdentity).Find(&rows); err != nil {
		return nil, errors.New(l.ctx, "查询文件夹路径失败", err, map[string]interface{}{
			"folder_identity": req.Identity,
		})
	}
	if len(rows) == 0 {
		return nil, errors.New(l.ctx, "查询文件夹路径失败", nil, map[string]interface{}{
			"folder_identity": req.Identity,
			"reason":          "folder does not exist",
		})
	}

	list := make([]*types.FolderPathItem, 0, len(rows)+1)
	list = append(list, root)
	list = append(list, rows...)
	return &types.UserFolderPathResponse{List: list}, nil
}
