package apperrors

import (
	"net/http"
)

var (
	ErrEmailTaken          = New(http.StatusConflict, "email already in use")
	ErrInvalidCredentials  = New(http.StatusUnauthorized, "invalid credentials")
	ErrUserNotFound        = New(http.StatusNotFound, "user not found")
	ErrProductNotFound     = New(http.StatusNotFound, "product not found")
	ErrItemNotFound        = New(http.StatusNotFound, "item not found in this order")
	ErrForbidden           = New(http.StatusForbidden, "forbidden")
	ErrUserProfileDeleted  = New(http.StatusForbidden, "profile deleted")
	ErrUnknownEntityType   = New(http.StatusBadRequest, "unknown entity type")
	ErrEntityNotFound      = New(http.StatusBadRequest, "entity not found")
	ErrInvalidImageFormat  = New(http.StatusBadRequest, "invalid image format: only PNG, JPG are allowed")
	ErrContentTypeMismatch = New(http.StatusBadRequest, "ext is not allowed for that content type")
	ErrInvalidUploadID     = New(http.StatusBadRequest, "invalid upload id")
	ErrInvalidEntityID     = New(http.StatusBadRequest, "invalid entity id")
	ErrFileAlreadyUploaded = New(http.StatusConflict, "file already uploaded")
	ErrInvalidQuantity     = New(http.StatusBadRequest, "quantity must be greater than zero")
	ErrProductNotActive    = New(http.StatusBadRequest, "product is not active")
	ErrInsufficientStock   = New(http.StatusBadRequest, "insufficient product stock")
	ErrInconsistentStock   = New(http.StatusConflict, "inconsistent product stock")
	ErrInvalidOrderStatus  = New(http.StatusConflict, "order status does not allow modifications")
	ErrEmptyOrder          = New(http.StatusBadRequest, "order is empty")

	ErrInvalidPassword   = New(http.StatusBadRequest, "invalid password")
	ErrInvalidToken      = New(http.StatusUnauthorized, "invalid or expired token")
	Err2FAAlreadyEnabled = New(http.StatusConflict, "2FA is already enabled; disable it first to reconfigure")
	Err2FANotEnabled     = New(http.StatusBadRequest, "2FA is not enabled")
	Err2FANotInitialized = New(http.StatusBadRequest, "2FA is not initialized")
	ErrInvalid2FACode    = New(http.StatusBadRequest, "invalid 2FA code")
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
