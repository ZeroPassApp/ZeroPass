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
