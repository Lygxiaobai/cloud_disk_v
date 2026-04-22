package logic

import (
	"cloud_disk/core/internal/models"
	"cloud_disk/core/internal/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildBatchRenameNameKeepsExtension(t *testing.T) {
	file := &models.UserRepository{Name: "report.pdf"}
	req := &types.UserFileBatchRenameRequest{
		Prefix:        "archived-",
		ApplySequence: true,
		StartIndex:    1,
		Padding:       2,
	}

	name := buildBatchRenameName(file, req, true, 1)
	require.Equal(t, "archived-report01.pdf", name)
}

func TestBuildBatchRenameNameForDirectory(t *testing.T) {
	file := &models.UserRepository{Name: "docs", IsDir: 1}
	req := &types.UserFileBatchRenameRequest{
		FindText:    "docs",
		ReplaceText: "manual",
		Suffix:      "-2026",
	}

	name := buildBatchRenameName(file, req, true, 1)
	require.Equal(t, "manual-2026", name)
}

func TestBuildBatchRenameNameFallsBackWhenEmpty(t *testing.T) {
	file := &models.UserRepository{Name: "demo.txt"}
	req := &types.UserFileBatchRenameRequest{
		FindText:    "demo",
		ReplaceText: "",
	}

	name := buildBatchRenameName(file, req, true, 1)
	require.Equal(t, "unnamed.txt", name)
}
