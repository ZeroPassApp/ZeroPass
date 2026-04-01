// Package crypto defines the core cryptographic interfaces and types for ZeroPass.
package crypto

// KDF defines a key derivation function.
type KDF interface {
	// DeriveKey derives a cryptographic key from a password and salt.
	DeriveKey(password []byte, salt []byte) ([]byte, error)
}

// Cipher defines symmetric encryption/decryption operations.
type Cipher interface {
	// Encrypt encrypts plaintext using the provided key.
	// The returned ciphertext includes any required metadata (e.g., nonce).
	Encrypt(plaintext []byte, key []byte) ([]byte, error)
	Decrypt(ciphertext []byte, key []byte) ([]byte, error)
}

// RecoveryManager handles mnemonic-based key recovery.
type RecoveryManager interface {
	// GenerateMnemonic creates a new BIP-39 mnemonic phrase.
	GenerateMnemonic() (string, error)
	// DeriveKeyFromMnemonic derives a cryptographic key from a mnemonic phrase.
	DeriveKeyFromMnemonic(mnemonic string) ([]byte, error)
	// ValidateMnemonic checks whether a mnemonic phrase is valid.
	ValidateMnemonic(mnemonic string) bool
}

// GeneratorOptions configures password generation.
type GeneratorOptions struct {
	Uppercase       bool
	Lowercase       bool
	Digits          bool
	Symbols         bool
	ExcludeAmbiguous bool // Exclude ambiguous characters (0, O, l, 1, I, etc.)
}

// StrengthResult holds the result of a password strength evaluation.
type StrengthResult struct {
	Score         int     // 0-4 (0=very weak, 4=very strong)
	Feedback      string  // Human-readable feedback
	CrackTimeSecs float64 // Estimated offline crack time in seconds
}

// PasswordGenerator generates and evaluates passwords.
type PasswordGenerator interface {
	// GenerateRandom generates a random password of the given length.
	GenerateRandom(length int, opts GeneratorOptions) (string, error)
	// GeneratePassphrase generates a passphrase with the given number of words.
	GeneratePassphrase(words int, separator string) (string, error)
	// ScoreStrength evaluates the strength of a password.
	ScoreStrength(password string) (StrengthResult, error)
}
