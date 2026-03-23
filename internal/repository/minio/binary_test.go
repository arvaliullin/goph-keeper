package minio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	retrypkg "github.com/arvaliullin/goph-keeper/internal/pkg/retry"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMinioClient struct {
	bucketExistsFunc func(ctx context.Context, bucketName string) (bool, error)
	makeBucketFunc   func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	putObjectFunc    func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	getObjectFunc    func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	removeObjectFunc func(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

func (m stubMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	if m.bucketExistsFunc != nil {
		return m.bucketExistsFunc(ctx, bucketName)
	}
	return true, nil
}
func (m stubMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	if m.makeBucketFunc != nil {
		return m.makeBucketFunc(ctx, bucketName, opts)
	}
	return nil
}
func (m stubMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, bucketName, objectName, reader, objectSize, opts)
	}
	return minio.UploadInfo{}, nil
}
func (m stubMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, bucketName, objectName, opts)
	}
	return nil, nil
}
func (m stubMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	if m.removeObjectFunc != nil {
		return m.removeObjectFunc(ctx, bucketName, objectName, opts)
	}
	return nil
}

func TestBinaryRepository_UploadDownloadDelete(t *testing.T) {
	client := stubMinioClient{
		bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
			return false, nil
		},
	}
	repo, err := NewBinaryRepositoryWithStrategy(
		context.Background(),
		client,
		"test",
		retrypkg.NewStrategy([]time.Duration{0}, IsRetryable),
	)
	assert.NoError(t, err)

	err = repo.Upload(context.Background(), "obj", bytes.NewReader([]byte("payload")), int64(len("payload")))
	assert.NoError(t, err)

	_, err = repo.Download(context.Background(), "obj")
	assert.NoError(t, err)

	err = repo.Delete(context.Background(), "obj")
	assert.NoError(t, err)
}

func TestBinaryRepository_Exists(t *testing.T) {
	client := stubMinioClient{
		bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
			return true, nil
		},
	}
	repo, err := NewBinaryRepository(context.Background(), client, "test")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestBinaryRepository_RetryBucketExists(t *testing.T) {
	attempts := 0
	client := stubMinioClient{
		bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
			attempts++
			if attempts == 1 {
				return false, temporaryNetError{}
			}
			return true, nil
		},
	}

	repo, err := NewBinaryRepositoryWithStrategy(
		context.Background(),
		client,
		"test",
		retrypkg.NewStrategy([]time.Duration{0}, IsRetryable),
	)
	require.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, 2, attempts)
}

func TestBinaryRepository_RetryUploadReusesPayload(t *testing.T) {
	attempts := 0
	var payloads [][]byte
	client := stubMinioClient{
		putObjectFunc: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
			attempts++
			payload, err := io.ReadAll(reader)
			require.NoError(t, err)
			payloads = append(payloads, payload)
			if attempts == 1 {
				return minio.UploadInfo{}, temporaryNetError{}
			}
			return minio.UploadInfo{}, nil
		},
	}

	repo, err := NewBinaryRepositoryWithStrategy(
		context.Background(),
		client,
		"test",
		retrypkg.NewStrategy([]time.Duration{0}, IsRetryable),
	)
	require.NoError(t, err)

	err = repo.Upload(context.Background(), "obj", bytes.NewReader([]byte("payload")), int64(len("payload")))
	require.NoError(t, err)
	require.Len(t, payloads, 2)
	assert.Equal(t, []byte("payload"), payloads[0])
	assert.Equal(t, []byte("payload"), payloads[1])
}

func TestBinaryRepository_RetryDelete(t *testing.T) {
	attempts := 0
	client := stubMinioClient{
		removeObjectFunc: func(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
			attempts++
			if attempts == 1 {
				return minio.ErrorResponse{Code: "ServiceUnavailable", StatusCode: 503}
			}
			return nil
		},
	}

	repo, err := NewBinaryRepositoryWithStrategy(
		context.Background(),
		client,
		"test",
		retrypkg.NewStrategy([]time.Duration{0}, IsRetryable),
	)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), "obj")
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestIsRetryable(t *testing.T) {
	assert.True(t, IsRetryable(temporaryNetError{}))
	assert.True(t, IsRetryable(minio.ErrorResponse{Code: "InternalError", StatusCode: 500}))
	assert.False(t, IsRetryable(context.Canceled))
	assert.False(t, IsRetryable(errors.New("permanent error")))
	assert.False(t, IsRetryable(minio.ErrorResponse{Code: minio.NoSuchKey, StatusCode: 404}))
}

type temporaryNetError struct{}

func (temporaryNetError) Error() string   { return "temporary network error" }
func (temporaryNetError) Timeout() bool   { return true }
func (temporaryNetError) Temporary() bool { return true }

var _ net.Error = (*temporaryNetError)(nil)
