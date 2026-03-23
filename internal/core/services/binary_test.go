package services

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBinaryService_Upload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBinaryRepository(ctrl)
	mockSecretRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewBinaryService(mockRepo, mockSecretRepo, mockCrypto)

	data := []byte("binary data")
	reader := bytes.NewReader(data)

	mockSecretRepo.EXPECT().GetByID(gomock.Any(), "file1", int64(1)).Return(&domain.Secret{
		ID:   "file1",
		Type: domain.SecretTypeBinary,
	}, nil)
	mockCrypto.EXPECT().Encrypt(data).Return([]byte("enc_binary_data"), nil)
	mockRepo.EXPECT().Upload(gomock.Any(), "1/file1", gomock.Any(), int64(len("enc_binary_data"))).Return(nil)

	err := service.Upload(context.Background(), 1, "file1", reader, int64(len(data)))
	assert.NoError(t, err)
}

func TestBinaryService_Download(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBinaryRepository(ctrl)
	mockSecretRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewBinaryService(mockRepo, mockSecretRepo, mockCrypto)

	readCloser := io.NopCloser(bytes.NewReader([]byte("enc_binary_data")))

	mockSecretRepo.EXPECT().GetByID(gomock.Any(), "file1", int64(1)).Return(&domain.Secret{
		ID:   "file1",
		Type: domain.SecretTypeBinary,
	}, nil)
	mockRepo.EXPECT().Download(gomock.Any(), "1/file1").Return(readCloser, nil)
	mockCrypto.EXPECT().Decrypt([]byte("enc_binary_data")).Return([]byte("binary data"), nil)

	res, err := service.Download(context.Background(), 1, "file1")
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestBinaryService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBinaryRepository(ctrl)
	mockSecretRepo := mocks.NewMockSecretRepository(ctrl)
	mockCrypto := mocks.NewMockCryptoService(ctrl)
	service := NewBinaryService(mockRepo, mockSecretRepo, mockCrypto)

	mockSecretRepo.EXPECT().GetByID(gomock.Any(), "file1", int64(1)).Return(&domain.Secret{
		ID:   "file1",
		Type: domain.SecretTypeBinary,
	}, nil)
	mockSecretRepo.EXPECT().Delete(gomock.Any(), "file1", int64(1)).Return(nil)
	mockRepo.EXPECT().Delete(gomock.Any(), "1/file1").Return(nil)

	err := service.Delete(context.Background(), 1, "file1")
	assert.NoError(t, err)
}
