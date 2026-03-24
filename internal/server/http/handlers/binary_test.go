package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBinaryHandler_Upload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBinaryService := mocks.NewMockBinaryService(ctrl)
	handler := NewBinaryHandler(mockBinaryService, 1024)

	data := []byte("binary data")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/file1", bytes.NewReader(data))
	req.ContentLength = int64(len(data))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
	req.SetPathValue("id", "file1")

	rec := httptest.NewRecorder()

	mockBinaryService.EXPECT().
		Upload(gomock.Any(), int64(1), "file1", gomock.Any(), int64(len(data))).
		Return(nil)

	handler.Upload(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBinaryHandler_Download(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBinaryService := mocks.NewMockBinaryService(ctrl)
	handler := NewBinaryHandler(mockBinaryService, 1024)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/binary/file1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
	req.SetPathValue("id", "file1")

	rec := httptest.NewRecorder()

	data := []byte("binary data")
	readCloser := io.NopCloser(bytes.NewReader(data))

	mockBinaryService.EXPECT().
		Download(gomock.Any(), int64(1), "file1").
		Return(readCloser, nil)

	handler.Download(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, data, rec.Body.Bytes())
}

func TestBinaryHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBinaryService := mocks.NewMockBinaryService(ctrl)
	handler := NewBinaryHandler(mockBinaryService, 1024)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/binary/file1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
	req.SetPathValue("id", "file1")

	rec := httptest.NewRecorder()

	mockBinaryService.EXPECT().Delete(gomock.Any(), int64(1), "file1").Return(nil)

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
