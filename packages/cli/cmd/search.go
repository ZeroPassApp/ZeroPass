package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/types"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for items in the vault",
	Long:  `Full-text search across all item fields. Results are displayed as a table.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

var (
	searchType string
	searchTag  string
)

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "", "filter by item type")
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "filter by tag")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
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

	// Search index
	ids, err := idx.Search(args[0])
	if err != nil {
		return err
	}

	var items []*types.Item
	for _, id := range ids {
		itm, e := mgr.GetItem(id)
		if e != nil {
			continue
		}
		// Apply filters
		if searchType != "" && string(itm.Type) != searchType {
			continue
		}
		if searchTag != "" {
			found := false
			for _, t := range itm.Tags {
				if t == searchTag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		items = append(items, itm)
	}

	printItemsTable(items)
	return nil
}
