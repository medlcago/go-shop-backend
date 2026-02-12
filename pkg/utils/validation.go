package utils

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func HumanizeValidationError(err error) map[string]string {
	result := make(map[string]string)

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return result
	}

	for _, e := range ve {
		field := ToSnakeCase(e.Field())
		result[field] = messageByTag(e)
	}

	return result
}

func messageByTag(e validator.FieldError) string {
	switch e.Tag() {

	case "required":
		return "This field is required"

	case "email":
		return "Invalid email format"

	case "min":
		return "Minimum length is " + e.Param()

	case "max":
		return "Maximum length is " + e.Param()

	case "gte":
		return "Must be ≥ " + e.Param()

	case "lte":
		return "Must be ≤ " + e.Param()

	case "oneof":
		return "Allowed values: " + join(e.Param())

	case "uuid":
		return "Invalid UUID"

	default:
		return "Invalid value"
	}
}

func join(param string) string {
	return strings.Join(strings.Fields(param), ", ")
}
