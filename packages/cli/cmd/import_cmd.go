package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/importexport"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/types"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import items from another password manager",
	Long:  `Import credentials from Chrome, Firefox, 1Password, Bitwarden, or a generic CSV file.`,
	RunE:  runImport,
}

var (
	importFrom string
	importFile string
)

func init() {
	importCmd.Flags().StringVar(&importFrom, "from", "", "source format: chrome|firefox|1password|bitwarden|csv")
	importCmd.Flags().StringVar(&importFile, "file", "", "input file path")
	_ = importCmd.MarkFlagRequired("from")
	_ = importCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
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

	f, err := os.Open(importFile)
	if err != nil {
		return fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	switch importFrom {
	case "chrome":
		imported, err := importexport.ImportChrome(f)
		if err != nil {
			return fmt.Errorf("import Chrome: %w", err)
		}
		return addImportedItems(cmd, mgr, imported)

	case "firefox":
		imported, err := importexport.ImportFirefox(f)
		if err != nil {
			return fmt.Errorf("import Firefox: %w", err)
		}
		return addImportedItems(cmd, mgr, imported)

	case "1password":
		imported, err := importexport.Import1Password(f)
		if err != nil {
			return fmt.Errorf("import 1Password: %w", err)
		}
		return addImportedItems(cmd, mgr, imported)

	case "bitwarden":
		imported, err := importexport.ImportBitwarden(f)
		if err != nil {
			return fmt.Errorf("import Bitwarden: %w", err)
		}
		return addImportedItems(cmd, mgr, imported)

	case "csv":
		imported, err := importexport.ImportCSV(f, importexport.CSVMapping{
			Name:     0,
			URL:      1,
			Username: 2,
			Password: 3,
			Notes:    4,
			Type:     -1,
			Skip:     1,
		})
		if err != nil {
			return fmt.Errorf("import CSV: %w", err)
		}
		return addImportedItems(cmd, mgr, imported)

	default:
		return fmt.Errorf("unknown import source: %s. Use chrome, firefox, 1password, bitwarden, or csv", importFrom)
	}
}

// addImportedItems adds imported items to the vault and prints a summary.
func addImportedItems(cmd *cobra.Command, mgr *item.Manager, imported []types.Item) error {
	if len(imported) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "No items found to import.")
		return nil
	}

	var added, failed int
	for i := range imported {
		itm := &imported[i]
		if err := mgr.AddItem(itm); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ Failed to import %q: %v\n", itm.Name, err)
			failed++
			continue
		}
		added++
	}

	if flagOutput == "json" {
		formatOutput(map[string]interface{}{
			"imported": added,
			"failed":   failed,
			"total":    len(imported),
		})
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Imported %d items", added)
		if failed > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), " (%d failed)", failed)
		}
		fmt.Fprintln(cmd.ErrOrStderr())
	}
	return nil
}
