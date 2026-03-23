package ports

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate mockgen -source=postgres.go -destination=mocks/postgres_mock.go -package=mocks
//go:generate mockgen -package=mocks -destination=mocks/pgx_mock.go github.com/jackc/pgx/v5 Row,Rows

// PostgresClient определяет минимально необходимые операции для работы с PostgreSQL.
type PostgresClient interface {
	// QueryRow выполняет запрос, возвращающий одну строку.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Query выполняет запрос, возвращающий набор строк.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	// Exec выполняет запрос без возврата строк.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}
