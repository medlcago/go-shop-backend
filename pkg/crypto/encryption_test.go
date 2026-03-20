package crypto

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAESGCMEncryptionManager(t *testing.T) {
	tests := []struct {
		name      string
		key       []byte
		expectErr bool
	}{
		{
			name:      "valid key",
			key:       make([]byte, 32),
			expectErr: false,
		},
		{
			name:      "invalid key length",
			key:       []byte("short"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewAESGCMEncryptionManager(tt.key)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, m)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, m)
			}
		})
	}
}

func TestNewAESGCMEncryptionManagerFromBase64(t *testing.T) {
	validKey := make([]byte, 32)
	validKeyBase64 := base64.StdEncoding.EncodeToString(validKey)

	tests := []struct {
		name      string
		keyBase64 string
		expectErr bool
	}{
		{
			name:      "valid base64 key",
			keyBase64: validKeyBase64,
			expectErr: false,
		},
		{
			name:      "invalid base64",
			keyBase64: "!!!invalid!!!",
			expectErr: true,
		},
		{
			name:      "decoded but wrong length",
			keyBase64: base64.StdEncoding.EncodeToString([]byte("short")),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewAESGCMEncryptionManagerFromBase64(tt.keyBase64)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, m)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, m)
			}
		})
	}
}

func TestAESGCMEncryptionManager_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	m, err := NewAESGCMEncryptionManager(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello"},
		{"empty string", ""},
		{"long text", "some very long text with symbols !@#$%^&*() 👋"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			encrypted, err := m.Encrypt(ctx, tt.plaintext)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			decrypted, err := m.Decrypt(ctx, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestAESGCMEncryptionManager_Context(t *testing.T) {
	key := make([]byte, 32)
	m, err := NewAESGCMEncryptionManager(key)
	require.NoError(t, err)

	tests := []struct {
		name        string
		fn          func(ctx context.Context) error
		expectError bool
	}{
		{
			name: "encrypt cancelled context",
			fn: func(ctx context.Context) error {
				_, err := m.Encrypt(ctx, "data")
				return err
			},
			expectError: true,
		},
		{
			name: "decrypt cancelled context",
			fn: func(ctx context.Context) error {
				_, err := m.Decrypt(ctx, "data")
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.fn(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAESGCMEncryptionManager_DecryptErrors(t *testing.T) {
	key := make([]byte, 32)
	m, err := NewAESGCMEncryptionManager(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "invalid base64",
			input:     "!!!not-base64!!!",
			expectErr: true,
		},
		{
			name:      "too short ciphertext",
			input:     base64.StdEncoding.EncodeToString([]byte("short")),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, err := m.Decrypt(ctx, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAESGCMEncryptionManager_NonceRandomness(t *testing.T) {
	key := make([]byte, 32)
	m, err := NewAESGCMEncryptionManager(key)
	require.NoError(t, err)

	ctx := context.Background()
	plaintext := "text"

	enc1, err := m.Encrypt(ctx, plaintext)
	assert.NoError(t, err)

	enc2, err := m.Encrypt(ctx, plaintext)
	assert.NoError(t, err)

	assert.NotEqual(t, enc1, enc2, "ciphertexts should differ due to random nonce")
}
