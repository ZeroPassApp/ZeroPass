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
