package testutils

import (
	"encoding/json"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/utils"
	structValidator "go-shop-backend/pkg/validator"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func CreateTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler:    middleware.ErrorHandler(slog.New(slog.DiscardHandler)),
		StructValidator: structValidator.New(validator.New()),
	})
}

func AssertJSONResponse(t *testing.T, expectedCode int, expectedJSON string, resp *http.Response) {
	t.Helper()
	assert.Equal(t, expectedCode, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.JSONEq(t, expectedJSON, string(body))
}

func StringJSON(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func validateStruct(v any) error {
	return structValidator.New(validator.New()).Validate(v)
}

func ValidationError(v any) *response.Response[struct{}] {
	return response.NewError(
		"Validation failed",
		utils.HumanizeValidationError(validateStruct(v)),
	)
}

func NewReadSeekCloser() io.ReadSeekCloser {
	reader := strings.NewReader("test")
	obj := struct {
		io.ReadSeeker
		io.Closer
	}{
		reader,
		io.NopCloser(reader),
	}

	return obj
}
