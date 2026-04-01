package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/vault/types"
)

func loginItem(name, pw string, updatedAt time.Time) *types.Item {
	return &types.Item{
		ID:   "id-" + name,
		Type: types.ItemTypeLogin,
		Name: name,
		Fields: map[string]string{
			types.FieldUsername: "user",
			types.FieldPassword: pw,
		},
		UpdatedAt: updatedAt,
	}
}

func TestAnalyzeEmptyVault(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	report := a.Analyze(nil)
	assert.Equal(t, 0, report.TotalItems)
	assert.Equal(t, 100, report.OverallScore)
	assert.Empty(t, report.Findings)
}

func TestAnalyzeStrongPasswords(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	items := []*types.Item{
		loginItem("GitHub", "Xk9!mNpQ2#vL8@wR", now),
		loginItem("AWS", "Yz7$bTfH4&jS6*cE", now),
	}
	report := a.Analyze(items)
	assert.Equal(t, 2, report.LoginItems)
	assert.Equal(t, 0, report.WeakCount)
	assert.Equal(t, 0, report.ReusedCount)
	assert.Equal(t, 100, report.OverallScore)
}

func TestAnalyzeWeakPasswords(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	items := []*types.Item{
		loginItem("Bad1", "123456", now),
		loginItem("Bad2", "password", now),
	}
	report := a.Analyze(items)
	assert.Equal(t, 2, report.WeakCount)
	assert.True(t, report.OverallScore < 100)

	// Check findings exist for weak passwords.
	var weakFindings []Finding
	for _, f := range report.Findings {
		if f.Category == "weak" {
			weakFindings = append(weakFindings, f)
		}
	}
	require.Len(t, weakFindings, 2)
}

func TestAnalyzeReusedPasswords(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	shared := "Xk9!mNpQ2#vL8@wR"
	items := []*types.Item{
		loginItem("Site1", shared, now),
		loginItem("Site2", shared, now),
		loginItem("Site3", "UniqueP@ss99!xyz", now),
	}
	report := a.Analyze(items)
	assert.Equal(t, 2, report.ReusedCount)

	var reusedFindings []Finding
	for _, f := range report.Findings {
		if f.Category == "reused" {
			reusedFindings = append(reusedFindings, f)
		}
	}
	assert.Len(t, reusedFindings, 2)
}

func TestAnalyzeOldPasswords(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	old := time.Now().Add(-120 * 24 * time.Hour) // 120 days ago
	items := []*types.Item{
		loginItem("OldSite", "Xk9!mNpQ2#vL8@wR", old),
	}
	report := a.Analyze(items)
	assert.Equal(t, 1, report.OldCount)

	var oldFindings []Finding
	for _, f := range report.Findings {
		if f.Category == "old" {
			oldFindings = append(oldFindings, f)
		}
	}
	assert.Len(t, oldFindings, 1)
}

func TestAnalyzeNonLoginItems(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	items := []*types.Item{
		{
			ID:   "note-1",
			Type: types.ItemTypeSecureNote,
			Name: "My Note",
		},
	}
	report := a.Analyze(items)
	assert.Equal(t, 1, report.TotalItems)
	assert.Equal(t, 0, report.LoginItems)
	assert.Equal(t, 100, report.OverallScore)
}

func TestAnalyzeMixedIssues(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	items := []*types.Item{
		loginItem("Weak", "abc", now),
		loginItem("Old", "Xk9!mNpQ2#vL8@wR", old),
		loginItem("Good", "Yz7$bTfH4&jS6*cE", now),
	}
	report := a.Analyze(items)
	assert.Equal(t, 3, report.LoginItems)
	assert.True(t, report.WeakCount > 0)
	assert.True(t, report.OldCount > 0)
	assert.True(t, report.OverallScore < 100)
	assert.True(t, report.OverallScore > 0)
}

func TestSeverityLevels(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	items := []*types.Item{
		loginItem("VeryWeak", "a", now),
	}
	report := a.Analyze(items)
	require.NotEmpty(t, report.Findings)
	// Score ≤ 1 should be critical
	found := false
	for _, f := range report.Findings {
		if f.Category == "weak" && f.Severity == SeverityCritical {
			found = true
		}
	}
	assert.True(t, found, "expected critical severity for very weak password")
}

func TestAnalyzeEmptyPassword(t *testing.T) {
	a := NewAnalyzer(DefaultAnalyzerConfig())
	items := []*types.Item{
		loginItem("NoPass", "", time.Now()),
	}
	report := a.Analyze(items)
	// Should not crash, should have no weak findings for empty password
	assert.Equal(t, 0, report.WeakCount)
}
