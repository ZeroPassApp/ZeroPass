package key

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	// ItemKeySize is the size of a derived item key in bytes.
	ItemKeySize = 32
	// hkdfInfo is the context info for item key derivation.
	hkdfInfo = "zeropass-item-key"
)

// DeriveItemKey derives a unique per-item encryption key from the vault key and item ID.
// Uses HKDF-SHA256 with the item ID as salt.
func DeriveItemKey(vaultKey []byte, itemID string) ([]byte, error) {
	if len(vaultKey) == 0 {
		return nil, errors.New("vault key must not be empty")
	}
	if itemID == "" {
		return nil, errors.New("item ID must not be empty")
	}

	hkdfReader := hkdf.New(sha256.New, vaultKey, []byte(itemID), []byte(hkdfInfo))

	itemKey := make([]byte, ItemKeySize)
	if _, err := io.ReadFull(hkdfReader, itemKey); err != nil {
		return nil, fmt.Errorf("derive item key: %w", err)
	}

	return itemKey, nil
}
