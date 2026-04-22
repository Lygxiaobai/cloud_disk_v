package test

import (
	"context"
	"testing"
	"time"

	appHelper "cloud_disk/core/internal/helper"
	"cloud_disk/core/internal/logic"
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/types"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRecentFilesIntegration(t *testing.T) {
	cfg := requireIntegration(t)
	svcCtx := newIntegrationServiceContext(t, cfg)
	ctx := context.Background()

	userIdentity := appHelper.UUID()
	repositoryA := appHelper.UUID()
	repositoryB := appHelper.UUID()
	fileA := appHelper.UUID()
	fileB := appHelper.UUID()

	_, err := svcCtx.Engine.Insert(
		&models.RepositoryPool{Identity: repositoryA, Name: "a.txt", Ext: ".txt", Size: 12, Path: "https://example.com/a.txt", Hash: appHelper.UUID()},
		&models.RepositoryPool{Identity: repositoryB, Name: "b.txt", Ext: ".txt", Size: 34, Path: "https://example.com/b.txt", Hash: appHelper.UUID()},
	)
	require.NoError(t, err)

	_, err = svcCtx.Engine.Insert(
		&models.UserRepository{Identity: fileA, UserIdentity: userIdentity, RepositoryIdentity: repositoryA, Name: "a.txt", Ext: ".txt"},
		&models.UserRepository{Identity: fileB, UserIdentity: userIdentity, RepositoryIdentity: repositoryB, Name: "b.txt", Ext: ".txt"},
	)
	require.NoError(t, err)

	key := "cloud_disk:recent:" + userIdentity
	now := time.Now()
	require.NoError(t, svcCtx.RDB.ZAdd(ctx, key,
		redis.Z{Score: float64(now.Add(-time.Minute).Unix()), Member: fileA},
		redis.Z{Score: float64(now.Unix()), Member: fileB},
	).Err())

	recentLogic := logic.NewUserRecentFileListLogic(ctx, svcCtx)
	resp, err := recentLogic.UserRecentFileList(&types.UserRecentFileListRequest{Limit: 10}, userIdentity)
	require.NoError(t, err)
	require.Len(t, resp.List, 2)
	require.Equal(t, fileB, resp.List[0].Identity)
	require.Equal(t, fileA, resp.List[1].Identity)
	require.NotEmpty(t, resp.List[0].LastAccessedAt)
}
