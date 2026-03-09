package models

import (
	"go-shop-backend/pkg/apperrors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduct_Available(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		reserved int
		want     int
	}{
		{
			name:     "positive stock no reserved",
			stock:    10,
			reserved: 0,
			want:     10,
		},
		{
			name:     "positive stock with reserved",
			stock:    10,
			reserved: 3,
			want:     7,
		},
		{
			name:     "zero stock",
			stock:    0,
			reserved: 0,
			want:     0,
		},
		{
			name:     "all stock reserved",
			stock:    5,
			reserved: 5,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{
				Stock:    tt.stock,
				Reserved: tt.reserved,
			}
			got := p.Available()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProduct_Reserve(t *testing.T) {
	tests := []struct {
		name          string
		setupProduct  func() *Product
		reserveQty    int
		wantErr       error
		wantReserved  int
		wantAvailable int
	}{
		{
			name: "successful reserve",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 0,
					IsActive: true,
				}
			},
			reserveQty:    3,
			wantErr:       nil,
			wantReserved:  3,
			wantAvailable: 7,
		},
		{
			name: "negative quantity",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 0,
					IsActive: true,
				}
			},
			reserveQty:    -1,
			wantErr:       errNegativeQty,
			wantReserved:  0,
			wantAvailable: 10,
		},
		{
			name: "product not active",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 0,
					IsActive: false,
				}
			},
			reserveQty:    3,
			wantErr:       apperrors.ErrProductNotActive,
			wantReserved:  0,
			wantAvailable: 10,
		},
		{
			name: "insufficient stock - reserve more than available",
			setupProduct: func() *Product {
				return &Product{
					Stock:    5,
					Reserved: 2,
					IsActive: true,
				}
			},
			reserveQty:    4,
			wantErr:       apperrors.ErrInsufficientStock,
			wantReserved:  2,
			wantAvailable: 3,
		},
		{
			name: "exact available stock",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 3,
					IsActive: true,
				}
			},
			reserveQty:    7,
			wantErr:       nil,
			wantReserved:  10,
			wantAvailable: 0,
		},
		{
			name: "zero quantity reserve",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 2,
					IsActive: true,
				}
			},
			reserveQty:    0,
			wantErr:       nil,
			wantReserved:  2,
			wantAvailable: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setupProduct()
			err := p.Reserve(tt.reserveQty)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantReserved, p.Reserved)
			assert.Equal(t, tt.wantAvailable, p.Available())
		})
	}
}

func TestProduct_Release(t *testing.T) {
	tests := []struct {
		name          string
		setupProduct  func() *Product
		releaseQty    int
		wantErr       error
		wantReserved  int
		wantAvailable int
	}{
		{
			name: "successful release",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			releaseQty:    3,
			wantErr:       nil,
			wantReserved:  2,
			wantAvailable: 8,
		},
		{
			name: "negative quantity",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			releaseQty:    -1,
			wantErr:       errNegativeQty,
			wantReserved:  5,
			wantAvailable: 5,
		},
		{
			name: "release more than reserved",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 3,
				}
			},
			releaseQty:    4,
			wantErr:       apperrors.ErrInconsistentStock,
			wantReserved:  3,
			wantAvailable: 7,
		},
		{
			name: "release all reserved",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			releaseQty:    5,
			wantErr:       nil,
			wantReserved:  0,
			wantAvailable: 10,
		},
		{
			name: "release zero quantity",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			releaseQty:    0,
			wantErr:       nil,
			wantReserved:  5,
			wantAvailable: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setupProduct()
			err := p.Release(tt.releaseQty)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantReserved, p.Reserved)
			assert.Equal(t, tt.wantAvailable, p.Available())
		})
	}
}

func TestProduct_Deduct(t *testing.T) {
	tests := []struct {
		name          string
		setupProduct  func() *Product
		deductQty     int
		wantErr       error
		wantStock     int
		wantReserved  int
		wantAvailable int
	}{
		{
			name: "successful deduction",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			deductQty:     3,
			wantErr:       nil,
			wantStock:     7,
			wantReserved:  2,
			wantAvailable: 5,
		},
		{
			name: "negative quantity",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			deductQty:     -1,
			wantErr:       errNegativeQty,
			wantStock:     10,
			wantReserved:  5,
			wantAvailable: 5,
		},
		{
			name: "deduct more than stock",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			deductQty:     11,
			wantErr:       apperrors.ErrInconsistentStock,
			wantStock:     10,
			wantReserved:  5,
			wantAvailable: 5,
		},
		{
			name: "deduct more than reserved",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 3,
				}
			},
			deductQty:     4,
			wantErr:       apperrors.ErrInconsistentStock,
			wantStock:     10,
			wantReserved:  3,
			wantAvailable: 7,
		},
		{
			name: "deduct exactly reserved amount",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 3,
				}
			},
			deductQty:     3,
			wantErr:       nil,
			wantStock:     7,
			wantReserved:  0,
			wantAvailable: 7,
		},
		{
			name: "deduct zero quantity",
			setupProduct: func() *Product {
				return &Product{
					Stock:    10,
					Reserved: 5,
				}
			},
			deductQty:     0,
			wantErr:       nil,
			wantStock:     10,
			wantReserved:  5,
			wantAvailable: 5,
		},
		{
			name: "deduct when reserved equals stock",
			setupProduct: func() *Product {
				return &Product{
					Stock:    5,
					Reserved: 5,
				}
			},
			deductQty:     3,
			wantErr:       nil,
			wantStock:     2,
			wantReserved:  2,
			wantAvailable: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setupProduct()
			err := p.Deduct(tt.deductQty)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantStock, p.Stock)
			assert.Equal(t, tt.wantReserved, p.Reserved)
			assert.Equal(t, tt.wantAvailable, p.Available())
		})
	}
}
