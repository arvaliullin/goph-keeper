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

func TestSecretService_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	t.Run("Create_EncryptDataErr", func(t *testing.T) {
		mockCrypto.EXPECT().Encrypt([]byte("data")).Return(nil, errors.New("err"))
		_, err := service.Create(context.Background(), 1, domain.SecretTypeLogin, []byte("data"), []byte("meta"))
		assert.Error(t, err)
	})

	t.Run("Create_EncryptMetaErr", func(t *testing.T) {
		mockCrypto.EXPECT().Encrypt([]byte("data")).Return([]byte("enc"), nil)
		mockCrypto.EXPECT().Encrypt([]byte("meta")).Return(nil, errors.New("err"))
		_, err := service.Create(context.Background(), 1, domain.SecretTypeLogin, []byte("data"), []byte("meta"))
		assert.Error(t, err)
	})

	t.Run("Create_RepoErr", func(t *testing.T) {
		mockCrypto.EXPECT().Encrypt([]byte("data")).Return([]byte("enc"), nil)
		mockCrypto.EXPECT().Encrypt([]byte("meta")).Return([]byte("enc_meta"), nil)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("err"))
		_, err := service.Create(context.Background(), 1, domain.SecretTypeLogin, []byte("data"), []byte("meta"))
		assert.Error(t, err)
	})

	t.Run("Update_EncryptDataErr", func(t *testing.T) {
		mockCrypto.EXPECT().Encrypt([]byte("data")).Return(nil, errors.New("err"))
		_, err := service.Update(context.Background(), 1, "1", []byte("data"), []byte("meta"))
		assert.Error(t, err)
	})

	t.Run("Get_RepoErr", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(gomock.Any(), "1", int64(1)).Return(nil, errors.New("err"))
		_, err := service.Get(context.Background(), 1, "1")
		assert.Error(t, err)
	})

	t.Run("List_DecryptErr", func(t *testing.T) {
		mockRepo.EXPECT().ListByUserID(gomock.Any(), int64(1)).Return([]*domain.Secret{{ID: "1", Data: []byte("d")}}, nil)
		mockCrypto.EXPECT().Decrypt([]byte("d")).Return(nil, errors.New("err"))
		_, err := service.List(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("Delete_NotFound", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(gomock.Any(), "missing", int64(1)).Return(nil, domain.ErrSecretNotFound)
		err := service.Delete(context.Background(), 1, "missing")
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})

	t.Run("Create_InvalidType", func(t *testing.T) {
		_, err := service.Create(context.Background(), 1, domain.SecretType("unknown"), []byte("data"), []byte("meta"))
		assert.ErrorIs(t, err, domain.ErrInvalidSecretType)
	})
}
