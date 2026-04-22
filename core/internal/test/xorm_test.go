package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLIntegrationPing(t *testing.T) {
	cfg := requireIntegration(t)
	svcCtx := newIntegrationServiceContext(t, cfg)
	require.NoError(t, svcCtx.Engine.Ping())
}
