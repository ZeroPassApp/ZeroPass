// Package cipher provides symmetric encryption for ZeroPass.
package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	// KeySize is the required AES-256 key size in bytes.
	KeySize = 32
	// NonceSize is the GCM nonce size in bytes.
	NonceSize = 12
)

// AESGCM implements the crypto.Cipher interface using AES-256-GCM.
type AESGCM struct{}

// New creates a new AES-256-GCM cipher.
func New() *AESGCM {
	return &AESGCM{}
}

// Encrypt encrypts plaintext using AES-256-GCM with the provided key.
// A unique 12-byte nonce is generated per operation and prepended to the ciphertext.
// Output format: [nonce (12 bytes)][ciphertext + GCM tag]
func (a *AESGCM) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends the ciphertext to nonce, so the result is [nonce][ciphertext+tag].
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// It extracts the nonce from the first 12 bytes, then decrypts the remainder.
func (a *AESGCM) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, errors.New("decryption failed: authentication error")
	}

	return plaintext, nil
}
