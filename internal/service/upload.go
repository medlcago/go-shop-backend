package service

import (
	"context"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/storage"
	"go-shop-backend/pkg/utils"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

var supportedImageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
}

type uploadService struct {
	storage       storage.Storage
	entityService EntityService
	uploadRepo    repository.UploadRepository
	uploadConfig  config.Upload
}

func NewUploadService(
	storage storage.Storage,
	entityService EntityService,
	uploadRepo repository.UploadRepository,
	uploadConfig config.Upload,
) UploadService {
	return &uploadService{
		storage:       storage,
		entityService: entityService,
		uploadRepo:    uploadRepo,
		uploadConfig:  uploadConfig,
	}
}

func (u *uploadService) SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error) {
	const op = "uploadService.SignURL"

	if err := validateExtAndContentType(req.Ext, req.ContentType); err != nil {
		return nil, err
	}

	exists, err := u.entityService.Exists(
		ctx,
		req.Entity.Type,
		req.Entity.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		return nil, apperrors.ErrEntityNotFound
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

	resp := &dto.SignURLResponse{
		UploadID:    uploadID,
		UploadURL:   result.URL,
		Filename:    objectKey,
		ContentType: req.ContentType,
		ExpireDate:  time.Now().UTC().Add(u.uploadConfig.PresignedUrlTTL),
		FormData:    result.Fields,
	}

	return resp, nil
}

func (u *uploadService) Save(ctx context.Context, req dto.UploadRequest) (*dto.UploadResponse, error) {
	const op = "uploadService.Save"

	entityExists, err := u.entityService.Exists(ctx, req.Entity.Type, req.Entity.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !entityExists {
		return nil, apperrors.ErrEntityNotFound
	}

	obj, err := u.storage.GetObjectInfo(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if obj.Metadata["Upload-Id"] != req.UploadID.String() {
		return nil, apperrors.ErrInvalidUploadID
	}

	if obj.Metadata["Entity-Id"] != req.Entity.ID.String() {
		return nil, apperrors.ErrInvalidEntityID
	}

	if obj.Metadata["Entity-Type"] != string(req.Entity.Type) {
		return nil, apperrors.ErrInvalidEntityID
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
	defer file.Close()

	detectedCT, err := utils.DetectContentType(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateDetectedContentType(detectedCT); err != nil {
		return nil, err
	}

	upload := &models.Upload{
		ObjectKey:   req.ObjectKey,
		EntityID:    req.Entity.ID,
		EntityType:  string(req.Entity.Type),
		FileSize:    obj.Size,
		ContentType: utils.Ptr(detectedCT),
		IsMain:      false,
	}

	if err := u.uploadRepo.Save(ctx, upload); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	url := u.PublicURL(ctx, req.ObjectKey)

	resp := &dto.UploadResponse{
		URL:         url,
		ContentType: upload.ContentType,
		IsMain:      upload.IsMain,
		CreatedAt:   upload.CreatedAt,
		UpdatedAt:   upload.UpdatedAt,
	}

	return resp, nil
}

func (u *uploadService) PublicURL(ctx context.Context, objectKey string) string {
	return u.storage.PublicURL(ctx, objectKey)
}

func validateExtAndContentType(ext, contentType string) error {
	if _, ok := supportedImageContentTypes[contentType]; !ok {
		return apperrors.ErrInvalidImageFormat
	}

	exts, err := utils.ContentTypeToExt(contentType)
	if err != nil {
		return err
	}

	for _, e := range exts {
		if strings.TrimPrefix(e, ".") == ext {
			return nil
		}
	}

	return apperrors.ErrContentTypeMismatch
}

func validateDetectedContentType(contentType string) error {
	baseType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}

	if _, ok := supportedImageContentTypes[baseType]; !ok {
		return apperrors.ErrInvalidImageFormat
	}

	return nil
}

func (u *uploadService) generateObjectKey(req dto.SignURLRequest, uploadID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		req.Entity.Type,
		req.Entity.ID,
		uploadID,
		req.Ext,
	)
}
