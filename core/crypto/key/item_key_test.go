package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveItemKey(t *testing.T) {
	vaultKey := generateKey(t)

	t.Run("derives 32-byte key", func(t *testing.T) {
		itemKey, err := DeriveItemKey(vaultKey, "item-001")
		require.NoError(t, err)
		assert.Len(t, itemKey, ItemKeySize)
	})

	t.Run("same inputs produce same key", func(t *testing.T) {
		key1, err := DeriveItemKey(vaultKey, "item-001")
		require.NoError(t, err)
		key2, err := DeriveItemKey(vaultKey, "item-001")
		require.NoError(t, err)
		assert.Equal(t, key1, key2)
	})

	t.Run("different item IDs produce different keys", func(t *testing.T) {
		key1, err := DeriveItemKey(vaultKey, "item-001")
		require.NoError(t, err)
		key2, err := DeriveItemKey(vaultKey, "item-002")
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("different vault keys produce different keys", func(t *testing.T) {
		vk2 := generateKey(t)
		key1, err := DeriveItemKey(vaultKey, "item-001")
		require.NoError(t, err)
		key2, err := DeriveItemKey(vk2, "item-001")
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("empty vault key returns error", func(t *testing.T) {
		_, err := DeriveItemKey([]byte{}, "item-001")
		assert.Error(t, err)
	})

	t.Run("nil vault key returns error", func(t *testing.T) {
		_, err := DeriveItemKey(nil, "item-001")
		assert.Error(t, err)
	})

	t.Run("empty item ID returns error", func(t *testing.T) {
		_, err := DeriveItemKey(vaultKey, "")
		assert.Error(t, err)
	})
}
