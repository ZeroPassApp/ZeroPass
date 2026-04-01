// Package cmd implements the ZeroPass CLI commands using cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0-alpha"

var (
	flagVaultPath string
	flagOutput    string
	flagNoColor   bool
)

var rootCmd = &cobra.Command{
	Use:     "zeropass",
	Short:   "ZeroPass — developer-first zero-knowledge credential manager",
	Long:    `ZeroPass is a developer-first zero-knowledge credential manager that keeps your secrets safe with end-to-end encryption.`,
	Version: version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagVaultPath, "vault-path", "", "path to vault directory (default: ~/.zeropass/vaults/default)")
	rootCmd.PersistentFlags().StringVar(&flagOutput, "output", "text", "output format: text or json")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable colored output")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
