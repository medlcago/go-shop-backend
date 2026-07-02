package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AddressServiceTestSuite struct {
	suite.Suite
	addressRepo    *repoMocks.MockAddressRepository
	addressService *addressService

	ctx       context.Context
	addressID uuid.UUID
	userID    uuid.UUID
}

func (suite *AddressServiceTestSuite) SetupTest() {
	suite.addressRepo = repoMocks.NewMockAddressRepository(suite.T())
	suite.addressService = NewAddressService(suite.addressRepo)

	suite.ctx = context.Background()
	suite.addressID = uuid.New()
	suite.userID = uuid.New()
}

func TestAddressServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AddressServiceTestSuite))
}

// ==================== CreateAddress Tests ====================

func (suite *AddressServiceTestSuite) TestCreateAddress_Success() {
	req := dto.CreateAddressRequest{
		Name: "Address 1",
	}

	suite.addressRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(address *models.Address) bool {
		return address.Name == req.Name && address.UserID == suite.userID
	})).Return(nil).Once()

	response, err := suite.addressService.CreateAddress(suite.ctx, suite.userID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(req.Name, response.Name)
}

func (suite *AddressServiceTestSuite) TestCreateAddress_RepositoryError() {
	req := dto.CreateAddressRequest{
		Name: "Address 1",
	}

	repoErr := errors.New("repoErr")

	suite.addressRepo.EXPECT().Create(suite.ctx, mock.AnythingOfType("*models.Address")).
		Return(repoErr).Once()

	response, err := suite.addressService.CreateAddress(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "addressService.CreateAddress")
}

// ==================== ListAddresses Tests ====================

func (suite *AddressServiceTestSuite) TestListAddresses_Success() {
	suite.addressRepo.EXPECT().ListByUser(suite.ctx, suite.userID).
		Return([]*models.Address{{ID: uuid.New()}, {ID: uuid.New()}}, nil).Once()

	response, err := suite.addressService.ListAddresses(suite.ctx, suite.userID)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Len(response, 2)
}

func (suite *AddressServiceTestSuite) TestListAddresses_RepositoryError() {
	repoErr := errors.New("repoErr")

	suite.addressRepo.EXPECT().ListByUser(suite.ctx, suite.userID).
		Return(nil, repoErr).Once()

	response, err := suite.addressService.ListAddresses(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "addressService.ListAddresses")
}

// ==================== GetAddress Tests ====================

func (suite *AddressServiceTestSuite) TestGetAddress_Success() {
	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(&models.Address{ID: suite.addressID, UserID: suite.userID}, nil).Once()

	response, err := suite.addressService.GetAddress(suite.ctx, suite.addressID, suite.userID)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(response.ID, suite.addressID)
}

func (suite *AddressServiceTestSuite) TestGetAddress_AddressNotFound() {
	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.addressService.GetAddress(suite.ctx, suite.addressID, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrAddressNotFound)
	suite.ErrorContains(err, "addressService.GetAddress")
}

func (suite *AddressServiceTestSuite) TestGetAddress_RepositoryError() {
	repoErr := errors.New("repoErr")

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(nil, repoErr).Once()

	response, err := suite.addressService.GetAddress(suite.ctx, suite.addressID, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "addressService.GetAddress")
}

// ==================== UpdateAddress Tests ====================

func (suite *AddressServiceTestSuite) TestUpdateAddress_Success() {
	req := dto.UpdateAddressRequest{
		CreateAddressRequest: dto.CreateAddressRequest{
			Name:    "Address 123",
			Country: "Russia",
			Comment: new("test"),
		},
	}

	address := &models.Address{
		ID:        suite.addressID,
		UserID:    suite.userID,
		Name:      "Test123",
		IsDefault: true,
		Apartment: new("14"),
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.addressRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(address *models.Address) bool {
		return address.Name == req.Name &&
			address.Country == req.Country &&
			*address.Comment == *req.Comment &&
			address.ID == suite.addressID &&
			address.UserID == suite.userID &&
			address.IsDefault &&
			address.Apartment == nil
	})).Return(nil).Once()

	response, err := suite.addressService.UpdateAddress(suite.ctx, suite.addressID, suite.userID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(address.ID, response.ID)
}

// ==================== DeleteAddress Tests ====================

func (suite *AddressServiceTestSuite) TestDeleteAddress_Success() {
	suite.addressRepo.EXPECT().Delete(suite.ctx, suite.addressID, suite.userID).
		Return(nil).Once()

	err := suite.addressService.DeleteAddress(suite.ctx, suite.addressID, suite.userID)
	suite.NoError(err)
}

func (suite *AddressServiceTestSuite) TestDeleteAddress_AddressNotFound() {
	suite.addressRepo.EXPECT().Delete(suite.ctx, suite.addressID, suite.userID).
		Return(repository.ErrRecordNotFound).Once()

	err := suite.addressService.DeleteAddress(suite.ctx, suite.addressID, suite.userID)

	suite.ErrorIs(err, apperror.ErrAddressNotFound)
	suite.ErrorContains(err, "addressService.DeleteAddress")
}

func (suite *AddressServiceTestSuite) TestDeleteAddress_RepositoryError() {
	repoErr := errors.New("repoErr")

	suite.addressRepo.EXPECT().Delete(suite.ctx, suite.addressID, suite.userID).
		Return(repoErr).Once()

	err := suite.addressService.DeleteAddress(suite.ctx, suite.addressID, suite.userID)

	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "addressService.DeleteAddress")
}

// ==================== SetDefault Tests ====================

func (suite *AddressServiceTestSuite) TestSetDefault_Success() {
	suite.addressRepo.EXPECT().SetDefault(suite.ctx, suite.addressID, suite.userID).
		Return(nil).Once()

	err := suite.addressService.SetDefault(suite.ctx, suite.addressID, suite.userID)
	suite.NoError(err)
}

func (suite *AddressServiceTestSuite) TestSetDefault_AddressNotFound() {
	suite.addressRepo.EXPECT().SetDefault(suite.ctx, suite.addressID, suite.userID).
		Return(repository.ErrRecordNotFound).Once()

	err := suite.addressService.SetDefault(suite.ctx, suite.addressID, suite.userID)

	suite.ErrorIs(err, apperror.ErrAddressNotFound)
	suite.ErrorContains(err, "addressService.SetDefault")
}

func (suite *AddressServiceTestSuite) TestSetDefault_RepositoryError() {
	repoErr := errors.New("repoErr")

	suite.addressRepo.EXPECT().SetDefault(suite.ctx, suite.addressID, suite.userID).
		Return(repoErr).Once()

	err := suite.addressService.SetDefault(suite.ctx, suite.addressID, suite.userID)

	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "addressService.SetDefault")
}
