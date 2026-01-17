package service

import (
	"context"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/storage"
	"time"

	"github.com/google/uuid"
)

const (
	imageFormatJPEG = "jpg"
	imageFormatPNG  = "png"

	maxFileSize     = 5 << 20 // 5MB
	presignedURLTTL = 20 * time.Minute
)

var supportedImageFormats = map[string]string{
	imageFormatJPEG: "image/jpeg",
	imageFormatPNG:  "image/png",
}

type uploadService struct {
	storage       storage.Storage
	entityService EntityService
}

func NewUploadService(storage storage.Storage, entityService EntityService) UploadService {
	return &uploadService{
		storage:       storage,
		entityService: entityService,
	}
}

func (u *uploadService) SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error) {
	expectedContentType, err := u.validateImageFormat(req)
	if err != nil {
		return nil, err
	}

	exists, err := u.entityService.Exists(
		ctx,
		req.Entity.Type,
		req.Entity.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("sign url: %w", err)
	}

	if !exists {
		return nil, apperrors.ErrEntityNotFound
	}

	uploadID := uuid.New()
	objectKey := u.generateObjectKey(req, uploadID)

	result, err := u.storage.CreatePresignedPost(ctx, storage.PresignedPostOptions{
		ObjectKey:   objectKey,
		ContentType: expectedContentType,
		MaxSize:     maxFileSize,
		ExpiresIn:   presignedURLTTL,
		Metadata: map[string]string{
			"upload_id": uploadID.String(),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("sign url: %w", err)
	}

	resp := &dto.SignURLResponse{
		ID:          uploadID,
		URL:         result.URL,
		Filename:    objectKey,
		ContentType: req.ContentType,
		ExpireDate:  time.Now().UTC().Add(presignedURLTTL),
		FormData:    result.Fields,
	}

	return resp, nil
}

func (u *uploadService) validateImageFormat(req dto.SignURLRequest) (string, error) {
	expectedCT, ok := supportedImageFormats[req.Ext]
	if !ok {
		return "", apperrors.ErrInvalidImageFormat
	}

	if req.ContentType != expectedCT {
		return "", apperrors.ErrContentTypeMismatch
	}

	return expectedCT, nil
}

func (u *uploadService) generateObjectKey(req dto.SignURLRequest, uploadID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		req.Entity.Type,
		req.Entity.ID,
		uploadID,
		req.Ext,
	)
}
