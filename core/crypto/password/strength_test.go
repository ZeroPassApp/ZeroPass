package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_ScoreStrength(t *testing.T) {
	g := NewGenerator()

	t.Run("empty password", func(t *testing.T) {
		result, err := g.ScoreStrength("")
		require.NoError(t, err)
		assert.Equal(t, 0, result.Score)
		assert.Contains(t, result.Feedback, "empty")
	})

	t.Run("weak password", func(t *testing.T) {
		result, err := g.ScoreStrength("password")
		require.NoError(t, err)
		assert.LessOrEqual(t, result.Score, 1)
	})

	t.Run("common password", func(t *testing.T) {
		result, err := g.ScoreStrength("123456")
		require.NoError(t, err)
		assert.LessOrEqual(t, result.Score, 1)
	})

	t.Run("strong password", func(t *testing.T) {
		result, err := g.ScoreStrength("c0rr3ct-h0rs3-b@tt3ry-st@pl3!")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Score, 3)
	})

	t.Run("very strong random", func(t *testing.T) {
		result, err := g.ScoreStrength("Xk9$mP2#qL7!wN4@vR6&hJ8*")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Score, 3)
	})

	t.Run("short password feedback", func(t *testing.T) {
		result, err := g.ScoreStrength("Ab1!")
		require.NoError(t, err)
		assert.Contains(t, result.Feedback, "too short")
	})

	t.Run("crack time is positive for non-empty", func(t *testing.T) {
		result, err := g.ScoreStrength("some-password-here")
		require.NoError(t, err)
		assert.Greater(t, result.CrackTimeSecs, float64(0))
	})
}

func TestStrengthFeedbackAllScores(t *testing.T) {
	g := NewGenerator()

	// Use passwords that are >= 12 chars to test all score-based branches
	t.Run("score 0 feedback", func(t *testing.T) {
		result := StrengthResult{Score: 0}
		// Score 0 with long password — test feedback text
		feedback := strengthFeedback(0, 14)
		assert.Contains(t, feedback, "Very weak")
		_ = result
	})

	t.Run("score 1 feedback", func(t *testing.T) {
		feedback := strengthFeedback(1, 14)
		assert.Contains(t, feedback, "Weak")
	})

	t.Run("score 2 feedback", func(t *testing.T) {
		feedback := strengthFeedback(2, 14)
		assert.Contains(t, feedback, "Fair")
	})

	t.Run("score 3 feedback", func(t *testing.T) {
		feedback := strengthFeedback(3, 14)
		assert.Contains(t, feedback, "Strong")
	})

	t.Run("score 4 feedback", func(t *testing.T) {
		feedback := strengthFeedback(4, 14)
		assert.Contains(t, feedback, "Very strong")
	})

	t.Run("invalid score feedback", func(t *testing.T) {
		feedback := strengthFeedback(5, 14)
		assert.Contains(t, feedback, "Unable to evaluate")
	})

	// Ensure ScoreStrength works with a moderate-length repeating password (triggers score ~1-2)
	t.Run("moderate password", func(t *testing.T) {
		result, err := g.ScoreStrength("aaaaaaaaaaaa") // 12 a's - weak but long
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Score, 0)
	})
}

func TestIsSecure(t *testing.T) {
	t.Run("secure password", func(t *testing.T) {
		result := StrengthResult{Score: 4}
		assert.True(t, IsSecure(result, 20))
	})

	t.Run("insecure score", func(t *testing.T) {
		result := StrengthResult{Score: 2}
		assert.False(t, IsSecure(result, 20))
	})

	t.Run("too short", func(t *testing.T) {
		result := StrengthResult{Score: 4}
		assert.False(t, IsSecure(result, 8))
	})

	t.Run("minimum acceptable", func(t *testing.T) {
		result := StrengthResult{Score: 3}
		assert.True(t, IsSecure(result, 12))
	})

	t.Run("just below minimum score", func(t *testing.T) {
		result := StrengthResult{Score: 2}
		assert.False(t, IsSecure(result, 12))
	})

	t.Run("just below minimum length", func(t *testing.T) {
		result := StrengthResult{Score: 3}
		assert.False(t, IsSecure(result, 11))
	})
}
