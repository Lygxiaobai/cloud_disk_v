package errors

import (
	"context"
	"net/http"
)

const (
	CodeInvalidParam       = 10001
	CodeValidationFailed   = 10002
	CodeVerificationFailed = 10003

	CodeUnauthorized = 20001
	CodeAuthFailed   = 20002
	CodeForbidden    = 20003
	CodeLoginLocked  = 20004
	CodeRateLimited  = 20005

	CodeResourceNotFound = 30000
	CodeFileNotFound     = 30001
	CodeShareNotFound    = 30002
	CodeFolderNotFound   = 30003
	CodeAlreadyExists    = 30010

	CodeBusinessRejected = 40000
	CodeInternalError    = 50000
)

func NewWithCode(ctx context.Context, status int, code int, message string, err error, extra map[string]interface{}) error {
	baseErr := New(ctx, message, err, extra)
	if logicErr, ok := baseErr.(*LogicError); ok {
		logicErr.status = status
		logicErr.code = code
		return logicErr
	}
	return baseErr
}

func InvalidParam(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusBadRequest, CodeInvalidParam, message, err, extra)
}

func ValidationFailed(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusBadRequest, CodeValidationFailed, message, err, extra)
}

func VerificationFailed(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusBadRequest, CodeVerificationFailed, message, err, extra)
}

func Unauthorized(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusUnauthorized, CodeUnauthorized, message, err, extra)
}

func AuthFailed(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusUnauthorized, CodeAuthFailed, message, err, extra)
}

func ForbiddenError(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusForbidden, CodeForbidden, message, err, extra)
}

func LoginLocked(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusTooManyRequests, CodeLoginLocked, message, err, extra)
}

func RateLimited(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusTooManyRequests, CodeRateLimited, message, err, extra)
}

func NotFound(ctx context.Context, code int, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusNotFound, code, message, err, extra)
}

func Conflict(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusConflict, CodeAlreadyExists, message, err, extra)
}

func Business(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusBadRequest, CodeBusinessRejected, message, err, extra)
}

func Internal(ctx context.Context, message string, err error, extra map[string]interface{}) error {
	return NewWithCode(ctx, http.StatusInternalServerError, CodeInternalError, message, err, extra)
}
