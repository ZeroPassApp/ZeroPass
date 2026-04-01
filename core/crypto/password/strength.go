package password

import (
	"fmt"

	zxcvbn "github.com/nbutton23/zxcvbn-go"
)

// StrengthResult holds the result of a password strength evaluation.
type StrengthResult struct {
	Score         int     // 0-4 (0=very weak, 4=very strong)
	Feedback      string  // Human-readable feedback
	CrackTimeSecs float64 // Estimated offline crack time in seconds
}

const (
	// MinSecureLength is the minimum password length for a secure password.
	MinSecureLength = 12
	// MinSecureScore is the minimum zxcvbn score for a secure password.
	MinSecureScore = 3
)

// ScoreStrength evaluates the strength of a password.
func (g *Generator) ScoreStrength(password string) (StrengthResult, error) {
	if len(password) == 0 {
		return StrengthResult{
			Score:    0,
			Feedback: "Password is empty",
		}, nil
	}

	result := zxcvbn.PasswordStrength(password, nil)

	feedback := strengthFeedback(result.Score, len(password))

	return StrengthResult{
		Score:         result.Score,
		Feedback:      feedback,
		CrackTimeSecs: result.CrackTime,
	}, nil
}

// IsSecure checks whether a password meets the minimum security requirements.
func IsSecure(result StrengthResult, passwordLen int) bool {
	return passwordLen >= MinSecureLength && result.Score >= MinSecureScore
}

// strengthFeedback generates human-readable feedback based on score and length.
func strengthFeedback(score int, length int) string {
	if length < MinSecureLength {
		return fmt.Sprintf("Password is too short. Use at least %d characters.", MinSecureLength)
	}

	switch score {
	case 0:
		return "Very weak password. Use a mix of uppercase, lowercase, digits, and symbols."
	case 1:
		return "Weak password. Consider adding more varied characters or increasing length."
	case 2:
		return "Fair password. Consider adding more complexity."
	case 3:
		return "Strong password."
	case 4:
		return "Very strong password."
	default:
		return "Unable to evaluate password strength."
	}
}
