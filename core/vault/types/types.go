// Package types defines data structures and constants for vault items.
package types

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

// ItemType represents the category of a vault item.
type ItemType string

const (
	ItemTypeLogin      ItemType = "login"
	ItemTypeAPIKey     ItemType = "apikey"
	ItemTypeSSHKey     ItemType = "sshkey"
	ItemTypeSecureNote ItemType = "note"
	ItemTypeCreditCard ItemType = "creditcard"
	ItemTypeIdentity   ItemType = "identity"
	ItemTypePasskey    ItemType = "passkey"
	ItemTypeCustom     ItemType = "custom"
)

// ValidItemTypes is the set of all valid item types.
var ValidItemTypes = map[ItemType]bool{
	ItemTypeLogin:      true,
	ItemTypeAPIKey:     true,
	ItemTypeSSHKey:     true,
	ItemTypeSecureNote: true,
	ItemTypeCreditCard: true,
	ItemTypeIdentity:   true,
	ItemTypePasskey:    true,
	ItemTypeCustom:     true,
}

// Login field name constants.
const (
	FieldUsername = "username"
	FieldPassword = "password"
	FieldURL      = "url"
	FieldTOTP     = "totp"
)

// API key field name constants.
const (
	FieldAPIKey    = "api_key"
	FieldAPISecret = "api_secret"
	FieldEndpoint  = "endpoint"
)

// SSH key field name constants.
const (
	FieldPublicKey   = "public_key"
	FieldPrivateKey  = "private_key"
	FieldPassphrase  = "passphrase"
	FieldFingerprint = "fingerprint"
)

// Credit card field name constants.
const (
	FieldCardNumber = "card_number"
	FieldExpiry     = "expiry"
	FieldCVV        = "cvv"
	FieldCardHolder = "card_holder"
)

// Identity field name constants.
const (
	FieldFirstName = "first_name"
	FieldLastName  = "last_name"
	FieldEmail     = "email"
	FieldPhone     = "phone"
	FieldAddress   = "address"
)

// Passkey field name constants.
const (
	FieldCredentialID     = "credential_id"
	FieldPasskeyPublicKey = "passkey_public_key"
	FieldRelyingPartyID   = "rp_id"
	FieldUserHandle       = "user_handle"
	FieldSignCount        = "sign_count"
)

// Item represents a single vault entry.
type Item struct {
	ID             string            `json:"id"`
	Type           ItemType          `json:"type"`
	Name           string            `json:"name"`
	Fields         map[string]string `json:"fields"`
	Notes          string            `json:"notes"`
	Tags           []string          `json:"tags"`
	Favorite       bool              `json:"favorite"`
	CustomFields   map[string]string `json:"custom_fields"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	LastAccessedAt time.Time         `json:"last_accessed_at"`
	Version        int               `json:"version"`
}

// EncryptedItem is the on-disk format for an encrypted vault item.
type EncryptedItem struct {
	ID       string `json:"id"`
	Data     string `json:"data"` // base64-encoded encrypted JSON
	Version  int    `json:"version"`
	Checksum string `json:"checksum"` // SHA-256 of encrypted data
}

// ItemFilter specifies criteria for listing items.
type ItemFilter struct {
	Type        ItemType `json:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Favorite    *bool    `json:"favorite,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"`
	SortOrder   string   `json:"sort_order,omitempty"` // "asc" or "desc"
}

// SortBy constants.
const (
	SortByName         = "name"
	SortByCreatedAt    = "created_at"
	SortByUpdatedAt    = "updated_at"
	SortByType         = "type"
	SortByLastAccessed = "last_accessed_at"
)

// SortOrder constants.
const (
	SortAsc  = "asc"
	SortDesc = "desc"
)

// ValidateItemID ensures item IDs are safe to use as filenames and cannot
// escape the vault items directory.
func ValidateItemID(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("item ID must not be empty")
	}
	if strings.TrimSpace(id) != id {
		return errors.New("item ID must not contain leading or trailing whitespace")
	}
	if strings.Contains(id, "..") {
		return errors.New("item ID must not contain traversal sequences")
	}
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return errors.New("item ID contains unsupported characters")
	}
	return nil
}
