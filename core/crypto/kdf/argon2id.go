// Package kdf provides key derivation functions for ZeroPass.
package kdf

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// DefaultMemory is the memory parameter in KiB (64 MB).
	DefaultMemory uint32 = 65536
	// DefaultIterations is the number of passes over the memory.
	DefaultIterations uint32 = 3
	// DefaultParallelism is the number of threads to use.
	DefaultParallelism uint8 = 4
	// DefaultKeyLen is the desired key length in bytes.
	DefaultKeyLen uint32 = 32
	// DefaultSaltLen is the default salt length in bytes.
	DefaultSaltLen = 16
)

// Argon2idParams holds the parameters for the Argon2id KDF.
type Argon2idParams struct {
	Memory      uint32 // Memory in KiB
	Iterations  uint32 // Number of passes
	Parallelism uint8  // Degree of parallelism
	KeyLen      uint32 // Desired key length in bytes
	SaltLen     int    // Salt length in bytes
}

// DefaultParams returns the default Argon2id parameters.
func DefaultParams() Argon2idParams {
	return Argon2idParams{
		Memory:      DefaultMemory,
		Iterations:  DefaultIterations,
		Parallelism: DefaultParallelism,
		KeyLen:      DefaultKeyLen,
		SaltLen:     DefaultSaltLen,
	}
}

// Argon2id implements the crypto.KDF interface using Argon2id.
type Argon2id struct {
	params Argon2idParams
}

// New creates a new Argon2id KDF with the given parameters.
func New(params Argon2idParams) *Argon2id {
	return &Argon2id{params: params}
}

// NewDefault creates a new Argon2id KDF with default parameters.
func NewDefault() *Argon2id {
	return New(DefaultParams())
}

// DeriveKey derives a key from the password and salt using Argon2id.
func (a *Argon2id) DeriveKey(password []byte, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("password must not be empty")
	}
	if len(salt) == 0 {
		return nil, errors.New("salt must not be empty")
	}
	if a.params.Memory == 0 || a.params.Iterations == 0 || a.params.Parallelism == 0 || a.params.KeyLen == 0 {
		return nil, errors.New("invalid KDF parameters")
	}

	key := argon2.IDKey(password, salt, a.params.Iterations, a.params.Memory, a.params.Parallelism, a.params.KeyLen)
	return key, nil
}

// GenerateSalt generates a cryptographically secure random salt.
func (a *Argon2id) GenerateSalt() ([]byte, error) {
	saltLen := a.params.SaltLen
	if saltLen <= 0 {
		saltLen = DefaultSaltLen
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

// Params returns the current Argon2id parameters.
func (a *Argon2id) Params() Argon2idParams {
	return a.params
}
