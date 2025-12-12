package middleware

import (
	"encoding/json"
	"errors"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/testutils"
	structValidator "go-shop-backend/pkg/validator"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		exceptedBody   *response.Response[struct{}]
	}{
		{
			name:           "Fiber Error - Not Found",
			err:            fiber.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			exceptedBody:   response.NewResponse(struct{}{}, http.StatusText(http.StatusNotFound)),
		},
		{
			name:           "Fiber Error - Bad Request",
			err:            fiber.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
			exceptedBody:   response.NewResponse(struct{}{}, http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "Fiber Error with custom message",
			err:            fiber.NewError(http.StatusForbidden, "access denied"),
			expectedStatus: http.StatusForbidden,
			exceptedBody:   response.NewResponse(struct{}{}, "access denied"),
		},
		{
			name:           "AppError",
			err:            apperrors.New(http.StatusConflict, "resource already exists"),
			expectedStatus: http.StatusConflict,
			exceptedBody:   response.NewResponse(struct{}{}, "resource already exists"),
		},
		{
			name: "Validator ValidationErrors",
			err: func() error {
				type TestStruct struct {
					Name string `validate:"required"`
				}

				v := structValidator.New(validator.New())
				err := v.Validate(TestStruct{Name: ""})
				return err
			}(),
			expectedStatus: http.StatusBadRequest,
			exceptedBody:   response.NewResponse(struct{}{}, "Key: 'TestStruct.Name' Error:Field validation for 'Name' failed on the 'required' tag"),
		},
		{
			name:           "JSON UnmarshalTypeError",
			err:            &json.UnmarshalTypeError{Field: "age", Type: reflect.TypeOf(struct{}{}), Offset: 10},
			expectedStatus: http.StatusBadRequest,
			exceptedBody:   response.NewResponse(struct{}{}, http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "JSON SyntaxError",
			err:            &json.SyntaxError{Offset: 5},
			expectedStatus: http.StatusBadRequest,
			exceptedBody:   response.NewResponse(struct{}{}, http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "Generic error",
			err:            errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			exceptedBody:   response.NewResponse(struct{}{}, http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.CreateTestApp(fiber.Config{
				ErrorHandler: ErrorHandler(testutils.DiscardSlog),
			})
			testCtx, cleanup := testutils.PrepareTestContext(app, "/", []byte("data"))
			defer cleanup()

			assert.NoError(t, app.Config().ErrorHandler(testCtx, tt.err))

			exceptedBody, _ := json.Marshal(tt.exceptedBody)

			assert.Equal(t, tt.expectedStatus, testCtx.Response().StatusCode())
			assert.Equal(t, exceptedBody, testCtx.Response().Body())
		})
	}
}
