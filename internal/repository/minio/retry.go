package minio

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"

	miniogo "github.com/minio/minio-go/v7"
)

var retryableStatusCodes = map[int]struct{}{
	http.StatusRequestTimeout:      {},
	http.StatusTooManyRequests:     {},
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	http.StatusGatewayTimeout:      {},
}

var retryableErrorCodes = map[string]struct{}{
	miniogo.InternalError: {},
	"RequestTimeout":      {},
	"SlowDown":            {},
	"ServiceUnavailable":  {},
}

// IsRetryable определяет, относится ли ошибка MinIO к временным сбоям.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	errResp := miniogo.ToErrorResponse(err)
	if _, ok := retryableStatusCodes[errResp.StatusCode]; ok {
		return true
	}
	_, ok := retryableErrorCodes[errResp.Code]
	return ok
}
