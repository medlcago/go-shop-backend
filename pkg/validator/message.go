package validator

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

func HumanizeValidationError(err error) map[string]string {
	result := make(map[string]string)

	ve, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return result
	}

	for _, e := range ve {
		namespace := e.StructNamespace()
		fieldPath := formatNamespace(namespace)
		result[fieldPath] = messageByTag(e)
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

	case "gt":
		return "Must be > " + e.Param()

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

func formatNamespace(ns string) string {
	parts := strings.Split(ns, ".")
	if len(parts) <= 1 {
		return ""
	}

	parts = parts[1:]

	for i, part := range parts {
		parts[i] = strcase.ToSnake(part)
	}

	return strings.Join(parts, ".")
}
