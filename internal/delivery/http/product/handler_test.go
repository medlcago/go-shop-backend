package product

import (
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductHandler_GetProductByID(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.ProductServiceMock)
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("GetProductByID", mock.Anything, uuid.MustParse("757c7dff-6d2f-44dc-9a22-ce16dabcaa2d")).
					Return(&dto.ProductResponse{}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(&dto.ProductResponse{}),
		},
		{
			name: "not found",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("GetProductByID", mock.Anything, uuid.MustParse("757c7dff-6d2f-44dc-9a22-ce16dabcaa2d")).
					Return(&dto.ProductResponse{}, apperrors.ErrProductNotFound).Once()
			},
			expectedCode: http.StatusNotFound,
			expectedBody: response.NewError(apperrors.ErrProductNotFound.Message),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("GetProductByID", mock.Anything, uuid.MustParse("757c7dff-6d2f-44dc-9a22-ce16dabcaa2d")).
					Return(&dto.ProductResponse{}, errors.New("unexpected error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.ProductServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			productHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Get("/:id<guid>", productHandler.GetProductByID)

			req := httptest.NewRequest(http.MethodGet, "/757c7dff-6d2f-44dc-9a22-ce16dabcaa2d", nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)

		})
	}
}

func TestProductHandler_ListProducts(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.ProductServiceMock)
		query        string
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{}).
					Return([]*dto.ProductResponse{}, 2, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 2)),
		},
		{
			name: "empty list",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{}).
					Return([]*dto.ProductResponse{}, 0, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 0)),
		},
		{
			name: "default parameters",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{
					Limit:     0,
					Offset:    0,
					OrderBy:   "",
					OrderDesc: false,
				}).Return([]*dto.ProductResponse{}, 3, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 3)),
		},
		{
			name: "with pagination parameters",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{
					Limit:     20,
					Offset:    10,
					OrderBy:   "created_at",
					OrderDesc: true,
				}).Return([]*dto.ProductResponse{}, 5, nil).Once()
			},
			query:        "limit=20&offset=10&order_by=created_at&order_desc=true",
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 5)),
		},
		{
			name: "with category filter",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{
					CategoryID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				}).Return([]*dto.ProductResponse{}, 3, nil).Once()
			},
			query:        "category_id=123e4567-e89b-12d3-a456-426614174000",
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 3)),
		},
		{
			name:         "invalid category uuid",
			setupMock:    nil,
			query:        "category_id=not-a-uuid",
			expectedCode: http.StatusBadRequest,
			expectedBody: response.NewError(fiber.ErrBadRequest.Message),
		},
		{
			name:         "invalid query",
			setupMock:    nil,
			query:        "limit=test",
			expectedCode: http.StatusBadRequest,
			expectedBody: response.NewError(fiber.ErrBadRequest.Message),
		},
		{
			name:  "negative limit and offset",
			query: "limit=-10&offset=-10",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{
					Limit:  -10,
					Offset: -10,
				}).Return([]*dto.ProductResponse{}, 2, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductResponse{}, 2)),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("ListProducts", mock.Anything, dto.ListProductRequest{}).
					Return([]*dto.ProductResponse{}, 0, errors.New("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.ProductServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			productHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Get("/", productHandler.ListProducts)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", tt.query), nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.ProductServiceMock)
		requestBody  any
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("CreateProduct", mock.Anything, dto.ProductCreateRequest{
					Name:  "test product",
					Price: 100,
					Stock: 10,
				}).
					Return(&dto.ProductResponse{Name: "test product"}, nil).Once()
			},
			requestBody: dto.ProductCreateRequest{
				Name:  "test product",
				Price: 100,
				Stock: 10,
			},
			expectedCode: http.StatusCreated,
			expectedBody: response.NewResponse(&dto.ProductResponse{Name: "test product"}),
		},
		{
			name:         "validation error",
			setupMock:    nil,
			requestBody:  dto.ProductCreateRequest{},
			expectedCode: http.StatusBadRequest,
			expectedBody: response.NewError(fiber.ErrBadRequest.Message),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.ProductServiceMock) {
				serviceMock.On("CreateProduct", mock.Anything, dto.ProductCreateRequest{
					Name:  "test product",
					Price: 100,
					Stock: 10,
				}).
					Return(&dto.ProductResponse{}, errors.New("unexpected error")).Once()
			},
			requestBody: dto.ProductCreateRequest{
				Name:  "test product",
				Price: 100,
				Stock: 10,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.ProductServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			productHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", productHandler.CreateProduct)

			body := testutils.StringJSON(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}
