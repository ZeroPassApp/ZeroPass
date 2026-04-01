package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/crypto/key"
)

var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Manage recovery phrase",
	Long:  `View or validate the recovery mnemonic phrase for the vault.`,
	RunE:  runRecovery,
}

var (
	recoveryValidate string
)

func init() {
	recoveryCmd.Flags().StringVar(&recoveryValidate, "validate", "", "validate a mnemonic phrase")
	rootCmd.AddCommand(recoveryCmd)
}

func runRecovery(cmd *cobra.Command, args []string) error {
	if recoveryValidate != "" {
		mgr := key.NewRecoveryKeyManager()
		valid := mgr.ValidateMnemonic(recoveryValidate)
		if flagOutput == "json" {
			formatOutput(map[string]interface{}{
				"mnemonic": recoveryValidate,
				"valid":    valid,
			})
		} else {
			if valid {
				fmt.Fprintln(cmd.ErrOrStderr(), "✓ Mnemonic is valid.")
			} else {
				fmt.Fprintln(cmd.ErrOrStderr(), "✗ Mnemonic is invalid.")
			}
		}
		return nil
	}

	// Test unlock with recovery to verify it works
	vaultPath := getVaultPath()
	v, err := openVault(vaultPath)
	if err != nil {
		return err
	}

	mnemonic := promptLine("Enter recovery mnemonic to test: ")
	if mnemonic == "" {
		return fmt.Errorf("no mnemonic provided")
	}

	mgr := key.NewRecoveryKeyManager()
	if !mgr.ValidateMnemonic(mnemonic) {
		return fmt.Errorf("invalid mnemonic phrase")
	}

	if err := v.UnlockWithRecovery(mnemonic); err != nil {
		return fmt.Errorf("recovery phrase does not match this vault: %w", err)
	}
	defer v.Lock()

	if flagOutput == "json" {
		formatOutput(map[string]string{"status": "valid", "message": "Recovery phrase is valid for this vault"})
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "✓ Recovery phrase is valid for this vault.")
	}
	return nil
}
