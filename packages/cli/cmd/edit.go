package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/types"
)

var editCmd = &cobra.Command{
	Use:   "edit <query>",
	Short: "Edit an existing item",
	Long:  `Find an item by name or ID and update its fields. Only specified fields are changed.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

var (
	editName     string
	editUsername string
	editPassword string
	editURL      string
	editTags     string
	editFavorite string
	editNotes    string
	// API key
	editAPIKey    string
	editAPISecret string
	// Credit card
	editCardNumber string
	editCardHolder string
	editCardExpiry string
	editCardCVV    string
	// Identity
	editFirstName string
	editLastName  string
	editEmail     string
	editPhone     string
	editAddress   string
)

func init() {
	editCmd.Flags().StringVar(&editName, "name", "", "update name")
	editCmd.Flags().StringVar(&editUsername, "username", "", "update username")
	editCmd.Flags().StringVar(&editPassword, "password", "", "update password")
	editCmd.Flags().StringVar(&editURL, "url", "", "update URL")
	editCmd.Flags().StringVar(&editTags, "tags", "", "update tags (comma-separated)")
	editCmd.Flags().StringVar(&editFavorite, "favorite", "", "set favorite (true/false)")
	editCmd.Flags().StringVar(&editNotes, "notes", "", "update notes")
	editCmd.Flags().StringVar(&editAPIKey, "key", "", "update API key")
	editCmd.Flags().StringVar(&editAPISecret, "secret", "", "update API secret")
	editCmd.Flags().StringVar(&editCardNumber, "card-number", "", "update card number")
	editCmd.Flags().StringVar(&editCardHolder, "card-holder", "", "update cardholder name")
	editCmd.Flags().StringVar(&editCardExpiry, "card-expiry", "", "update card expiry")
	editCmd.Flags().StringVar(&editCardCVV, "card-cvv", "", "update card CVV")
	editCmd.Flags().StringVar(&editFirstName, "first-name", "", "update first name")
	editCmd.Flags().StringVar(&editLastName, "last-name", "", "update last name")
	editCmd.Flags().StringVar(&editEmail, "email", "", "update email")
	editCmd.Flags().StringVar(&editPhone, "phone", "", "update phone")
	editCmd.Flags().StringVar(&editAddress, "address", "", "update address")
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	// Apply updates
	updated := false

	if editName != "" {
		itm.Name = editName
		updated = true
	}
	if editNotes != "" {
		itm.Notes = editNotes
		updated = true
	}
	if editTags != "" {
		itm.Tags = parseTags(editTags)
		updated = true
	}
	if editFavorite != "" {
		itm.Favorite = editFavorite == "true"
		updated = true
	}

	if itm.Fields == nil {
		itm.Fields = make(map[string]string)
	}

	fieldUpdates := map[string]string{
		types.FieldUsername:   editUsername,
		types.FieldPassword:  editPassword,
		types.FieldURL:       editURL,
		types.FieldAPIKey:    editAPIKey,
		types.FieldAPISecret: editAPISecret,
		types.FieldCardNumber: editCardNumber,
		types.FieldCardHolder: editCardHolder,
		types.FieldExpiry:     editCardExpiry,
		types.FieldCVV:        editCardCVV,
		types.FieldFirstName:  editFirstName,
		types.FieldLastName:   editLastName,
		types.FieldEmail:      editEmail,
		types.FieldPhone:      editPhone,
		types.FieldAddress:    editAddress,
	}

	for field, val := range fieldUpdates {
		if val != "" {
			itm.Fields[field] = val
			updated = true
		}
	}

	if !updated {
		return fmt.Errorf("no fields specified to update")
	}

	if err := mgr.UpdateItem(itm.ID, itm); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	if flagOutput == "json" {
		printItem(itm, false)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Item updated: %s (version %d)\n", itm.Name, itm.Version)
		printItem(itm, false)
	}
	return nil
}
