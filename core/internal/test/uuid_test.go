package test

import (
	"testing"

	"cloud_disk/core/internal/helper"
	"github.com/stretchr/testify/require"
)

func TestUUID(t *testing.T) {
	first := helper.UUID()
	second := helper.UUID()

	require.NotEmpty(t, first)
	require.NotEmpty(t, second)
	require.NotEqual(t, first, second)
}
