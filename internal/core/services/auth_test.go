package services

import (
	"context"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockUserRepo, "secret")

	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	token, err := service.Register(context.Background(), "test", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockUserRepo, "secret")

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           1,
		Login:        "test",
		PasswordHash: string(hash),
	}

	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "test").Return(user, nil)

	token, err := service.Login(context.Background(), "test", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockUserRepo, "secret")

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           1,
		Login:        "test",
		PasswordHash: string(hash),
	}

	mockUserRepo.EXPECT().GetByLogin(gomock.Any(), "test").Return(user, nil)

	token, err := service.Login(context.Background(), "test", "wrong")
	assert.Error(t, err)
	assert.Empty(t, token)
}
