package apperror

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = New(http.StatusNotFound, "user not found")
	ErrUserProfileDeleted = New(http.StatusForbidden, "profile deleted")
	ErrEmailTaken         = New(http.StatusConflict, "email already in use")

	ErrInvalidCredentials = New(http.StatusUnauthorized, "invalid credentials")
	ErrInvalidSessionID   = New(http.StatusUnauthorized, "invalid or missing session id")
	ErrNotFound           = New(http.StatusNotFound, "not found")
	ErrForbidden          = New(http.StatusForbidden, "forbidden")
	ErrGatewayTimeout     = New(http.StatusGatewayTimeout, "gateway timeout")

	ErrInvalidFileType     = New(http.StatusBadRequest, "invalid file type")
	ErrContentTypeMismatch = New(http.StatusBadRequest, "ext is not allowed for that content type")
	ErrInvalidUploadID     = New(http.StatusBadRequest, "invalid upload id")
	ErrInvalidEntityID     = New(http.StatusBadRequest, "invalid entity id")
	ErrInvalidEntityType   = New(http.StatusBadRequest, "invalid entity type")
	ErrFileAlreadyUploaded = New(http.StatusConflict, "file already uploaded")
	ErrFileTooLarge        = New(http.StatusRequestEntityTooLarge, "file too large")

	ErrProductNotFound  = New(http.StatusNotFound, "product not found")
	ErrProductNotActive = New(http.StatusBadRequest, "product is not active")
	ErrItemNotFound     = New(http.StatusNotFound, "item not found in this order")

	ErrInvalidQuantity    = New(http.StatusBadRequest, "quantity must be greater than zero")
	ErrInsufficientStock  = New(http.StatusBadRequest, "insufficient product stock")
	ErrInconsistentStock  = New(http.StatusBadRequest, "inconsistent product stock")
	ErrInvalidOrderStatus = New(http.StatusConflict, "order status does not allow modifications")
	ErrEmptyOrder         = New(http.StatusBadRequest, "order is empty")
	ErrOrderNotFound      = New(http.StatusNotFound, "order not found")

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

type UnavailableItem struct {
	ProductID    uuid.UUID
	RequestedQty int
	AvailableQty int
	Action       string
	Reason       string
}

type ItemsUnavailableError struct {
	Items []UnavailableItem
}

func (e *ItemsUnavailableError) Error() string {
	return fmt.Sprintf("%d items unavailable", len(e.Items))
}
