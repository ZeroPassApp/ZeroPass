// Package key provides key management for ZeroPass.
package key

import (
	"errors"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/crypto/kdf"
)

// MasterKey holds the derived master key and its associated salt.
type MasterKey struct {
	key  []byte
	salt []byte
}

// DeriveNewMasterKey derives a new master key from a password, generating a fresh salt.
func DeriveNewMasterKey(password []byte, k *kdf.Argon2id) (*MasterKey, error) {
	if len(password) == 0 {
		return nil, errors.New("password must not be empty")
	}

	salt, err := k.GenerateSalt()
	if err != nil {
		return nil, err
	}

	key, err := k.DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}

	return &MasterKey{key: key, salt: salt}, nil
}

// DeriveMasterKeyWithSalt derives a master key from a password and existing salt.
// Used when unlocking an existing vault.
func DeriveMasterKeyWithSalt(password []byte, salt []byte, k *kdf.Argon2id) (*MasterKey, error) {
	if len(password) == 0 {
		return nil, errors.New("password must not be empty")
	}
	if len(salt) == 0 {
		return nil, errors.New("salt must not be empty")
	}

	key, err := k.DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}

	return &MasterKey{key: key, salt: salt}, nil
}

// Key returns the raw master key bytes.
// Caller should zero this after use with crypto.ZeroBytes.
func (mk *MasterKey) Key() []byte {
	return mk.key
}

// Salt returns the salt used to derive this master key.
func (mk *MasterKey) Salt() []byte {
	return mk.salt
}

// Zero securely zeroes the master key material from memory.
func (mk *MasterKey) Zero() {
	crypto.ZeroBytes(mk.key)
	mk.key = nil
}
