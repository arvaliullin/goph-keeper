package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)

	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
		*dest[0].(*int64) = 1
		*dest[1].(*time.Time) = time.Now()
		return nil
	})
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewUserRepository(mockDB)
	err := repo.Create(context.Background(), &domain.User{Login: "test", PasswordHash: "hash"})
	assert.NoError(t, err)
}

func TestUserRepository_GetByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)

	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
		*dest[0].(*int64) = 1
		*dest[1].(*string) = "test"
		*dest[2].(*string) = "hash"
		*dest[3].(*time.Time) = time.Now()
		return nil
	})
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewUserRepository(mockDB)
	user, err := repo.GetByLogin(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepository_GetByLogin_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)

	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewUserRepository(mockDB)
	user, err := repo.GetByLogin(context.Background(), "test")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, user)
}
