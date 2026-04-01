package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/crypto/password"
	"github.com/zeropass/zeropass/core/vault/types"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new item to the vault",
	Long:  `Add a new credential, API key, SSH key, note, credit card, or identity to the vault.`,
	RunE:  runAdd,
}

var (
	addType             string
	addName             string
	addUsername          string
	addPassword         string
	addURL              string
	addTags             string
	addFavorite         bool
	addGeneratePassword bool
	addNotes            string
	// API key
	addAPIKey    string
	addAPISecret string
	// SSH key
	addPrivateKeyFile string
	addPublicKeyFile  string
	// Credit card
	addCardNumber string
	addCardHolder string
	addCardExpiry string
	addCardCVV    string
	// Identity
	addFirstName string
	addLastName  string
	addEmail     string
	addPhone     string
	addAddress   string
)

func init() {
	addCmd.Flags().StringVar(&addType, "type", "login", "item type: login|apikey|sshkey|note|creditcard|identity|custom")
	addCmd.Flags().StringVar(&addName, "name", "", "item name")
	addCmd.Flags().StringVar(&addUsername, "username", "", "username")
	addCmd.Flags().StringVar(&addPassword, "password", "", "password (will prompt if not provided)")
	addCmd.Flags().StringVar(&addURL, "url", "", "URL")
	addCmd.Flags().StringVar(&addTags, "tags", "", "comma-separated tags")
	addCmd.Flags().BoolVar(&addFavorite, "favorite", false, "mark as favorite")
	addCmd.Flags().BoolVar(&addGeneratePassword, "generate-password", false, "auto-generate password")
	addCmd.Flags().StringVar(&addNotes, "notes", "", "notes")
	// API key
	addCmd.Flags().StringVar(&addAPIKey, "key", "", "API key")
	addCmd.Flags().StringVar(&addAPISecret, "secret", "", "API secret")
	// SSH key
	addCmd.Flags().StringVar(&addPrivateKeyFile, "private-key-file", "", "path to SSH private key file")
	addCmd.Flags().StringVar(&addPublicKeyFile, "public-key-file", "", "path to SSH public key file")
	// Credit card
	addCmd.Flags().StringVar(&addCardNumber, "card-number", "", "credit card number")
	addCmd.Flags().StringVar(&addCardHolder, "card-holder", "", "cardholder name")
	addCmd.Flags().StringVar(&addCardExpiry, "card-expiry", "", "card expiry (MM/YY)")
	addCmd.Flags().StringVar(&addCardCVV, "card-cvv", "", "card CVV")
	// Identity
	addCmd.Flags().StringVar(&addFirstName, "first-name", "", "first name")
	addCmd.Flags().StringVar(&addLastName, "last-name", "", "last name")
	addCmd.Flags().StringVar(&addEmail, "email", "", "email address")
	addCmd.Flags().StringVar(&addPhone, "phone", "", "phone number")
	addCmd.Flags().StringVar(&addAddress, "address", "", "address")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	itemType := types.ItemType(addType)
	if !types.ValidItemTypes[itemType] {
		return fmt.Errorf("invalid item type: %s. Valid types: login, apikey, sshkey, note, creditcard, identity, custom", addType)
	}

	// Prompt for name if not provided
	name := addName
	if name == "" {
		name = promptLine("Item name: ")
		if name == "" {
			return fmt.Errorf("item name is required")
		}
	}

	fields := make(map[string]string)

	switch itemType {
	case types.ItemTypeLogin:
		if err := buildLoginFields(cmd, fields); err != nil {
			return err
		}
	case types.ItemTypeAPIKey:
		buildAPIKeyFields(fields)
	case types.ItemTypeSSHKey:
		if err := buildSSHKeyFields(fields); err != nil {
			return err
		}
	case types.ItemTypeCreditCard:
		buildCreditCardFields(fields)
	case types.ItemTypeIdentity:
		buildIdentityFields(fields)
	case types.ItemTypeSecureNote:
		// notes handled below
	case types.ItemTypeCustom:
		// custom fields handled below
	}

	notes := addNotes
	if notes == "" && itemType == types.ItemTypeSecureNote {
		notes = promptLine("Note content: ")
	}

	itm := &types.Item{
		Name:     name,
		Type:     itemType,
		Fields:   fields,
		Notes:    notes,
		Tags:     parseTags(addTags),
		Favorite: addFavorite,
	}

	if err := mgr.AddItem(itm); err != nil {
		return fmt.Errorf("add item: %w", err)
	}

	if flagOutput == "json" {
		printItem(itm, false)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✓ Item added: %s (%s)\n", itm.Name, itm.ID)
		printItem(itm, false)
	}
	return nil
}

func buildLoginFields(cmd *cobra.Command, fields map[string]string) error {
	username := addUsername
	if username == "" {
		username = promptLine("Username: ")
	}
	if username != "" {
		fields[types.FieldUsername] = username
	}

	if addURL != "" {
		fields[types.FieldURL] = addURL
	}

	if addGeneratePassword {
		gen := password.NewGenerator()
		pw, err := gen.GenerateRandom(32, password.GeneratorOptions{
			Uppercase: true,
			Lowercase: true,
			Digits:    true,
			Symbols:   true,
		})
		if err != nil {
			return fmt.Errorf("generate password: %w", err)
		}
		fields[types.FieldPassword] = pw
		fmt.Fprintf(cmd.ErrOrStderr(), "Generated password: %s\n", pw)
	} else {
		pw := addPassword
		if pw == "" {
			var err error
			pw, err = promptPassword("Password: ")
			if err != nil {
				return err
			}
		}
		if pw != "" {
			fields[types.FieldPassword] = pw
		}
	}
	return nil
}

func buildAPIKeyFields(fields map[string]string) {
	key := addAPIKey
	if key == "" {
		key = promptLine("API key: ")
	}
	if key != "" {
		fields[types.FieldAPIKey] = key
	}

	secret := addAPISecret
	if secret == "" {
		secret = promptLine("API secret: ")
	}
	if secret != "" {
		fields[types.FieldAPISecret] = secret
	}

	if addURL != "" {
		fields[types.FieldEndpoint] = addURL
	}
}

func buildSSHKeyFields(fields map[string]string) error {
	if addPrivateKeyFile != "" {
		data, err := os.ReadFile(addPrivateKeyFile)
		if err != nil {
			return fmt.Errorf("read private key file: %w", err)
		}
		fields[types.FieldPrivateKey] = string(data)
	}

	if addPublicKeyFile != "" {
		data, err := os.ReadFile(addPublicKeyFile)
		if err != nil {
			return fmt.Errorf("read public key file: %w", err)
		}
		fields[types.FieldPublicKey] = string(data)
	}

	return nil
}

func buildCreditCardFields(fields map[string]string) {
	num := addCardNumber
	if num == "" {
		num = promptLine("Card number: ")
	}
	if num != "" {
		fields[types.FieldCardNumber] = num
	}

	holder := addCardHolder
	if holder == "" {
		holder = promptLine("Cardholder name: ")
	}
	if holder != "" {
		fields[types.FieldCardHolder] = holder
	}

	expiry := addCardExpiry
	if expiry == "" {
		expiry = promptLine("Expiry (MM/YY): ")
	}
	if expiry != "" {
		fields[types.FieldExpiry] = expiry
	}

	cvv := addCardCVV
	if cvv == "" {
		cvv = promptLine("CVV: ")
	}
	if cvv != "" {
		fields[types.FieldCVV] = cvv
	}
}

func buildIdentityFields(fields map[string]string) {
	if addFirstName != "" {
		fields[types.FieldFirstName] = addFirstName
	}
	if addLastName != "" {
		fields[types.FieldLastName] = addLastName
	}
	if addEmail != "" {
		fields[types.FieldEmail] = addEmail
	}
	if addPhone != "" {
		fields[types.FieldPhone] = addPhone
	}
	if addAddress != "" {
		fields[types.FieldAddress] = addAddress
	}
}
