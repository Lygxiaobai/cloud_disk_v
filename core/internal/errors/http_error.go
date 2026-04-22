package errors

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}

func ErrorResponse(ctx context.Context, err error) (int, any) {
	if err == nil {
		return http.StatusInternalServerError, Response{
			Code:    CodeInternalError,
			Message: "internal server error",
			TraceID: traceIDFromContext(ctx),
		}
	}

	var logicErr *LogicError
	if stderrors.As(err, &logicErr) {
		return logicErr.StatusCode(), Response{
			Code:    logicErr.BusinessCode(),
			Message: logicErr.Error(),
			TraceID: traceIDFromContext(ctx),
		}
	}

	status, code := classify(err.Error(), err)
	return status, Response{
		Code:    code,
		Message: err.Error(),
		TraceID: traceIDFromContext(ctx),
	}
}

func classify(message string, cause error) (int, int) {
	lower := strings.ToLower(strings.TrimSpace(message))

	switch {
	case lower == "":
		return http.StatusInternalServerError, CodeInternalError
	case strings.Contains(lower, "too many requests"),
		strings.Contains(lower, "rate limit"),
		strings.Contains(lower, "retry-after"),
		strings.Contains(lower, "locked"):
		return http.StatusTooManyRequests, CodeRateLimited
	case strings.Contains(lower, "identity or authorization is required"),
		strings.Contains(lower, "auth"),
		strings.Contains(lower, "authorization"),
		strings.Contains(lower, "token"):
		return http.StatusUnauthorized, CodeAuthFailed
	case strings.Contains(lower, "permission"),
		strings.Contains(lower, "forbidden"):
		return http.StatusForbidden, CodeForbidden
	case strings.Contains(lower, "folder does not exist"):
		return http.StatusNotFound, CodeFolderNotFound
	case strings.Contains(lower, "file does not exist"):
		return http.StatusNotFound, CodeFileNotFound
	case strings.Contains(lower, "share does not exist"):
		return http.StatusNotFound, CodeShareNotFound
	case strings.Contains(lower, "does not exist"),
		strings.Contains(lower, "not found"):
		return http.StatusNotFound, CodeResourceNotFound
	case strings.Contains(lower, "already exists"),
		strings.Contains(lower, "already registered"),
		strings.Contains(lower, "same name"),
		strings.Contains(lower, "duplicate"):
		return http.StatusConflict, CodeAlreadyExists
	case strings.Contains(lower, "verification"),
		strings.Contains(lower, "access code"):
		return http.StatusBadRequest, CodeVerificationFailed
	case strings.Contains(lower, "invalid"),
		strings.Contains(lower, "expired"),
		strings.Contains(lower, "mismatch"),
		strings.Contains(lower, "too large"),
		strings.Contains(lower, "exceeds max upload size"),
		strings.Contains(lower, "unsupported"),
		strings.Contains(lower, "does not support"):
		return http.StatusBadRequest, CodeInvalidParam
	case cause != nil:
		return http.StatusInternalServerError, CodeInternalError
	default:
		return http.StatusBadRequest, CodeBusinessRejected
	}
}

func traceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value("trace_id").(string)
	return traceID
}
