package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/jackc/pgx/v5"
)

// SecretRepository реализация хранилища секретов в PostgreSQL.
type SecretRepository struct {
	Repository[domain.Secret]
}

// NewSecretRepository создает новый экземпляр SecretRepository.
func NewSecretRepository(db ports.PostgresClient) *SecretRepository {
	return &SecretRepository{
		Repository: Repository[domain.Secret]{
			DB:   db,
			Scan: scanSecret,
		},
	}
}

func scanSecret(row pgx.Row) (*domain.Secret, error) {
	s := &domain.Secret{}
	err := row.Scan(&s.ID, &s.UserID, &s.Type, &s.Data, &s.Metadata, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSecretNotFound
		}
		return nil, err
	}
	return s, nil
}

// Create сохраняет новый секрет в БД.
func (r *SecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, type, data, metadata, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	now := time.Now()
	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = now
	}
	secret.UpdatedAt = now

	_, err := r.DB.Exec(
		ctx,
		query,
		secret.ID,
		secret.UserID,
		secret.Type,
		secret.Data,
		secret.Metadata,
		secret.CreatedAt,
		secret.UpdatedAt,
		secret.DeletedAt,
	)
	return err
}

// Update обновляет существующий секрет.
func (r *SecretRepository) Update(ctx context.Context, secret *domain.Secret) error {
	query := `
		UPDATE secrets 
		SET data = $1, metadata = $2, updated_at = $3 
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL`

	secret.UpdatedAt = time.Now()
	res, err := r.DB.Exec(ctx, query, secret.Data, secret.Metadata, secret.UpdatedAt, secret.ID, secret.UserID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}

// Delete удаляет секрет по ID.
func (r *SecretRepository) Delete(ctx context.Context, id string, userID int64) error {
	query := `
		UPDATE secrets
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL`
	deletedAt := time.Now()
	res, err := r.DB.Exec(ctx, query, deletedAt, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}

// GetByID возвращает секрет по ID для конкретного пользователя.
func (r *SecretRepository) GetByID(ctx context.Context, id string, userID int64) (*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, data, metadata, created_at, updated_at, deleted_at
		FROM secrets 
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	return r.QueryOne(ctx, query, id, userID)
}

// ListByUserID возвращает все секреты пользователя.
func (r *SecretRepository) ListByUserID(ctx context.Context, userID int64) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, data, metadata, created_at, updated_at, deleted_at
		FROM secrets 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC`
	return r.QueryMany(ctx, query, userID)
}

// SyncByUserID возвращает измененные секреты пользователя, включая tombstone записи.
func (r *SecretRepository) SyncByUserID(ctx context.Context, userID int64, updatedAfter time.Time) ([]*domain.Secret, error) {
	query := `
		SELECT id, user_id, type, data, metadata, created_at, updated_at, deleted_at
		FROM secrets
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at DESC`
	return r.QueryMany(ctx, query, userID, updatedAfter)
}
