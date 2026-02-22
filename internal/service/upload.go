package service

import (
	"context"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/storage"
	"mime"
	"time"

	"github.com/google/uuid"
)

var allowedTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
}

type uploadService struct {
	storage       storage.Storage
	entityService EntityService
	uploadRepo    repository.UploadRepository
	uploadConfig  config.Upload
	ctDetector    contenttype.Detector
}

func NewUploadService(
	storage storage.Storage,
	entityService EntityService,
	uploadRepo repository.UploadRepository,
	uploadConfig config.Upload,
	ctDetector contenttype.Detector,
) *uploadService {
	return &uploadService{
		storage:       storage,
		entityService: entityService,
		uploadRepo:    uploadRepo,
		uploadConfig:  uploadConfig,
		ctDetector:    ctDetector,
	}
}

func (u *uploadService) SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error) {
	const op = "uploadService.SignURL"

	if err := validateExtAndContentType(req.Ext, req.ContentType); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err := u.entityService.Exists(ctx, req.Entity.Type, req.Entity.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uploadID := uuid.New()
	objectKey := u.generateObjectKey(req, uploadID)

	result, err := u.storage.CreatePresignedPost(ctx, storage.PresignedPostOptions{
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		MaxSize:     u.uploadConfig.MaxFileSize,
		ExpiresIn:   u.uploadConfig.PresignedUrlTTL,
		Metadata: map[string]string{
			"Upload-Id":   uploadID.String(),
			"Entity-Type": string(req.Entity.Type),
			"Entity-Id":   req.Entity.ID.String(),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response := &dto.SignURLResponse{
		UploadID:    uploadID,
		UploadURL:   result.URL,
		Filename:    objectKey,
		ContentType: req.ContentType,
		ExpireDate:  time.Now().UTC().Add(u.uploadConfig.PresignedUrlTTL),
		FormData:    result.Fields,
	}

	return response, nil
}

func (u *uploadService) Save(ctx context.Context, req dto.UploadRequest) (*dto.UploadResponse, error) {
	const op = "uploadService.Create"

	err := u.entityService.Exists(ctx, req.Entity.Type, req.Entity.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	obj, err := u.storage.GetObjectInfo(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateMetadata(req, obj.Metadata); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uploadExists, err := u.uploadRepo.Exists(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if uploadExists {
		return nil, apperrors.ErrFileAlreadyUploaded
	}

	file, err := u.storage.Open(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detectedCT, err := u.ctDetector.Detect(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateDetectedContentType(detectedCT); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	upload := &models.Upload{
		ObjectKey:   req.ObjectKey,
		EntityID:    req.Entity.ID,
		EntityType:  string(req.Entity.Type),
		FileSize:    obj.Size,
		ContentType: &detectedCT,
		IsMain:      false,
	}

	if err := u.uploadRepo.Create(ctx, upload); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	url := u.PublicURL(ctx, req.ObjectKey)

	response := &dto.UploadResponse{
		URL:         url,
		ContentType: upload.ContentType,
		IsMain:      upload.IsMain,
		CreatedAt:   upload.CreatedAt,
		UpdatedAt:   upload.UpdatedAt,
	}

	return response, nil
}

func (u *uploadService) PublicURL(ctx context.Context, objectKey string) string {
	return u.storage.PublicURL(ctx, objectKey)
}

func (u *uploadService) generateObjectKey(req dto.SignURLRequest, uploadID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		req.Entity.Type,
		req.Entity.ID,
		uploadID,
		req.Ext,
	)
}

func validateMetadata(req dto.UploadRequest, metadata map[string]string) error {
	if metadata["Upload-Id"] != req.UploadID.String() {
		return apperrors.ErrInvalidUploadID
	}

	if metadata["Entity-Id"] != req.Entity.ID.String() {
		return apperrors.ErrInvalidEntityID
	}

	if metadata["Entity-Type"] != string(req.Entity.Type) {
		return apperrors.ErrInvalidEntityID
	}

	return nil
}

func validateExtAndContentType(ext, contentType string) error {
	extension, ok := allowedTypes[contentType]
	if !ok {
		return apperrors.ErrInvalidImageFormat
	}

	if extension != ext {
		return apperrors.ErrContentTypeMismatch
	}

	return nil
}

func validateDetectedContentType(contentType string) error {
	baseType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}

	if _, ok := allowedTypes[baseType]; !ok {
		return apperrors.ErrInvalidImageFormat
	}

	return nil
}
