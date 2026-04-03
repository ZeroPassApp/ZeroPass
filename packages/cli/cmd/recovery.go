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
	recoveryValidate   string
	recoveryRegenerate bool
)

func init() {
	recoveryCmd.Flags().StringVar(&recoveryValidate, "validate", "", "validate a mnemonic phrase")
	recoveryCmd.Flags().BoolVar(&recoveryRegenerate, "regenerate", false, "generate new recovery phrase (requires vault unlock)")
	rootCmd.AddCommand(recoveryCmd)
}

func runRecovery(cmd *cobra.Command, args []string) error {
	if recoveryRegenerate {
		v, err := openAndUnlockVault()
		if err != nil {
			return err
		}
		defer v.Lock()

		mnemonic, err := v.RegenerateRecovery()
		if err != nil {
			return fmt.Errorf("regenerate recovery: %w", err)
		}

		if flagOutput == "json" {
			formatOutput(map[string]string{
				"status":   "regenerated",
				"mnemonic": mnemonic,
			})
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: Your previous recovery phrase is now INVALID.")
			fmt.Fprintln(cmd.ErrOrStderr(), "")
			fmt.Fprintln(cmd.ErrOrStderr(), "New recovery phrase:")
			fmt.Fprintln(cmd.ErrOrStderr(), "  "+mnemonic)
			fmt.Fprintln(cmd.ErrOrStderr(), "")
			fmt.Fprintln(cmd.ErrOrStderr(), "Store this phrase safely. You will need it if you forget your master password.")
		}
		return nil
	}

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

	if err := v.ValidateRecovery(mnemonic); err != nil {
		return fmt.Errorf("recovery phrase does not match this vault: %w", err)
	}

	if flagOutput == "json" {
		formatOutput(map[string]string{"status": "valid", "message": "Recovery phrase is valid for this vault"})
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "✓ Recovery phrase is valid for this vault.")
	}
	return nil
}
