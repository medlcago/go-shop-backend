package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, opts UploadOptions) (string, error)
	Delete(ctx context.Context, objectKey string) error
	TemporaryURL(ctx context.Context, objectKey string, expires time.Duration) (string, error)
	TemporaryUploadURL(ctx context.Context, opts TemporaryUploadURLOptions) (*TemporaryUploadURL, error)
	PublicURL(objectKey string) string
	Exists(ctx context.Context, objectKey string) error
	Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error)
	GetObjectInfo(ctx context.Context, objectKey string) (*ObjectInfo, error)
}
