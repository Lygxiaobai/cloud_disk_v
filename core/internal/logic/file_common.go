package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/models"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

const (
	recentKeyPrefix = "cloud_disk:recent"
	recentKeepCount = 100
)

var fileTypeExtMap = map[string][]string{
	"image":    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico"},
	"video":    {".mp4", ".mov", ".avi", ".mkv", ".wmv", ".m4v", ".webm"},
	"audio":    {".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a"},
	"document": {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv", ".json"},
	"archive":  {".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"},
	"code":     {".go", ".js", ".ts", ".vue", ".java", ".py", ".c", ".cpp", ".h", ".hpp", ".php", ".html", ".css", ".scss", ".sql", ".xml", ".yaml", ".yml", ".sh"},
}

func resolveParentID(ctx context.Context, sess *xorm.Session, userIdentity string, parentID int64, parentIdentity string, forUpdate bool) (int64, error) {
	if parentIdentity == "" && parentID == 0 {
		return 0, nil
	}

	var (
		sql  string
		args []interface{}
	)
	if parentIdentity != "" {
		sql = "SELECT * FROM user_repository WHERE identity = ? AND user_identity = ? AND is_dir = 1 AND deleted_at IS NULL LIMIT 1"
		args = []interface{}{parentIdentity, userIdentity}
	} else {
		sql = "SELECT * FROM user_repository WHERE id = ? AND user_identity = ? AND is_dir = 1 AND deleted_at IS NULL LIMIT 1"
		args = []interface{}{parentID, userIdentity}
	}
	if forUpdate {
		sql += " FOR UPDATE"
	}

	var folder models.UserRepository
	has, err := sess.SQL(sql, args...).Get(&folder)
	if err != nil {
		return 0, errors.New(ctx, "query parent folder failed", err, map[string]interface{}{
			"parent_id":       parentID,
			"parent_identity": parentIdentity,
		})
	}
	if !has {
		return 0, errors.New(ctx, "parent folder does not exist", nil, map[string]interface{}{
			"parent_id":       parentID,
			"parent_identity": parentIdentity,
		})
	}
	return int64(folder.Id), nil
}

func getUserRepository(ctx context.Context, sess *xorm.Session, userIdentity string, identity string, includeDeleted bool, forUpdate bool) (*models.UserRepository, error) {
	rows, err := getUserRepositoriesByIdentities(ctx, sess, userIdentity, []string{identity}, includeDeleted, forUpdate)
	if err != nil {
		return nil, err
	}
	row, ok := rows[identity]
	if !ok {
		return nil, errors.New(ctx, "file does not exist", nil, map[string]interface{}{
			"identity": identity,
		})
	}
	return row, nil
}

func getUserRepositoriesByIdentities(ctx context.Context, sess *xorm.Session, userIdentity string, identities []string, includeDeleted bool, forUpdate bool) (map[string]*models.UserRepository, error) {
	identities = uniqueStrings(identities)
	result := make(map[string]*models.UserRepository, len(identities))
	if len(identities) == 0 {
		return result, nil
	}

	sql := fmt.Sprintf(
		"SELECT * FROM user_repository WHERE user_identity = ? AND identity IN (%s)",
		placeholders(len(identities)),
	)
	args := make([]interface{}, 0, len(identities)+1)
	args = append(args, userIdentity)
	for _, identity := range identities {
		args = append(args, identity)
	}
	if !includeDeleted {
		sql += " AND deleted_at IS NULL"
	}
	sql += " ORDER BY id ASC"
	if forUpdate {
		sql += " FOR UPDATE"
	}

	rows := make([]models.UserRepository, 0, len(identities))
	if err := sess.SQL(sql, args...).Find(&rows); err != nil {
		return nil, errors.New(ctx, "query files failed", err, map[string]interface{}{
			"identities": identities,
		})
	}
	for index := range rows {
		row := rows[index]
		result[row.Identity] = &row
	}
	return result, nil
}

func getUserRepositoryByIDRaw(ctx context.Context, sess *xorm.Session, userIdentity string, id int64, forUpdate bool) (*models.UserRepository, error) {
	row := &models.UserRepository{}
	sql := "SELECT * FROM user_repository WHERE id = ? AND user_identity = ? LIMIT 1"
	if forUpdate {
		sql += " FOR UPDATE"
	}
	has, err := sess.SQL(sql, id, userIdentity).Get(row)
	if err != nil {
		return nil, errors.New(ctx, "query file failed", err, map[string]interface{}{
			"id": id,
		})
	}
	if !has {
		return nil, errors.New(ctx, "file does not exist", nil, map[string]interface{}{
			"id": id,
		})
	}
	return row, nil
}

func ensureNameAvailable(ctx context.Context, sess *xorm.Session, userIdentity string, parentID int64, name string, excludeIdentity string) error {
	sql := "SELECT COUNT(1) AS count FROM user_repository WHERE user_identity = ? AND parent_id = ? AND name = ? AND deleted_at IS NULL"
	args := []interface{}{userIdentity, parentID, name}
	if excludeIdentity != "" {
		sql += " AND identity <> ?"
		args = append(args, excludeIdentity)
	}
	sql += " FOR UPDATE"

	var row struct {
		Count int64 `xorm:"'count'"`
	}
	has, err := sess.SQL(sql, args...).Get(&row)
	if err != nil {
		return errors.New(ctx, "check duplicate name failed", err, map[string]interface{}{
			"parent_id": parentID,
			"name":      name,
		})
	}
	if has && row.Count > 0 {
		return errors.New(ctx, "same name already exists in current folder", nil, map[string]interface{}{
			"parent_id": parentID,
			"name":      name,
		})
	}
	return nil
}

func loadExistingNames(ctx context.Context, sess *xorm.Session, userIdentity string, parentID int64, names []string, includeDeleted bool, forUpdate bool) (map[string][]models.UserRepository, error) {
	result := make(map[string][]models.UserRepository)
	names = uniqueStrings(names)
	if len(names) == 0 {
		return result, nil
	}

	sql := fmt.Sprintf(
		"SELECT * FROM user_repository WHERE user_identity = ? AND parent_id = ? AND name IN (%s)",
		placeholders(len(names)),
	)
	args := make([]interface{}, 0, len(names)+2)
	args = append(args, userIdentity, parentID)
	for _, name := range names {
		args = append(args, name)
	}
	if !includeDeleted {
		sql += " AND deleted_at IS NULL"
	}
	sql += " ORDER BY id ASC"
	if forUpdate {
		sql += " FOR UPDATE"
	}

	rows := make([]models.UserRepository, 0)
	if err := sess.SQL(sql, args...).Find(&rows); err != nil {
		return nil, errors.New(ctx, "query sibling files failed", err, map[string]interface{}{
			"parent_id": parentID,
			"names":     names,
		})
	}
	for _, row := range rows {
		result[row.Name] = append(result[row.Name], row)
	}
	return result, nil
}

func collectSubtree(ctx context.Context, sess *xorm.Session, userIdentity string, rootIdentity string, includeDeleted bool, forUpdate bool) ([]models.UserRepository, error) {
	root, err := getUserRepository(ctx, sess, userIdentity, rootIdentity, includeDeleted, forUpdate)
	if err != nil {
		return nil, err
	}

	condition := ""
	if !includeDeleted {
		condition = " AND deleted_at IS NULL"
	}

	sql := fmt.Sprintf(`
WITH RECURSIVE subtree AS (
	SELECT *
	FROM user_repository
	WHERE id = ? AND user_identity = ?%s
	UNION ALL
	SELECT child.*
	FROM user_repository child
	INNER JOIN subtree parent ON child.parent_id = parent.id
	WHERE child.user_identity = ?%s
)
SELECT *
FROM subtree
ORDER BY id ASC`, condition, condition)
	if forUpdate {
		sql += " FOR UPDATE"
	}

	rows := make([]models.UserRepository, 0)
	if err := sess.SQL(sql, root.Id, userIdentity, userIdentity).Find(&rows); err != nil {
		return nil, errors.New(ctx, "query child files failed", err, map[string]interface{}{
			"root_identity": rootIdentity,
		})
	}
	return rows, nil
}

func collectSubtreeIdentities(ctx context.Context, sess *xorm.Session, userIdentity string, rootIdentities []string, includeDeleted bool, forUpdate bool) ([]string, error) {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, identity := range uniqueStrings(rootIdentities) {
		items, err := collectSubtree(ctx, sess, userIdentity, identity, includeDeleted, forUpdate)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if _, ok := seen[item.Identity]; ok {
				continue
			}
			seen[item.Identity] = struct{}{}
			result = append(result, item.Identity)
		}
	}
	return result, nil
}

func buildFileTypeCondition(fileType string, fileAlias string) string {
	fileType = strings.ToLower(strings.TrimSpace(fileType))
	switch fileType {
	case "", "all":
		return ""
	case "dir", "folder":
		return fmt.Sprintf(" AND %s.is_dir = 1", fileAlias)
	case "file":
		return fmt.Sprintf(" AND %s.is_dir = 0", fileAlias)
	}

	allExts := make([]string, 0)
	for _, exts := range fileTypeExtMap {
		allExts = append(allExts, exts...)
	}

	if fileType == "other" {
		return fmt.Sprintf(" AND %s.is_dir = 0 AND (LOWER(%s.ext) = '' OR LOWER(%s.ext) NOT IN (%s))",
			fileAlias,
			fileAlias,
			fileAlias,
			quoteStrings(allExts),
		)
	}

	exts, ok := fileTypeExtMap[fileType]
	if !ok {
		return ""
	}

	return fmt.Sprintf(" AND %s.is_dir = 0 AND LOWER(%s.ext) IN (%s)", fileAlias, fileAlias, quoteStrings(exts))
}

func buildOrderClause(orderBy string, orderDir string, userAlias string, repoAlias string, trash bool) string {
	orderBy = strings.ToLower(strings.TrimSpace(orderBy))
	orderDir = strings.ToUpper(strings.TrimSpace(orderDir))
	if orderDir != "ASC" {
		orderDir = "DESC"
	}

	columnMap := map[string]string{
		"name":       userAlias + ".name",
		"size":       repoAlias + ".size",
		"created_at": userAlias + ".created_at",
		"updated_at": userAlias + ".updated_at",
		"deleted_at": userAlias + ".deleted_at",
	}

	column, ok := columnMap[orderBy]
	if !ok {
		if trash {
			return fmt.Sprintf(" ORDER BY %s.deleted_at DESC, %s.id DESC", userAlias, userAlias)
		}
		return fmt.Sprintf(" ORDER BY %s.is_dir DESC, %s.updated_at DESC, %s.id DESC", userAlias, userAlias, userAlias)
	}

	if orderBy == "size" {
		return fmt.Sprintf(" ORDER BY %s.is_dir DESC, COALESCE(%s, 0) %s, %s.id DESC", userAlias, column, orderDir, userAlias)
	}
	return fmt.Sprintf(" ORDER BY %s %s, %s.id DESC", column, orderDir, userAlias)
}

func isPreviewText(ext string) bool {
	textTypes := []string{
		".txt", ".md", ".json", ".xml", ".yaml", ".yml", ".log", ".ini", ".csv",
		".go", ".js", ".ts", ".vue", ".java", ".py", ".c", ".cpp", ".h", ".hpp",
		".css", ".scss", ".html", ".sql", ".sh",
	}
	return slices.Contains(textTypes, strings.ToLower(ext))
}

func previewKind(ext string) string {
	ext = strings.ToLower(ext)
	switch {
	case slices.Contains(fileTypeExtMap["image"], ext):
		return "image"
	case slices.Contains(fileTypeExtMap["video"], ext):
		return "video"
	case slices.Contains(fileTypeExtMap["audio"], ext):
		return "audio"
	case ext == ".pdf":
		return "pdf"
	case isPreviewText(ext):
		return "text"
	default:
		return "download"
	}
}

func addRecentFile(ctx context.Context, rdb *redis.Client, userIdentity string, fileIdentity string) {
	if rdb == nil || userIdentity == "" || fileIdentity == "" {
		return
	}

	key := recentRedisKey(userIdentity)
	now := float64(time.Now().Unix())
	rdb.ZAdd(ctx, key, redis.Z{Score: now, Member: fileIdentity})
	rdb.ZRemRangeByRank(ctx, key, 0, -recentKeepCount-1)
}

func recentRedisKey(userIdentity string) string {
	return recentKeyPrefix + ":" + userIdentity
}

func quoteStrings(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("'%s'", strings.ToLower(value)))
	}
	return strings.Join(quoted, ",")
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		values = append(values, "?")
	}
	return strings.Join(values, ",")
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
