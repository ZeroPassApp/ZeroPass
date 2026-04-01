package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/crypto/password"
	"github.com/zeropass/zeropass/core/vault/store"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long:  `Create a new ZeroPass vault with a master password and generate a recovery mnemonic.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath := getVaultPath()

	// Check if vault already exists
	if _, err := store.Open(vaultPath); err == nil {
		return fmt.Errorf("vault already exists at %s", vaultPath)
	}

	// Prompt for master password with confirmation
	pw, err := promptConfirmPassword()
	if err != nil {
		return err
	}
	defer zeroString(pw)

	// Check password strength
	gen := password.NewGenerator()
	strength, _ := gen.ScoreStrength(pw)
	if len(pw) < password.MinSecureLength {
		return fmt.Errorf("master password must be at least %d characters", password.MinSecureLength)
	}
	if strength.Score < password.MinSecureScore {
		return fmt.Errorf("master password is too weak (%s). Use a stronger password", strength.Feedback)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Password strength: %s\n", strength.Feedback)

	// Create vault
	cfg := store.DefaultConfig()
	v, result, err := store.Create(pw, vaultPath, cfg)
	if err != nil {
		return fmt.Errorf("create vault: %w", err)
	}
	defer v.Lock()

	// Display recovery mnemonic
	fmt.Fprintln(cmd.ErrOrStderr())
	fmt.Fprintln(cmd.ErrOrStderr(), "╔══════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(cmd.ErrOrStderr(), "║                    RECOVERY PHRASE                          ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "║                                                              ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "║  Write down these words in order. They are your ONLY way     ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "║  to recover your vault if you forget your master password.   ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "║                                                              ║")
	fmt.Fprintf(cmd.ErrOrStderr(), "║  %s\n", result.Mnemonic)
	fmt.Fprintln(cmd.ErrOrStderr(), "║                                                              ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "║  ⚠ NEVER share this phrase. Store it in a safe place.        ║")
	fmt.Fprintln(cmd.ErrOrStderr(), "╚══════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(cmd.ErrOrStderr())

	if !promptYesNo("Have you saved your recovery phrase?") {
		fmt.Fprintln(cmd.ErrOrStderr(), "⚠ Please save your recovery phrase before continuing.")
		fmt.Fprintln(cmd.ErrOrStderr(), "Your recovery phrase is shown above. Write it down NOW.")
		if !promptYesNo("I confirm I have saved my recovery phrase") {
			fmt.Fprintln(cmd.ErrOrStderr(), "Vault created but recovery phrase may not be saved.")
		}
	}

	if flagOutput == "json" {
		formatOutput(map[string]interface{}{
			"vault_path": vaultPath,
			"status":     "created",
		})
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Vault created at %s\n", vaultPath)
	}
	return nil
}
