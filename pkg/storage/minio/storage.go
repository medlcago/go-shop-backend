package minio

import (
	"context"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/pkg/storage"
	"io"
	"path"
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
		return nil, fmt.Errorf("new minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket exists: %w", err)
	}

	if !ok {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &Storage{
		cli:     client,
		bucket:  cfg.Bucket,
		baseURL: cfg.BaseURL,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, opts storage.UploadOptions) (string, error) {
	_, err := s.cli.PutObject(ctx, s.bucket, objectKey, r, size, minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	})

	if err != nil {
		return "", fmt.Errorf("upload object: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, objectKey)
	return path.Clean(url), nil
}

func (s *Storage) Delete(ctx context.Context, objectKey string) error {
	err := s.cli.RemoveObject(
		ctx,
		s.bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func (s *Storage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.cli.PresignedGetObject(ctx, s.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("get presigned url: %w", err)
	}

	return url.String(), nil
}

func (s *Storage) CreatePresignedPost(ctx context.Context, opts storage.PresignedPostOptions) (*storage.PresignedPost, error) {
	policy := minio.NewPostPolicy()
	err := policy.SetBucket(s.bucket)
	if err != nil {
		return nil, fmt.Errorf("set bucket: %w", err)
	}

	if opts.KeyPrefixOnly {
		if err := policy.SetKeyStartsWith(opts.ObjectKey); err != nil {
			return nil, fmt.Errorf("set object key prefix: %w", err)
		}
	} else {
		if err := policy.SetKey(opts.ObjectKey); err != nil {
			return nil, fmt.Errorf("set object key: %w", err)
		}
	}

	if opts.ContentType != "" {
		if err := policy.SetContentType(opts.ContentType); err != nil {
			return nil, fmt.Errorf("set content type: %w", err)
		}
	}

	if opts.MaxSize > 0 {
		if err := policy.SetContentLengthRange(1, opts.MaxSize); err != nil {
			return nil, fmt.Errorf("set content length range: %w", err)
		}
	}

	if opts.ExpiresIn > 0 {
		if err := policy.SetExpires(time.Now().Add(opts.ExpiresIn)); err != nil {
			return nil, fmt.Errorf("set expiry: %w", err)
		}
	}

	for k, v := range opts.Metadata {
		if err := policy.SetUserMetadata(k, v); err != nil {
			return nil, fmt.Errorf("set user metadata: %w", err)
		}
	}

	url, formData, err := s.cli.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, fmt.Errorf("presigned post policy: %w", err)
	}

	return &storage.PresignedPost{
		URL:    url.String(),
		Fields: formData,
	}, nil
}

func (s *Storage) GetURL(_ context.Context, objectKey string) (string, error) {
	url := fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, objectKey)
	return url, nil
}

func (s *Storage) Exists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.cli.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return false, fmt.Errorf("stat object: %w", err)
	}

	return true, nil
}

func (s *Storage) Open(ctx context.Context, objectKey string) (io.ReadSeekCloser, error) {
	object, err := s.cli.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	return object, nil
}

func (s *Storage) GetObjectInfo(ctx context.Context, objectKey string) (*storage.ObjectInfo, error) {
	info, err := s.cli.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("stat object: %w", err)
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
