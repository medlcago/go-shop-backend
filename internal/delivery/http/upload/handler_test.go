package upload

import (
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service/mocks"
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

func TestUploadHandler_SignURL(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.UploadServiceMock)
		requestBody  any
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.UploadServiceMock) {
				serviceMock.On("SignURL", mock.Anything, dto.SignURLRequest{
					ContentType: "image/jpeg",
					Entity: dto.Entity{
						ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
						Type: dto.EntityProduct,
					},
					Ext: "jpg",
				}).Return(&dto.SignURLResponse{}, nil).Once()
			},
			requestBody: dto.SignURLRequest{
				ContentType: "image/jpeg",
				Entity: dto.Entity{
					ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					Type: dto.EntityProduct,
				},
				Ext: "jpg",
			},
			expectedCode: http.StatusCreated,
			expectedBody: response.NewResponse(&dto.SignURLResponse{}),
		},
		{
			name: "invalid ext",
			requestBody: dto.SignURLRequest{
				ContentType: "image/jpeg",
				Entity: dto.Entity{
					ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					Type: dto.EntityProduct,
				},
				Ext: "mp4",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: testutils.ValidationError(
				dto.SignURLRequest{
					ContentType: "image/jpeg",
					Entity: dto.Entity{
						ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
						Type: dto.EntityProduct,
					},
					Ext: "mp4",
				},
			),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.UploadServiceMock) {
				serviceMock.On("SignURL", mock.Anything, mock.Anything).
					Return(&dto.SignURLResponse{}, errors.New("internal server error")).Once()
			},
			requestBody: dto.SignURLRequest{
				ContentType: "image/jpeg",
				Entity: dto.Entity{
					ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					Type: dto.EntityProduct,
				},
				Ext: "jpg",
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(fiber.ErrInternalServerError.Message),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.UploadServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			uploadHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", uploadHandler.SignURL)

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

func TestUploadHandler_Save(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.UploadServiceMock)
		requestBody  any
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.UploadServiceMock) {
				serviceMock.On("Save", mock.Anything, dto.UploadRequest{
					UploadID:  uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					ObjectKey: "key",
					Entity: dto.Entity{
						ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
						Type: dto.EntityProduct,
					},
				}).Return(&dto.UploadResponse{}, nil).Once()
			},
			requestBody: dto.UploadRequest{
				UploadID:  uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
				ObjectKey: "key",
				Entity: dto.Entity{
					ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					Type: dto.EntityProduct,
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(&dto.UploadResponse{}),
		},
		{
			name:         "validation error",
			requestBody:  dto.UploadRequest{ObjectKey: "key"},
			expectedCode: http.StatusBadRequest,
			expectedBody: testutils.ValidationError(dto.UploadRequest{ObjectKey: "key"}),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.UploadServiceMock) {
				serviceMock.On("Save", mock.Anything, mock.Anything).
					Return(&dto.UploadResponse{}, errors.New("internal server error")).Once()
			},
			requestBody: dto.UploadRequest{
				UploadID:  uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
				ObjectKey: "key",
				Entity: dto.Entity{
					ID:   uuid.MustParse("e0261a9b-7cb0-4e62-8d6f-72f1fe12fad6"),
					Type: dto.EntityProduct,
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(fiber.ErrInternalServerError.Message),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.UploadServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			uploadHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", uploadHandler.Save)

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
