package testutils

import (
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/utils"
	structValidator "go-shop-backend/pkg/validator"

	"github.com/go-playground/validator/v10"
)

func validateStruct(v any) error {
	return structValidator.New(validator.New()).Validate(v)
}

func ValidationError(v any) *response.Response[struct{}] {
	return response.NewError(
		"Validation failed",
		utils.HumanizeValidationError(validateStruct(v)),
	)
}
