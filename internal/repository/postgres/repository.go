package postgres

import (
	"context"

	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/jackc/pgx/v5"
)

// ScanFunc функция сканирования строки результата в экземпляр T.
type ScanFunc[T any] func(row pgx.Row) (*T, error)

// Repository обобщенный репозиторий для PostgreSQL, устраняющий дублирование
// логики подключения и итерации по строкам результата.
type Repository[T any] struct {
	// DB клиент PostgreSQL для выполнения запросов.
	DB ports.PostgresClient
	// Scan функция сканирования строки результата в экземпляр T.
	Scan ScanFunc[T]
}

// QueryOne выполняет запрос и сканирует единственную строку результата.
func (r *Repository[T]) QueryOne(ctx context.Context, query string, args ...any) (*T, error) {
	return r.Scan(r.DB.QueryRow(ctx, query, args...))
}

// QueryMany выполняет запрос и возвращает коллекцию результатов.
func (r *Repository[T]) QueryMany(ctx context.Context, query string, args ...any) ([]*T, error) {
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*T
	for rows.Next() {
		item, scanErr := r.Scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
