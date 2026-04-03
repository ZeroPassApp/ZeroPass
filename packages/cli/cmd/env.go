package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/types"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments",
	Long: `Manage named environments for your secrets.
Environments are implemented as tags on items (e.g., "env:production").`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE:  runEnvList,
}

var envCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvCreate,
}

var envUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "List items in an environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvUse,
}

func init() {
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envCreateCmd)
	envCmd.AddCommand(envUseCmd)
	rootCmd.AddCommand(envCmd)
}

const envTagPrefix = "env:"

func runEnvList(cmd *cobra.Command, args []string) error {
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
		return err
	}

	// Collect unique environments
	envs := make(map[string]int) // env name -> item count
	for _, itm := range items {
		for _, tag := range itm.Tags {
			if strings.HasPrefix(tag, envTagPrefix) {
				envName := strings.TrimPrefix(tag, envTagPrefix)
				envs[envName]++
			}
		}
	}

	if flagOutput == "json" {
		envList := make([]map[string]interface{}, 0, len(envs))
		for name, count := range envs {
			envList = append(envList, map[string]interface{}{
				"name":       name,
				"item_count": count,
			})
		}
		formatOutput(envList)
		return nil
	}

	if len(envs) == 0 {
		fmt.Println("No environments found.")
		fmt.Printf("Create one with: %s\n", cliUsage("env create <name>"))
		return nil
	}

	fmt.Println("Environments:")
	for name, count := range envs {
		fmt.Printf("  • %s (%d items)\n", name, count)
	}
	return nil
}

func runEnvCreate(cmd *cobra.Command, args []string) error {
	envName := args[0]

	// Validate environment name
	if strings.ContainsAny(envName, " \t\n,:") {
		return fmt.Errorf("environment name cannot contain spaces, commas, or colons")
	}

	if flagOutput == "json" {
		formatOutput(map[string]string{
			"status":      "created",
			"environment": envName,
			"tag":         envTagPrefix + envName,
		})
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Environment %q created.\n", envName)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Tag items with %q to add them to this environment.\n", envTagPrefix+envName)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Example: %s\n", cliUsage(fmt.Sprintf("add --name myapi --type apikey --tags %s", envTagPrefix+envName)))
	}
	return nil
}

func runEnvUse(cmd *cobra.Command, args []string) error {
	envName := args[0]
	envTag := envTagPrefix + envName

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

	items, err := mgr.ListItems(types.ItemFilter{
		Tags: []string{envTag},
	})
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "No items found in environment %q.\n", envName)
		return nil
	}

	if flagOutput == "json" {
		printItemsTable(items)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "Items in environment %q:\n\n", envName)
		printItemsTable(items)
	}
	return nil
}
