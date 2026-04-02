package key

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryKeyManager_GenerateMnemonic(t *testing.T) {
	mgr := NewRecoveryKeyManager()

	t.Run("generates 12-word mnemonic", func(t *testing.T) {
		mnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)
		words := strings.Fields(mnemonic)
		assert.Len(t, words, 12)
	})

	t.Run("generates unique mnemonics", func(t *testing.T) {
		m1, err := mgr.GenerateMnemonic()
		require.NoError(t, err)
		m2, err := mgr.GenerateMnemonic()
		require.NoError(t, err)
		assert.NotEqual(t, m1, m2)
	})

	t.Run("generated mnemonic is valid", func(t *testing.T) {
		mnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)
		assert.True(t, mgr.ValidateMnemonic(mnemonic))
	})
}

func TestRecoveryKeyManager_DeriveKeyFromMnemonic(t *testing.T) {
	mgr := NewRecoveryKeyManager()

	t.Run("derives 32-byte key", func(t *testing.T) {
		mnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)

		key, err := mgr.DeriveKeyFromMnemonic(mnemonic)
		require.NoError(t, err)
		assert.Len(t, key, RecoveryKeySize)
	})

	t.Run("same mnemonic produces same key", func(t *testing.T) {
		mnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)

		key1, err := mgr.DeriveKeyFromMnemonic(mnemonic)
		require.NoError(t, err)
		key2, err := mgr.DeriveKeyFromMnemonic(mnemonic)
		require.NoError(t, err)
		assert.Equal(t, key1, key2)
	})

	t.Run("different mnemonics produce different keys", func(t *testing.T) {
		m1, _ := mgr.GenerateMnemonic()
		m2, _ := mgr.GenerateMnemonic()

		key1, err := mgr.DeriveKeyFromMnemonic(m1)
		require.NoError(t, err)
		key2, err := mgr.DeriveKeyFromMnemonic(m2)
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("invalid mnemonic returns error", func(t *testing.T) {
		_, err := mgr.DeriveKeyFromMnemonic("not a valid mnemonic")
		assert.Error(t, err)
	})

	t.Run("empty mnemonic returns error", func(t *testing.T) {
		_, err := mgr.DeriveKeyFromMnemonic("")
		assert.Error(t, err)
	})
}

func TestRecoveryKeyManager_ValidateMnemonic(t *testing.T) {
	mgr := NewRecoveryKeyManager()

	t.Run("valid mnemonic", func(t *testing.T) {
		mnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)
		assert.True(t, mgr.ValidateMnemonic(mnemonic))
	})

	t.Run("invalid mnemonic", func(t *testing.T) {
		assert.False(t, mgr.ValidateMnemonic("invalid words here"))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.False(t, mgr.ValidateMnemonic(""))
	})

	// Known valid test vector (BIP-39)
	t.Run("known valid mnemonic", func(t *testing.T) {
		valid := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		assert.True(t, mgr.ValidateMnemonic(valid))
	})
}

func TestEncryptDecryptVaultKeyWithRecovery(t *testing.T) {
	mgr := NewRecoveryKeyManager()
	vaultKey, err := GenerateVaultKey()
	require.NoError(t, err)

	mnemonic, err := mgr.GenerateMnemonic()
	require.NoError(t, err)

	t.Run("round trip with recovery", func(t *testing.T) {
		encrypted, err := EncryptVaultKeyWithRecovery(vaultKey, mnemonic)
		require.NoError(t, err)

		decrypted, err := DecryptVaultKeyWithRecovery(encrypted, mnemonic)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted)
	})

	t.Run("wrong mnemonic fails", func(t *testing.T) {
		encrypted, err := EncryptVaultKeyWithRecovery(vaultKey, mnemonic)
		require.NoError(t, err)

		otherMnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)

		_, err = DecryptVaultKeyWithRecovery(encrypted, otherMnemonic)
		assert.Error(t, err)
	})

	t.Run("invalid mnemonic for encrypt", func(t *testing.T) {
		_, err := EncryptVaultKeyWithRecovery(vaultKey, "invalid mnemonic")
		assert.Error(t, err)
	})

	t.Run("invalid mnemonic for decrypt", func(t *testing.T) {
		encrypted, err := EncryptVaultKeyWithRecovery(vaultKey, mnemonic)
		require.NoError(t, err)
		_, err = DecryptVaultKeyWithRecovery(encrypted, "invalid mnemonic")
		assert.Error(t, err)
	})
}

func TestRegenerateRecoveryKey(t *testing.T) {
	t.Run("success generates valid mnemonic and decryptable key", func(t *testing.T) {
		vaultKey, err := GenerateVaultKey()
		require.NoError(t, err)

		newMnemonic, encVK, err := RegenerateRecoveryKey(vaultKey)
		require.NoError(t, err)
		assert.NotEmpty(t, newMnemonic)
		assert.NotNil(t, encVK)

		// Mnemonic should be valid BIP-39
		mgr := NewRecoveryKeyManager()
		assert.True(t, mgr.ValidateMnemonic(newMnemonic))

		// Should be a 12-word mnemonic
		words := strings.Fields(newMnemonic)
		assert.Len(t, words, 12)

		// Decrypt with the new mnemonic should recover the vault key
		decrypted, err := DecryptVaultKeyWithRecovery(encVK, newMnemonic)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted)
	})

	t.Run("empty vault key returns error", func(t *testing.T) {
		_, _, err := RegenerateRecoveryKey([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vault key must not be empty")
	})

	t.Run("nil vault key returns error", func(t *testing.T) {
		_, _, err := RegenerateRecoveryKey(nil)
		assert.Error(t, err)
	})

	t.Run("produces different mnemonic each time", func(t *testing.T) {
		vaultKey, err := GenerateVaultKey()
		require.NoError(t, err)

		m1, _, err := RegenerateRecoveryKey(vaultKey)
		require.NoError(t, err)
		m2, _, err := RegenerateRecoveryKey(vaultKey)
		require.NoError(t, err)

		assert.NotEqual(t, m1, m2)
	})

	t.Run("full round trip regenerate and decrypt", func(t *testing.T) {
		// Generate original vault key and recovery
		vaultKey, err := GenerateVaultKey()
		require.NoError(t, err)

		mgr := NewRecoveryKeyManager()
		origMnemonic, err := mgr.GenerateMnemonic()
		require.NoError(t, err)

		origEncVK, err := EncryptVaultKeyWithRecovery(vaultKey, origMnemonic)
		require.NoError(t, err)

		// Verify original works
		decrypted, err := DecryptVaultKeyWithRecovery(origEncVK, origMnemonic)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted)

		// Regenerate recovery key
		newMnemonic, newEncVK, err := RegenerateRecoveryKey(vaultKey)
		require.NoError(t, err)
		assert.NotEqual(t, origMnemonic, newMnemonic)

		// New mnemonic decrypts the new encrypted vault key
		decrypted2, err := DecryptVaultKeyWithRecovery(newEncVK, newMnemonic)
		require.NoError(t, err)
		assert.Equal(t, vaultKey, decrypted2)

		// Old mnemonic should NOT decrypt the new encrypted vault key
		_, err = DecryptVaultKeyWithRecovery(newEncVK, origMnemonic)
		assert.Error(t, err)
	})
}
