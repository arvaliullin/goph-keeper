package services

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
)

// BinaryService реализация сервиса управления бинарными файлами.
type BinaryService struct {
	repo          ports.BinaryRepository
	secretRepo    ports.SecretRepository
	cryptoService ports.CryptoService
}

// NewBinaryService создает новый экземпляр BinaryService.
func NewBinaryService(
	repo ports.BinaryRepository,
	secretRepo ports.SecretRepository,
	cryptoService ports.CryptoService,
) *BinaryService {
	return &BinaryService{
		repo:          repo,
		secretRepo:    secretRepo,
		cryptoService: cryptoService,
	}
}

// Upload загружает бинарный файл в S3.
func (s *BinaryService) Upload(ctx context.Context, userID int64, id string, reader io.Reader, size int64) error {
	if _, err := s.ensureBinarySecret(ctx, userID, id); err != nil {
		return err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	encryptedData, err := s.cryptoService.Encrypt(data)
	if err != nil {
		return err
	}

	return s.repo.Upload(ctx, binaryObjectName(userID, id), bytes.NewReader(encryptedData), int64(len(encryptedData)))
}

// Download скачивает бинарный файл из S3.
func (s *BinaryService) Download(ctx context.Context, userID int64, id string) (io.ReadCloser, error) {
	if _, err := s.ensureBinarySecret(ctx, userID, id); err != nil {
		return nil, err
	}

	reader, err := s.repo.Download(ctx, binaryObjectName(userID, id))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	encryptedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	data, err := s.cryptoService.Decrypt(encryptedData)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Delete удаляет бинарный секрет и его содержимое.
func (s *BinaryService) Delete(ctx context.Context, userID int64, id string) error {
	if _, err := s.ensureBinarySecret(ctx, userID, id); err != nil {
		return err
	}
	if err := s.secretRepo.Delete(ctx, id, userID); err != nil {
		return err
	}
	if delErr := s.repo.Delete(ctx, binaryObjectName(userID, id)); delErr != nil {
		log.Printf("failed to delete binary object %s: %v", binaryObjectName(userID, id), delErr)
	}
	return nil
}

func (s *BinaryService) ensureBinarySecret(ctx context.Context, userID int64, id string) (*domain.Secret, error) {
	secret, err := s.secretRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if secret.Type != domain.SecretTypeBinary {
		return nil, domain.ErrBinarySecretRequired
	}
	return secret, nil
}
