package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretRepository_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	errDB := errors.New("db error")

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag(""), errDB).Times(1)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag(""), errDB).Times(1)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag(""), errDB).Times(1)

	mockRowErr := mocks.NewMockRow(ctrl)
	mockRowErr.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errDB).Times(1)
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRowErr).Times(1)

	mockDB.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errDB).Times(1)
	mockDB.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errDB).Times(1)

	repo := NewSecretRepository(mockDB)

	err := repo.Create(context.Background(), &domain.Secret{})
	assert.Error(t, err)

	err = repo.Update(context.Background(), &domain.Secret{})
	assert.Error(t, err)

	err = repo.Delete(context.Background(), "1", 1)
	assert.Error(t, err)

	_, err = repo.GetByID(context.Background(), "1", 1)
	assert.Error(t, err)

	_, err = repo.ListByUserID(context.Background(), 1)
	assert.Error(t, err)

	_, err = repo.SyncByUserID(context.Background(), 1, time.Now())
	assert.Error(t, err)
}

func TestSecretRepository_Update_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 0"), nil)

	repo := NewSecretRepository(mockDB)
	err := repo.Update(context.Background(), &domain.Secret{})
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestSecretRepository_GetByID_NoRows(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewSecretRepository(mockDB)
	sec, err := repo.GetByID(context.Background(), "sec1", 1)
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
	assert.Nil(t, sec)
}

func TestUserRepository_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	errDB := errors.New("db error")

	mockDB := mocks.NewMockPostgresClient(ctrl)

	mockRowCreate := mocks.NewMockRow(ctrl)
	mockRowCreate.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(errDB)
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRowCreate).Times(1)

	mockRowGet := mocks.NewMockRow(ctrl)
	mockRowGet.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errDB)
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRowGet).Times(1)

	repo := NewUserRepository(mockDB)

	err := repo.Create(context.Background(), &domain.User{})
	assert.Error(t, err)

	_, err = repo.GetByLogin(context.Background(), "test")
	assert.Error(t, err)
}

func TestSecretRepository_ListByUserID_RowErr(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRows := mocks.NewMockRows(ctrl)

	mockRows.EXPECT().Next().Return(true).Times(1)
	mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*string) = "sec1"
			*dest[1].(*int64) = int64(1)
			*dest[2].(*domain.SecretType) = domain.SecretTypeLogin
			*dest[3].(*[]byte) = nil
			*dest[4].(*[]byte) = nil
			*dest[5].(*time.Time) = time.Time{}
			*dest[6].(*time.Time) = time.Time{}
			*dest[7].(**time.Time) = nil
			return nil
		}).Times(1)
	mockRows.EXPECT().Next().Return(false).Times(1)
	mockRows.EXPECT().Close().Times(1)
	mockRows.EXPECT().Err().Return(errors.New("row err")).Times(1)

	mockDB.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRows, nil)

	repo := NewSecretRepository(mockDB)
	_, err := repo.ListByUserID(context.Background(), 1)
	assert.Error(t, err)
}

func TestSecretRepository_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 0"), nil)

	repo := NewSecretRepository(mockDB)
	err := repo.Delete(context.Background(), "sec1", 1)
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestUserRepository_Create_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewUserRepository(mockDB)
	err := repo.Create(context.Background(), &domain.User{})
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}
