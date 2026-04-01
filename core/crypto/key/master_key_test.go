package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/crypto/kdf"
)

func TestDeriveNewMasterKey(t *testing.T) {
	k := kdf.New(kdf.Argon2idParams{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		KeyLen:      32,
		SaltLen:     16,
	})

	t.Run("derives master key", func(t *testing.T) {
		mk, err := DeriveNewMasterKey([]byte("strong-password"), k)
		require.NoError(t, err)
		assert.Len(t, mk.Key(), 32)
		assert.Len(t, mk.Salt(), 16)
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := DeriveNewMasterKey([]byte{}, k)
		assert.Error(t, err)
	})

	t.Run("nil password", func(t *testing.T) {
		_, err := DeriveNewMasterKey(nil, k)
		assert.Error(t, err)
	})

	t.Run("different calls produce different salts", func(t *testing.T) {
		mk1, err := DeriveNewMasterKey([]byte("password"), k)
		require.NoError(t, err)
		mk2, err := DeriveNewMasterKey([]byte("password"), k)
		require.NoError(t, err)
		assert.NotEqual(t, mk1.Salt(), mk2.Salt())
	})
}

func TestDeriveMasterKeyWithSalt(t *testing.T) {
	k := kdf.New(kdf.Argon2idParams{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		KeyLen:      32,
		SaltLen:     16,
	})

	t.Run("reproduces same key", func(t *testing.T) {
		mk1, err := DeriveNewMasterKey([]byte("password"), k)
		require.NoError(t, err)

		mk2, err := DeriveMasterKeyWithSalt([]byte("password"), mk1.Salt(), k)
		require.NoError(t, err)

		assert.Equal(t, mk1.Key(), mk2.Key())
	})

	t.Run("wrong password produces different key", func(t *testing.T) {
		mk1, err := DeriveNewMasterKey([]byte("password1"), k)
		require.NoError(t, err)

		mk2, err := DeriveMasterKeyWithSalt([]byte("password2"), mk1.Salt(), k)
		require.NoError(t, err)

		assert.NotEqual(t, mk1.Key(), mk2.Key())
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := DeriveMasterKeyWithSalt([]byte{}, []byte("salt1234salt1234"), k)
		assert.Error(t, err)
	})

	t.Run("empty salt", func(t *testing.T) {
		_, err := DeriveMasterKeyWithSalt([]byte("password"), []byte{}, k)
		assert.Error(t, err)
	})
}

func TestMasterKey_Zero(t *testing.T) {
	k := kdf.New(kdf.Argon2idParams{
		Memory:      4096,
		Iterations:  1,
		Parallelism: 1,
		KeyLen:      32,
		SaltLen:     16,
	})

	mk, err := DeriveNewMasterKey([]byte("password"), k)
	require.NoError(t, err)

	// Store a reference before zeroing
	keyRef := mk.Key()
	assert.Len(t, keyRef, 32)

	mk.Zero()

	// Key should be nil after zeroing
	assert.Nil(t, mk.Key())

	// The underlying bytes should be zeroed
	for i, b := range keyRef {
		assert.Equal(t, byte(0), b, "byte %d should be zero", i)
	}
}
