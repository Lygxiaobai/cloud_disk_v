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

// ShareBasicListLogic 用于“我的分享”管理视图。
// 这里返回分享口令、下载开关、点击量和过期状态，方便前端集中治理。
type ShareBasicListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareBasicListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareBasicListLogic {
	return &ShareBasicListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareBasicListLogic) ShareBasicList(req *types.ShareBasicListRequest, userIdentity string) (*types.ShareBasicListResponse, error) {
	size := req.Size
	if size == 0 {
		size = define.PageSize
	}
	page := req.Page
	if page == 0 {
		page = define.Page
	}
	offset := (page - 1) * size

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	whereSQL := `
FROM share_basic sb
LEFT JOIN user_repository ur ON sb.user_repository_identity = ur.identity
LEFT JOIN repository_pool rp ON sb.repository_identity = rp.identity
WHERE sb.user_identity = ?
  AND sb.deleted_at IS NULL`
	args := []interface{}{userIdentity}
	if query := strings.TrimSpace(req.Query); query != "" {
		whereSQL += " AND ur.name LIKE ?"
		args = append(args, "%"+query+"%")
	}

	listSQL := `
SELECT
  sb.identity,
  sb.user_repository_identity AS user_file_identity,
  COALESCE(ur.name, rp.name, '') AS name,
  COALESCE(ur.ext, rp.ext, '') AS ext,
  COALESCE(rp.size, 0) AS size,
  sb.click_num,
  sb.allow_download,
  CASE WHEN sb.access_code <> '' THEN true ELSE false END AS access_code_set,
  DATE_FORMAT(sb.created_at, '%Y-%m-%d %H:%i:%s') AS created_at,
  CASE
    WHEN sb.expired_time <= 0 THEN ''
    ELSE DATE_FORMAT(DATE_ADD(sb.created_at, INTERVAL sb.expired_time SECOND), '%Y-%m-%d %H:%i:%s')
  END AS expires_at,
  CASE
    WHEN sb.expired_time > 0 AND DATE_ADD(sb.created_at, INTERVAL sb.expired_time SECOND) < NOW() THEN true
    ELSE false
  END AS expired
` + whereSQL + " ORDER BY sb.created_at DESC, sb.id DESC" + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	list := make([]*types.ShareBasicListItem, 0)
	if err := sess.SQL(listSQL, args...).Find(&list); err != nil {
		return nil, errors.New(l.ctx, "query share list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
		})
	}

	var total struct {
		Count int64 `xorm:"'count'"`
	}
	if has, err := sess.SQL("SELECT COUNT(1) AS count "+whereSQL, args...).Get(&total); err != nil {
		return nil, errors.New(l.ctx, "count share list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
		})
	} else if !has {
		total.Count = 0
	}

	return &types.ShareBasicListResponse{
		List:  list,
		Count: total.Count,
	}, nil
}
