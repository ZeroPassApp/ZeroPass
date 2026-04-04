// Package item provides CRUD operations for vault items, handling encryption,
// disk persistence, and integration with the search index.
package item

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/crypto/cipher"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/types"
)

// generateID produces a new UUID-like item ID.
func generateID() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Indexer is the interface that the search index must satisfy.
type Indexer interface {
	AddToIndex(item *types.Item) error
	UpdateIndex(item *types.Item) error
	RemoveFromIndex(id string) error
}

// Manager manages item CRUD against a vault directory.
type Manager struct {
	itemsPath string
	vaultKey  func() ([]byte, error)
	indexer   Indexer // may be nil
}

// NewManager creates an item manager.
// vaultKeyFn returns the current vault key; it must error if the vault is locked.
func NewManager(itemsPath string, vaultKeyFn func() ([]byte, error), indexer Indexer) *Manager {
	return &Manager{
		itemsPath: itemsPath,
		vaultKey:  vaultKeyFn,
		indexer:   indexer,
	}
}

// AddItem validates, encrypts, and persists a new item.
func (m *Manager) AddItem(item *types.Item) error {
	if item == nil {
		return errors.New("item must not be nil")
	}
	normalizeItem(item)

	// Auto-detect item type if empty.
	if item.Type == "" {
		item.Type = AutoCategorize(item.Fields)
	}

	if err := validateItem(item); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = generateID()
	}
	if err := types.ValidateItemID(item.ID); err != nil {
		return fmt.Errorf("validate item ID: %w", err)
	}
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Version = 1

	if err := m.encryptAndSave(item); err != nil {
		return err
	}

	if m.indexer != nil {
		if err := m.indexer.AddToIndex(item); err != nil {
			return fmt.Errorf("index add: %w", err)
		}
	}
	return nil
}

// GetItem reads and decrypts an item from disk.
// It also updates the LastAccessedAt timestamp (best-effort, does not fail on save error).
func (m *Manager) GetItem(id string) (*types.Item, error) {
	if err := types.ValidateItemID(id); err != nil {
		return nil, err
	}
	item, err := m.readAndDecrypt(id)
	if err != nil {
		return nil, err
	}

	// Update last accessed timestamp (best-effort).
	item.LastAccessedAt = time.Now().UTC()
	_ = m.encryptAndSave(item) // save without incrementing version; ignore errors

	return item, nil
}

// UpdateItem increments the version, re-encrypts, and saves.
func (m *Manager) UpdateItem(id string, item *types.Item) error {
	if err := types.ValidateItemID(id); err != nil {
		return err
	}
	if item == nil {
		return errors.New("item must not be nil")
	}
	normalizeItem(item)
	if err := validateItem(item); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	item.ID = id
	item.UpdatedAt = time.Now().UTC()
	item.Version++

	if err := m.encryptAndSave(item); err != nil {
		return err
	}

	if m.indexer != nil {
		if err := m.indexer.UpdateIndex(item); err != nil {
			return fmt.Errorf("index update: %w", err)
		}
	}
	return nil
}

// DeleteItem removes an item file and its version history from disk.
func (m *Manager) DeleteItem(id string) error {
	if err := types.ValidateItemID(id); err != nil {
		return err
	}

	fp := m.itemFilePath(id)
	if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove item: %w", err)
	}

	// Remove version history file if present.
	vfp := m.versionFilePath(id)
	_ = os.Remove(vfp)

	if m.indexer != nil {
		if err := m.indexer.RemoveFromIndex(id); err != nil {
			return fmt.Errorf("index remove: %w", err)
		}
	}
	return nil
}

// ListItems returns items matching the given filter.
func (m *Manager) ListItems(filter types.ItemFilter) ([]*types.Item, error) {
	entries, err := os.ReadDir(m.itemsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*types.Item{}, nil
		}
		return nil, fmt.Errorf("read items dir: %w", err)
	}

	items := make([]*types.Item, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".versions.json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if err := types.ValidateItemID(id); err != nil {
			continue
		}
		item, err := m.readAndDecrypt(id)
		if err != nil {
			continue // skip unreadable items
		}
		if matchesFilter(item, filter) {
			items = append(items, item)
		}
	}

	sortItems(items, filter.SortBy, filter.SortOrder)
	return items, nil
}

// AllItems returns all items (unfiltered).
func (m *Manager) AllItems() ([]*types.Item, error) {
	return m.ListItems(types.ItemFilter{})
}

// --- encryption / persistence ---

func (m *Manager) encryptAndSave(item *types.Item) error {
	if err := types.ValidateItemID(item.ID); err != nil {
		return fmt.Errorf("validate item ID: %w", err)
	}

	vk, err := m.vaultKey()
	if err != nil {
		return fmt.Errorf("get vault key: %w", err)
	}

	itemKey, err := key.DeriveItemKey(vk, item.ID)
	if err != nil {
		return fmt.Errorf("derive item key: %w", err)
	}
	defer crypto.ZeroBytes(itemKey)

	plaintext, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	c := cipher.New()
	ciphertext, err := c.Encrypt(plaintext, itemKey)
	if err != nil {
		return fmt.Errorf("encrypt item: %w", err)
	}

	dataB64 := encoding.Base64StdEncode(ciphertext)
	checksum := fmt.Sprintf("%x", sha256.Sum256(ciphertext))

	enc := types.EncryptedItem{
		ID:       item.ID,
		Data:     dataB64,
		Version:  item.Version,
		Checksum: checksum,
	}

	data, err := json.MarshalIndent(enc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal encrypted item: %w", err)
	}

	if err := os.MkdirAll(m.itemsPath, 0700); err != nil {
		return fmt.Errorf("ensure items dir: %w", err)
	}

	fp := m.itemFilePath(item.ID)
	if err := os.WriteFile(fp, data, 0600); err != nil {
		return fmt.Errorf("write item: %w", err)
	}
	return nil
}

func (m *Manager) readAndDecrypt(id string) (*types.Item, error) {
	if err := types.ValidateItemID(id); err != nil {
		return nil, err
	}

	fp := m.itemFilePath(id)
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("read item file: %w", err)
	}

	var enc types.EncryptedItem
	if err := json.Unmarshal(data, &enc); err != nil {
		return nil, fmt.Errorf("unmarshal encrypted item: %w", err)
	}

	// Verify checksum.
	ciphertext, err := encoding.Base64StdDecode(enc.Data)
	if err != nil {
		return nil, fmt.Errorf("decode item data: %w", err)
	}
	checksum := fmt.Sprintf("%x", sha256.Sum256(ciphertext))
	if checksum != enc.Checksum {
		return nil, errors.New("item checksum mismatch — data may be corrupted")
	}

	vk, err := m.vaultKey()
	if err != nil {
		return nil, fmt.Errorf("get vault key: %w", err)
	}

	itemKey, err := key.DeriveItemKey(vk, id)
	if err != nil {
		return nil, fmt.Errorf("derive item key: %w", err)
	}
	defer crypto.ZeroBytes(itemKey)

	c := cipher.New()
	plaintext, err := c.Decrypt(ciphertext, itemKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt item: %w", err)
	}

	var item types.Item
	if err := json.Unmarshal(plaintext, &item); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}
	normalizeItem(&item)
	return &item, nil
}

func (m *Manager) itemFilePath(id string) string {
	return filepath.Join(m.itemsPath, id+".json")
}

func (m *Manager) versionFilePath(id string) string {
	return filepath.Join(m.itemsPath, id+".versions.json")
}

// --- validation ---

func normalizeItem(item *types.Item) {
	if item.Fields == nil {
		item.Fields = map[string]string{}
	}
	if item.CustomFields == nil {
		item.CustomFields = map[string]string{}
	}
	if item.Tags == nil {
		item.Tags = []string{}
	}
}

func validateItem(item *types.Item) error {
	if item.Name == "" {
		return errors.New("item name must not be empty")
	}
	if item.Type == "" {
		return errors.New("item type must not be empty")
	}
	if !types.ValidItemTypes[item.Type] {
		return fmt.Errorf("invalid item type: %s", item.Type)
	}
	return nil
}

// --- filtering ---

func matchesFilter(item *types.Item, f types.ItemFilter) bool {
	if f.Type != "" && item.Type != f.Type {
		return false
	}
	if f.Favorite != nil && item.Favorite != *f.Favorite {
		return false
	}
	if len(f.Tags) > 0 && !hasAnyTag(item.Tags, f.Tags) {
		return false
	}
	if f.SearchQuery != "" {
		q := strings.ToLower(f.SearchQuery)
		if !strings.Contains(strings.ToLower(item.Name), q) &&
			!strings.Contains(strings.ToLower(item.Notes), q) &&
			!containsFieldValue(item.Fields, q) &&
			!containsFieldValue(item.CustomFields, q) {
			return false
		}
	}
	return true
}

func hasAnyTag(itemTags, filterTags []string) bool {
	set := make(map[string]bool, len(itemTags))
	for _, t := range itemTags {
		set[t] = true
	}
	for _, t := range filterTags {
		if set[t] {
			return true
		}
	}
	return false
}

func containsFieldValue(fields map[string]string, q string) bool {
	for _, v := range fields {
		if strings.Contains(strings.ToLower(v), q) {
			return true
		}
	}
	return false
}

// --- sorting ---

func sortItems(items []*types.Item, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = types.SortByName
	}
	desc := strings.ToLower(sortOrder) == types.SortDesc

	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case types.SortByCreatedAt:
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		case types.SortByUpdatedAt:
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		case types.SortByType:
			less = items[i].Type < items[j].Type
		case types.SortByLastAccessed:
			less = items[i].LastAccessedAt.Before(items[j].LastAccessedAt)
		default: // name
			less = strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
		if desc {
			return !less
		}
		return less
	})
}

// AutoCategorize infers an ItemType from the provided fields map.
func AutoCategorize(fields map[string]string) types.ItemType {
	has := func(key string) bool {
		_, ok := fields[key]
		return ok
	}

	switch {
	case has(types.FieldURL) && has(types.FieldUsername) && has(types.FieldPassword):
		return types.ItemTypeLogin
	case has(types.FieldAPIKey) || has(types.FieldAPISecret):
		return types.ItemTypeAPIKey
	case has(types.FieldPrivateKey) || has(types.FieldPublicKey):
		return types.ItemTypeSSHKey
	case has(types.FieldCardNumber):
		return types.ItemTypeCreditCard
	case has(types.FieldFirstName) && has(types.FieldLastName):
		return types.ItemTypeIdentity
	case has(types.FieldCredentialID) && has(types.FieldRelyingPartyID):
		return types.ItemTypePasskey
	default:
		return types.ItemTypeCustom
	}
}
