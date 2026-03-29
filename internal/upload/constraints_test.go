package upload_test

import (
	"go-shop-backend/internal/upload"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileConstraints_IsValidExt(t *testing.T) {
	constraints := upload.FileConstraints{
		AllowedFormats: []upload.Format{
			{
				ContentType: "image/jpeg",
				Extensions:  []string{"jpg", "jpeg"},
			},
			{
				ContentType: "image/png",
				Extensions:  []string{"png"},
			},
		},
	}

	tests := []struct {
		name string
		ext  string
		ct   string
		want bool
	}{
		{
			name: "valid jpeg jpg",
			ext:  "jpg",
			ct:   "image/jpeg",
			want: true,
		},
		{
			name: "valid jpeg jpeg",
			ext:  "jpeg",
			ct:   "image/jpeg",
			want: true,
		},
		{
			name: "invalid ext for jpeg",
			ext:  "png",
			ct:   "image/jpeg",
			want: false,
		},
		{
			name: "valid ext but wrong content type",
			ext:  "jpg",
			ct:   "image/png",
			want: false,
		},
		{
			name: "unknown content type",
			ext:  "jpg",
			ct:   "application/pdf",
			want: false,
		},
		{
			name: "empty extension",
			ext:  "",
			ct:   "image/jpeg",
			want: false,
		},
		{
			name: "empty content type",
			ext:  "jpg",
			ct:   "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constraints.IsValidExt(tt.ext, tt.ct)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileConstraints_IsValidType(t *testing.T) {
	constraints := upload.FileConstraints{
		AllowedFormats: []upload.Format{
			{
				ContentType: "image/jpeg",
				Extensions:  []string{"jpg", "jpeg"},
			},
			{
				ContentType: "image/png",
				Extensions:  []string{"png"},
			},
		},
	}

	tests := []struct {
		name string
		ct   string
		want bool
	}{
		{
			name: "valid jpeg",
			ct:   "image/jpeg",
			want: true,
		},
		{
			name: "valid png",
			ct:   "image/png",
			want: true,
		},
		{
			name: "invalid type",
			ct:   "application/pdf",
			want: false,
		},
		{
			name: "empty type",
			ct:   "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constraints.IsValidType(tt.ct)
			assert.Equal(t, tt.want, got)
		})
	}
}
