package minio

import (
	"context"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/pkg/storage"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ storage.Storage = (*Storage)(nil)

type Storage struct {
	cli     *minio.Client
	bucket  string
	baseURL string
}

func New(cfg config.Minio) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket exists: %w", err)
	}

	if !ok {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Storage{
		cli:     client,
		bucket:  cfg.Bucket,
		baseURL: cfg.BaseURL,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, opts storage.UploadOptions) (string, error) {
	const op = "minio.Storage.Upload"

	_, err := s.cli.PutObject(ctx, s.bucket, objectKey, r, size, minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return s.PublicURL(objectKey), nil
}

func (s *Storage) Delete(ctx context.Context, objectKey string) error {
	const op = "minio.Storage.Delete"

	err := s.cli.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	const op = "minio.Storage.GetPresignedURL"

	u, err := s.cli.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return u.String(), nil
}

func (s *Storage) CreatePresignedPost(ctx context.Context, opts storage.PresignedPostOptions) (*storage.PresignedPost, error) {
	const op = "minio.Storage.CreatePresignedPost"

	policy := minio.NewPostPolicy()

	if err := policy.SetBucket(s.bucket); err != nil {
		return nil, fmt.Errorf("%s: set bucket: %w", op, err)
	}

	if err := policy.SetKey(opts.ObjectKey); err != nil {
		return nil, fmt.Errorf("%s: set key: %w", op, err)
	}

	if opts.ContentType != "" {
		if err := policy.SetContentType(opts.ContentType); err != nil {
			return nil, fmt.Errorf("%s: set content type: %w", op, err)
		}
	}

	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = 5 << 20 // 5MB
	}
	if err := policy.SetContentLengthRange(1, maxSize); err != nil {
		return nil, fmt.Errorf("%s: set length range: %w", op, err)
	}

	expireTime := time.Now().UTC().Add(10 * time.Minute)
	if opts.ExpiresIn > 0 {
		expireTime = time.Now().UTC().Add(opts.ExpiresIn)
	}
	if err := policy.SetExpires(expireTime); err != nil {
		return nil, fmt.Errorf("%s: set expiry: %w", op, err)
	}

	for k, v := range opts.Metadata {
		if err := policy.SetUserMetadata(k, v); err != nil {
			return nil, fmt.Errorf("%s: set metadata %s: %w", op, k, err)
		}
	}

	u, formData, err := s.cli.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, fmt.Errorf("%s: generate policy: %w", op, err)
	}

	return &storage.PresignedPost{
		URL:    u.String(),
		Fields: formData,
	}, nil
}

func (s *Storage) PublicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, objectKey)
}
func (s *Storage) Exists(ctx context.Context, objectKey string) error {
	const op = "minio.Storage.Exists"

	_, err := s.cli.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == minio.NoSuchKey {
			return storage.ErrNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	const op = "minio.Storage.Open"

	obj, err := s.cli.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()

		errResp := minio.ToErrorResponse(err)
		if errResp.Code == minio.NoSuchKey {
			return nil, storage.ErrNotFound
		}

		return nil, fmt.Errorf("%s: stat failed: %w", op, err)
	}

	return obj, nil
}

func (s *Storage) GetObjectInfo(ctx context.Context, objectKey string) (*storage.ObjectInfo, error) {
	const op = "minio.Storage.GetObjectInfo"

	info, err := s.cli.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &storage.ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
		ETag:         info.ETag,
		Metadata:     info.UserMetadata,
	}, nil
}
