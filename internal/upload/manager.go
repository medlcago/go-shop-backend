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

const (
	MethodPOST = "POST"
)

type Manager interface {
	SignURL(ctx context.Context, req dto.GeneratePresignedURLRequest) (*dto.GeneratePresignedURLResponse, error)
	Attach(ctx context.Context, req dto.AttachFileRequest) (*dto.AttachFileResponse, error)
	PublicURL(objectKey string) string
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

func (m *uploadManager) SignURL(ctx context.Context, req dto.GeneratePresignedURLRequest) (*dto.GeneratePresignedURLResponse, error) {
	const op = "uploadManager.SignURL"

	reg, err := m.resolveRegistry(ctx, op, Purpose(req.Purpose))
	if err != nil {
		return nil, err
	}

	if err := m.validateSignURLRequest(req, reg.Policy); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := reg.Handler.CanSignURL(ctx, req.Entity.Type, req.Entity.ID); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	uploadID, expireDate, options := m.prepareUploadOptions(req, reg.Policy)

	result, err := m.storage.TemporaryUploadURL(ctx, options)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response := &dto.GeneratePresignedURLResponse{
		UploadID:    uploadID,
		UploadURL:   result.URL,
		Method:      MethodPOST,
		Filename:    options.ObjectKey,
		ContentType: req.ContentType,
		ExpireDate:  expireDate,
		FormData:    result.Fields,
	}

	return response, nil
}

func (m *uploadManager) Attach(ctx context.Context, req dto.AttachFileRequest) (*dto.AttachFileResponse, error) {
	const op = "uploadManager.Attach"

	reg, err := m.resolveRegistry(ctx, op, Purpose(req.Purpose))
	if err != nil {
		return nil, err
	}

	obj, err := m.fetchObjectInfo(ctx, op, req.ObjectKey)
	if err != nil {
		return nil, err
	}

	if err := m.validateMetadata(req, obj.Metadata); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := m.ensureNotDuplicate(ctx, req.ObjectKey); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := m.ensureWithinSize(ctx, req.ObjectKey, obj.Size, reg.Policy); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	contentType, err := m.resolveContentType(ctx, req.ObjectKey, reg.Policy)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := reg.Handler.CanAttach(ctx, req.Entity.Type, req.Entity.ID); err != nil {
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
		return nil, apperror.Wrap(op, err)
	}

	url := m.PublicURL(req.ObjectKey)

	response := &dto.AttachFileResponse{
		ID:  upload.ID,
		URL: url,
	}

	return response, nil
}

func (m *uploadManager) PublicURL(objectKey string) string {
	return m.storage.PublicURL(objectKey)
}

func (m *uploadManager) resolveRegistry(ctx context.Context, op string, purpose Purpose) (*Registration, error) {
	reg, err := m.registry.Get(purpose)
	if err != nil {
		m.logger.InfoContext(
			ctx,
			"unknown purpose",
			slog.String("purpose", string(purpose)),
			logger.Err(err),
			logger.Op(op),
		)

		if errors.Is(err, ErrPurposeNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidPurpose)
		}

		return nil, apperror.Wrap(op, err)
	}

	return reg, nil
}

func (m *uploadManager) fetchObjectInfo(ctx context.Context, op string, objectKey string) (*storage.ObjectInfo, error) {
	obj, err := m.storage.GetObjectInfo(ctx, objectKey)
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"failed to get object info",
			logger.Err(err),
			logger.Op(op),
		)

		if errors.Is(err, storage.ErrNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrFileNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return obj, nil
}

func (m *uploadManager) ensureWithinSize(ctx context.Context, objectKey string, size int64, policy Policy) error {
	const op = "uploadManager.ensureWithinSize"

	effectiveMax := policy.CalculateEffectiveMaxSize(m.uploadConfig.MaxFileSize)
	if size <= effectiveMax {
		return nil
	}

	m.delete(ctx, objectKey)
	return apperror.Wrap(op, apperror.ErrFileTooLarge)
}

func (m *uploadManager) prepareUploadOptions(
	req dto.GeneratePresignedURLRequest,
	policy Policy,
) (uuid.UUID, time.Time, storage.TemporaryUploadURLOptions) {
	uploadID := uuid.New()
	expireDate := time.Now().UTC().Add(m.uploadConfig.PresignedURLTTL)
	objectKey := m.generateObjectKey(req.Entity, uploadID, req.Ext)
	effectiveMaxSize := policy.CalculateEffectiveMaxSize(m.uploadConfig.MaxFileSize)

	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Type": req.Entity.Type,
		"Entity-Id":   req.Entity.ID.String(),
		"Purpose":     req.Purpose,
	}

	options := storage.TemporaryUploadURLOptions{
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		MinSize:     policy.MinSize,
		MaxSize:     effectiveMaxSize,
		Expires:     expireDate,
		Metadata:    metadata,
	}

	return uploadID, expireDate, options
}

func (m *uploadManager) validateSignURLRequest(req dto.GeneratePresignedURLRequest, policy Policy) error {
	const op = "uploadManager.validateSignURLRequest"

	if !policy.IsValidExt(req.Ext, req.ContentType) {
		return apperror.Wrap(op, apperror.ErrContentTypeMismatch)
	}

	return nil
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

func (m *uploadManager) resolveContentType(
	ctx context.Context,
	objectKey string,
	policy Policy,
) (string, error) {
	const op = "uploadManager.resolveContentType"

	contentType, err := m.detectContentType(ctx, objectKey)
	if err != nil {
		return "", apperror.Wrap(op, err)
	}

	if !policy.IsValidContentType(contentType) {
		m.delete(ctx, objectKey)
		return "", apperror.Wrap(op, apperror.ErrInvalidFileType)
	}

	return contentType, nil
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

	if metadata["Purpose"] != req.Purpose {
		return apperror.Wrap(op, apperror.ErrInvalidPurpose)
	}

	return nil
}
