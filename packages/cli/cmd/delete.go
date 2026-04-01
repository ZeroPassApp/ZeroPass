package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <query>",
	Short: "Delete an item from the vault",
	Long:  `Find and delete an item by name or ID. Requires confirmation unless --force is used.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var deleteForce bool

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	itm, err := findItemByQuery(mgr, idx, args[0])
	if err != nil {
		return err
	}

	if !deleteForce {
		fmt.Fprintf(cmd.ErrOrStderr(), "Delete item %q (%s, type: %s)?\n", itm.Name, itm.ID, itm.Type)
		if !promptYesNo("Are you sure?") {
			fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
			return nil
		}
	}

	if err := mgr.DeleteItem(itm.ID); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	if flagOutput == "json" {
		formatOutput(map[string]string{"status": "deleted", "id": itm.ID, "name": itm.Name})
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Item deleted: %s (%s)\n", itm.Name, itm.ID)
	}
	return nil
}
