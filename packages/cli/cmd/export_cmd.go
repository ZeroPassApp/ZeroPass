package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/importexport"
	"github.com/zeropass/zeropass/core/vault/types"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export vault items",
	Long:  `Export vault items in JSON, CSV, or encrypted format.`,
	RunE:  runExport,
}

var (
	exportFormat     string
	exportOutputFile string
	exportForce      bool
)

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "export format: json|csv|encrypted")
	exportCmd.Flags().StringVar(&exportOutputFile, "file", "", "output file path (default: stdout)")
	exportCmd.Flags().BoolVar(&exportForce, "force", false, "skip plaintext export warning")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
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

	// Convert []*types.Item to []types.Item for export functions
	plainItems := make([]types.Item, len(items))
	for i, itm := range items {
		plainItems[i] = *itm
	}

	// Warn before plaintext export
	if exportFormat != "encrypted" && !exportForce {
		fmt.Fprintln(cmd.ErrOrStderr(), "⚠ WARNING: This exports your vault in PLAINTEXT.")
		fmt.Fprintln(cmd.ErrOrStderr(), "  Anyone with this file can read ALL your secrets.")
		if !promptYesNo("Continue?") {
			return fmt.Errorf("export cancelled")
		}
	}

	// Determine output writer
	writer := os.Stdout
	if exportOutputFile != "" {
		f, err := os.Create(exportOutputFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	switch exportFormat {
	case "json":
		if err := importexport.ExportJSON(plainItems, writer); err != nil {
			return fmt.Errorf("export JSON: %w", err)
		}
	case "csv":
		if err := importexport.ExportCSV(plainItems, writer); err != nil {
			return fmt.Errorf("export CSV: %w", err)
		}
	case "encrypted":
		vaultKey, err := v.VaultKey()
		if err != nil {
			return fmt.Errorf("get vault key: %w", err)
		}
		if err := importexport.ExportEncrypted(plainItems, vaultKey, writer); err != nil {
			return fmt.Errorf("export encrypted: %w", err)
		}
	default:
		return fmt.Errorf("unknown export format: %s. Use json, csv, or encrypted", exportFormat)
	}

	if exportOutputFile != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Exported %d items to %s (%s)\n", len(items), exportOutputFile, exportFormat)
	}
	return nil
}
