package kdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2id_DeriveKey(t *testing.T) {
	kdf := NewDefault()

	t.Run("derives 32-byte key", func(t *testing.T) {
		salt, err := kdf.GenerateSalt()
		require.NoError(t, err)

		key, err := kdf.DeriveKey([]byte("password123"), salt)
		require.NoError(t, err)
		assert.Len(t, key, 32)
	})

	t.Run("same input produces same output", func(t *testing.T) {
		salt := []byte("fixed-salt-16!!")
		key1, err := kdf.DeriveKey([]byte("password"), salt)
		require.NoError(t, err)
		key2, err := kdf.DeriveKey([]byte("password"), salt)
		require.NoError(t, err)
		assert.Equal(t, key1, key2)
	})

	t.Run("different password produces different key", func(t *testing.T) {
		salt := []byte("fixed-salt-16!!")
		key1, err := kdf.DeriveKey([]byte("password1"), salt)
		require.NoError(t, err)
		key2, err := kdf.DeriveKey([]byte("password2"), salt)
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("different salt produces different key", func(t *testing.T) {
		key1, err := kdf.DeriveKey([]byte("password"), []byte("salt-one-16byte"))
		require.NoError(t, err)
		key2, err := kdf.DeriveKey([]byte("password"), []byte("salt-two-16byte"))
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("empty password returns error", func(t *testing.T) {
		_, err := kdf.DeriveKey([]byte{}, []byte("somesalt"))
		assert.Error(t, err)
	})

	t.Run("nil password returns error", func(t *testing.T) {
		_, err := kdf.DeriveKey(nil, []byte("somesalt"))
		assert.Error(t, err)
	})

	t.Run("empty salt returns error", func(t *testing.T) {
		_, err := kdf.DeriveKey([]byte("password"), []byte{})
		assert.Error(t, err)
	})

	t.Run("nil salt returns error", func(t *testing.T) {
		_, err := kdf.DeriveKey([]byte("password"), nil)
		assert.Error(t, err)
	})
}

func TestArgon2id_CustomParams(t *testing.T) {
	params := Argon2idParams{
		Memory:      32768,
		Iterations:  2,
		Parallelism: 2,
		KeyLen:      64,
		SaltLen:     32,
	}
	kdf := New(params)

	t.Run("custom key length", func(t *testing.T) {
		salt, err := kdf.GenerateSalt()
		require.NoError(t, err)
		assert.Len(t, salt, 32)

		key, err := kdf.DeriveKey([]byte("password"), salt)
		require.NoError(t, err)
		assert.Len(t, key, 64)
	})

	t.Run("returns params", func(t *testing.T) {
		assert.Equal(t, params, kdf.Params())
	})
}

func TestArgon2id_InvalidParams(t *testing.T) {
	kdf := New(Argon2idParams{
		Memory:      0,
		Iterations:  0,
		Parallelism: 0,
		KeyLen:      0,
		SaltLen:     16,
	})

	_, err := kdf.DeriveKey([]byte("password"), []byte("somesalt12345678"))
	assert.Error(t, err)
}

func TestArgon2id_GenerateSalt(t *testing.T) {
	kdf := NewDefault()

	t.Run("generates salt of correct length", func(t *testing.T) {
		salt, err := kdf.GenerateSalt()
		require.NoError(t, err)
		assert.Len(t, salt, DefaultSaltLen)
	})

	t.Run("generates unique salts", func(t *testing.T) {
		salt1, err := kdf.GenerateSalt()
		require.NoError(t, err)
		salt2, err := kdf.GenerateSalt()
		require.NoError(t, err)
		assert.NotEqual(t, salt1, salt2)
	})

	t.Run("zero salt len uses default", func(t *testing.T) {
		k := New(Argon2idParams{
			Memory:      4096,
			Iterations:  1,
			Parallelism: 1,
			KeyLen:      32,
			SaltLen:     0,
		})
		salt, err := k.GenerateSalt()
		require.NoError(t, err)
		assert.Len(t, salt, DefaultSaltLen)
	})

	t.Run("negative salt len uses default", func(t *testing.T) {
		k := New(Argon2idParams{
			Memory:      4096,
			Iterations:  1,
			Parallelism: 1,
			KeyLen:      32,
			SaltLen:     -1,
		})
		salt, err := k.GenerateSalt()
		require.NoError(t, err)
		assert.Len(t, salt, DefaultSaltLen)
	})
}

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()
	assert.Equal(t, uint32(65536), p.Memory)
	assert.Equal(t, uint32(3), p.Iterations)
	assert.Equal(t, uint8(4), p.Parallelism)
	assert.Equal(t, uint32(32), p.KeyLen)
	assert.Equal(t, 16, p.SaltLen)
}
