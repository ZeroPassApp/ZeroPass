package cipher

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	c := New()
	key := generateTestKey(t)

	t.Run("round trip", func(t *testing.T) {
		plaintext := []byte("Hello, ZeroPass!")
		ciphertext, err := c.Encrypt(plaintext, key)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := c.Decrypt(ciphertext, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("empty plaintext round trip", func(t *testing.T) {
		plaintext := []byte{}
		ciphertext, err := c.Encrypt(plaintext, key)
		require.NoError(t, err)
		// Should have at least nonce + tag
		assert.True(t, len(ciphertext) > NonceSize)

		decrypted, err := c.Decrypt(ciphertext, key)
		require.NoError(t, err)
		assert.Empty(t, decrypted)
	})

	t.Run("large plaintext", func(t *testing.T) {
		plaintext := make([]byte, 1<<16) // 64 KiB
		_, err := rand.Read(plaintext)
		require.NoError(t, err)

		ciphertext, err := c.Encrypt(plaintext, key)
		require.NoError(t, err)

		decrypted, err := c.Decrypt(ciphertext, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("unique nonce per encryption", func(t *testing.T) {
		plaintext := []byte("same plaintext")
		ct1, err := c.Encrypt(plaintext, key)
		require.NoError(t, err)
		ct2, err := c.Encrypt(plaintext, key)
		require.NoError(t, err)

		// Nonces (first 12 bytes) should differ
		assert.NotEqual(t, ct1[:NonceSize], ct2[:NonceSize])
		// Full ciphertexts should differ
		assert.NotEqual(t, ct1, ct2)
	})
}

func TestAESGCM_InvalidKey(t *testing.T) {
	c := New()

	t.Run("encrypt short key", func(t *testing.T) {
		_, err := c.Encrypt([]byte("test"), []byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key size")
	})

	t.Run("encrypt long key", func(t *testing.T) {
		longKey := make([]byte, 64)
		_, err := c.Encrypt([]byte("test"), longKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key size")
	})

	t.Run("decrypt short key", func(t *testing.T) {
		_, err := c.Decrypt(make([]byte, 28), []byte("short"))
		assert.Error(t, err)
	})

	t.Run("decrypt wrong key", func(t *testing.T) {
		key1 := generateTestKey(t)
		key2 := generateTestKey(t)

		ciphertext, err := c.Encrypt([]byte("secret"), key1)
		require.NoError(t, err)

		_, err = c.Decrypt(ciphertext, key2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication error")
	})
}

func TestAESGCM_CorruptedCiphertext(t *testing.T) {
	c := New()
	key := generateTestKey(t)

	t.Run("ciphertext too short", func(t *testing.T) {
		_, err := c.Decrypt([]byte("short"), key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})

	t.Run("tampered ciphertext", func(t *testing.T) {
		ciphertext, err := c.Encrypt([]byte("secret data"), key)
		require.NoError(t, err)

		// Tamper with the ciphertext (after nonce)
		ciphertext[NonceSize+1] ^= 0xFF

		_, err = c.Decrypt(ciphertext, key)
		assert.Error(t, err)
	})

	t.Run("tampered nonce", func(t *testing.T) {
		ciphertext, err := c.Encrypt([]byte("secret data"), key)
		require.NoError(t, err)

		// Tamper with the nonce
		ciphertext[0] ^= 0xFF

		_, err = c.Decrypt(ciphertext, key)
		assert.Error(t, err)
	})

	t.Run("truncated ciphertext", func(t *testing.T) {
		ciphertext, err := c.Encrypt([]byte("secret data"), key)
		require.NoError(t, err)

		// Truncate the ciphertext
		_, err = c.Decrypt(ciphertext[:NonceSize+2], key)
		assert.Error(t, err)
	})
}

func TestAESGCM_NilInputs(t *testing.T) {
	c := New()
	key := generateTestKey(t)

	t.Run("nil plaintext encrypts successfully", func(t *testing.T) {
		ciphertext, err := c.Encrypt(nil, key)
		require.NoError(t, err)

		decrypted, err := c.Decrypt(ciphertext, key)
		require.NoError(t, err)
		// nil plaintext decrypts to empty
		assert.Empty(t, decrypted)
	})

	t.Run("nil key", func(t *testing.T) {
		_, err := c.Encrypt([]byte("test"), nil)
		assert.Error(t, err)
	})

	t.Run("nil ciphertext decrypt", func(t *testing.T) {
		_, err := c.Decrypt(nil, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})

	t.Run("decrypt nil key", func(t *testing.T) {
		ciphertext, err := c.Encrypt([]byte("data"), key)
		require.NoError(t, err)
		_, err = c.Decrypt(ciphertext, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key size")
	})

	t.Run("decrypt long key", func(t *testing.T) {
		longKey := make([]byte, 64)
		_, err := c.Decrypt(make([]byte, 28), longKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key size")
	})

	t.Run("decrypt exact nonce size", func(t *testing.T) {
		// Exactly NonceSize bytes: nonce parsed but encrypted part is empty
		ciphertext := make([]byte, NonceSize)
		_, err := c.Decrypt(ciphertext, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication error")
	})

	t.Run("encrypt zero key", func(t *testing.T) {
		zeroKey := make([]byte, KeySize)
		ct, err := c.Encrypt([]byte("hello"), zeroKey)
		require.NoError(t, err)
		pt, err := c.Decrypt(ct, zeroKey)
		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), pt)
	})
}
