package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap_Example(t *testing.T) {
	t.Parallel()
	baseErr := errors.New("sql: no rows in result set")

	err1 := Wrap("repository.GetWishlist", baseErr)
	err2 := Wrap("service.GetWishlist", err1)
	err3 := Wrap("handler.GetWishlist", err2)

	fmt.Println("Message:", err3.Message)
	fmt.Println("Error():", err3.Error())
	fmt.Println("Full error:", err3.Err)
	fmt.Println(errors.Is(err3, baseErr))

	if appErr, ok := errors.AsType[*AppError](err3); ok {
		fmt.Println("Code:", appErr.Code)
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()
	baseErr := errors.New("something went wrong")

	tests := []struct {
		name string
		op   string
		err  error

		wantNil     bool
		wantCode    int
		wantMessage string
		wantErr     string
		wantDetails any
	}{
		{
			name:    "nil error",
			op:      "create user",
			err:     nil,
			wantNil: true,
		},
		{
			name:        "regular error",
			op:          "create user",
			err:         baseErr,
			wantCode:    http.StatusInternalServerError,
			wantMessage: http.StatusText(http.StatusInternalServerError),
			wantErr:     "create user: something went wrong",
		},
		{
			name: "app error",
			op:   "create user",
			err: &AppError{
				Code:    http.StatusBadRequest,
				Message: "invalid user",
			},
			wantCode:    http.StatusBadRequest,
			wantMessage: "invalid user",
			wantErr:     "create user: invalid user (code=400)",
		},
		{
			name: "app error with details",
			op:   "validate user",
			err: &AppError{
				Code:    http.StatusUnprocessableEntity,
				Message: "validation failed",
				Details: map[string]any{
					"email": "invalid email",
					"name":  "required",
				},
			},
			wantCode:    http.StatusUnprocessableEntity,
			wantMessage: "validation failed",
			wantErr:     "validate user: validation failed (code=422)",
			wantDetails: map[string]any{
				"email": "invalid email",
				"name":  "required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Wrap(tt.op, tt.err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)

			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantMessage, got.Message)
			assert.Equal(t, tt.wantErr, got.Err.Error())

			if tt.wantDetails != nil {
				assert.Equal(t, tt.wantDetails, got.Details)
			}
		})
	}
}

func TestWrap_UnwrapsOriginalError(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("database unavailable")

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "regular error",
			err:  baseErr,
		},
		{
			name: "wrapped regular error",
			err:  fmt.Errorf("query failed: %w", baseErr),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Wrap("create user", tt.err)

			require.NotNil(t, got)
			assert.ErrorIs(t, got, baseErr)
		})
	}
}

func TestWrap_AppError(t *testing.T) {
	t.Parallel()

	details := map[string]any{
		"field": "email",
	}

	original := &AppError{
		Code:    http.StatusBadRequest,
		Message: "invalid request",
		Details: details,
	}

	got := Wrap("create user", original)

	require.NotNil(t, got)

	assert.NotSame(t, original, got)
	assert.Equal(t, original.Code, got.Code)
	assert.Equal(t, original.Message, got.Message)
	assert.Equal(t, original.Details, got.Details)
	assert.Equal(t, "create user: invalid request (code=400)", got.Err.Error())
}
