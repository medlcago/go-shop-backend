package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

func Wrap(op string, err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := errors.AsType[*AppError](err); ok {
		return &AppError{
			Code:    appErr.Code,
			Message: appErr.Message,
			Err:     fmt.Errorf("%s: %w", op, appErr),
			Details: appErr.Details,
		}
	}

	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Err:     fmt.Errorf("%s: %w", op, err),
	}
}
