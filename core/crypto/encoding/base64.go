// Package encoding provides encoding/decoding utilities for ZeroPass.
package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Base64URLEncode encodes data using URL-safe base64 encoding without padding.
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes a URL-safe base64 encoded string without padding.
func Base64URLDecode(s string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64url decode: %w", err)
	}
	return data, nil
}

// Base64StdEncode encodes data using standard base64 encoding with padding.
func Base64StdEncode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64StdDecode decodes a standard base64 encoded string with padding.
func Base64StdDecode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	return data, nil
}

// HexEncode encodes data as a hexadecimal string.
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode decodes a hexadecimal string to bytes.
func HexDecode(s string) ([]byte, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("hex decode: %w", err)
	}
	return data, nil
}
