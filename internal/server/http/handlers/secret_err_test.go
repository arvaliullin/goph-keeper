package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretHandler_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	t.Run("Create_NoAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader([]byte("{}")))
		rec := httptest.NewRecorder()
		handler.Create(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Create_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.Create(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Create_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader([]byte("{}")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		mockSecretService.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrInvalidSecretType)
		handler.Create(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Update_MissingID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/", bytes.NewReader([]byte("{}")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.Update(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Update_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/1", bytes.NewReader([]byte("{}")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockSecretService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrSecretNotFound)
		handler.Update(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Get_MissingID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.Get(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockSecretService.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, domain.ErrSecretNotFound)
		handler.Get(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Delete_MissingID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.Delete(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Delete_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		mockSecretService.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.ErrSecretNotFound)
		handler.Delete(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("List_ServiceErr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		mockSecretService.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("err"))
		handler.List(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("List_InvalidUpdatedAfter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets?updated_after=broken", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		rec := httptest.NewRecorder()
		handler.List(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
