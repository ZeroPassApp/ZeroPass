// Package password provides password generation and strength evaluation.
package password

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	// MinPasswordLength is the minimum allowed password length.
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum allowed password length.
	MaxPasswordLength = 128
	// MinPassphraseWords is the minimum number of words in a passphrase.
	MinPassphraseWords = 4
	// MaxPassphraseWords is the maximum number of words in a passphrase.
	MaxPassphraseWords = 8
)

// Character sets for password generation.
const (
	lowercaseChars = "abcdefghijkmnopqrstuvwxyz"          // no 'l'
	uppercaseChars = "ABCDEFGHJKLMNPQRSTUVWXYZ"            // no 'I', 'O'
	digitChars     = "23456789"                             // no '0', '1'
	symbolChars    = "!@#$%^&*()-_=+[]{}|;:,.<>?"

	// Full sets (with ambiguous chars)
	lowercaseCharsFull = "abcdefghijklmnopqrstuvwxyz"
	uppercaseCharsFull = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitCharsFull     = "0123456789"
)

// Generator implements the crypto.PasswordGenerator interface.
type Generator struct{}

// NewGenerator creates a new password generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateRandom generates a random password of the given length with configurable options.
func (g *Generator) GenerateRandom(length int, opts GeneratorOptions) (string, error) {
	if length < MinPasswordLength || length > MaxPasswordLength {
		return "", fmt.Errorf("password length must be between %d and %d", MinPasswordLength, MaxPasswordLength)
	}

	// Build the character pool based on options
	var charset string
	var required []string

	if opts.Lowercase {
		if opts.ExcludeAmbiguous {
			charset += lowercaseChars
			required = append(required, lowercaseChars)
		} else {
			charset += lowercaseCharsFull
			required = append(required, lowercaseCharsFull)
		}
	}
	if opts.Uppercase {
		if opts.ExcludeAmbiguous {
			charset += uppercaseChars
			required = append(required, uppercaseChars)
		} else {
			charset += uppercaseCharsFull
			required = append(required, uppercaseCharsFull)
		}
	}
	if opts.Digits {
		if opts.ExcludeAmbiguous {
			charset += digitChars
			required = append(required, digitChars)
		} else {
			charset += digitCharsFull
			required = append(required, digitCharsFull)
		}
	}
	if opts.Symbols {
		charset += symbolChars
		required = append(required, symbolChars)
	}

	// Default to all character types if none specified
	if charset == "" {
		charset = lowercaseCharsFull + uppercaseCharsFull + digitCharsFull + symbolChars
		required = []string{lowercaseCharsFull, uppercaseCharsFull, digitCharsFull, symbolChars}
	}

	if len(required) > length {
		return "", errors.New("password length too short to satisfy all character type requirements")
	}

	// Generate password ensuring at least one char from each required set
	password := make([]byte, length)

	// First, place one character from each required set at random positions
	positions := make([]int, length)
	for i := range positions {
		positions[i] = i
	}
	shuffleSlice(positions)

	for i, req := range required {
		idx, err := secureRandomInt(len(req))
		if err != nil {
			return "", err
		}
		password[positions[i]] = req[idx]
	}

	// Fill remaining positions with random chars from the full charset
	for i := len(required); i < length; i++ {
		idx, err := secureRandomInt(len(charset))
		if err != nil {
			return "", err
		}
		password[positions[i]] = charset[idx]
	}

	return string(password), nil
}

// GeneratePassphrase generates a passphrase with the given number of words separated by the specified separator.
func (g *Generator) GeneratePassphrase(words int, separator string) (string, error) {
	if words < MinPassphraseWords || words > MaxPassphraseWords {
		return "", fmt.Errorf("word count must be between %d and %d", MinPassphraseWords, MaxPassphraseWords)
	}

	wordList := getWordList()
	selected := make([]string, words)

	for i := 0; i < words; i++ {
		idx, err := secureRandomInt(len(wordList))
		if err != nil {
			return "", fmt.Errorf("select word: %w", err)
		}
		selected[i] = wordList[idx]
	}

	return strings.Join(selected, separator), nil
}

// GeneratorOptions configures password generation (re-exported from crypto package).
type GeneratorOptions struct {
	Uppercase        bool
	Lowercase        bool
	Digits           bool
	Symbols          bool
	ExcludeAmbiguous bool
}

// secureRandomInt returns a cryptographically secure random integer in [0, max).
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be positive")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("secure random: %w", err)
	}
	return int(n.Int64()), nil
}

// shuffleSlice performs a Fisher-Yates shuffle using crypto/rand.
func shuffleSlice(s []int) {
	for i := len(s) - 1; i > 0; i-- {
		j, err := secureRandomInt(i + 1)
		if err != nil {
			// Fallback: swap with previous (still deterministic but acceptable for position shuffling)
			j = i - 1
		}
		s[i], s[j] = s[j], s[i]
	}
}
