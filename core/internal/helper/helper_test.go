package helper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildMultipartRangesSingleChunk(t *testing.T) {
	ranges, err := buildMultipartRanges(2, multipartChunkSize)
	require.NoError(t, err)
	require.Len(t, ranges, 1)
	require.Equal(t, multipartRange{PartNumber: 1, Start: 0, EndExclusive: 2}, ranges[0])
}

func TestBuildMultipartRangesSplitByChunkSize(t *testing.T) {
	ranges, err := buildMultipartRanges(11, 5)
	require.NoError(t, err)
	require.Equal(t, []multipartRange{
		{PartNumber: 1, Start: 0, EndExclusive: 5},
		{PartNumber: 2, Start: 5, EndExclusive: 10},
		{PartNumber: 3, Start: 10, EndExclusive: 11},
	}, ranges)
}

func TestBuildMultipartRangesRejectsEmptyFile(t *testing.T) {
	_, err := buildMultipartRanges(0, multipartChunkSize)
	require.Error(t, err)
}

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret-123")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, "secret-123", hash)
	require.NoError(t, CheckPassword(hash, "secret-123"))
	require.Error(t, CheckPassword(hash, "wrong"))
}

func TestGenerateAndAnalyzeToken(t *testing.T) {
	token, err := GenerateToken(7, "u-1", "tester", "admin", "secret-key", 60)
	require.NoError(t, err)

	claims, err := AnalyzeToken(token, "secret-key")
	require.NoError(t, err)
	require.Equal(t, 7, claims.ID)
	require.Equal(t, "u-1", claims.Identity)
	require.Equal(t, "tester", claims.Name)
	require.Equal(t, "admin", claims.Role)
}

func TestValidateUploadExtRejectsBlockedExtension(t *testing.T) {
	err := ValidateUploadExt(".exe", []string{".exe", ".bat"})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ".exe"))
}
