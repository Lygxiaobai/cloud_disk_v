package test

import (
	"testing"

	"cloud_disk/core/internal/helper"
	"github.com/stretchr/testify/require"
)

func TestRandCode(t *testing.T) {
	code := helper.RandCode(6)
	require.Len(t, code, 6)
	require.Regexp(t, "^[0-9]{6}$", code)
}
