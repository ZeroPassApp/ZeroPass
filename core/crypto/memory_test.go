package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroBytes(t *testing.T) {
	t.Run("zeroes all bytes", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04, 0xFF}
		ZeroBytes(data)
		for i, b := range data {
			assert.Equal(t, byte(0), b, "byte at index %d should be zero", i)
		}
	})

	t.Run("handles nil slice", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ZeroBytes(nil)
		})
	})

	t.Run("handles empty slice", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ZeroBytes([]byte{})
		})
	})

	t.Run("handles single byte", func(t *testing.T) {
		data := []byte{0xFF}
		ZeroBytes(data)
		assert.Equal(t, byte(0), data[0])
	})

	t.Run("handles large slice", func(t *testing.T) {
		data := make([]byte, 4096)
		for i := range data {
			data[i] = 0xFF
		}
		ZeroBytes(data)
		for i, b := range data {
			assert.Equal(t, byte(0), b, "byte at index %d should be zero", i)
		}
	})
}

func TestSecureCompare(t *testing.T) {
	t.Run("equal slices returns true", func(t *testing.T) {
		a := []byte{0x01, 0x02, 0x03, 0x04}
		b := []byte{0x01, 0x02, 0x03, 0x04}
		assert.True(t, SecureCompare(a, b))
	})

	t.Run("different slices returns false", func(t *testing.T) {
		a := []byte{0x01, 0x02, 0x03, 0x04}
		b := []byte{0x01, 0x02, 0x03, 0x05}
		assert.False(t, SecureCompare(a, b))
	})

	t.Run("different lengths returns false", func(t *testing.T) {
		a := []byte{0x01, 0x02, 0x03}
		b := []byte{0x01, 0x02, 0x03, 0x04}
		assert.False(t, SecureCompare(a, b))
	})

	t.Run("empty slices returns true", func(t *testing.T) {
		assert.True(t, SecureCompare([]byte{}, []byte{}))
	})

	t.Run("nil slices returns true", func(t *testing.T) {
		assert.True(t, SecureCompare(nil, nil))
	})

	t.Run("one nil one empty returns true", func(t *testing.T) {
		// Both have length 0, so they are equal
		assert.True(t, SecureCompare(nil, []byte{}))
	})

	t.Run("completely different content returns false", func(t *testing.T) {
		a := []byte{0x00, 0x00, 0x00, 0x00}
		b := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		assert.False(t, SecureCompare(a, b))
	})
}

func TestSecureCompareStrings(t *testing.T) {
	t.Run("equal strings returns true", func(t *testing.T) {
		assert.True(t, SecureCompareStrings("password123", "password123"))
	})

	t.Run("different strings returns false", func(t *testing.T) {
		assert.False(t, SecureCompareStrings("password123", "password456"))
	})

	t.Run("different length strings returns false", func(t *testing.T) {
		assert.False(t, SecureCompareStrings("short", "a longer string"))
	})

	t.Run("empty strings returns true", func(t *testing.T) {
		assert.True(t, SecureCompareStrings("", ""))
	})

	t.Run("one empty one non-empty returns false", func(t *testing.T) {
		assert.False(t, SecureCompareStrings("", "notempty"))
	})
}
