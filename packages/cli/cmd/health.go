package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/health"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Analyze password health",
	Long:  `Run a security analysis on all vault items, checking for weak, reused, and old passwords.`,
	RunE:  runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	v, err := openAndUnlockVault()
	if err != nil {
		return err
	}
	defer v.Lock()

	mgr, idx, err := newItemManager(v)
	if err != nil {
		return err
	}
	defer idx.Close()

	items, err := mgr.AllItems()
	if err != nil {
		return fmt.Errorf("list items: %w", err)
	}

	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(items)

	if flagOutput == "json" {
		formatOutput(report)
		return nil
	}

	// Human-readable output
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║         PASSWORD HEALTH REPORT       ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Overall Score: %s\n", scoreDisplay(report.OverallScore))
	fmt.Printf("Total Items:   %d\n", report.TotalItems)
	fmt.Printf("Login Items:   %d\n", report.LoginItems)
	fmt.Println()

	if report.WeakCount > 0 {
		fmt.Printf("%s Weak passwords:   %d\n", severityIcon(health.SeverityCritical), report.WeakCount)
	}
	if report.ReusedCount > 0 {
		fmt.Printf("%s Reused passwords: %d\n", severityIcon(health.SeverityWarning), report.ReusedCount)
	}
	if report.OldCount > 0 {
		fmt.Printf("%s Old passwords:    %d\n", severityIcon(health.SeverityInfo), report.OldCount)
	}

	if len(report.Findings) == 0 {
		fmt.Println("\n✓ No issues found. All passwords are healthy!")
		return nil
	}

	fmt.Println()
	fmt.Println("Findings:")
	fmt.Println(strings.Repeat("─", 60))

	for _, f := range report.Findings {
		icon := severityIcon(f.Severity)
		fmt.Printf("  %s [%s] %s\n", icon, f.Category, f.ItemName)
		fmt.Printf("    %s\n", f.Message)
	}

	return nil
}

func scoreDisplay(score int) string {
	switch {
	case score >= 90:
		return fmt.Sprintf("%d/100 ✓ Excellent", score)
	case score >= 70:
		return fmt.Sprintf("%d/100 Good", score)
	case score >= 50:
		return fmt.Sprintf("%d/100 ⚠ Fair", score)
	default:
		return fmt.Sprintf("%d/100 ✗ Poor", score)
	}
}

func severityIcon(sev health.Severity) string {
	switch sev {
	case health.SeverityCritical:
		return "🔴"
	case health.SeverityWarning:
		return "🟡"
	case health.SeverityInfo:
		return "🔵"
	default:
		return "⚪"
	}
}
