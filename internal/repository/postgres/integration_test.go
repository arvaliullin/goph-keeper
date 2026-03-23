//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/repository/postgres/testhelpers"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testContainer *testhelpers.PostgresContainer
	testPool      *pgxpool.Pool
)

// TestMain настраивает тестовый контейнер PostgreSQL для всех интеграционных тестов.
func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testContainer, err = testhelpers.NewPostgresContainer(ctx)
	if err != nil {
		panic("не удалось запустить тестовый контейнер: " + err.Error())
	}

	testPool, err = pgxpool.New(ctx, testContainer.DSN)
	if err != nil {
		testContainer.Terminate(ctx)
		panic("не удалось создать пул соединений: " + err.Error())
	}

	code := m.Run()

	testPool.Close()
	testContainer.Terminate(ctx)

	os.Exit(code)
}

// setupTest очищает таблицы перед каждым тестом для изоляции.
func setupTest(t *testing.T) {
	t.Helper()
	if err := testContainer.CleanupTables(context.Background()); err != nil {
		t.Fatalf("не удалось очистить таблицы: %v", err)
	}
}
