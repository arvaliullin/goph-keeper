package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	secretReq := secretRequest{
		Type:     domain.SecretTypeLogin,
		Data:     []byte("data"),
		Metadata: []byte("meta"),
	}
	body, _ := json.Marshal(secretReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().
		Create(gomock.Any(), int64(1), domain.SecretTypeLogin, []byte("data"), []byte("meta")).
		Return(&domain.Secret{ID: "sec1"}, nil)

	handler.Create(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSecretHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/sec1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "sec1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().
		Get(gomock.Any(), int64(1), "sec1").
		Return(&domain.Secret{ID: "sec1", CreatedAt: time.Now()}, nil)

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSecretHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	secretReq := secretRequest{
		Type:     domain.SecretTypeLogin,
		Data:     []byte("newdata"),
		Metadata: []byte("newmeta"),
	}
	body, _ := json.Marshal(secretReq)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/sec1", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "sec1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().
		Update(gomock.Any(), int64(1), "sec1", []byte("newdata"), []byte("newmeta")).
		Return(&domain.Secret{ID: "sec1"}, nil)

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSecretHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/sec1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "sec1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().Delete(gomock.Any(), int64(1), "sec1").Return(nil)

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestSecretHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().List(gomock.Any(), int64(1)).Return([]*domain.Secret{
		{ID: "sec1", CreatedAt: time.Now()},
	}, nil)

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSecretHandler_ListSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretService := mocks.NewMockSecretService(ctrl)
	handler := NewSecretHandler(mockSecretService)

	updatedAfter := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets?updated_after="+updatedAfter, nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	rec := httptest.NewRecorder()

	mockSecretService.EXPECT().Sync(gomock.Any(), int64(1), gomock.Any()).Return([]*domain.Secret{
		{ID: "sec1", CreatedAt: time.Now()},
	}, nil)

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
