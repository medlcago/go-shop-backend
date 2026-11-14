package upload_test

import (
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/apperror"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPolicy(t *testing.T) {
	const mb = int64(1 << 20)

	tests := []struct {
		name         string
		minSize      int64
		maxSize      int64
		maxSizeLimit int64
		mimes        []string
		wantErr      bool
		errContains  string
		wantMaxSize  int64
	}{
		{
			name:        "valid policy",
			minSize:     1,
			maxSize:     5 * mb,
			wantMaxSize: 5 * mb,
			mimes:       []string{"image/jpeg", "image/png"},
		},
		{
			name:        "zero min size",
			minSize:     0,
			maxSize:     1024,
			wantErr:     true,
			errContains: "min size must be > 0",
			mimes:       []string{"image/jpeg"},
		},
		{
			name:        "negative min size",
			minSize:     -1,
			maxSize:     1024,
			wantErr:     true,
			errContains: "min size must be > 0",
			mimes:       []string{"image/jpeg"},
		},
		{
			name:        "zero max size (no limit)",
			minSize:     1024,
			maxSize:     0,
			wantErr:     true,
			errContains: "max size must be > 0",
			mimes:       []string{"image/jpeg"},
		},
		{
			name:        "negative max size (no limit)",
			minSize:     1024,
			maxSize:     -1,
			wantErr:     true,
			errContains: "max size must be > 0",
			mimes:       []string{"image/jpeg"},
		},
		{
			name:        "min size greater than max size",
			minSize:     2048,
			maxSize:     1024,
			wantErr:     true,
			errContains: "min size 2048 > max size 1024",
			mimes:       []string{"image/jpeg"},
		},
		{
			name:        "empty allowed mimes",
			minSize:     1024,
			maxSize:     2048,
			wantErr:     true,
			errContains: "allowed mimes must not be empty",
			mimes:       []string{},
		},
		{
			name:        "nil allowed mimes",
			minSize:     1024,
			maxSize:     2048,
			wantErr:     true,
			errContains: "allowed mimes must not be empty",
			mimes:       nil,
		},
		{
			name:        "min size equals max size",
			minSize:     1024,
			maxSize:     1024,
			wantMaxSize: 1024,
			mimes:       []string{"image/png"},
		},

		{
			name:         "maxSizeLimit caps maxSize",
			minSize:      1024,
			maxSize:      10 * mb,
			maxSizeLimit: 1 * mb,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},
		{
			name:         "maxSizeLimit greater than maxSize is ignored",
			minSize:      1024,
			maxSize:      1 * mb,
			maxSizeLimit: 10 * mb,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},
		{
			name:         "maxSizeLimit equals maxSize",
			minSize:      1024,
			maxSize:      1 * mb,
			maxSizeLimit: 1 * mb,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},
		{
			name:         "maxSizeLimit applies when maxSize is zero",
			minSize:      1024,
			maxSize:      0,
			maxSizeLimit: 1 * mb,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},
		{
			name:         "maxSizeLimit applies when maxSize is negative",
			minSize:      1024,
			maxSize:      -1,
			maxSizeLimit: 1 * mb,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},

		{
			name:         "negative maxSizeLimit is ignored",
			minSize:      1024,
			maxSize:      1 * mb,
			maxSizeLimit: -1,
			wantMaxSize:  1 * mb,
			mimes:        []string{"image/png"},
		},
		{
			name:         "maxSizeLimit pushes effective max below minSize",
			minSize:      1024,
			maxSize:      10 * mb,
			maxSizeLimit: 512,
			wantErr:      true,
			errContains:  "min size 1024 > max size 512",
			mimes:        []string{"image/png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := upload.NewPolicy(tt.minSize, tt.maxSize, tt.maxSizeLimit, tt.mimes)

			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errContains)
				require.Empty(t, p)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.minSize, p.MinSize)
			require.Equal(t, tt.wantMaxSize, p.MaxSize)
			require.Equal(t, tt.mimes, p.AllowedMimes)
		})
	}
}

func TestPolicy_Validate(t *testing.T) {
	policy, err := upload.NewPolicy(
		100,
		1024,
		0,
		[]string{"image/jpeg", "image/png"},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		file    upload.File
		wantErr error
	}{
		{
			name: "valid jpeg file",
			file: upload.File{
				Size:        500,
				ContentType: "image/jpeg",
			},
			wantErr: nil,
		},
		{
			name: "valid png file",
			file: upload.File{
				Size:        500,
				ContentType: "image/png",
			},
			wantErr: nil,
		},
		{
			name: "file size equals min size",
			file: upload.File{
				Size:        100,
				ContentType: "image/jpeg",
			},
			wantErr: nil,
		},
		{
			name: "file size equals max size",
			file: upload.File{
				Size:        1024,
				ContentType: "image/jpeg",
			},
			wantErr: nil,
		},
		{
			name: "file size less than min size",
			file: upload.File{
				Size:        99,
				ContentType: "image/jpeg",
			},
			wantErr: apperror.ErrInvalidFileSize,
		},
		{
			name: "file size greater than max size",
			file: upload.File{
				Size:        1025,
				ContentType: "image/jpeg",
			},
			wantErr: apperror.ErrInvalidFileSize,
		},
		{
			name: "file size is zero",
			file: upload.File{
				Size:        0,
				ContentType: "image/jpeg",
			},
			wantErr: apperror.ErrInvalidFileSize,
		},
		{
			name: "invalid mime type",
			file: upload.File{
				Size:        500,
				ContentType: "application/pdf",
			},
			wantErr: apperror.ErrInvalidFileType,
		},
		{
			name: "empty mime type",
			file: upload.File{
				Size:        500,
				ContentType: "",
			},
			wantErr: apperror.ErrInvalidFileType,
		},
		{
			name: "invalid size and mime type",
			file: upload.File{
				Size:        50,
				ContentType: "application/pdf",
			},
			wantErr: apperror.ErrInvalidFileSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.file)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
