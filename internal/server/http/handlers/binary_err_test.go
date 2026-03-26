package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBinaryHandler_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBinaryService := mocks.NewMockBinaryService(ctrl)
	handler := NewBinaryHandler(mockBinaryService, 4)

	t.Run("Upload_NoAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/1", bytes.NewReader([]byte{}))
		rec := httptest.NewRecorder()
		handler.Upload(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Upload_MissingID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/", bytes.NewReader([]byte{}))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.Upload(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Upload_NoContentLength", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/1", bytes.NewReader([]byte{}))
		req.ContentLength = 0
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		handler.Upload(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Upload_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/1", bytes.NewReader([]byte("data")))
		req.ContentLength = 4
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockBinaryService.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrBinarySecretRequired)
		handler.Upload(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Download_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/binary/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockBinaryService.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrSecretNotFound)
		handler.Download(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Upload_TooLarge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/binary/1", bytes.NewReader([]byte("large")))
		req.ContentLength = 5
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		handler.Upload(rec, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})

	t.Run("Delete_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/binary/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockBinaryService.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrSecretNotFound)
		handler.Delete(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
