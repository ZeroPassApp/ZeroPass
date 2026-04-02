package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItemTypeConstants(t *testing.T) {
	assert.Equal(t, ItemType("login"), ItemTypeLogin)
	assert.Equal(t, ItemType("apikey"), ItemTypeAPIKey)
	assert.Equal(t, ItemType("sshkey"), ItemTypeSSHKey)
	assert.Equal(t, ItemType("note"), ItemTypeSecureNote)
	assert.Equal(t, ItemType("creditcard"), ItemTypeCreditCard)
	assert.Equal(t, ItemType("identity"), ItemTypeIdentity)
	assert.Equal(t, ItemType("passkey"), ItemTypePasskey)
	assert.Equal(t, ItemType("custom"), ItemTypeCustom)
}

func TestValidItemTypes(t *testing.T) {
	for _, it := range []ItemType{
		ItemTypeLogin, ItemTypeAPIKey, ItemTypeSSHKey,
		ItemTypeSecureNote, ItemTypeCreditCard, ItemTypeIdentity,
		ItemTypePasskey, ItemTypeCustom,
	} {
		assert.True(t, ValidItemTypes[it], "expected %s to be valid", it)
	}
	assert.False(t, ValidItemTypes["invalid"])
}

func TestFieldConstants(t *testing.T) {
	assert.Equal(t, "username", FieldUsername)
	assert.Equal(t, "password", FieldPassword)
	assert.Equal(t, "url", FieldURL)
	assert.Equal(t, "api_key", FieldAPIKey)
	assert.Equal(t, "public_key", FieldPublicKey)
	assert.Equal(t, "private_key", FieldPrivateKey)
	assert.Equal(t, "card_number", FieldCardNumber)
	assert.Equal(t, "first_name", FieldFirstName)
}

func TestItemStruct(t *testing.T) {
	now := time.Now()
	item := Item{
		ID:       "test-id",
		Type:     ItemTypeLogin,
		Name:     "My Login",
		Fields:   map[string]string{FieldUsername: "user", FieldPassword: "pass"},
		Notes:    "some notes",
		Tags:     []string{"work", "dev"},
		Favorite: true,
		CustomFields: map[string]string{"env": "production"},
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	assert.Equal(t, "test-id", item.ID)
	assert.Equal(t, ItemTypeLogin, item.Type)
	assert.Equal(t, "My Login", item.Name)
	assert.Equal(t, "user", item.Fields[FieldUsername])
	assert.Equal(t, "pass", item.Fields[FieldPassword])
	assert.Equal(t, "some notes", item.Notes)
	assert.Equal(t, []string{"work", "dev"}, item.Tags)
	assert.True(t, item.Favorite)
	assert.Equal(t, "production", item.CustomFields["env"])
	assert.Equal(t, 1, item.Version)
}

func TestEncryptedItemStruct(t *testing.T) {
	ei := EncryptedItem{
		ID:       "uuid-1",
		Data:     "base64data",
		Version:  2,
		Checksum: "sha256hash",
	}
	assert.Equal(t, "uuid-1", ei.ID)
	assert.Equal(t, "base64data", ei.Data)
	assert.Equal(t, 2, ei.Version)
	assert.Equal(t, "sha256hash", ei.Checksum)
}

func TestItemFilter(t *testing.T) {
	fav := true
	f := ItemFilter{
		Type:        ItemTypeLogin,
		Tags:        []string{"work"},
		Favorite:    &fav,
		SearchQuery: "github",
		SortBy:      SortByName,
		SortOrder:   SortAsc,
	}
	assert.Equal(t, ItemTypeLogin, f.Type)
	assert.Equal(t, []string{"work"}, f.Tags)
	assert.NotNil(t, f.Favorite)
	assert.True(t, *f.Favorite)
	assert.Equal(t, "github", f.SearchQuery)
}

func TestPasskeyFieldConstants(t *testing.T) {
	assert.Equal(t, "credential_id", FieldCredentialID)
	assert.Equal(t, "passkey_public_key", FieldPasskeyPublicKey)
	assert.Equal(t, "rp_id", FieldRelyingPartyID)
	assert.Equal(t, "user_handle", FieldUserHandle)
	assert.Equal(t, "sign_count", FieldSignCount)
}

func TestSortByLastAccessedConstant(t *testing.T) {
	assert.Equal(t, "last_accessed_at", SortByLastAccessed)
}

func TestItemLastAccessedAt(t *testing.T) {
	now := time.Now()
	item := Item{
		ID:             "test",
		Type:           ItemTypeLogin,
		Name:           "Test",
		LastAccessedAt: now,
	}
	assert.Equal(t, now, item.LastAccessedAt)
}

func TestPasskeyItemType(t *testing.T) {
	item := Item{
		ID:   "passkey-1",
		Type: ItemTypePasskey,
		Name: "My Passkey",
		Fields: map[string]string{
			FieldCredentialID:    "cred-123",
			FieldPasskeyPublicKey: "pk-456",
			FieldRelyingPartyID:  "example.com",
			FieldUserHandle:      "user-789",
			FieldSignCount:       "42",
		},
	}
	assert.Equal(t, ItemTypePasskey, item.Type)
	assert.True(t, ValidItemTypes[item.Type])
	assert.Equal(t, "cred-123", item.Fields[FieldCredentialID])
	assert.Equal(t, "example.com", item.Fields[FieldRelyingPartyID])
}
