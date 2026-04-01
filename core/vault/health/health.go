// Package health provides password health analysis for vault items.
package health

import (
	"fmt"
	"time"

	"github.com/zeropass/zeropass/core/crypto/password"
	"github.com/zeropass/zeropass/core/vault/types"
)

// Severity levels for health findings.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Finding represents a single health issue found in the vault.
type Finding struct {
	ItemID   string   `json:"item_id"`
	ItemName string   `json:"item_name"`
	Severity Severity `json:"severity"`
	Category string   `json:"category"`
	Message  string   `json:"message"`
}

// BreachFinding holds breach-specific details for a finding.
type BreachFinding struct {
	ItemID          string `json:"item_id"`
	ItemName        string `json:"item_name"`
	OccurrenceCount int    `json:"occurrence_count"`
	Recommendation  string `json:"recommendation"`
}

// HealthReport contains the results of a vault health analysis.
type HealthReport struct {
	TotalItems        int              `json:"total_items"`
	LoginItems        int              `json:"login_items"`
	WeakCount         int              `json:"weak_count"`
	ReusedCount       int              `json:"reused_count"`
	OldCount          int              `json:"old_count"`
	BreachedCount     int              `json:"breached_count"`
	BreachedPasswords []BreachFinding  `json:"breached_passwords,omitempty"`
	Findings          []Finding        `json:"findings"`
	AnalyzedAt        time.Time        `json:"analyzed_at"`
	OverallScore      int              `json:"overall_score"` // 0-100
}

// AnalyzerConfig holds configurable thresholds.
type AnalyzerConfig struct {
	MinScore       int           // minimum zxcvbn score (default 3)
	MaxPasswordAge time.Duration // max age before warning (default 90 days)
}

// DefaultAnalyzerConfig returns sensible defaults.
func DefaultAnalyzerConfig() AnalyzerConfig {
	return AnalyzerConfig{
		MinScore:       3,
		MaxPasswordAge: 90 * 24 * time.Hour,
	}
}

// Analyzer performs password health analysis.
type Analyzer struct {
	config AnalyzerConfig
	gen    *password.Generator
}

// NewAnalyzer creates a new health analyzer.
func NewAnalyzer(cfg AnalyzerConfig) *Analyzer {
	return &Analyzer{
		config: cfg,
		gen:    password.NewGenerator(),
	}
}

// Analyze inspects all provided items and returns a health report.
func (a *Analyzer) Analyze(items []*types.Item) *HealthReport {
	report := &HealthReport{
		TotalItems: len(items),
		AnalyzedAt: time.Now().UTC(),
	}

	// Collect login items.
	var logins []*types.Item
	for _, item := range items {
		if item.Type == types.ItemTypeLogin {
			logins = append(logins, item)
		}
	}
	report.LoginItems = len(logins)

	if len(logins) == 0 {
		report.OverallScore = 100
		return report
	}

	// Track password reuse.
	passwordUsers := make(map[string][]passwordRef)
	now := time.Now()

	for _, item := range logins {
		pw := item.Fields[types.FieldPassword]
		if pw == "" {
			continue
		}

		// Weak password check.
		result, _ := a.gen.ScoreStrength(pw)
		if result.Score < a.config.MinScore {
			sev := SeverityWarning
			if result.Score <= 1 {
				sev = SeverityCritical
			}
			report.Findings = append(report.Findings, Finding{
				ItemID:   item.ID,
				ItemName: item.Name,
				Severity: sev,
				Category: "weak",
				Message:  result.Feedback,
			})
			report.WeakCount++
		}

		// Collect for reuse check.
		passwordUsers[pw] = append(passwordUsers[pw], passwordRef{
			ID:   item.ID,
			Name: item.Name,
		})

		// Old password check.
		if !item.UpdatedAt.IsZero() && now.Sub(item.UpdatedAt) > a.config.MaxPasswordAge {
			report.Findings = append(report.Findings, Finding{
				ItemID:   item.ID,
				ItemName: item.Name,
				Severity: SeverityInfo,
				Category: "old",
				Message:  "Password has not been updated in over 90 days.",
			})
			report.OldCount++
		}
	}

	// Check for reused passwords.
	for _, refs := range passwordUsers {
		if len(refs) > 1 {
			for _, ref := range refs {
				report.Findings = append(report.Findings, Finding{
					ItemID:   ref.ID,
					ItemName: ref.Name,
					Severity: SeverityWarning,
					Category: "reused",
					Message:  "Password is reused across multiple items.",
				})
			}
			report.ReusedCount += len(refs)
		}
	}

	// Calculate overall score.
	report.OverallScore = calculateScore(report)
	return report
}

type passwordRef struct {
	ID   string
	Name string
}

func calculateScore(r *HealthReport) int {
	if r.LoginItems == 0 {
		return 100
	}

	// Breached passwords are critical — each counts double.
	issues := r.WeakCount + r.ReusedCount + r.OldCount + (r.BreachedCount * 2)
	if issues == 0 {
		return 100
	}

	// Each issue reduces score proportionally.
	penalty := float64(issues) / float64(r.LoginItems) * 100
	score := 100 - int(penalty)
	if score < 0 {
		score = 0
	}
	return score
}

// AnalyzeWithBreach performs a full health analysis including HIBP breach checks.
// The hibpClient must be non-nil; use NewHIBPClient to create one.
// Breach checking is opt-in: callers must explicitly call this method rather
// than the standard Analyze method.
func (a *Analyzer) AnalyzeWithBreach(items []*types.Item, hibpClient *HIBPClient) *HealthReport {
	report := a.Analyze(items)

	if hibpClient == nil {
		return report
	}

	// Collect passwords from login items.
	seen := make(map[string]bool)
	for _, item := range items {
		if item.Type != types.ItemTypeLogin {
			continue
		}
		pw := item.Fields[types.FieldPassword]
		if pw == "" || seen[item.ID] {
			continue
		}
		seen[item.ID] = true

		breached, count, err := hibpClient.CheckPassword(pw)
		if err != nil {
			// On error, skip this item rather than failing the whole report.
			continue
		}
		if !breached {
			continue
		}

		report.BreachedCount++
		recommendation := "This password has appeared in data breaches. Change it immediately and avoid reusing it elsewhere."
		report.BreachedPasswords = append(report.BreachedPasswords, BreachFinding{
			ItemID:          item.ID,
			ItemName:        item.Name,
			OccurrenceCount: count,
			Recommendation:  recommendation,
		})
		report.Findings = append(report.Findings, Finding{
			ItemID:   item.ID,
			ItemName: item.Name,
			Severity: SeverityCritical,
			Category: "breached",
			Message:  fmt.Sprintf("Password found in %d data breaches. Change immediately.", count),
		})
	}

	// Recalculate score with breach data.
	report.OverallScore = calculateScore(report)
	return report
}
