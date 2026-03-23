//go:build integration

package minio_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/repository/minio"
	"github.com/arvaliullin/goph-keeper/internal/repository/minio/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testContainer *testhelpers.MinioContainer

// TestMain настраивает тестовый контейнер MinIO для всех интеграционных тестов.
func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testContainer, err = testhelpers.NewMinioContainer(ctx)
	if err != nil {
		panic("не удалось запустить тестовый контейнер minio: " + err.Error())
	}

	code := m.Run()

	testContainer.Terminate(ctx)
	os.Exit(code)
}

func TestBinaryRepository_Integration_UploadDownloadDelete(t *testing.T) {
	ctx := context.Background()
	client, err := testContainer.NewClient()
	require.NoError(t, err)

	repo, err := minio.NewBinaryRepository(ctx, client, "test-bucket")
	require.NoError(t, err)

	payload := []byte("hello, gophkeeper!")
	objectName := "test/object1"

	t.Run("upload", func(t *testing.T) {
		err := repo.Upload(ctx, objectName, bytes.NewReader(payload), int64(len(payload)))
		assert.NoError(t, err)
	})

	t.Run("download", func(t *testing.T) {
		reader, err := repo.Download(ctx, objectName)
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, payload, data)
	})

	t.Run("delete", func(t *testing.T) {
		err := repo.Delete(ctx, objectName)
		assert.NoError(t, err)
	})

	t.Run("download after delete returns error", func(t *testing.T) {
		reader, err := repo.Download(ctx, objectName)
		if err != nil {
			return
		}
		defer reader.Close()
		_, err = io.ReadAll(reader)
		assert.Error(t, err)
	})
}

func TestBinaryRepository_Integration_BucketCreatedAutomatically(t *testing.T) {
	ctx := context.Background()
	client, err := testContainer.NewClient()
	require.NoError(t, err)

	repo, err := minio.NewBinaryRepository(ctx, client, "auto-bucket")
	require.NoError(t, err)
	assert.NotNil(t, repo)
}
