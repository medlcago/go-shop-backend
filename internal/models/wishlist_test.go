package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWishlist_CanView(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name     string
		wishlist Wishlist
		userID   uuid.UUID
		expected bool
	}{
		{
			name: "public wishlist - any user can view",
			wishlist: Wishlist{
				UserID:   ownerID,
				IsPublic: true,
			},
			userID:   otherUserID,
			expected: true,
		},
		{
			name: "private wishlist - owner can view",
			wishlist: Wishlist{
				UserID:   ownerID,
				IsPublic: false,
			},
			userID:   ownerID,
			expected: true,
		},
		{
			name: "private wishlist - other user cannot view",
			wishlist: Wishlist{
				UserID:   ownerID,
				IsPublic: false,
			},
			userID:   otherUserID,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.wishlist.CanView(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWishlist_CanEdit(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name     string
		wishlist Wishlist
		userID   uuid.UUID
		expected bool
	}{
		{
			name: "owner can edit",
			wishlist: Wishlist{
				UserID: ownerID,
			},
			userID:   ownerID,
			expected: true,
		},
		{
			name: "non-owner cannot edit",
			wishlist: Wishlist{
				UserID: ownerID,
			},
			userID:   otherUserID,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.wishlist.CanEdit(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
