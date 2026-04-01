package key

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/zeropass/zeropass/core/crypto"
	aesgcm "github.com/zeropass/zeropass/core/crypto/cipher"
)

const (
	// VaultKeySize is the size of a vault key in bytes.
	VaultKeySize = 32
)

// EncryptedVaultKey holds an encrypted vault key.
type EncryptedVaultKey struct {
	Ciphertext []byte
}

// GenerateVaultKey generates a new random vault key.
func GenerateVaultKey() ([]byte, error) {
	key := make([]byte, VaultKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate vault key: %w", err)
	}
	return key, nil
}

// EncryptVaultKey encrypts a vault key with a master key using AES-256-GCM.
func EncryptVaultKey(vaultKey []byte, masterKey []byte) (*EncryptedVaultKey, error) {
	if len(vaultKey) != VaultKeySize {
		return nil, errors.New("vault key must be 32 bytes")
	}
	if len(masterKey) != aesgcm.KeySize {
		return nil, errors.New("master key must be 32 bytes")
	}

	c := aesgcm.New()
	ciphertext, err := c.Encrypt(vaultKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt vault key: %w", err)
	}

	return &EncryptedVaultKey{Ciphertext: ciphertext}, nil
}

// DecryptVaultKey decrypts an encrypted vault key using the master key.
func DecryptVaultKey(encrypted *EncryptedVaultKey, masterKey []byte) ([]byte, error) {
	if encrypted == nil || len(encrypted.Ciphertext) == 0 {
		return nil, errors.New("encrypted vault key must not be empty")
	}
	if len(masterKey) != aesgcm.KeySize {
		return nil, errors.New("master key must be 32 bytes")
	}

	c := aesgcm.New()
	vaultKey, err := c.Decrypt(encrypted.Ciphertext, masterKey)
	if err != nil {
		return nil, errors.New("failed to decrypt vault key: authentication error")
	}

	if len(vaultKey) != VaultKeySize {
		crypto.ZeroBytes(vaultKey)
		return nil, errors.New("decrypted vault key has invalid length")
	}

	return vaultKey, nil
}

// RotateVaultKey re-encrypts a vault key with a new master key.
// The old master key is used to decrypt, and the new master key encrypts.
func RotateVaultKey(encrypted *EncryptedVaultKey, oldMasterKey []byte, newMasterKey []byte) (*EncryptedVaultKey, error) {
	// Decrypt with old key
	vaultKey, err := DecryptVaultKey(encrypted, oldMasterKey)
	if err != nil {
		return nil, fmt.Errorf("rotation decrypt: %w", err)
	}
	defer crypto.ZeroBytes(vaultKey)

	// Re-encrypt with new key
	newEncrypted, err := EncryptVaultKey(vaultKey, newMasterKey)
	if err != nil {
		return nil, fmt.Errorf("rotation encrypt: %w", err)
	}

	return newEncrypted, nil
}
