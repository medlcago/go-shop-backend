package upload

import (
	"context"
	"fmt"
	"go-shop-backend/config"
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
		policyProvider: policyProvider,
		logger:         logger,
	}
}

func (m *uploadManager) SignURL(ctx context.Context, req SignURLRequest, policy Policy) (*SignURLResponse, error) {
	const op = "uploadManager.SignURL"

	constraints, err := m.policyProvider.Get(policy)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := m.validateSignURLRequest(req, constraints); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	uploadID := uuid.New()
	metadata := map[string]string{
		"Upload-Id":   uploadID.String(),
		"Entity-Type": string(req.Entity.Type),
		"Entity-Id":   req.Entity.ID.String(),
	}

	objectKey := m.generateObjectKey(req.Entity, uploadID, req.Ext)
	effectiveMaxSize := m.effectiveMaxSize(constraints)
	expireDate := time.Now().UTC().Add(m.uploadConfig.PresignedUrlTTL)

	options := storage.TemporaryUploadURLOptions{
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		MaxSize:     effectiveMaxSize,
		Expires:     expireDate,
		Metadata:    metadata,
	}

	result, err := m.storage.TemporaryUploadURL(ctx, options)

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

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

	obj, err := m.storage.GetObjectInfo(ctx, req.ObjectKey)
	if err != nil {
		return nil, apperror.Wrap(op, apperror.ErrNotFound)
	}

	if err := m.validateMetadata(req, obj.Metadata); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := m.ensureNotDuplicate(ctx, req.ObjectKey); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	constraints, err := m.policyProvider.Get(policy)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	effectiveMaxSize := m.effectiveMaxSize(constraints)
	if obj.Size > effectiveMaxSize {
		m.delete(ctx, req.ObjectKey)
		return nil, apperror.Wrap(op, apperror.ErrFileTooLarge)
	}

	detectedCT, err := m.detectContentType(ctx, req.ObjectKey)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !constraints.IsValidType(detectedCT) {
		m.delete(ctx, req.ObjectKey)
		return nil, apperror.Wrap(op, apperror.ErrInvalidFileType)
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
		m.delete(ctx, req.ObjectKey)
		return nil, apperror.Wrap(op, err)
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

func (m *uploadManager) validateSignURLRequest(req SignURLRequest, constraints FileConstraints) error {
	const op = "uploadManager.validateSignURLRequest"

	if !constraints.IsValidExt(req.Ext, req.ContentType) {
		return apperror.Wrap(op, apperror.ErrContentTypeMismatch)
	}

	return nil
}

func (m *uploadManager) generateObjectKey(entity Entity, uploadID uuid.UUID, ext string) string {
	return fmt.Sprintf("%s/%s/%s.%s",
		entity.Type,
		entity.ID,
		uploadID,
		ext,
	)
}

func (m *uploadManager) ensureNotDuplicate(ctx context.Context, objectKey string) error {
	const op = "uploadManager.ensureNotDuplicate"

	exists, err := m.uploadRepo.Exists(ctx, objectKey)
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

	file, err := m.storage.Open(ctx, objectKey)
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
		m.logger.Error(
			"failed to delete object from storage",
			logger.Err(err),
			slog.String("op", op),
		)
	}
}

func (m *uploadManager) effectiveMaxSize(constraints FileConstraints) int64 {
	return min(constraints.MaxSize, m.uploadConfig.MaxFileSize)
}

func (m *uploadManager) validateMetadata(req SaveUploadRequest, metadata map[string]string) error {
	const op = "uploadManager.validateMetadata"

	if metadata["Upload-Id"] != req.UploadID.String() {
		return apperror.Wrap(op, apperror.ErrInvalidUploadID)
	}

	if metadata["Entity-Id"] != req.Entity.ID.String() {
		return apperror.Wrap(op, apperror.ErrInvalidEntityID)
	}

	if metadata["Entity-Type"] != string(req.Entity.Type) {
		return apperror.Wrap(op, apperror.ErrInvalidEntityType)
	}

	return nil
}
