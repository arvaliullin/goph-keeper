package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	tests := []struct {
		name           string
		reqBody        authRequest
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: authRequest{
				Login:    "test",
				Password: "password",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().Register(gomock.Any(), "test", "password").Return("mock_token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User Already Exists",
			reqBody: authRequest{
				Login:    "test",
				Password: "password",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().Register(gomock.Any(), "test", "password").Return("", domain.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	handler := NewAuthHandler(mockAuthService)

	tests := []struct {
		name           string
		reqBody        authRequest
		mockBehavior   func()
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: authRequest{
				Login:    "test",
				Password: "password",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().Login(gomock.Any(), "test", "password").Return("mock_token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Credentials",
			reqBody: authRequest{
				Login:    "test",
				Password: "wrong",
			},
			mockBehavior: func() {
				mockAuthService.EXPECT().Login(gomock.Any(), "test", "wrong").Return("", domain.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
