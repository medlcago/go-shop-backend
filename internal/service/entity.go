package service

import (
	"context"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"

	"github.com/google/uuid"
)

type entityService struct {
	productRepo repository.ProductRepository
}

func NewEntityService(productRepo repository.ProductRepository) *entityService {
	return &entityService{
		productRepo: productRepo,
	}
}

func (e *entityService) Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) error {
	const op = "entityService.Exists"

	var (
		exists bool
		err    error
	)

	switch entityType {
	case dto.EntityProduct:
		exists, err = e.productRepo.Exists(ctx, id)
	default:
		return apperrors.ErrUnknownEntityType
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		return apperrors.ErrEntityNotFound
	}

	return nil
}
