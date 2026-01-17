package apperrors

import (
	"fmt"
	"net/http"
)

var (
	ErrEmailTaken          = New(http.StatusConflict, "email already in use")
	ErrInvalidCredentials  = New(http.StatusUnauthorized, "invalid credentials")
	ErrUserNotFound        = New(http.StatusNotFound, "user not found")
	ErrProductNotFound     = New(http.StatusNotFound, "product not found")
	ErrForbidden           = New(http.StatusForbidden, "forbidden")
	ErrUserProfileDeleted  = New(http.StatusForbidden, "profile deleted")
	ErrUnknownEntityType   = New(http.StatusBadRequest, "unknown entity type")
	ErrEntityNotFound      = New(http.StatusBadRequest, "entity not found")
	ErrInvalidImageFormat  = New(http.StatusBadRequest, "invalid image format: only PNG, JPG are allowed")
	ErrContentTypeMismatch = New(http.StatusBadRequest, "ext is not allowed for that content type")
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
