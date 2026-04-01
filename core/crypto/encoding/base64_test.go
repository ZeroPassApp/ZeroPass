package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBase64URLRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"hello", []byte("hello")},
		{"binary data", []byte{0x00, 0xFF, 0x80, 0x7F, 0x01}},
		{"url-unsafe chars", []byte{0xFB, 0xFF, 0xFE}},
		{"long data", make([]byte, 1024)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Base64URLEncode(tc.input)
			decoded, err := Base64URLDecode(encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestBase64URLDecodeInvalid(t *testing.T) {
	_, err := Base64URLDecode("!!!invalid!!!")
	assert.Error(t, err)
}

func TestBase64StdRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"hello world", []byte("hello world")},
		{"binary", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Base64StdEncode(tc.input)
			decoded, err := Base64StdDecode(encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestBase64StdDecodeInvalid(t *testing.T) {
	_, err := Base64StdDecode("!!!invalid!!!")
	assert.Error(t, err)
}

func TestHexRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"deadbeef", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"zeros", []byte{0x00, 0x00, 0x00}},
		{"all bytes", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := HexEncode(tc.input)
			decoded, err := HexDecode(encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestHexDecodeInvalid(t *testing.T) {
	_, err := HexDecode("not-hex!")
	assert.Error(t, err)
}

func TestHexEncodeKnownVector(t *testing.T) {
	assert.Equal(t, "deadbeef", HexEncode([]byte{0xDE, 0xAD, 0xBE, 0xEF}))
}

func TestBase64URLEncodeKnownVector(t *testing.T) {
	// "Hello" -> "SGVsbG8" in raw URL base64
	assert.Equal(t, "SGVsbG8", Base64URLEncode([]byte("Hello")))
}
