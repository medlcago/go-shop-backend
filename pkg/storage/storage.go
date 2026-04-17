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

type UploadOptions struct {
	Metadata    map[string]string
	ContentType string
}

type TemporaryUploadURLOptions struct {
	ObjectKey   string
	ContentType string
	MaxSize     int64
	Expires     time.Time
	Metadata    map[string]string
}

type TemporaryUploadURL struct {
	URL    string
	Fields map[string]string
}

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
	Metadata     map[string]string
}

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
