package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/store"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the vault",
	Long:  `Lock the vault and clear all key material from memory.`,
	RunE:  runLock,
}

func init() {
	rootCmd.AddCommand(lockCmd)
}

func runLock(cmd *cobra.Command, args []string) error {
	vaultPath := getVaultPath()

	v, err := store.Open(vaultPath)
	if err != nil {
		return fmt.Errorf("open vault: %w", err)
	}

	v.Lock()

	if flagOutput == "json" {
		formatOutput(map[string]string{"status": "locked"})
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "✓ Vault locked.")
	}
	return nil
}
