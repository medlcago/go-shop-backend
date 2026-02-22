package category

import (
	"fmt"
	"go-shop-backend/internal/dto"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/testutils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCategoryHandler_ListCategories(t *testing.T) {
	tests := []struct {
		name         string
		categoryID   string
		setupMock    func(serviceMock *serviceMocks.MockCategoryService)
		query        string
		expectedCode int
		expectedBody any
	}{
		{
			name:       "success",
			categoryID: "",
			setupMock: func(serviceMock *serviceMocks.MockCategoryService) {
				serviceMock.EXPECT().ListCategories(mock.Anything, dto.ListCategoryRequest{}).
					Return([]*dto.ProductCategoryResponse{}, 3, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductCategoryResponse{}, 3)),
		},
		{
			name:       "success get subcategory",
			categoryID: "e267ab94-8c00-4dce-b44e-93dc546f631a",
			setupMock: func(serviceMock *serviceMocks.MockCategoryService) {
				serviceMock.EXPECT().ListCategories(mock.Anything, dto.ListCategoryRequest{
					ID: uuid.MustParse("e267ab94-8c00-4dce-b44e-93dc546f631a"),
				}).Return([]*dto.ProductCategoryResponse{}, 3, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(response.NewPaginated([]*dto.ProductCategoryResponse{}, 3)),
		},
		{
			name:         "invalid query",
			categoryID:   "",
			setupMock:    nil,
			query:        "limit=test",
			expectedCode: http.StatusBadRequest,
			expectedBody: response.NewError(http.StatusText(http.StatusBadRequest)),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *serviceMocks.MockCategoryService) {
				serviceMock.EXPECT().ListCategories(mock.Anything, dto.ListCategoryRequest{}).
					Return([]*dto.ProductCategoryResponse{}, 0, fmt.Errorf("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := serviceMocks.NewMockCategoryService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			categoryHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Get("/:id<guid>?", categoryHandler.ListCategories)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s?%s", tt.categoryID, tt.query), nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}
