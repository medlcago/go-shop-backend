package storage

import (
	"context"
	"io"
	"time"
)

type UploadOptions struct {
	Metadata    map[string]string
	ContentType string
}

type PresignedPostOptions struct {
	ObjectKey     string
	ContentType   string
	MaxSize       int64
	ExpiresIn     time.Duration
	Metadata      map[string]string
	KeyPrefixOnly bool
}

type PresignedPost struct {
	URL    string
	Fields map[string]string
}

type Storage interface {
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, opts UploadOptions) (string, error)
	Delete(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	CreatePresignedPost(ctx context.Context, opts PresignedPostOptions) (*PresignedPost, error)
	GetURL(ctx context.Context, objectKey string) (string, error)
	Exists(ctx context.Context, objectKey string) (bool, error)
}
