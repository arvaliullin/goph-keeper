package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/google/uuid"
)

// SecretService реализация сервиса управления секретами с серверным шифрованием.
type SecretService struct {
	repo          ports.SecretRepository
	binaryRepo    ports.BinaryRepository
	cryptoService ports.CryptoService
}

// NewSecretService создает новый экземпляр SecretService.
func NewSecretService(
	repo ports.SecretRepository,
	binaryRepo ports.BinaryRepository,
	cryptoService ports.CryptoService,
) *SecretService {
	return &SecretService{
		repo:          repo,
		binaryRepo:    binaryRepo,
		cryptoService: cryptoService,
	}
}

// Create создает новый секрет, предварительно зашифровав его данные и метаданные.
func (s *SecretService) Create(ctx context.Context, userID int64, secretType domain.SecretType, data, metadata []byte) (*domain.Secret, error) {
	if !secretType.Valid() {
		return nil, domain.ErrInvalidSecretType
	}

	encData, err := s.cryptoService.Encrypt(data)
	if err != nil {
		return nil, err
	}

	encMeta, err := s.cryptoService.Encrypt(metadata)
	if err != nil {
		return nil, err
	}

	secret := &domain.Secret{
		ID:       uuid.NewString(),
		UserID:   userID,
		Type:     secretType,
		Data:     encData,
		Metadata: encMeta,
	}

	if err := s.repo.Create(ctx, secret); err != nil {
		return nil, err
	}

	// Возвращаем клиенту расшифрованные данные, с которыми он работал
	secret.Data = data
	secret.Metadata = metadata

	return secret, nil
}

// Update обновляет существующий секрет, шифруя новые данные.
func (s *SecretService) Update(ctx context.Context, userID int64, id string, data, metadata []byte) (*domain.Secret, error) {
	encData, err := s.cryptoService.Encrypt(data)
	if err != nil {
		return nil, err
	}

	encMeta, err := s.cryptoService.Encrypt(metadata)
	if err != nil {
		return nil, err
	}

	secret := &domain.Secret{
		ID:       id,
		UserID:   userID,
		Data:     encData,
		Metadata: encMeta,
	}

	if err := s.repo.Update(ctx, secret); err != nil {
		return nil, err
	}

	secret.Data = data
	secret.Metadata = metadata

	return secret, nil
}

// Delete удаляет секрет по ID.
func (s *SecretService) Delete(ctx context.Context, userID int64, id string) error {
	secret, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}
	if secret.Type == domain.SecretTypeBinary && s.binaryRepo != nil {
		if delErr := s.binaryRepo.Delete(ctx, binaryObjectName(userID, id)); delErr != nil {
			log.Printf("failed to delete binary object %s: %v", binaryObjectName(userID, id), delErr)
		}
	}

	return nil
}

// Get возвращает секрет, расшифровывая его данные.
func (s *SecretService) Get(ctx context.Context, userID int64, id string) (*domain.Secret, error) {
	secret, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	decData, err := s.cryptoService.Decrypt(secret.Data)
	if err != nil {
		return nil, err
	}
	secret.Data = decData

	decMeta, err := s.cryptoService.Decrypt(secret.Metadata)
	if err != nil {
		return nil, err
	}
	secret.Metadata = decMeta

	return secret, nil
}

// List возвращает список всех секретов пользователя, расшифровывая их.
func (s *SecretService) List(ctx context.Context, userID int64) ([]*domain.Secret, error) {
	secrets, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.decryptSecrets(secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

// Sync возвращает изменения пользователя, произошедшие после указанного времени.
func (s *SecretService) Sync(ctx context.Context, userID int64, updatedAfter time.Time) ([]*domain.Secret, error) {
	secrets, err := s.repo.SyncByUserID(ctx, userID, updatedAfter)
	if err != nil {
		return nil, err
	}

	if err := s.decryptSecrets(secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (s *SecretService) decryptSecrets(secrets []*domain.Secret) error {
	for _, sec := range secrets {
		if sec.DeletedAt != nil {
			sec.Data = nil
			sec.Metadata = nil
			continue
		}

		decData, err := s.cryptoService.Decrypt(sec.Data)
		if err != nil {
			return err
		}
		sec.Data = decData

		decMeta, err := s.cryptoService.Decrypt(sec.Metadata)
		if err != nil {
			return err
		}
		sec.Metadata = decMeta
	}

	return nil
}

func binaryObjectName(userID int64, id string) string {
	return fmt.Sprintf("%d/%s", userID, id)
}
