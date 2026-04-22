package errors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorResponseUsesLogicErrorCode(t *testing.T) {
	err := NotFound(context.Background(), CodeFileNotFound, "file does not exist", nil, nil)

	status, body := ErrorResponse(context.Background(), err)
	resp := body.(Response)

	require.Equal(t, 404, status)
	require.Equal(t, CodeFileNotFound, resp.Code)
	require.Equal(t, "file does not exist", resp.Message)
}

func TestErrorResponseClassifiesGenericError(t *testing.T) {
	status, body := ErrorResponse(context.Background(), context.DeadlineExceeded)
	resp := body.(Response)

	require.Equal(t, 500, status)
	require.Equal(t, CodeInternalError, resp.Code)
}

func TestErrorResponseUsesExplicitRateLimitCode(t *testing.T) {
	err := RateLimited(context.Background(), "too many requests", nil, nil)

	status, body := ErrorResponse(context.Background(), err)
	resp := body.(Response)

	require.Equal(t, 429, status)
	require.Equal(t, CodeRateLimited, resp.Code)
	require.Equal(t, "too many requests", resp.Message)
}
