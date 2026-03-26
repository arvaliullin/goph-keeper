package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

	repo := NewSecretRepository(mockDB)
	err := repo.Create(context.Background(), &domain.Secret{ID: "sec1", UserID: 1})
	assert.NoError(t, err)
}

func TestSecretRepository_Update(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("UPDATE 1"), nil)

	repo := NewSecretRepository(mockDB)
	err := repo.Update(context.Background(), &domain.Secret{ID: "sec1", UserID: 1})
	assert.NoError(t, err)
}

func TestSecretRepository_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgconn.NewCommandTag("DELETE 1"), nil)

	repo := NewSecretRepository(mockDB)
	err := repo.Delete(context.Background(), "sec1", 1)
	assert.NoError(t, err)
}

func TestSecretRepository_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	mockRow := mocks.NewMockRow(ctrl)

	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*string) = "sec1"
			*dest[1].(*int64) = 1
			*dest[2].(*domain.SecretType) = domain.SecretTypeLogin
			*dest[3].(*[]byte) = []byte("data")
			*dest[4].(*[]byte) = []byte("meta")
			*dest[5].(*time.Time) = time.Now()
			*dest[6].(*time.Time) = time.Now()
			*dest[7].(**time.Time) = nil
			return nil
		})
	mockDB.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockRow)

	repo := NewSecretRepository(mockDB)
	sec, err := repo.GetByID(context.Background(), "sec1", 1)
	assert.NoError(t, err)
	assert.Equal(t, "sec1", sec.ID)
}

func setupMockRows(ctrl *gomock.Controller, data []*domain.Secret) *mocks.MockRows {
	mockRows := mocks.NewMockRows(ctrl)
	cursor := 0

	mockRows.EXPECT().Next().DoAndReturn(func() bool {
		if cursor < len(data) {
			cursor++
			return true
		}
		return false
	}).Times(len(data) + 1)

	for i := range data {
		sec := data[i]
		mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = sec.ID
				*dest[1].(*int64) = sec.UserID
				*dest[2].(*domain.SecretType) = sec.Type
				*dest[3].(*[]byte) = sec.Data
				*dest[4].(*[]byte) = sec.Metadata
				*dest[5].(*time.Time) = sec.CreatedAt
				*dest[6].(*time.Time) = sec.UpdatedAt
				*dest[7].(**time.Time) = sec.DeletedAt
				return nil
			}).Times(1)
	}

	mockRows.EXPECT().Close().Times(1)
	mockRows.EXPECT().Err().Return(nil).Times(1)

	return mockRows
}

func TestSecretRepository_ListByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	rows := setupMockRows(ctrl, []*domain.Secret{
		{ID: "sec1", UserID: 1, Type: domain.SecretTypeLogin},
	})
	mockDB.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(rows, nil)

	repo := NewSecretRepository(mockDB)
	secs, err := repo.ListByUserID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, secs, 1)
}

func TestSecretRepository_SyncByUserID(t *testing.T) {
	deletedAt := time.Now()
	ctrl := gomock.NewController(t)

	mockDB := mocks.NewMockPostgresClient(ctrl)
	rows := setupMockRows(ctrl, []*domain.Secret{
		{ID: "sec1", UserID: 1, Type: domain.SecretTypeLogin},
		{ID: "sec2", UserID: 1, Type: domain.SecretTypeBinary, DeletedAt: &deletedAt},
	})
	mockDB.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(rows, nil)

	repo := NewSecretRepository(mockDB)
	secs, err := repo.SyncByUserID(context.Background(), 1, time.Now().Add(-time.Hour))
	assert.NoError(t, err)
	assert.Len(t, secs, 2)
	assert.NotNil(t, secs[1].DeletedAt)
}
