package postgres

import (
	"context"
	"errors"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository реализация хранилища пользователей в PostgreSQL.
type UserRepository struct {
	Repository[domain.User]
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db ports.PostgresClient) *UserRepository {
	return &UserRepository{
		Repository: Repository[domain.User]{
			DB:   db,
			Scan: scanUser,
		},
	}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// Create создает нового пользователя.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`
	err := r.DB.QueryRow(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

// GetByLogin возвращает пользователя по логину.
func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	return r.QueryOne(ctx, query, login)
}
