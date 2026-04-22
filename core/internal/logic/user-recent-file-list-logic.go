package logic

import (
	"cloud_disk/core/internal/errors"
	"cloud_disk/core/internal/svc"
	"cloud_disk/core/internal/types"
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRecentFileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRecentFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRecentFileListLogic {
	return &UserRecentFileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRecentFileListLogic) UserRecentFileList(req *types.UserRecentFileListRequest, userIdentity string) (*types.UserRecentFileListResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	key := recentRedisKey(userIdentity)
	values, err := l.svcCtx.RDB.ZRevRangeWithScores(l.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, errors.New(l.ctx, "query recent files failed", err, nil)
	}

	identities := make([]string, 0, len(values))
	accessedAtMap := make(map[string]string, len(values))
	for _, item := range values {
		identity, ok := item.Member.(string)
		if !ok || identity == "" {
			continue
		}
		identities = append(identities, identity)
		accessedAtMap[identity] = time.Unix(int64(item.Score), 0).Format("2006-01-02 15:04:05")
	}
	if len(identities) == 0 {
		return &types.UserRecentFileListResponse{List: []*types.UserFile{}}, nil
	}

	sql := `
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
FROM user_repository ur
LEFT JOIN repository_pool rp ON ur.repository_identity = rp.identity
WHERE ur.user_identity = ?
  AND ur.identity IN (` + placeholders(len(identities)) + `)
  AND ur.deleted_at IS NULL`

	args := make([]interface{}, 0, len(identities)+1)
	args = append(args, userIdentity)
	for _, identity := range identities {
		args = append(args, identity)
	}

	rows := make([]*types.UserFile, 0, len(identities))
	if err := l.svcCtx.Engine.SQL(sql, args...).Find(&rows); err != nil {
		return nil, errors.New(l.ctx, "query recent file detail failed", err, map[string]interface{}{
			"identities": identities,
		})
	}

	byIdentity := make(map[string]*types.UserFile, len(rows))
	for _, row := range rows {
		row.LastAccessedAt = accessedAtMap[row.Identity]
		byIdentity[row.Identity] = row
	}

	list := make([]*types.UserFile, 0, len(identities))
	for _, identity := range identities {
		if row, ok := byIdentity[identity]; ok {
			list = append(list, row)
		}
	}

	return &types.UserRecentFileListResponse{List: list}, nil
}
