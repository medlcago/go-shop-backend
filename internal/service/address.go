package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/mapper"

	"github.com/google/uuid"
)

type addressService struct {
	addressRepo repository.AddressRepository
}

func NewAddressService(addressRepo repository.AddressRepository) *addressService {
	return &addressService{
		addressRepo: addressRepo,
	}
}

func (a *addressService) CreateAddress(ctx context.Context, userID uuid.UUID, req dto.CreateAddressRequest) (*dto.AddressResponse, error) {
	const op = "addressService.CreateAddress"

	address, err := mapper.MapOne[dto.CreateAddressRequest, models.Address](req)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}
	address.UserID = userID

	if err := a.addressRepo.Create(ctx, address); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := a.mapAddress(address)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (a *addressService) ListAddresses(ctx context.Context, userID uuid.UUID) ([]*dto.AddressResponse, error) {
	const op = "addressService.ListAddresses"

	address, err := a.addressRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := a.mapAddresses(address)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (a *addressService) GetAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*dto.AddressResponse, error) {
	const op = "addressService.GetAddress"

	address, err := a.getAddressByID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := a.mapAddress(address)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (a *addressService) UpdateAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID, req dto.UpdateAddressRequest) (*dto.AddressResponse, error) {
	const op = "addressService.UpdateAddress"

	address, err := a.getAddressByID(ctx, id, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := mapper.Copy(address, req, false); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := a.addressRepo.Update(ctx, address); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := a.mapAddress(address)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (a *addressService) DeleteAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const op = "addressService.DeleteAddress"

	err := a.addressRepo.Delete(ctx, id, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return apperror.Wrap(op, apperror.ErrAddressNotFound)
		}

		return apperror.Wrap(op, err)
	}

	return nil
}

func (a *addressService) SetDefault(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const op = "addressService.SetDefault"

	err := a.addressRepo.SetDefault(ctx, id, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return apperror.Wrap(op, apperror.ErrAddressNotFound)
		}

		return apperror.Wrap(op, err)
	}

	return nil
}

func (a *addressService) getAddressByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Address, error) {
	const op = "addressService.getAddressByID"

	address, err := a.addressRepo.GetByID(ctx, id, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrAddressNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return address, nil
}

func (a *addressService) mapAddress(address *models.Address) (*dto.AddressResponse, error) {
	const op = "addressService.mapAddress"

	response, err := mapper.MapOne[*models.Address, dto.AddressResponse](address)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (a *addressService) mapAddresses(addresses []*models.Address) ([]*dto.AddressResponse, error) {
	const op = "addressService.mapAddress"

	response, err := mapper.MapList[*models.Address, *dto.AddressResponse](addresses)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}
