package middleware

import (
	"encoding/json"
	"errors"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/response"
	structValidator "go-shop-backend/pkg/validator"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		expectedBody   *response.Response[struct{}]
		isBindingError bool
	}{
		{
			name:           "Fiber Error - Not Found",
			err:            fiber.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   response.NewError(http.StatusText(http.StatusNotFound)),
		},
		{
			name:           "Fiber Error - Bad Request",
			err:            fiber.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.NewError(http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "Fiber Error with custom message",
			err:            fiber.NewError(http.StatusForbidden, "access denied"),
			expectedStatus: http.StatusForbidden,
			expectedBody:   response.NewError("access denied"),
		},
		{
			name:           "AppError",
			err:            apperrors.New(http.StatusConflict, "resource already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody:   response.NewError("resource already exists"),
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
			expectedBody: response.NewError(
				"Validation failed",
				map[string]string{
					"name": "This field is required",
				},
			),
		},
		{
			name:           "JSON UnmarshalTypeError",
			err:            &json.UnmarshalTypeError{Field: "age", Type: reflect.TypeOf(struct{}{}), Offset: 10},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.NewError(http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "JSON SyntaxError",
			err:            &json.SyntaxError{Offset: 5},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.NewError(http.StatusText(http.StatusBadRequest)),
		},
		{
			name:           "Generic error",
			err:            errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
		{
			name:           "Binding Error",
			isBindingError: true,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.NewError(http.StatusText(http.StatusBadRequest)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t)

			app.Post("/", func(c fiber.Ctx) error {
				if tt.isBindingError {
					tt.err = c.Bind().Query(&struct {
						Test int `query:"test"`
					}{})
				}

				return c.App().ErrorHandler(c, tt.err)
			})

			req := httptest.NewRequest(http.MethodPost, "/?test=test123", nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedBody, _ := json.Marshal(tt.expectedBody)
			actualBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, string(expectedBody), string(actualBody))
		})
	}
}

func newTestApp(t *testing.T) *fiber.App {
	t.Helper()

	return fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(slog.New(slog.DiscardHandler)),
	})
}
