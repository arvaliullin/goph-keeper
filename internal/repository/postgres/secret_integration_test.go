//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) *domain.User {
	t.Helper()
	repo := postgres.NewUserRepository(testPool)
	user := &domain.User{Login: "user-" + uuid.NewString(), PasswordHash: "hash"}
	require.NoError(t, repo.Create(context.Background(), user))
	return user
}

func TestSecretRepository_Integration_Create(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	secret := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeLogin,
		Data:   []byte("encrypted-data"),
	}
	err := repo.Create(context.Background(), secret)
	require.NoError(t, err)
	assert.NotZero(t, secret.CreatedAt)
	assert.NotZero(t, secret.UpdatedAt)
}

func TestSecretRepository_Integration_GetByID(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	secret := &domain.Secret{
		ID:       uuid.NewString(),
		UserID:   user.ID,
		Type:     domain.SecretTypeText,
		Data:     []byte("data"),
		Metadata: []byte("meta"),
	}
	require.NoError(t, repo.Create(context.Background(), secret))

	t.Run("найден", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), secret.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, secret.ID, found.ID)
		assert.Equal(t, domain.SecretTypeText, found.Type)
		assert.Equal(t, []byte("data"), found.Data)
	})

	t.Run("не найден", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.NewString(), user.ID)
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})

	t.Run("чужой секрет недоступен", func(t *testing.T) {
		other := createTestUser(t)
		_, err := repo.GetByID(context.Background(), secret.ID, other.ID)
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})
}

func TestSecretRepository_Integration_Update(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	secret := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeLogin,
		Data:   []byte("old-data"),
	}
	require.NoError(t, repo.Create(context.Background(), secret))

	t.Run("успешное обновление", func(t *testing.T) {
		secret.Data = []byte("new-data")
		secret.Metadata = []byte("new-meta")
		err := repo.Update(context.Background(), secret)
		require.NoError(t, err)

		found, err := repo.GetByID(context.Background(), secret.ID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, []byte("new-data"), found.Data)
	})

	t.Run("не найден", func(t *testing.T) {
		err := repo.Update(context.Background(), &domain.Secret{ID: uuid.NewString(), UserID: user.ID})
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})
}

func TestSecretRepository_Integration_Delete(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	secret := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeCard,
		Data:   []byte("card-data"),
	}
	require.NoError(t, repo.Create(context.Background(), secret))

	t.Run("успешное удаление", func(t *testing.T) {
		err := repo.Delete(context.Background(), secret.ID, user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), secret.ID, user.ID)
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})

	t.Run("повторное удаление", func(t *testing.T) {
		err := repo.Delete(context.Background(), secret.ID, user.ID)
		assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	})
}

func TestSecretRepository_Integration_List(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	for i := range 3 {
		s := &domain.Secret{
			ID:     uuid.NewString(),
			UserID: user.ID,
			Type:   domain.SecretTypeText,
			Data:   []byte("data"),
		}
		_ = i
		require.NoError(t, repo.Create(context.Background(), s))
	}

	deleted := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeText,
		Data:   []byte("data"),
	}
	require.NoError(t, repo.Create(context.Background(), deleted))
	require.NoError(t, repo.Delete(context.Background(), deleted.ID, user.ID))

	secrets, err := repo.ListByUserID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 3)
}

func TestSecretRepository_Integration_Sync(t *testing.T) {
	setupTest(t)
	user := createTestUser(t)
	repo := postgres.NewSecretRepository(testPool)

	before := time.Now()
	time.Sleep(5 * time.Millisecond)

	created := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeLogin,
		Data:   []byte("data"),
	}
	require.NoError(t, repo.Create(context.Background(), created))

	deleted := &domain.Secret{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Type:   domain.SecretTypeText,
		Data:   []byte("data"),
	}
	require.NoError(t, repo.Create(context.Background(), deleted))
	require.NoError(t, repo.Delete(context.Background(), deleted.ID, user.ID))

	t.Run("возвращает созданные и удаленные", func(t *testing.T) {
		synced, err := repo.SyncByUserID(context.Background(), user.ID, before)
		require.NoError(t, err)
		assert.Len(t, synced, 2)

		var hasDeleted bool
		for _, s := range synced {
			if s.ID == deleted.ID {
				assert.NotNil(t, s.DeletedAt)
				hasDeleted = true
			}
		}
		assert.True(t, hasDeleted)
	})

	t.Run("пустой результат при нет изменений", func(t *testing.T) {
		synced, err := repo.SyncByUserID(context.Background(), user.ID, time.Now())
		require.NoError(t, err)
		assert.Empty(t, synced)
	})
}
