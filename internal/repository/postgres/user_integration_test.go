//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Integration_Create(t *testing.T) {
	setupTest(t)
	repo := postgres.NewUserRepository(testPool)

	t.Run("успешное создание пользователя", func(t *testing.T) {
		user := &domain.User{Login: "alice", PasswordHash: "hash1"}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("ошибка при дублировании логина", func(t *testing.T) {
		user1 := &domain.User{Login: "duplicate", PasswordHash: "hash1"}
		require.NoError(t, repo.Create(context.Background(), user1))

		user2 := &domain.User{Login: "duplicate", PasswordHash: "hash2"}
		err := repo.Create(context.Background(), user2)
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	})
}

func TestUserRepository_Integration_GetByLogin(t *testing.T) {
	setupTest(t)
	repo := postgres.NewUserRepository(testPool)

	t.Run("пользователь найден", func(t *testing.T) {
		user := &domain.User{Login: "bob", PasswordHash: "bobhash"}
		require.NoError(t, repo.Create(context.Background(), user))

		found, err := repo.GetByLogin(context.Background(), "bob")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "bob", found.Login)
		assert.Equal(t, "bobhash", found.PasswordHash)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		_, err := repo.GetByLogin(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}
