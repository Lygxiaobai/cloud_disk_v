package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRedisIntegrationSetGet(t *testing.T) {
	cfg := requireIntegration(t)
	svcCtx := newIntegrationServiceContext(t, cfg)

	err := svcCtx.RDB.Set(context.Background(), "key", "value", 10*time.Second).Err()
	require.NoError(t, err)

	val, err := svcCtx.RDB.Get(context.Background(), "key").Result()
	require.NoError(t, err)
	require.Equal(t, "value", val)
}
