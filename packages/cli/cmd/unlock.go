package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/store"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock the vault",
	Long:  `Unlock the vault with your master password or recovery phrase.`,
	RunE:  runUnlock,
}

var unlockRecovery bool

func init() {
	unlockCmd.Flags().BoolVar(&unlockRecovery, "recovery", false, "unlock using recovery mnemonic")
	rootCmd.AddCommand(unlockCmd)
}

func runUnlock(cmd *cobra.Command, args []string) error {
	vaultPath := getVaultPath()

	v, err := openVault(vaultPath)
	if err != nil {
		return err
	}

	if unlockRecovery {
		mnemonic := promptLine("Enter recovery mnemonic: ")
		if err := v.UnlockWithRecovery(mnemonic); err != nil {
			return fmt.Errorf("unlock with recovery: %w", err)
		}
	} else {
		pw, err := promptPassword("Enter master password: ")
		if err != nil {
			return err
		}
		defer zeroString(pw)

		if err := v.Unlock(pw); err != nil {
			return fmt.Errorf("incorrect master password")
		}
	}

	if flagOutput == "json" {
		formatOutput(map[string]string{"status": "unlocked"})
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "✓ Vault unlocked.")
	}
	return nil
}

// openVault opens a vault without unlocking it.
func openVault(vaultPath string) (*store.Vault, error) {
	v, err := store.Open(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("open vault: %w (run 'zeropass init' to create a vault)", err)
	}
	return v, nil
}
