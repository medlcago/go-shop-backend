package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrInvalidKeyLength  = errors.New("invalid key length, must be 32 bytes for AES-256")
)

type EncryptionManager interface {
	Encrypt(ctx context.Context, plaintext string) (string, error)
	Decrypt(ctx context.Context, encrypted string) (string, error)
}

type AESGCMEncryptionManager struct {
	key []byte
}

func NewAESGCMEncryptionManager(key []byte) (*AESGCMEncryptionManager, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes, want 32", ErrInvalidKeyLength, len(key))
	}
	return &AESGCMEncryptionManager{key: key}, nil
}

func NewAESGCMEncryptionManagerFromBase64(keyBase64 string) (*AESGCMEncryptionManager, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %w", err)
	}
	return NewAESGCMEncryptionManager(key)
}

func (e *AESGCMEncryptionManager) Encrypt(ctx context.Context, plaintext string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	finalCiphertext := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(finalCiphertext), nil
}

func (e *AESGCMEncryptionManager) Decrypt(ctx context.Context, encrypted string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("%w: failed to decode base64: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: failed to decrypt: %v", ErrInvalidCiphertext, err)
	}

	return string(plaintext), nil
}
