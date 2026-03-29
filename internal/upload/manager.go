package upload

import (
	"context"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/storage"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Manager interface {
	SignURL(ctx context.Context, req SignURLRequest, policy Policy) (*SignURLResponse, error)
	Save(ctx context.Context, req SaveUploadRequest, policy Policy) (*ContentResponse, error)
	PublicURL(objectKey string) string
}

type uploadManager struct {
	storage        storage.Storage
	uploadRepo     repository.UploadRepository
	uploadConfig   config.Upload
	ctDetector     contenttype.Detector
	policyProvider PolicyProvider
	logger         *slog.Logger
}

func NewManager(
	storage storage.Storage,
	uploadRepo repository.UploadRepository,
	uploadConfig config.Upload,
	ctDetector contenttype.Detector,
	policyProvider PolicyProvider,
	logger *slog.Logger,
) *uploadManager {
	return &uploadManager{
		storage:        storage,
		uploadRepo:     uploadRepo,
		uploadConfig:   uploadConfig,
		ctDetector:     ctDetector,
		logger:         logger,
		policyProvider: policyProvider,
	}
}

func (m *uploadManager) SignURL(ctx context.Context, req SignURLRequest, policy Policy) (*SignURLResponse, error) {
	const op = "uploadManager.SignURL"

	constraints, err := m.policyProvider.Get(policy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !constraints.IsValidExt(req.Ext, req.ContentType) {
		return nil, apperrors.ErrContentTypeMismatch
	}

	uploadID := uuid.New()
	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Type": string(req.Entity.Type),
		"Entity-Id":   req.Entity.ID.String(),
	}

	objectKey := m.generateObjectKey(req, uploadID)
	maxSize := min(constraints.MaxSize, m.uploadConfig.MaxFileSize)

	options := storage.PresignedPostOptions{
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		MaxSize:     maxSize,
		ExpiresIn:   m.uploadConfig.PresignedUrlTTL,
		Metadata:    metadata,
	}

	result, err := m.storage.CreatePresignedPost(ctx, options)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	expireDate := time.Now().UTC().Add(m.uploadConfig.PresignedUrlTTL)

	response := &SignURLResponse{
		UploadID:    uploadID,
		UploadURL:   result.URL,
		Filename:    objectKey,
		ContentType: req.ContentType,
		ExpireDate:  expireDate,
		FormData:    result.Fields,
	}

	return response, nil
}

func (m *uploadManager) Save(ctx context.Context, req SaveUploadRequest, policy Policy) (*ContentResponse, error) {
	const op = "uploadManager.Save"

	log := m.logger.With(
		"op", op,
		"object_key", req.ObjectKey,
	)

	obj, err := m.storage.GetObjectInfo(ctx, req.ObjectKey)
	if err != nil {
		log.Error("storage.GetObjectInfo failed", logger.AppErr(apperrors.ErrNotFound), logger.Err(err))
		return nil, apperrors.ErrNotFound
	}

	if err := validateMetadata(req, obj.Metadata); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uploadExists, err := m.uploadRepo.Exists(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if uploadExists {
		return nil, apperrors.ErrFileAlreadyUploaded
	}

	constraints, err := m.policyProvider.Get(policy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	maxSize := min(constraints.MaxSize, m.uploadConfig.MaxFileSize)

	if obj.Size > maxSize {
		if delErr := m.storage.Delete(ctx, req.ObjectKey); delErr != nil {
			log.Error("storage.Delete failed", logger.AppErr(apperrors.ErrFileTooLarge), logger.Err(delErr))
		}
		return nil, apperrors.ErrFileTooLarge
	}

	file, err := m.storage.Open(ctx, req.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		_ = file.Close()
	}()

	detectedCT, err := m.ctDetector.Detect(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !constraints.IsValidType(detectedCT) {
		if delErr := m.storage.Delete(ctx, req.ObjectKey); delErr != nil {
			log.Error("storage.Delete failed", logger.AppErr(apperrors.ErrInvalidFileType), logger.Err(delErr))
		}
		return nil, apperrors.ErrInvalidFileType
	}

	upload := &models.Upload{
		ObjectKey:   req.ObjectKey,
		EntityID:    req.Entity.ID,
		EntityType:  string(req.Entity.Type),
		FileSize:    obj.Size,
		ContentType: &detectedCT,
		IsMain:      req.IsMain,
	}

	if err := m.uploadRepo.Create(ctx, upload); err != nil {
		if delErr := m.storage.Delete(ctx, upload.ObjectKey); delErr != nil {
			m.logger.Error("storage.Delete failed", logger.AppErr(err), logger.Err(delErr))
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	url := m.PublicURL(req.ObjectKey)

	response := &ContentResponse{
		URL:         url,
		ContentType: upload.ContentType,
		IsMain:      upload.IsMain,
		CreatedAt:   upload.CreatedAt,
		UpdatedAt:   upload.UpdatedAt,
	}

	return response, nil
}

func (m *uploadManager) PublicURL(objectKey string) string {
	return m.storage.PublicURL(objectKey)
}

func (m *uploadManager) generateObjectKey(req SignURLRequest, uploadID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		req.Entity.Type,
		req.Entity.ID,
		uploadID,
		req.Ext,
	)
}

func validateMetadata(req SaveUploadRequest, metadata map[string]string) error {
	if metadata["Upload-Id"] != req.UploadID.String() {
		return apperrors.ErrInvalidUploadID
	}

	if metadata["Entity-Id"] != req.Entity.ID.String() {
		return apperrors.ErrInvalidEntityID
	}

	if metadata["Entity-Type"] != string(req.Entity.Type) {
		return apperrors.ErrInvalidEntityType
	}

	return nil
}
