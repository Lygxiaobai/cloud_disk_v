package logic

import (
	"cloud_disk/core/internal/define"
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserRecycleListLogic 负责回收站列表查询。
type UserRecycleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRecycleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRecycleListLogic {
	return &UserRecycleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserRecycleList 和正常文件列表类似，但查询对象限定为 deleted_at 不为空的记录。
func (l *UserRecycleListLogic) UserRecycleList(req *types.UserRecycleListRequest, userIdentity string) (*types.UserRecycleListResponse, error) {
	size := req.Size
	if size == 0 {
		size = define.PageSize
	}
	page := req.Page
	if page == 0 {
		page = define.Page
	}
	offset := (page - 1) * size

	whereSQL := `
FROM user_repository ur
LEFT JOIN repository_pool rp ON ur.repository_identity = rp.identity
WHERE ur.user_identity = ?
  AND ur.deleted_at IS NOT NULL`
	args := []interface{}{userIdentity}

	if query := strings.TrimSpace(req.Query); query != "" {
		whereSQL += " AND ur.name LIKE ?"
		args = append(args, "%"+query+"%")
	}

	listSQL := `
SELECT
  ur.id,
  ur.identity,
  ur.repository_identity,
  ur.name,
  ur.ext,
  COALESCE(rp.path, '') AS path,
  COALESCE(rp.size, 0) AS size,
  ur.is_dir,
  ur.is_favorite,
  DATE_FORMAT(ur.created_at, '%Y-%m-%d %H:%i:%s') AS created_at,
  DATE_FORMAT(ur.updated_at, '%Y-%m-%d %H:%i:%s') AS updated_at,
  DATE_FORMAT(ur.deleted_at, '%Y-%m-%d %H:%i:%s') AS deleted_at
` + whereSQL + buildOrderClause(req.OrderBy, req.OrderDir, "ur", "rp", true) + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	list := make([]*types.UserFile, 0)
	if err := l.svcCtx.Engine.SQL(listSQL, args...).Find(&list); err != nil {
		return nil, errors.New(l.ctx, "query recycle list failed", err, nil)
	}

	var total struct {
		Count int64 `xorm:"'count'"`
	}
	if has, err := l.svcCtx.Engine.SQL("SELECT COUNT(1) AS count "+whereSQL, args...).Get(&total); err != nil {
		return nil, errors.New(l.ctx, "count recycle list failed", err, nil)
	} else if !has {
		total.Count = 0
	}

	return &types.UserRecycleListResponse{
		List:  list,
		Count: total.Count,
	}, nil
}
