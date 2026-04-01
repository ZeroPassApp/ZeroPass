package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/types"
)

var getCmd = &cobra.Command{
	Use:   "get <query>",
	Short: "Get an item from the vault",
	Long:  `Retrieve a credential by name or ID. Use --copy to copy the password to clipboard.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var (
	getCopy  bool
	getField string
	getJSON  bool
)

func init() {
	getCmd.Flags().BoolVar(&getCopy, "copy", false, "copy password/key to clipboard")
	getCmd.Flags().StringVar(&getField, "field", "", "specific field to retrieve")
	getCmd.Flags().BoolVar(&getJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
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

	// If specific field requested
	if getField != "" {
		val, ok := itm.Fields[getField]
		if !ok {
			val, ok = itm.CustomFields[getField]
		}
		if !ok {
			return fmt.Errorf("field %q not found in item %q", getField, itm.Name)
		}
		if getCopy {
			return copyToClipboard(val, v.Config().ClipboardClearSec)
		}
		if getJSON || flagOutput == "json" {
			formatOutput(map[string]string{"field": getField, "value": val})
		} else {
			fmt.Println(val)
		}
		return nil
	}

	// Copy default secret field
	if getCopy {
		secret := defaultSecretField(itm)
		if secret == "" {
			return fmt.Errorf("no secret field found to copy for item type %s", itm.Type)
		}
		return copyToClipboard(secret, v.Config().ClipboardClearSec)
	}

	// Print full item
	if getJSON {
		flagOutput = "json"
	}
	printItem(itm, true)
	return nil
}

// defaultSecretField returns the primary secret field value for a given item type.
func defaultSecretField(itm *types.Item) string {
	switch itm.Type {
	case types.ItemTypeLogin:
		return itm.Fields[types.FieldPassword]
	case types.ItemTypeAPIKey:
		if s := itm.Fields[types.FieldAPISecret]; s != "" {
			return s
		}
		return itm.Fields[types.FieldAPIKey]
	case types.ItemTypeSSHKey:
		return itm.Fields[types.FieldPrivateKey]
	case types.ItemTypeCreditCard:
		return itm.Fields[types.FieldCardNumber]
	case types.ItemTypeSecureNote:
		return itm.Notes
	default:
		// Try common fields
		if pw := itm.Fields[types.FieldPassword]; pw != "" {
			return pw
		}
		return ""
	}
}
