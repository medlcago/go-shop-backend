package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	txManager   database.TxManager
}

func NewCartService(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	txManager database.TxManager,
) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		txManager:   txManager,
	}
}

func (c *cartService) GetCart(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID) (*dto.CartResponse, error) {
	const op = "cartService.GetCart"

	cart, err := c.getOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return c.mapToDTO(cart), nil
}

func (c *cartService) getOrCreateCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *uuid.UUID,
) (*models.Cart, error) {

	if userID == nil && sessionID == nil {
		return nil, apperrors.ErrInvalidCart
	}

	var (
		cart *models.Cart
		err  error
	)

	if userID != nil {
		sessionID = nil
		cart, err = c.cartRepo.GetByUserID(ctx, *userID)
	} else {
		userID = nil
		cart, err = c.cartRepo.GetBySessionID(ctx, *sessionID)
	}

	if err == nil {
		return cart, nil
	}

	cart = &models.Cart{
		UserID:    userID,
		SessionID: sessionID,
	}

	if err := c.cartRepo.Create(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (c *cartService) AddItem(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, req dto.AddItemRequest) (*dto.CartResponse, error) {
	const op = "cartService.AddItem"

	if req.Quantity <= 0 {
		return nil, apperrors.ErrInvalidQuantity
	}

	var (
		cart *models.Cart
		err  error
	)

	addItem := func(ctx context.Context) error {
		cart, err = c.getOrCreateCart(ctx, userID, sessionID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		product, err := c.productRepo.GetByID(ctx, req.ProductID, false)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return apperrors.ErrProductNotFound
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		if !product.IsActive {
			return apperrors.ErrProductNotActive
		}

		if product.Stock < req.Quantity {
			return apperrors.ErrInsufficientStock
		}

		itemFound := false
		for i, item := range cart.Items {
			if item.ProductID == req.ProductID {
				cart.Items[i].Quantity = req.Quantity
				itemFound = true
				break
			}
		}

		if !itemFound {
			cart.Items = append(cart.Items, models.CartItem{
				CartID:    cart.ID,
				ProductID: req.ProductID,
				Quantity:  req.Quantity,
				UnitPrice: product.Price,
			})
		}

		if err := c.cartRepo.Save(ctx, cart); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	}

	err = c.txManager.Wrap(ctx, addItem)

	if err != nil {
		return nil, err
	}

	return c.mapToDTO(cart), nil
}

func (c *cartService) DeleteItem(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, productID uuid.UUID) (*dto.CartResponse, error) {
	const op = "cartService.DeleteItem"

	cart, err := c.getOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.cartRepo.DeleteItem(ctx, cart.ID, productID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	cart, err = c.getOrCreateCart(ctx, userID, sessionID)

	return c.mapToDTO(cart), nil
}

func (c *cartService) mapToDTO(cart *models.Cart) *dto.CartResponse {
	response := &dto.CartResponse{
		ID:    cart.ID,
		Items: make([]dto.ItemResponse, 0, len(cart.Items)),
	}

	var total float64
	for _, item := range cart.Items {
		total += item.UnitPrice * float64(item.Quantity)
		response.Items = append(response.Items, dto.ItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	response.TotalCost = total
	return response
}
