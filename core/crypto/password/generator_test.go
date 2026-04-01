package password

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateRandom(t *testing.T) {
	g := NewGenerator()

	t.Run("generates password of correct length", func(t *testing.T) {
		pw, err := g.GenerateRandom(16, GeneratorOptions{
			Lowercase: true, Uppercase: true, Digits: true, Symbols: true,
		})
		require.NoError(t, err)
		assert.Len(t, pw, 16)
	})

	t.Run("minimum length", func(t *testing.T) {
		pw, err := g.GenerateRandom(8, GeneratorOptions{Lowercase: true})
		require.NoError(t, err)
		assert.Len(t, pw, 8)
	})

	t.Run("maximum length", func(t *testing.T) {
		pw, err := g.GenerateRandom(128, GeneratorOptions{Lowercase: true})
		require.NoError(t, err)
		assert.Len(t, pw, 128)
	})

	t.Run("too short", func(t *testing.T) {
		_, err := g.GenerateRandom(7, GeneratorOptions{Lowercase: true})
		assert.Error(t, err)
	})

	t.Run("too long", func(t *testing.T) {
		_, err := g.GenerateRandom(129, GeneratorOptions{Lowercase: true})
		assert.Error(t, err)
	})

	t.Run("contains required character types", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			pw, err := g.GenerateRandom(20, GeneratorOptions{
				Lowercase: true, Uppercase: true, Digits: true, Symbols: true,
			})
			require.NoError(t, err)

			hasLower := false
			hasUpper := false
			hasDigit := false
			hasSymbol := false
			for _, c := range pw {
				switch {
				case unicode.IsLower(c):
					hasLower = true
				case unicode.IsUpper(c):
					hasUpper = true
				case unicode.IsDigit(c):
					hasDigit = true
				default:
					hasSymbol = true
				}
			}
			assert.True(t, hasLower, "should contain lowercase")
			assert.True(t, hasUpper, "should contain uppercase")
			assert.True(t, hasDigit, "should contain digit")
			assert.True(t, hasSymbol, "should contain symbol")
		}
	})

	t.Run("lowercase only", func(t *testing.T) {
		pw, err := g.GenerateRandom(20, GeneratorOptions{Lowercase: true})
		require.NoError(t, err)
		for _, c := range pw {
			assert.True(t, unicode.IsLower(c), "should only contain lowercase, got: %c", c)
		}
	})

	t.Run("uppercase only", func(t *testing.T) {
		pw, err := g.GenerateRandom(20, GeneratorOptions{Uppercase: true})
		require.NoError(t, err)
		for _, c := range pw {
			assert.True(t, unicode.IsUpper(c), "should only contain uppercase, got: %c", c)
		}
	})

	t.Run("digits only", func(t *testing.T) {
		pw, err := g.GenerateRandom(20, GeneratorOptions{Digits: true})
		require.NoError(t, err)
		for _, c := range pw {
			assert.True(t, unicode.IsDigit(c), "should only contain digits, got: %c", c)
		}
	})

	t.Run("exclude ambiguous chars", func(t *testing.T) {
		ambiguous := "0O1lI"
		for i := 0; i < 20; i++ {
			pw, err := g.GenerateRandom(64, GeneratorOptions{
				Lowercase: true, Uppercase: true, Digits: true,
				ExcludeAmbiguous: true,
			})
			require.NoError(t, err)
			for _, c := range pw {
				assert.False(t, strings.ContainsRune(ambiguous, c),
					"should not contain ambiguous char: %c", c)
			}
		}
	})

	t.Run("default options when none set", func(t *testing.T) {
		pw, err := g.GenerateRandom(20, GeneratorOptions{})
		require.NoError(t, err)
		assert.Len(t, pw, 20)
	})

	t.Run("generates unique passwords", func(t *testing.T) {
		passwords := make(map[string]bool)
		for i := 0; i < 100; i++ {
			pw, err := g.GenerateRandom(20, GeneratorOptions{
				Lowercase: true, Uppercase: true, Digits: true, Symbols: true,
			})
			require.NoError(t, err)
			passwords[pw] = true
		}
		// With 20 chars from ~80 char pool, collisions are astronomically unlikely
		assert.Equal(t, 100, len(passwords), "all passwords should be unique")
	})
}

func TestGenerator_GenerateRandom_LengthTooShortForRequiredSets(t *testing.T) {
	g := NewGenerator()
	// 4 required sets (lower, upper, digit, symbol) but length=8 should still work
	pw, err := g.GenerateRandom(8, GeneratorOptions{
		Lowercase: true, Uppercase: true, Digits: true, Symbols: true,
	})
	require.NoError(t, err)
	assert.Len(t, pw, 8)
}

func TestGenerator_GeneratePassphrase(t *testing.T) {
	g := NewGenerator()

	t.Run("generates correct word count", func(t *testing.T) {
		pp, err := g.GeneratePassphrase(4, "-")
		require.NoError(t, err)
		words := strings.Split(pp, "-")
		assert.Len(t, words, 4)
	})

	t.Run("custom separator", func(t *testing.T) {
		pp, err := g.GeneratePassphrase(5, ".")
		require.NoError(t, err)
		words := strings.Split(pp, ".")
		assert.Len(t, words, 5)
	})

	t.Run("space separator", func(t *testing.T) {
		pp, err := g.GeneratePassphrase(4, " ")
		require.NoError(t, err)
		words := strings.Fields(pp)
		assert.Len(t, words, 4)
	})

	t.Run("minimum words", func(t *testing.T) {
		pp, err := g.GeneratePassphrase(4, "-")
		require.NoError(t, err)
		assert.NotEmpty(t, pp)
	})

	t.Run("maximum words", func(t *testing.T) {
		pp, err := g.GeneratePassphrase(8, "-")
		require.NoError(t, err)
		words := strings.Split(pp, "-")
		assert.Len(t, words, 8)
	})

	t.Run("too few words", func(t *testing.T) {
		_, err := g.GeneratePassphrase(3, "-")
		assert.Error(t, err)
	})

	t.Run("too many words", func(t *testing.T) {
		_, err := g.GeneratePassphrase(9, "-")
		assert.Error(t, err)
	})

	t.Run("generates unique passphrases", func(t *testing.T) {
		passphrases := make(map[string]bool)
		for i := 0; i < 50; i++ {
			pp, err := g.GeneratePassphrase(6, "-")
			require.NoError(t, err)
			passphrases[pp] = true
		}
		assert.Equal(t, 50, len(passphrases), "all passphrases should be unique")
	})
}
