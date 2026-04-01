package key

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	aesgcm "github.com/zeropass/zeropass/core/crypto/cipher"
)

func generateKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestGenerateVaultKey(t *testing.T) {
	t.Run("generates 32-byte key", func(t *testing.T) {
		vk, err := GenerateVaultKey()
		require.NoError(t, err)
		assert.Len(t, vk, VaultKeySize)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		vk1, err := GenerateVaultKey()
		require.NoError(t, err)
		vk2, err := GenerateVaultKey()
		require.NoError(t, err)
		assert.NotEqual(t, vk1, vk2)
	})
}

func TestEncryptDecryptVaultKey(t *testing.T) {
	masterKey := generateKey(t)
	vaultKey, err := GenerateVaultKey()
	require.NoError(t, err)

	t.Run("round trip", func(t *testing.T) {
		encrypted, err := EncryptVaultKey(vaultKey, masterKey)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted.Ciphertext)

		decrypted, err := DecryptVaultKey(encrypted, masterKey)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted)
	})

	t.Run("wrong master key fails", func(t *testing.T) {
		encrypted, err := EncryptVaultKey(vaultKey, masterKey)
		require.NoError(t, err)

		wrongKey := generateKey(t)
		_, err = DecryptVaultKey(encrypted, wrongKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication error")
	})

	t.Run("invalid vault key length", func(t *testing.T) {
		_, err := EncryptVaultKey([]byte("short"), masterKey)
		assert.Error(t, err)
	})

	t.Run("invalid master key length for encrypt", func(t *testing.T) {
		_, err := EncryptVaultKey(vaultKey, []byte("short"))
		assert.Error(t, err)
	})

	t.Run("invalid master key length for decrypt", func(t *testing.T) {
		encrypted, err := EncryptVaultKey(vaultKey, masterKey)
		require.NoError(t, err)
		_, err = DecryptVaultKey(encrypted, []byte("short"))
		assert.Error(t, err)
	})

	t.Run("nil encrypted vault key", func(t *testing.T) {
		_, err := DecryptVaultKey(nil, masterKey)
		assert.Error(t, err)
	})

	t.Run("empty ciphertext", func(t *testing.T) {
		_, err := DecryptVaultKey(&EncryptedVaultKey{Ciphertext: []byte{}}, masterKey)
		assert.Error(t, err)
	})
}

func TestRotateVaultKey(t *testing.T) {
	oldMasterKey := generateKey(t)
	newMasterKey := generateKey(t)

	vaultKey, err := GenerateVaultKey()
	require.NoError(t, err)

	encrypted, err := EncryptVaultKey(vaultKey, oldMasterKey)
	require.NoError(t, err)

	t.Run("rotation preserves vault key", func(t *testing.T) {
		rotated, err := RotateVaultKey(encrypted, oldMasterKey, newMasterKey)
		require.NoError(t, err)

		// Decrypt with new key should give same vault key
		decrypted, err := DecryptVaultKey(rotated, newMasterKey)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted)

		// Old key should not decrypt rotated
		_, err = DecryptVaultKey(rotated, oldMasterKey)
		assert.Error(t, err)
	})

	t.Run("rotation with wrong old key fails", func(t *testing.T) {
		wrongKey := generateKey(t)
		_, err := RotateVaultKey(encrypted, wrongKey, newMasterKey)
		assert.Error(t, err)
	})

	t.Run("rotation with invalid new master key fails", func(t *testing.T) {
		_, err := RotateVaultKey(encrypted, oldMasterKey, []byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rotation encrypt")
	})
}

func TestDecryptVaultKeyInvalidDecryptedLength(t *testing.T) {
	masterKey := generateKey(t)

	// Encrypt a 16-byte value directly with AES-GCM (not 32 bytes),
	// then try to decrypt it as a vault key — should fail length check.
	c := aesgcm.New()
	shortData := make([]byte, 16) // not VaultKeySize (32)
	ciphertext, err := c.Encrypt(shortData, masterKey)
	require.NoError(t, err)

	encrypted := &EncryptedVaultKey{Ciphertext: ciphertext}
	_, err = DecryptVaultKey(encrypted, masterKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
}

func TestDecryptVaultKeyLongDecryptedData(t *testing.T) {
	masterKey := generateKey(t)

	// Encrypt a 48-byte value directly (longer than VaultKeySize)
	c := aesgcm.New()
	longData := make([]byte, 48)
	ciphertext, err := c.Encrypt(longData, masterKey)
	require.NoError(t, err)

	encrypted := &EncryptedVaultKey{Ciphertext: ciphertext}
	_, err = DecryptVaultKey(encrypted, masterKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
}
