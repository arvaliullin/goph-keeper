package ports

//go:generate mockgen -source=services.go -destination=mocks/services_mock.go -package=mocks

import (
	"context"
	"io"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
)

// AuthService контракт сервиса аутентификации.
type AuthService interface {
	// Register регистрирует нового пользователя и возвращает JWT-токен.
	Register(ctx context.Context, login, password string) (string, error)
	// Login аутентифицирует пользователя и возвращает JWT-токен.
	Login(ctx context.Context, login, password string) (string, error)
}

// SecretService контракт сервиса управления секретами.
type SecretService interface {
	// Create создает новый секрет указанного типа.
	Create(ctx context.Context, userID int64, secretType domain.SecretType, data, metadata []byte) (*domain.Secret, error)
	// Update обновляет данные и метаданные существующего секрета.
	Update(ctx context.Context, userID int64, id string, data, metadata []byte) (*domain.Secret, error)
	// Delete удаляет секрет (мягкое удаление).
	Delete(ctx context.Context, userID int64, id string) error
	// Get возвращает секрет по идентификатору.
	Get(ctx context.Context, userID int64, id string) (*domain.Secret, error)
	// List возвращает все секреты пользователя.
	List(ctx context.Context, userID int64) ([]*domain.Secret, error)
	// Sync возвращает секреты, измененные после указанного времени.
	Sync(ctx context.Context, userID int64, updatedAfter time.Time) ([]*domain.Secret, error)
}

// BinaryService контракт сервиса управления бинарными файлами.
type BinaryService interface {
	// Upload загружает бинарные данные для секрета.
	Upload(ctx context.Context, userID int64, id string, reader io.Reader, size int64) error
	// Download возвращает поток чтения бинарных данных секрета.
	Download(ctx context.Context, userID int64, id string) (io.ReadCloser, error)
	// Delete удаляет бинарные данные секрета.
	Delete(ctx context.Context, userID int64, id string) error
}

// CryptoService контракт для серверного шифрования данных.
type CryptoService interface {
	// Encrypt шифрует данные с использованием AES-GCM.
	Encrypt(plaintext []byte) ([]byte, error)
	// Decrypt расшифровывает данные, зашифрованные методом Encrypt.
	Decrypt(ciphertext []byte) ([]byte, error)
}
