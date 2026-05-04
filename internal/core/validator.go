package core

import (
	"go-shop-backend/pkg/validator"

	govalidator "github.com/go-playground/validator/v10"
)

func NewValidator() validator.Validator {
	v := govalidator.New()
	return validator.New(v)
}
