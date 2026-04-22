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

// UserFileListLogic 负责网盘主列表查询。
// 当前版本把搜索、筛选、排序都统一收口到了这个接口中。
type UserFileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserFileListLogic {
	return &UserFileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserFileList 查询用户当前目录下的文件列表，并支持：
// 1. 目录定位
// 2. 文件名搜索
// 3. 文件类型筛选
// 4. 收藏筛选
// 5. 排序
// 6. 分页
//
// 新增能力：
// 7. scope=all 时执行全局搜索 / 全局浏览
// 8. view=duplicates / large 时切到文件治理视图
func (l *UserFileListLogic) UserFileList(req *types.UserFileListRequest, userIdentity string) (resp *types.UserFileListResponse, err error) {
	size := req.Size
	if size == 0 {
		size = define.PageSize
	}
	page := req.Page
	if page == 0 {
		page = define.Page
	}
	offset := (page - 1) * size

	// “文件治理”视图不再限定当前目录，而是按全盘维度返回治理结果。
	switch strings.ToLower(strings.TrimSpace(req.View)) {
	case "duplicates":
		return l.listDuplicateFiles(req, userIdentity, size, offset)
	case "large":
		return l.listLargeFiles(req, userIdentity, size, offset)
	}

	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	// 先拼公共的 where 部分，后面再逐步追加搜索、收藏和类型过滤条件。
	whereSQL := `
FROM user_repository ur
LEFT JOIN repository_pool rp ON ur.repository_identity = rp.identity
WHERE ur.user_identity = ?
  AND ur.deleted_at IS NULL`
	args := []interface{}{userIdentity}

	// scope=all 时启用全局搜索 / 全局文件浏览，否则仍保留目录上下文。
	parentID := int64(0)
	globalScope := strings.ToLower(strings.TrimSpace(req.Scope)) == "all"
	if !globalScope {
		parentID, err = resolveParentID(l.ctx, sess, userIdentity, req.Id, req.Identity, false)
		if err != nil {
			return nil, err
		}
		whereSQL += " AND ur.parent_id = ?"
		args = append(args, parentID)
	}

	if query := strings.TrimSpace(req.Query); query != "" {
		whereSQL += " AND ur.name LIKE ?"
		args = append(args, "%"+query+"%")
	}
	if req.FavoriteOnly {
		whereSQL += " AND ur.is_favorite = 1"
	}
	whereSQL += buildFileTypeCondition(req.FileType, "ur")

	// 列表查询与总数查询共用同一套 where 条件，避免数据和统计不一致。
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
  DATE_FORMAT(ur.updated_at, '%Y-%m-%d %H:%i:%s') AS updated_at
` + whereSQL + buildOrderClause(req.OrderBy, req.OrderDir, "ur", "rp", false) + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	list := make([]*types.UserFile, 0)
	if err := sess.SQL(listSQL, args...).Find(&list); err != nil {
		errorFields := map[string]interface{}{
			"scope": req.Scope,
		}
		if !globalScope {
			errorFields["parent_id"] = parentID
		}
		return nil, errors.New(l.ctx, "query file list failed", err, errorFields)
	}

	var total struct {
		Count int64 `xorm:"'count'"`
	}
	if has, err := sess.SQL("SELECT COUNT(1) AS count "+whereSQL, args...).Get(&total); err != nil {
		errorFields := map[string]interface{}{
			"scope": req.Scope,
		}
		if !globalScope {
			errorFields["parent_id"] = parentID
		}
		return nil, errors.New(l.ctx, "count file list failed", err, errorFields)
	} else if !has {
		total.Count = 0
	}

	return &types.UserFileListResponse{
		List:  list,
		Count: total.Count,
	}, nil
}
