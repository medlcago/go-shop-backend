package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsTwoFAEnabled(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "both TwoFAEnabled true and TwoFASecret not nil should return true",
			user: &User{
				TwoFAEnabled: true,
				TwoFASecret:  new("secret123"),
			},
			expected: true,
		},
		{
			name: "TwoFAEnabled false but TwoFASecret not nil should return false",
			user: &User{
				TwoFAEnabled: false,
				TwoFASecret:  new("secret123"),
			},
			expected: false,
		},
		{
			name: "TwoFAEnabled true but TwoFASecret nil should return false",
			user: &User{
				TwoFAEnabled: true,
				TwoFASecret:  nil,
			},
			expected: false,
		},
		{
			name: "both TwoFAEnabled false and TwoFASecret nil should return false",
			user: &User{
				TwoFAEnabled: false,
				TwoFASecret:  nil,
			},
			expected: false,
		},
		{
			name: "TwoFAEnabled true and TwoFASecret empty string should return false",
			user: &User{
				TwoFAEnabled: true,
				TwoFASecret:  new(""),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsTwoFAEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}
