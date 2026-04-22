package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/types"
	"fmt"
	"strings"
)

// listDuplicateFiles 用于“重复文件”视图。
// 这里按 repository_identity 聚合同一用户引用到的同一物理文件，
// 只返回重复次数大于 1 的文件，方便前端做去重治理。
func (l *UserFileListLogic) listDuplicateFiles(req *types.UserFileListRequest, userIdentity string, size int, offset int) (*types.UserFileListResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	baseWhere := `
FROM user_repository ur
LEFT JOIN repository_pool rp ON ur.repository_identity = rp.identity
INNER JOIN (
	SELECT
	  ur2.repository_identity,
	  COUNT(1) AS duplicate_count
	FROM user_repository ur2
	WHERE ur2.user_identity = ?
	  AND ur2.deleted_at IS NULL
	  AND ur2.is_dir = 0
	GROUP BY ur2.repository_identity
	HAVING COUNT(1) > 1
) dup ON dup.repository_identity = ur.repository_identity
WHERE ur.user_identity = ?
  AND ur.deleted_at IS NULL
  AND ur.is_dir = 0`
	args := []interface{}{userIdentity, userIdentity}

	if query := strings.TrimSpace(req.Query); query != "" {
		baseWhere += " AND ur.name LIKE ?"
		args = append(args, "%"+query+"%")
	}
	if req.FavoriteOnly {
		baseWhere += " AND ur.is_favorite = 1"
	}
	baseWhere += buildFileTypeCondition(req.FileType, "ur")

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
  dup.duplicate_count AS duplicate_count,
  COALESCE(rp.size, 0) * dup.duplicate_count AS duplicate_group_size
` + baseWhere + buildInsightOrderClause(req.OrderBy, req.OrderDir, "duplicates") + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	list := make([]*types.UserFile, 0)
	if err := sess.SQL(listSQL, args...).Find(&list); err != nil {
		return nil, errors.New(l.ctx, "query duplicate file list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
		})
	}

	var total struct {
		Count int64 `xorm:"'count'"`
	}
	if has, err := sess.SQL("SELECT COUNT(1) AS count "+baseWhere, args...).Get(&total); err != nil {
		return nil, errors.New(l.ctx, "count duplicate file list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
		})
	} else if !has {
		total.Count = 0
	}

	return &types.UserFileListResponse{
		List:  list,
		Count: total.Count,
	}, nil
}

// listLargeFiles 用于“大文件”视图。
// 这里固定只看真实文件，并允许前端通过 min_size_mb 控制治理阈值。
func (l *UserFileListLogic) listLargeFiles(req *types.UserFileListRequest, userIdentity string, size int, offset int) (*types.UserFileListResponse, error) {
	sess := l.svcCtx.Engine.NewSession()
	defer sess.Close()

	minSizeMB := req.MinSizeMB
	if minSizeMB <= 0 {
		minSizeMB = 100
	}
	minSizeBytes := minSizeMB * 1024 * 1024

	baseWhere := `
FROM user_repository ur
LEFT JOIN repository_pool rp ON ur.repository_identity = rp.identity
WHERE ur.user_identity = ?
  AND ur.deleted_at IS NULL
  AND ur.is_dir = 0
  AND COALESCE(rp.size, 0) >= ?`
	args := []interface{}{userIdentity, minSizeBytes}

	if query := strings.TrimSpace(req.Query); query != "" {
		baseWhere += " AND ur.name LIKE ?"
		args = append(args, "%"+query+"%")
	}
	if req.FavoriteOnly {
		baseWhere += " AND ur.is_favorite = 1"
	}
	baseWhere += buildFileTypeCondition(req.FileType, "ur")

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
` + baseWhere + buildInsightOrderClause(req.OrderBy, req.OrderDir, "large") + fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

	list := make([]*types.UserFile, 0)
	if err := sess.SQL(listSQL, args...).Find(&list); err != nil {
		return nil, errors.New(l.ctx, "query large file list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
			"min_size_mb":   minSizeMB,
		})
	}

	var total struct {
		Count int64 `xorm:"'count'"`
	}
	if has, err := sess.SQL("SELECT COUNT(1) AS count "+baseWhere, args...).Get(&total); err != nil {
		return nil, errors.New(l.ctx, "count large file list failed", err, map[string]interface{}{
			"user_identity": userIdentity,
			"min_size_mb":   minSizeMB,
		})
	} else if !has {
		total.Count = 0
	}

	return &types.UserFileListResponse{
		List:  list,
		Count: total.Count,
	}, nil
}

// buildInsightOrderClause 给治理视图提供更贴合场景的默认排序：
// - 重复文件优先看重复次数和占用空间
// - 大文件优先按文件大小倒序
func buildInsightOrderClause(orderBy string, orderDir string, view string) string {
	orderBy = strings.ToLower(strings.TrimSpace(orderBy))
	orderDir = strings.ToUpper(strings.TrimSpace(orderDir))
	if orderDir != "ASC" {
		orderDir = "DESC"
	}

	switch view {
	case "duplicates":
		switch orderBy {
		case "name":
			return fmt.Sprintf(" ORDER BY ur.name %s, ur.id DESC", orderDir)
		case "updated_at":
			return fmt.Sprintf(" ORDER BY ur.updated_at %s, ur.id DESC", orderDir)
		case "created_at":
			return fmt.Sprintf(" ORDER BY ur.created_at %s, ur.id DESC", orderDir)
		case "size":
			return fmt.Sprintf(" ORDER BY COALESCE(rp.size, 0) %s, dup.duplicate_count DESC, ur.id DESC", orderDir)
		default:
			return " ORDER BY dup.duplicate_count DESC, COALESCE(rp.size, 0) DESC, ur.id DESC"
		}
	case "large":
		if orderBy == "name" {
			return fmt.Sprintf(" ORDER BY ur.name %s, ur.id DESC", orderDir)
		}
		if orderBy == "updated_at" {
			return fmt.Sprintf(" ORDER BY ur.updated_at %s, ur.id DESC", orderDir)
		}
		if orderBy == "created_at" {
			return fmt.Sprintf(" ORDER BY ur.created_at %s, ur.id DESC", orderDir)
		}
		return fmt.Sprintf(" ORDER BY COALESCE(rp.size, 0) %s, ur.id DESC", orderDir)
	default:
		return ""
	}
}
