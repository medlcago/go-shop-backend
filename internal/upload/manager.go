package upload

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/storage"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Manager interface {
	SignURL(ctx context.Context, req dto.GeneratePresignedURLRequest, uploadType Type) (*dto.GeneratePresignedURLResponse, error)
	Attach(ctx context.Context, req dto.AttachFileRequest, uploadType Type) (*dto.UploadResponse, error)
	PublicURL(objectKey string) string
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type uploadManager struct {
	storage      storage.Storage
	uploadRepo   repository.UploadRepository
	uploadConfig config.Upload
	ctDetector   contenttype.Detector
	registry     Registry
	logger       *slog.Logger
}

func New(
	storage storage.Storage,
	uploadRepo repository.UploadRepository,
	uploadConfig config.Upload,
	ctDetector contenttype.Detector,
	registry Registry,
	logger *slog.Logger,
) *uploadManager {
	return &uploadManager{
		storage:      storage,
		uploadRepo:   uploadRepo,
		uploadConfig: uploadConfig,
		ctDetector:   ctDetector,
		registry:     registry,
		logger:       logger,
	}
}

func (m *uploadManager) SignURL(ctx context.Context, req dto.GeneratePresignedURLRequest, uploadType Type) (*dto.GeneratePresignedURLResponse, error) {
	const op = "uploadManager.SignURL"

	policy, err := m.registry.Get(uploadType)
	if err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidUploadType)
		}

		return nil, apperror.Wrap(op, err)
	}

	if err := m.validateSignURLRequest(req, policy); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	uploadID := uuid.New()
	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Type": req.Entity.Type,
		"Entity-Id":   req.Entity.ID.String(),
	}

	objectKey := m.generateObjectKey(req.Entity, uploadID, req.Ext)
	effectiveMaxSize := policy.CalculateEffectiveMaxSize(m.uploadConfig.MaxFileSize)
	expireDate := time.Now().UTC().Add(m.uploadConfig.PresignedUrlTTL)

	options := storage.TemporaryUploadURLOptions{
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		MinSize:     policy.MinSize,
		MaxSize:     effectiveMaxSize,
		Expires:     expireDate,
		Metadata:    metadata,
	}

	result, err := m.storage.TemporaryUploadURL(ctx, options)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response := &dto.GeneratePresignedURLResponse{
		UploadID:    uploadID,
		UploadURL:   result.URL,
		Filename:    objectKey,
		ContentType: req.ContentType,
		ExpireDate:  expireDate,
		FormData:    result.Fields,
	}

	return response, nil
}

func (m *uploadManager) Attach(ctx context.Context, req dto.AttachFileRequest, uploadType Type) (*dto.UploadResponse, error) {
	const op = "uploadManager.Attach"

	obj, err := m.storage.GetObjectInfo(ctx, req.ObjectKey)
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"failed to get object info",
			logger.Err(err),
			logger.Op(op),
		)

		return nil, apperror.Wrap(op, apperror.ErrNotFound)
	}

	policy, err := m.registry.Get(uploadType)
	if err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidUploadType)
		}

		return nil, apperror.Wrap(op, err)
	}

	contentType, err := m.prepareAttach(ctx, obj, req, policy)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	upload := &models.Upload{
		ObjectKey:   req.ObjectKey,
		EntityID:    req.Entity.ID,
		EntityType:  models.EntityType(req.Entity.Type),
		FileSize:    obj.Size,
		ContentType: &contentType,
		MediaType:   models.UploadMediaTypeDefault,
	}

	if err := m.uploadRepo.Create(ctx, upload); err != nil {
		m.delete(ctx, req.ObjectKey)
		return nil, apperror.Wrap(op, err)
	}

	url := m.PublicURL(req.ObjectKey)

	response := &dto.UploadResponse{
		ID:          upload.ID,
		URL:         url,
		ContentType: upload.ContentType,
		MediaType:   string(upload.MediaType),
		CreatedAt:   upload.CreatedAt,
		UpdatedAt:   upload.UpdatedAt,
	}

	return response, nil
}

func (m *uploadManager) PublicURL(objectKey string) string {
	return m.storage.PublicURL(objectKey)
}

func (m *uploadManager) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const op = "uploadManager.Delete"

	if err := m.uploadRepo.DeleteByID(ctx, id); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (m *uploadManager) validateSignURLRequest(req dto.GeneratePresignedURLRequest, policy Policy) error {
	const op = "uploadManager.validateSignURLRequest"

	if !policy.IsValidExt(req.Ext, req.ContentType) {
		return apperror.Wrap(op, apperror.ErrContentTypeMismatch)
	}

	return nil
}

func (m *uploadManager) prepareAttach(
	ctx context.Context,
	obj *storage.ObjectInfo,
	req dto.AttachFileRequest,
	policy Policy,
) (contentType string, err error) {
	const op = "uploadManager.prepareAttach"

	if err = m.validateMetadata(req, obj.Metadata); err != nil {
		return "", apperror.Wrap(op, err)
	}

	if err = m.ensureNotDuplicate(ctx, req.ObjectKey); err != nil {
		return "", apperror.Wrap(op, err)
	}

	effectiveMaxSize := policy.CalculateEffectiveMaxSize(m.uploadConfig.MaxFileSize)
	if obj.Size > effectiveMaxSize {
		m.delete(ctx, req.ObjectKey)
		return "", apperror.Wrap(op, apperror.ErrFileTooLarge)
	}

	contentType, err = m.detectContentType(ctx, req.ObjectKey)
	if err != nil {
		return "", apperror.Wrap(op, err)
	}

	if !policy.IsValidContentType(contentType) {
		m.delete(ctx, req.ObjectKey)
		return "", apperror.Wrap(op, apperror.ErrInvalidFileType)
	}

	return contentType, nil
}

func (m *uploadManager) generateObjectKey(entity dto.UploadEntity, uploadID uuid.UUID, ext string) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		entity.Type,
		entity.ID,
		uploadID,
		ext,
	)
}

func (m *uploadManager) ensureNotDuplicate(ctx context.Context, objectKey string) error {
	const op = "uploadManager.ensureNotDuplicate"

	exists, err := m.uploadRepo.ExistsByObjectKey(ctx, objectKey)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if exists {
		return apperror.Wrap(op, apperror.ErrFileAlreadyUploaded)
	}

	return nil
}

func (m *uploadManager) detectContentType(ctx context.Context, objectKey string) (string, error) {
	const op = "uploadManager.detectContentType"

	file, err := m.storage.Get(ctx, objectKey)
	if err != nil {
		return "", apperror.Wrap(op, err)
	}
	defer func() {
		_ = file.Close()
	}()

	ct, err := m.ctDetector.Detect(file)
	if err != nil {
		return "", apperror.Wrap(op, err)
	}

	return ct, nil
}

func (m *uploadManager) delete(ctx context.Context, objectKey string) {
	const op = "uploadManager.delete"

	if err := m.storage.Delete(ctx, objectKey); err != nil {
		m.logger.ErrorContext(
			ctx,
			"failed to delete object from storage",
			logger.Err(err),
			logger.Op(op),
		)
	}
}

func (m *uploadManager) validateMetadata(req dto.AttachFileRequest, metadata map[string]string) error {
	const op = "uploadManager.validateMetadata"

	if metadata["Upload-Id"] != req.UploadID.String() {
		return apperror.Wrap(op, apperror.ErrInvalidUploadID)
	}

	if metadata["Entity-Id"] != req.Entity.ID.String() {
		return apperror.Wrap(op, apperror.ErrInvalidEntityID)
	}

	if metadata["Entity-Type"] != req.Entity.Type {
		return apperror.Wrap(op, apperror.ErrInvalidEntityType)
	}

	return nil
}
