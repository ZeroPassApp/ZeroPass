package key

import (
	"errors"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"github.com/zeropass/zeropass/core/crypto"
	aesgcm "github.com/zeropass/zeropass/core/crypto/cipher"
)

const (
	// RecoveryKeySize is the size of the derived recovery key in bytes.
	RecoveryKeySize = 32
	// mnemonicBitSize is the entropy bits for a 12-word mnemonic.
	mnemonicBitSize = 128
)

// RecoveryKeyManager implements the crypto.RecoveryManager interface
// using BIP-39 mnemonics.
type RecoveryKeyManager struct{}

// NewRecoveryKeyManager creates a new RecoveryKeyManager.
func NewRecoveryKeyManager() *RecoveryKeyManager {
	return &RecoveryKeyManager{}
}

// GenerateMnemonic generates a new 12-word BIP-39 mnemonic phrase.
func (r *RecoveryKeyManager) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(mnemonicBitSize)
	if err != nil {
		return "", fmt.Errorf("generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// DeriveKeyFromMnemonic derives a 32-byte key from a BIP-39 mnemonic.
// Uses the BIP-39 seed derivation (PBKDF2) with an empty passphrase,
// then takes the first 32 bytes.
func (r *RecoveryKeyManager) DeriveKeyFromMnemonic(mnemonic string) ([]byte, error) {
	if !r.ValidateMnemonic(mnemonic) {
		return nil, errors.New("invalid mnemonic phrase")
	}

	// BIP-39 NewSeed uses PBKDF2-HMAC-SHA512 with 2048 iterations
	// and returns a 64-byte seed. We take the first 32 bytes for our key.
	seed := bip39.NewSeed(mnemonic, "")
	key := make([]byte, RecoveryKeySize)
	copy(key, seed[:RecoveryKeySize])

	// Zero the full seed after extracting our key
	crypto.ZeroBytes(seed)

	return key, nil
}

// ValidateMnemonic checks whether a mnemonic phrase is valid BIP-39.
func (r *RecoveryKeyManager) ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// EncryptVaultKeyWithRecovery encrypts a vault key using a recovery key derived from a mnemonic.
// This provides a backup decryption path for the vault key.
func EncryptVaultKeyWithRecovery(vaultKey []byte, mnemonic string) (*EncryptedVaultKey, error) {
	mgr := NewRecoveryKeyManager()
	recoveryKey, err := mgr.DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("derive recovery key: %w", err)
	}
	defer crypto.ZeroBytes(recoveryKey)

	return EncryptVaultKey(vaultKey, recoveryKey)
}

// RegenerateRecoveryKey generates a new BIP-39 mnemonic and re-encrypts the vault key with it.
// Returns the new mnemonic and the new encrypted vault key.
// The caller is responsible for persisting the new encrypted recovery key to vault metadata.
func RegenerateRecoveryKey(vaultKey []byte) (string, *EncryptedVaultKey, error) {
	if len(vaultKey) == 0 {
		return "", nil, errors.New("vault key must not be empty")
	}

	mgr := NewRecoveryKeyManager()
	mnemonic, err := mgr.GenerateMnemonic()
	if err != nil {
		return "", nil, fmt.Errorf("generate new mnemonic: %w", err)
	}

	encVK, err := EncryptVaultKeyWithRecovery(vaultKey, mnemonic)
	if err != nil {
		return "", nil, fmt.Errorf("encrypt vault key with new recovery: %w", err)
	}

	return mnemonic, encVK, nil
}

// DecryptVaultKeyWithRecovery decrypts a vault key using a mnemonic recovery phrase.
func DecryptVaultKeyWithRecovery(encrypted *EncryptedVaultKey, mnemonic string) ([]byte, error) {
	mgr := NewRecoveryKeyManager()
	recoveryKey, err := mgr.DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("derive recovery key: %w", err)
	}
	defer crypto.ZeroBytes(recoveryKey)

	c := aesgcm.New()
	vaultKey, err := c.Decrypt(encrypted.Ciphertext, recoveryKey)
	if err != nil {
		return nil, errors.New("failed to decrypt vault key with recovery phrase")
	}

	return vaultKey, nil
}
