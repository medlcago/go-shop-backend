package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = New(http.StatusNotFound, "user not found")
	ErrUserProfileDeleted = New(http.StatusForbidden, "account has been deleted, contact support")
	ErrEmailTaken         = New(http.StatusConflict, "email already in use")

	ErrInvalidCredentials = New(http.StatusUnauthorized, "invalid credentials")
	ErrInvalidSessionID   = New(http.StatusUnauthorized, "invalid or missing session id")
	ErrNotFound           = New(http.StatusNotFound, "not found")
	ErrForbidden          = New(http.StatusForbidden, "forbidden")
	ErrInvalidTokenType   = New(http.StatusUnauthorized, "invalid token type")

	ErrInvalidFileType     = New(http.StatusBadRequest, "invalid file type")
	ErrContentTypeMismatch = New(http.StatusBadRequest, "ext is not allowed for that content type")
	ErrInvalidUploadID     = New(http.StatusBadRequest, "invalid upload id")
	ErrInvalidEntityID     = New(http.StatusBadRequest, "invalid entity id")
	ErrInvalidEntityType   = New(http.StatusBadRequest, "invalid entity type")
	ErrFileAlreadyUploaded = New(http.StatusConflict, "file already uploaded")
	ErrFileTooLarge        = New(http.StatusRequestEntityTooLarge, "file too large")

	ErrProductNotFound       = New(http.StatusNotFound, "product not found")
	ErrProductNotActive      = New(http.StatusBadRequest, "product is not active")
	ErrItemNotFound          = New(http.StatusNotFound, "item not found in this order")
	ErrStockLessThanReserved = New(http.StatusBadRequest, "stock cannot be less than reserved quantity")

	ErrInvalidQuantity    = New(http.StatusBadRequest, "quantity must be greater than zero")
	ErrInsufficientStock  = New(http.StatusBadRequest, "insufficient product stock")
	ErrInconsistentStock  = New(http.StatusBadRequest, "inconsistent product stock")
	ErrInvalidOrderStatus = New(http.StatusConflict, "order status does not allow this action")
	ErrEmptyOrder         = New(http.StatusBadRequest, "order is empty")
	ErrOrderNotFound      = New(http.StatusNotFound, "order not found")
	ErrOrderExpired       = New(http.StatusForbidden, "order is expired")

	ErrPaymentAlreadyCreated = New(http.StatusConflict, "payment already created for this order")

	Err2FAAlreadyEnabled = New(http.StatusConflict, "2FA is already enabled; disable it first to reconfigure")
	Err2FANotEnabled     = New(http.StatusBadRequest, "2FA is not enabled")
	Err2FANotInitialized = New(http.StatusBadRequest, "2FA is not initialized")
	ErrInvalid2FACode    = New(http.StatusBadRequest, "invalid 2FA code")
	Err2FACodeRequired   = New(http.StatusUnauthorized, "2FA code is required")

	ErrWishlistNotFound         = New(http.StatusNotFound, "wishlist not found or is private")
	ErrWishlistItemNotFound     = New(http.StatusNotFound, "wishlist item not found")
	ErrProductAlreadyInWishlist = New(http.StatusConflict, "product already in wishlist")

	ErrEmailConfirmationCodeAlreadySent = New(http.StatusTooManyRequests, "email confirmation code already sent, please wait before requesting a new one")
	ErrInvalidCode                      = New(http.StatusBadRequest, "invalid code")
	ErrEmailAlreadyConfirmed            = New(http.StatusConflict, "email already confirmed")
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Details any    `json:"details"`
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s (code=%d): %v", e.Message, e.Code, e.Err)
	}

	return fmt.Sprintf("%s (code=%d)", e.Message, e.Code)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) HttpStatusCode() int {
	return e.Code
}

type UnavailableItem struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	RequestedQty int       `json:"requested_qty"`
	AvailableQty int       `json:"available_qty"`
	Action       string    `json:"action"`
	Reason       string    `json:"reason"`
}

var (
	errUnavailableItems = errors.New("unavailable items")
)

func UnavailableItemsError(unavailableItems []UnavailableItem) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: errUnavailableItems.Error(),
		Err:     errUnavailableItems,
		Details: map[string]any{
			"unavailable_items": unavailableItems,
		},
	}
}

func GetUnavailableItemsFromError(err error) ([]UnavailableItem, bool) {
	if appErr, ok := errors.AsType[*AppError](err); ok && errors.Is(appErr, errUnavailableItems) {
		details, ok := appErr.Details.(map[string]any)
		if !ok {
			return nil, false
		}

		itemsRaw, exists := details["unavailable_items"]
		if !exists {
			return nil, false
		}

		items, ok := itemsRaw.([]UnavailableItem)
		return items, ok
	}

	return nil, false
}
