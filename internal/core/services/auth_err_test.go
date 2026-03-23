package services

import (
	"context"
	"errors"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthService_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockUserRepo, "secret")

	_, err := service.Register(context.Background(), "", "")
	assert.Error(t, err)

	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
	_, err = service.Register(context.Background(), "user", "pass")
	assert.Error(t, err)
}

func TestAuthService_Login_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockUserRepo, "secret")

	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "user").Return(nil, errors.New("db err"))
	_, err := service.Login(context.Background(), "user", "pass")
	assert.Error(t, err)

	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "user").Return(nil, domain.ErrUserNotFound)
	_, err = service.Login(context.Background(), "user", "pass")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
