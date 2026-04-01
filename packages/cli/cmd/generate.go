package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/crypto/password"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a random password or passphrase",
	Long:  `Generate a cryptographically secure random password or passphrase.`,
	RunE:  runGenerate,
}

var (
	genLength      int
	genNoSymbols   bool
	genNoDigits    bool
	genNoUppercase bool
	genPassphrase  bool
	genWords       int
	genSeparator   string
	genCopy        bool
)

func init() {
	generateCmd.Flags().IntVar(&genLength, "length", 32, "password length")
	generateCmd.Flags().BoolVar(&genNoSymbols, "no-symbols", false, "exclude symbols")
	generateCmd.Flags().BoolVar(&genNoDigits, "no-digits", false, "exclude digits")
	generateCmd.Flags().BoolVar(&genNoUppercase, "no-uppercase", false, "exclude uppercase")
	generateCmd.Flags().BoolVar(&genPassphrase, "passphrase", false, "generate a passphrase instead")
	generateCmd.Flags().IntVar(&genWords, "words", 4, "number of words in passphrase")
	generateCmd.Flags().StringVar(&genSeparator, "separator", "-", "word separator for passphrase")
	generateCmd.Flags().BoolVar(&genCopy, "copy", false, "copy to clipboard")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	gen := password.NewGenerator()

	var result string
	var err error

	if genPassphrase {
		result, err = gen.GeneratePassphrase(genWords, genSeparator)
		if err != nil {
			return fmt.Errorf("generate passphrase: %w", err)
		}
	} else {
		opts := password.GeneratorOptions{
			Lowercase: true,
			Uppercase: !genNoUppercase,
			Digits:    !genNoDigits,
			Symbols:   !genNoSymbols,
		}
		result, err = gen.GenerateRandom(genLength, opts)
		if err != nil {
			return fmt.Errorf("generate password: %w", err)
		}
	}

	// Score the generated password
	strength, _ := gen.ScoreStrength(result)

	if genCopy {
		if err := copyToClipboard(result, 30); err != nil {
			return err
		}
	}

	if flagOutput == "json" {
		formatOutput(map[string]interface{}{
			"password": result,
			"strength": map[string]interface{}{
				"score":    strength.Score,
				"feedback": strength.Feedback,
			},
		})
	} else {
		fmt.Println(result)
		fmt.Fprintf(cmd.ErrOrStderr(), "Strength: %s (score: %d/4)\n", strength.Feedback, strength.Score)
	}
	return nil
}
