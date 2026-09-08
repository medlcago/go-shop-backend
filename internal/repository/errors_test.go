package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestHandleError(t *testing.T) {
	t.Parallel()

	sentinelErr := errors.New("some error")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil",
			err:  nil,
			want: nil,
		},
		{
			name: "record not found",
			err:  gorm.ErrRecordNotFound,
			want: ErrRecordNotFound,
		},
		{
			name: "wrapped record not found",
			err:  fmt.Errorf("query failed: %w", gorm.ErrRecordNotFound),
			want: ErrRecordNotFound,
		},
		{
			name: "other error",
			err:  sentinelErr,
			want: sentinelErr,
		},
		{
			name: "wrapped other error",
			err:  fmt.Errorf("query failed: %w", sentinelErr),
			want: sentinelErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := HandleError(tt.err)

			if tt.want == nil {
				assert.NoError(t, got)
				return
			}

			assert.ErrorIs(t, got, tt.want)
		})
	}
}
