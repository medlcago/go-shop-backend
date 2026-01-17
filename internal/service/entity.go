package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"

	"github.com/google/uuid"
)

type entityService struct {
	productRepo repository.ProductRepository
}

func NewEntityService(productRepo repository.ProductRepository) EntityService {
	return &entityService{
		productRepo: productRepo,
	}
}

func (e *entityService) Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) (bool, error) {
	switch entityType {
	case dto.EntityProduct:
		return e.productRepo.Exists(ctx, id)
	default:
		return false, apperrors.ErrUnknownEntityType
	}
}
