package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/types"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items in the vault",
	Long:  `List all vault items with optional filters for type, tag, favorite status, and sort order.`,
	RunE:  runList,
}

var (
	listType     string
	listTag      string
	listFavorite bool
	listSort     string
	listJSON     bool
)

func init() {
	listCmd.Flags().StringVar(&listType, "type", "", "filter by item type")
	listCmd.Flags().StringVar(&listTag, "tag", "", "filter by tag")
	listCmd.Flags().BoolVar(&listFavorite, "favorite", false, "show only favorites")
	listCmd.Flags().StringVar(&listSort, "sort", "name", "sort by: name|date|type")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	filter := types.ItemFilter{}
	if listType != "" {
		filter.Type = types.ItemType(listType)
	}
	if listTag != "" {
		filter.Tags = []string{listTag}
	}
	if listFavorite {
		fav := true
		filter.Favorite = &fav
	}

	switch listSort {
	case "date":
		filter.SortBy = types.SortByUpdatedAt
	case "type":
		filter.SortBy = types.SortByType
	default:
		filter.SortBy = types.SortByName
	}

	items, err := mgr.ListItems(filter)
	if err != nil {
		return err
	}

	if listJSON {
		flagOutput = "json"
	}

	printItemsTable(items)
	return nil
}
