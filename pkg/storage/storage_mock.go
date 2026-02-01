package storage

import (
	"context"
	"io"
	"time"

	"github.com/stretchr/testify/mock"
)

var _ Storage = (*MockStorage)(nil)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, opts UploadOptions) (string, error) {
	args := m.Called(ctx, objectKey, r, size, opts)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

func (m *MockStorage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) CreatePresignedPost(ctx context.Context, opts PresignedPostOptions) (*PresignedPost, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*PresignedPost), args.Error(1)
}

func (m *MockStorage) PublicURL(ctx context.Context, objectKey string) string {
	args := m.Called(ctx, objectKey)
	return args.String(0)
}

func (m *MockStorage) Exists(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

func (m *MockStorage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	args := m.Called(ctx, objectKey)
	return args.Get(0).(io.ReadSeekCloser), args.Error(1)
}

func (m *MockStorage) GetObjectInfo(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	args := m.Called(ctx, objectKey)
	return args.Get(0).(*ObjectInfo), args.Error(1)
}
