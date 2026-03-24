// Package testhelpers предоставляет вспомогательные утилиты для интеграционных тестов PostgreSQL.
package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/arvaliullin/goph-keeper/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer представляет тестовый контейнер PostgreSQL.
type PostgresContainer struct {
	*postgres.PostgresContainer
	// DSN строка подключения к тестовой базе данных.
	DSN string
}

// NewPostgresContainer создаёт и запускает контейнер PostgreSQL для тестирования.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("запуск контейнера postgres: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("получение строки подключения: %w", err)
	}

	if err := runMigrations(ctx, dsn); err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("применение миграций: %w", err)
	}

	return &PostgresContainer{
		PostgresContainer: container,
		DSN:               dsn,
	}, nil
}

// Terminate останавливает и удаляет контейнер.
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.PostgresContainer.Terminate(ctx)
}

// CleanupTables очищает все таблицы для изоляции тестов.
func (c *PostgresContainer) CleanupTables(ctx context.Context) error {
	db, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, table := range []string{"secrets", "users"} {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			return fmt.Errorf("очистка таблицы %s: %w", table, err)
		}
	}
	return nil
}

func runMigrations(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("открытие БД: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("установка диалекта: %w", err)
	}

	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("применение миграций: %w", err)
	}
	return nil
}
