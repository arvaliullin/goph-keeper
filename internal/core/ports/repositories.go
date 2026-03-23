package ports

//go:generate mockgen -source=repositories.go -destination=mocks/repositories_mock.go -package=mocks

import (
	"context"
	"io"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
)

// UserRepository определяет контракт для хранилища пользователей.
type UserRepository interface {
	// Create сохраняет нового пользователя.
	Create(ctx context.Context, user *domain.User) error
	// GetByLogin возвращает пользователя по логину.
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
}

// SecretRepository определяет контракт для хранилища секретов.
type SecretRepository interface {
	// Create сохраняет новый секрет.
	Create(ctx context.Context, secret *domain.Secret) error
	// Update обновляет существующий секрет.
	Update(ctx context.Context, secret *domain.Secret) error
	// Delete выполняет мягкое удаление секрета.
	Delete(ctx context.Context, id string, userID int64) error
	// GetByID возвращает секрет по идентификатору и владельцу.
	GetByID(ctx context.Context, id string, userID int64) (*domain.Secret, error)
	// ListByUserID возвращает все секреты пользователя.
	ListByUserID(ctx context.Context, userID int64) ([]*domain.Secret, error)
	// SyncByUserID возвращает секреты, обновленные после указанного времени.
	SyncByUserID(ctx context.Context, userID int64, updatedAfter time.Time) ([]*domain.Secret, error)
}

// BinaryRepository определяет контракт для объектного хранилища бинарных данных (S3).
type BinaryRepository interface {
	// Upload загружает объект в хранилище.
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64) error
	// Download возвращает поток чтения объекта из хранилища.
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)
	// Delete удаляет объект из хранилища.
	Delete(ctx context.Context, objectName string) error
}
