package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type inventoryService struct {
	productRepo repository.ProductRepository
	txManager   database.TxManager
}

func NewInventoryService(
	productRepo repository.ProductRepository,
	txManager database.TxManager,
) *inventoryService {
	return &inventoryService{
		productRepo: productRepo,
		txManager:   txManager,
	}
}

func (i *inventoryService) CheckProduct(ctx context.Context, productID uuid.UUID, quantity int) (*models.Product, error) {
	const op = "inventoryService.CheckProduct"

	product, err := i.productRepo.GetByID(ctx, productID, false)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrProductNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	if !product.IsActive {
		return nil, apperror.Wrap(op, apperror.ErrProductNotActive)
	}

	if product.Available() < quantity {
		return nil, apperror.Wrap(op, apperror.ErrInsufficientStock)
	}

	return product, nil
}

func (i *inventoryService) ReserveItems(ctx context.Context, items []dto.InventoryItem) error {
	const op = "inventoryService.ReserveItems"

	err := i.txManager.Wrap(ctx, func(ctx context.Context) error {
		return i.applyOnProducts(ctx, items, "reserve",
			func(p *models.Product, qty int) error {
				return p.Reserve(qty)
			},
		)
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (i *inventoryService) ReleaseItems(ctx context.Context, items []dto.InventoryItem) error {
	const op = "inventoryService.ReleaseItems"

	err := i.txManager.Wrap(ctx, func(ctx context.Context) error {
		return i.applyOnProducts(ctx, items, "release",
			func(p *models.Product, qty int) error {
				return p.Release(qty)
			},
		)
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (i *inventoryService) DeductItems(ctx context.Context, items []dto.InventoryItem) error {
	const op = "inventoryService.DeductItems"

	err := i.txManager.Wrap(ctx, func(ctx context.Context) error {
		return i.applyOnProducts(ctx, items, "deduct",
			func(p *models.Product, qty int) error {
				return p.Deduct(qty)
			},
		)
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (i *inventoryService) applyOnProducts(
	ctx context.Context,
	items []dto.InventoryItem,
	actionName string,
	action func(p *models.Product, qty int) error,
) error {
	const op = "orderService.applyOnProducts"

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	products, err := i.productRepo.GetByIDsForUpdate(ctx, productIDs)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	productMap := make(map[uuid.UUID]*models.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var unavailableItems []apperror.UnavailableItem

	for _, item := range items {
		product := productMap[item.ProductID]
		if product == nil {
			unavailableItems = append(unavailableItems, apperror.UnavailableItem{
				ID:           item.ItemID,
				ProductID:    item.ProductID,
				RequestedQty: item.Quantity,
				Action:       actionName,
				Reason:       "PRODUCT_NOT_FOUND",
			})
			continue
		}

		if err := action(product, item.Quantity); err != nil {
			unavailableItems = append(unavailableItems, apperror.UnavailableItem{
				ID:           item.ItemID,
				ProductID:    product.ID,
				RequestedQty: item.Quantity,
				AvailableQty: product.Available(),
				Action:       actionName,
				Reason:       err.Error(),
			})
			continue
		}
	}

	if len(unavailableItems) > 0 {
		return apperror.Wrap(op, apperror.UnavailableItemsError(unavailableItems))
	}

	if err := i.productRepo.BulkUpsert(ctx, products); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}
