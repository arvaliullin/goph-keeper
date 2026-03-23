package services

import (
	"context"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	mockCrypto.EXPECT().Encrypt([]byte("data")).Return([]byte("enc_data"), nil)
	mockCrypto.EXPECT().Encrypt([]byte("meta")).Return([]byte("enc_meta"), nil)

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	secret, err := service.Create(context.Background(), 1, domain.SecretTypeLogin, []byte("data"), []byte("meta"))
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, []byte("data"), secret.Data)
}

func TestSecretService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	encSecret := &domain.Secret{
		ID:       "sec1",
		Data:     []byte("enc_data"),
		Metadata: []byte("enc_meta"),
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), "sec1", int64(1)).Return(encSecret, nil)

	mockCrypto.EXPECT().Decrypt([]byte("enc_data")).Return([]byte("data"), nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_meta")).Return([]byte("meta"), nil)

	secret, err := service.Get(context.Background(), 1, "sec1")
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, []byte("data"), secret.Data)
}

func TestSecretService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	mockCrypto.EXPECT().Encrypt([]byte("newdata")).Return([]byte("enc_newdata"), nil)
	mockCrypto.EXPECT().Encrypt([]byte("newmeta")).Return([]byte("enc_newmeta"), nil)

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	secret, err := service.Update(context.Background(), 1, "sec1", []byte("newdata"), []byte("newmeta"))
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, []byte("newdata"), secret.Data)
}

func TestSecretService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	mockRepo.EXPECT().GetByID(gomock.Any(), "sec1", int64(1)).Return(&domain.Secret{ID: "sec1", Type: domain.SecretTypeLogin}, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), "sec1", int64(1)).Return(nil)

	err := service.Delete(context.Background(), 1, "sec1")
	assert.NoError(t, err)
}

func TestSecretService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	secrets := []*domain.Secret{
		{ID: "sec1", Data: []byte("enc_data1"), Metadata: []byte("enc_meta1"), CreatedAt: time.Now()},
		{ID: "sec2", Data: []byte("enc_data2"), Metadata: []byte("enc_meta2"), CreatedAt: time.Now()},
	}

	mockRepo.EXPECT().ListByUserID(gomock.Any(), int64(1)).Return(secrets, nil)

	mockCrypto.EXPECT().Decrypt([]byte("enc_data1")).Return([]byte("data1"), nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_meta1")).Return([]byte("meta1"), nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_data2")).Return([]byte("data2"), nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_meta2")).Return([]byte("meta2"), nil)

	res, err := service.List(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, []byte("data1"), res[0].Data)
}

func TestSecretService_Sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewSecretService(mockRepo, nil, mockCrypto)

	updatedAfter := time.Now().Add(-time.Hour)
	deletedAt := time.Now()
	secrets := []*domain.Secret{
		{ID: "sec1", Data: []byte("enc_data1"), Metadata: []byte("enc_meta1"), CreatedAt: time.Now()},
		{ID: "sec2", DeletedAt: &deletedAt, CreatedAt: time.Now()},
	}

	mockRepo.EXPECT().SyncByUserID(gomock.Any(), int64(1), updatedAfter).Return(secrets, nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_data1")).Return([]byte("data1"), nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_meta1")).Return([]byte("meta1"), nil)

	res, err := service.Sync(context.Background(), 1, updatedAfter)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, []byte("data1"), res[0].Data)
	assert.Nil(t, res[1].Data)
	assert.NotNil(t, res[1].DeletedAt)
}
