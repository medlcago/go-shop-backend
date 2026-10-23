package handlers

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/auth"

	"github.com/google/uuid"
)

type ProductChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type UploadCounter interface {
	Count(ctx context.Context, entityType string, entityID uuid.UUID) (int64, error)
}

type ProductImageHandler struct {
	productChecker ProductChecker
	uploadCounter  UploadCounter
	maxImages      int64
}

func NewProductImageHandler(
	productChecker ProductChecker,
	uploadCounter UploadCounter,
	maxImages int64,
) *ProductImageHandler {
	return &ProductImageHandler{
		productChecker: productChecker,
		uploadCounter:  uploadCounter,
		maxImages:      maxImages,
	}
}

func (h *ProductImageHandler) CanSignURL(ctx context.Context, entityType string, entityID uuid.UUID) error {
	const op = "ProductImageHandler.CanSignURL"

	if err := h.validate(ctx, entityType, entityID); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (h *ProductImageHandler) CanAttach(ctx context.Context, entityType string, entityID uuid.UUID) error {
	const op = "ProductImageHandler.CanAttach"

	if err := h.validate(ctx, entityType, entityID); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (h *ProductImageHandler) validate(ctx context.Context, entityType string, entityID uuid.UUID) error {
	if entityID == uuid.Nil || entityType != string(models.EntityTypeProduct) {
		return apperror.ErrInvalidEntityType
	}

	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return apperror.ErrForbidden
	}

	if user.Role != string(models.UserRoleAdmin) && user.Role != string(models.UserRoleSeller) {
		return apperror.ErrForbidden
	}

	exists, err := h.productChecker.Exists(ctx, entityID)
	if err != nil {
		return err
	}
	if !exists {
		return apperror.ErrProductNotFound
	}

	count, err := h.uploadCounter.Count(ctx, entityType, entityID)
	if err != nil {
		return err
	}

	if count >= h.maxImages {
		return apperror.ErrProductImageLimitExceeded
	}

	return nil
}
