package item

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/types"
)

// testManager creates a Manager backed by a temp dir with a static vault key.
func testManager(t *testing.T) *Manager {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)

	vkFn := func() ([]byte, error) { return vaultKey, nil }
	return NewManager(dir, vkFn, nil)
}

func sampleLogin() *types.Item {
	return &types.Item{
		Type: types.ItemTypeLogin,
		Name: "GitHub",
		Fields: map[string]string{
			types.FieldUsername: "devuser",
			types.FieldPassword: "s3cur3P@ss!",
			types.FieldURL:      "https://github.com",
		},
		Tags: []string{"dev", "work"},
		Notes: "main account",
	}
}

func TestAddAndGetItem(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()

	err := m.AddItem(item)
	require.NoError(t, err)
	assert.NotEmpty(t, item.ID)
	assert.Equal(t, 1, item.Version)
	assert.False(t, item.CreatedAt.IsZero())

	got, err := m.GetItem(item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, "GitHub", got.Name)
	assert.Equal(t, "devuser", got.Fields[types.FieldUsername])
	assert.Equal(t, "s3cur3P@ss!", got.Fields[types.FieldPassword])
	assert.Equal(t, []string{"dev", "work"}, got.Tags)
}

func TestAddItemValidation(t *testing.T) {
	m := testManager(t)

	err := m.AddItem(nil)
	assert.Error(t, err)

	err = m.AddItem(&types.Item{Type: types.ItemTypeLogin})
	assert.Error(t, err) // no name

	// Empty type is now auto-categorized to "custom", so this should succeed.
	err = m.AddItem(&types.Item{Name: "x"})
	assert.NoError(t, err)

	err = m.AddItem(&types.Item{Name: "x", Type: "bogus"})
	assert.Error(t, err) // invalid type
}

func TestUpdateItem(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))
	id := item.ID

	item.Name = "GitHub Enterprise"
	item.Fields[types.FieldPassword] = "newP@ssw0rd"
	err := m.UpdateItem(id, item)
	require.NoError(t, err)
	assert.Equal(t, 2, item.Version)

	got, err := m.GetItem(id)
	require.NoError(t, err)
	assert.Equal(t, "GitHub Enterprise", got.Name)
	assert.Equal(t, "newP@ssw0rd", got.Fields[types.FieldPassword])
	assert.Equal(t, 2, got.Version)
}

func TestDeleteItem(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	err := m.DeleteItem(item.ID)
	require.NoError(t, err)

	_, err = m.GetItem(item.ID)
	assert.Error(t, err)
}

func TestDeleteNonExistent(t *testing.T) {
	m := testManager(t)
	err := m.DeleteItem("does-not-exist")
	assert.NoError(t, err) // idempotent
}

func TestListItems(t *testing.T) {
	m := testManager(t)

	login := sampleLogin()
	require.NoError(t, m.AddItem(login))

	note := &types.Item{
		Type:  types.ItemTypeSecureNote,
		Name:  "Secret Note",
		Notes: "confidential info",
		Tags:  []string{"personal"},
	}
	require.NoError(t, m.AddItem(note))

	apiKey := &types.Item{
		Type: types.ItemTypeAPIKey,
		Name: "AWS Key",
		Fields: map[string]string{
			types.FieldAPIKey: "AKIA...",
		},
		Favorite: true,
	}
	require.NoError(t, m.AddItem(apiKey))

	// All items
	all, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Filter by type
	logins, err := m.ListItems(types.ItemFilter{Type: types.ItemTypeLogin})
	require.NoError(t, err)
	assert.Len(t, logins, 1)
	assert.Equal(t, "GitHub", logins[0].Name)

	// Filter by tag
	personal, err := m.ListItems(types.ItemFilter{Tags: []string{"personal"}})
	require.NoError(t, err)
	assert.Len(t, personal, 1)

	// Filter by favorite
	fav := true
	favs, err := m.ListItems(types.ItemFilter{Favorite: &fav})
	require.NoError(t, err)
	assert.Len(t, favs, 1)
	assert.Equal(t, "AWS Key", favs[0].Name)

	// Search query
	results, err := m.ListItems(types.ItemFilter{SearchQuery: "github"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestListItemsSorting(t *testing.T) {
	m := testManager(t)

	items := []string{"Charlie", "Alpha", "Bravo"}
	for _, name := range items {
		require.NoError(t, m.AddItem(&types.Item{
			Type: types.ItemTypeLogin,
			Name: name,
			Fields: map[string]string{types.FieldPassword: "pass"},
		}))
	}

	// Sort ascending by name (default)
	sorted, err := m.ListItems(types.ItemFilter{SortBy: types.SortByName, SortOrder: types.SortAsc})
	require.NoError(t, err)
	require.Len(t, sorted, 3)
	assert.Equal(t, "Alpha", sorted[0].Name)
	assert.Equal(t, "Bravo", sorted[1].Name)
	assert.Equal(t, "Charlie", sorted[2].Name)

	// Sort descending
	sorted, err = m.ListItems(types.ItemFilter{SortBy: types.SortByName, SortOrder: types.SortDesc})
	require.NoError(t, err)
	assert.Equal(t, "Charlie", sorted[0].Name)
}

func TestGetItemEmptyID(t *testing.T) {
	m := testManager(t)
	_, err := m.GetItem("")
	assert.Error(t, err)
}

func TestUpdateItemEmptyID(t *testing.T) {
	m := testManager(t)
	err := m.UpdateItem("", sampleLogin())
	assert.Error(t, err)
}

func TestDeleteItemEmptyID(t *testing.T) {
	m := testManager(t)
	err := m.DeleteItem("")
	assert.Error(t, err)
}

func TestCRUDLifecycle(t *testing.T) {
	m := testManager(t)

	// Create
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))
	id := item.ID

	// Read
	got, err := m.GetItem(id)
	require.NoError(t, err)
	assert.Equal(t, "GitHub", got.Name)

	// Update
	got.Name = "Updated"
	require.NoError(t, m.UpdateItem(id, got))

	got2, err := m.GetItem(id)
	require.NoError(t, err)
	assert.Equal(t, "Updated", got2.Name)
	assert.Equal(t, 2, got2.Version)

	// Delete
	require.NoError(t, m.DeleteItem(id))
	_, err = m.GetItem(id)
	assert.Error(t, err)
}

func TestEmptyVault(t *testing.T) {
	m := testManager(t)
	items, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestAllItems(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(sampleLogin()))
	items, err := m.AllItems()
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestSortByCreatedAt(t *testing.T) {
	m := testManager(t)
	names := []string{"Zeta", "Alpha", "Mid"}
	for _, n := range names {
		item := &types.Item{
			Type:   types.ItemTypeLogin,
			Name:   n,
			Fields: map[string]string{types.FieldPassword: "p"},
		}
		require.NoError(t, m.AddItem(item))
	}
	items, err := m.ListItems(types.ItemFilter{SortBy: types.SortByCreatedAt, SortOrder: types.SortAsc})
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

func TestSortByUpdatedAt(t *testing.T) {
	m := testManager(t)
	for _, n := range []string{"B", "A"} {
		require.NoError(t, m.AddItem(&types.Item{
			Type: types.ItemTypeLogin, Name: n,
			Fields: map[string]string{types.FieldPassword: "p"},
		}))
	}
	items, err := m.ListItems(types.ItemFilter{SortBy: types.SortByUpdatedAt})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestSortByType(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{Type: types.ItemTypeSecureNote, Name: "Note"}))
	require.NoError(t, m.AddItem(&types.Item{Type: types.ItemTypeLogin, Name: "Login", Fields: map[string]string{types.FieldPassword: "p"}}))
	items, err := m.ListItems(types.ItemFilter{SortBy: types.SortByType})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type) // "login" < "note"
}

func TestFilterBySearchInCustomFields(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{
		Type:         types.ItemTypeCustom,
		Name:         "Custom",
		CustomFields: map[string]string{"env": "production"},
	}))
	items, err := m.ListItems(types.ItemFilter{SearchQuery: "production"})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestFilterBySearchInNotes(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{
		Type:  types.ItemTypeSecureNote,
		Name:  "Note",
		Notes: "contains unique marker xyz987",
	}))
	items, err := m.ListItems(types.ItemFilter{SearchQuery: "xyz987"})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestUpdateItemNil(t *testing.T) {
	m := testManager(t)
	err := m.UpdateItem("some-id", nil)
	assert.Error(t, err)
}

func TestUpdateItemValidation(t *testing.T) {
	m := testManager(t)
	err := m.UpdateItem("id", &types.Item{Type: types.ItemTypeLogin})
	assert.Error(t, err) // name empty
}

// mockIndexer implements item.Indexer for testing
type mockIndexer struct {
	added   []*types.Item
	updated []*types.Item
	removed []string
}

func (mi *mockIndexer) AddToIndex(item *types.Item) error {
	mi.added = append(mi.added, item)
	return nil
}

func (mi *mockIndexer) UpdateIndex(item *types.Item) error {
	mi.updated = append(mi.updated, item)
	return nil
}

func (mi *mockIndexer) RemoveFromIndex(id string) error {
	mi.removed = append(mi.removed, id)
	return nil
}

func testManagerWithIndexer(t *testing.T) (*Manager, *mockIndexer) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)

	vkFn := func() ([]byte, error) { return vaultKey, nil }
	idx := &mockIndexer{}
	return NewManager(dir, vkFn, idx), idx
}

func TestAddItemWithIndexer(t *testing.T) {
	m, idx := testManagerWithIndexer(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))
	assert.Len(t, idx.added, 1)
	assert.Equal(t, item.ID, idx.added[0].ID)
}

func TestUpdateItemWithIndexer(t *testing.T) {
	m, idx := testManagerWithIndexer(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	item.Name = "Updated"
	require.NoError(t, m.UpdateItem(item.ID, item))
	assert.Len(t, idx.updated, 1)
	assert.Equal(t, "Updated", idx.updated[0].Name)
}

func TestDeleteItemWithIndexer(t *testing.T) {
	m, idx := testManagerWithIndexer(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	require.NoError(t, m.DeleteItem(item.ID))
	assert.Len(t, idx.removed, 1)
	assert.Equal(t, item.ID, idx.removed[0])
}

// errorIndexer always returns errors
type errorIndexer struct{}

func (ei *errorIndexer) AddToIndex(item *types.Item) error    { return errors.New("index add error") }
func (ei *errorIndexer) UpdateIndex(item *types.Item) error   { return errors.New("index update error") }
func (ei *errorIndexer) RemoveFromIndex(id string) error      { return errors.New("index remove error") }

func TestAddItemIndexerError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)
	vkFn := func() ([]byte, error) { return vaultKey, nil }
	m := NewManager(dir, vkFn, &errorIndexer{})

	err = m.AddItem(sampleLogin())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index add")
}

func TestUpdateItemIndexerError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)
	vkFn := func() ([]byte, error) { return vaultKey, nil }
	m := NewManager(dir, vkFn, &errorIndexer{})

	item := sampleLogin()
	// Temporarily use nil indexer to add successfully
	m.indexer = nil
	require.NoError(t, m.AddItem(item))
	m.indexer = &errorIndexer{}

	err = m.UpdateItem(item.ID, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index update")
}

func TestDeleteItemIndexerError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)
	vkFn := func() ([]byte, error) { return vaultKey, nil }
	m := NewManager(dir, vkFn, &errorIndexer{})

	item := sampleLogin()
	m.indexer = nil
	require.NoError(t, m.AddItem(item))
	m.indexer = &errorIndexer{}

	err = m.DeleteItem(item.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index remove")
}

func TestGetItemNonExistent(t *testing.T) {
	m := testManager(t)
	_, err := m.GetItem("does-not-exist-at-all")
	assert.Error(t, err)
}

func TestVaultKeyError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vkFn := func() ([]byte, error) { return nil, errors.New("vault is locked") }
	m := NewManager(dir, vkFn, nil)

	err := m.AddItem(sampleLogin())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get vault key")
}

func TestListItemsNonExistentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	vaultKey, _ := key.GenerateVaultKey()
	vkFn := func() ([]byte, error) { return vaultKey, nil }
	m := NewManager(dir, vkFn, nil)

	items, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestListItemsSkipsDirectories(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Create a subdirectory inside items dir — should be skipped
	require.NoError(t, os.MkdirAll(filepath.Join(m.itemsPath, "subdir"), 0700))

	// Create a non-json file — should be skipped
	require.NoError(t, os.WriteFile(filepath.Join(m.itemsPath, "readme.txt"), []byte("hi"), 0600))

	items, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1) // only the valid item
}

func TestListItemsSkipsVersionFiles(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Create a .versions.json file — should be skipped
	versionsPath := filepath.Join(m.itemsPath, item.ID+".versions.json")
	require.NoError(t, os.WriteFile(versionsPath, []byte("[]"), 0600))

	items, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1) // only the regular item, not the versions file
}

func TestListItemsSkipsCorruptedFiles(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Create a corrupted json item file
	corruptedPath := filepath.Join(m.itemsPath, "corrupted-id.json")
	require.NoError(t, os.WriteFile(corruptedPath, []byte("{invalid json"), 0600))

	items, err := m.ListItems(types.ItemFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1) // corrupted file silently skipped
}

func TestReadAndDecryptChecksumMismatch(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Tamper with the checksum in the on-disk file
	fp := filepath.Join(m.itemsPath, item.ID+".json")
	data, err := os.ReadFile(fp)
	require.NoError(t, err)

	var enc types.EncryptedItem
	require.NoError(t, json.Unmarshal(data, &enc))
	enc.Checksum = "0000000000000000000000000000000000000000000000000000000000000000"
	tampered, err := json.Marshal(enc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(fp, tampered, 0600))

	_, err = m.GetItem(item.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestListItemsFilterCombination(t *testing.T) {
	m := testManager(t)

	// Add items with different types and tags
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeLogin, Name: "GitHub",
		Fields: map[string]string{types.FieldPassword: "p"},
		Tags: []string{"dev", "work"},
	}))
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeLogin, Name: "GitLab",
		Fields: map[string]string{types.FieldPassword: "p"},
		Tags: []string{"dev"},
	}))
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeSecureNote, Name: "MyNote",
		Tags: []string{"personal"},
	}))

	// Filter by type + tags
	items, err := m.ListItems(types.ItemFilter{
		Type: types.ItemTypeLogin,
		Tags: []string{"work"},
	})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "GitHub", items[0].Name)

	// Filter type + search
	items, err = m.ListItems(types.ItemFilter{
		Type:        types.ItemTypeLogin,
		SearchQuery: "gitlab",
	})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "GitLab", items[0].Name)

	// No match for combined filters
	items, err = m.ListItems(types.ItemFilter{
		Type: types.ItemTypeSecureNote,
		Tags: []string{"dev"},
	})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFilterNotFavorite(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeLogin, Name: "Fav",
		Fields: map[string]string{types.FieldPassword: "p"},
		Favorite: true,
	}))
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeLogin, Name: "NotFav",
		Fields: map[string]string{types.FieldPassword: "p"},
		Favorite: false,
	}))

	notFav := false
	items, err := m.ListItems(types.ItemFilter{Favorite: &notFav})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "NotFav", items[0].Name)
}

func TestSearchInFields(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{
		Type: types.ItemTypeLogin, Name: "AWS",
		Fields: map[string]string{
			types.FieldUsername: "admin@company.com",
			types.FieldPassword: "secret",
		},
	}))

	items, err := m.ListItems(types.ItemFilter{SearchQuery: "admin@company"})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestSortDescByUpdatedAt(t *testing.T) {
	m := testManager(t)
	for _, n := range []string{"A", "B"} {
		require.NoError(t, m.AddItem(&types.Item{
			Type: types.ItemTypeLogin, Name: n,
			Fields: map[string]string{types.FieldPassword: "p"},
		}))
	}
	items, err := m.ListItems(types.ItemFilter{
		SortBy:    types.SortByUpdatedAt,
		SortOrder: types.SortDesc,
	})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestSortDescByCreatedAt(t *testing.T) {
	m := testManager(t)
	for _, n := range []string{"A", "B"} {
		require.NoError(t, m.AddItem(&types.Item{
			Type: types.ItemTypeLogin, Name: n,
			Fields: map[string]string{types.FieldPassword: "p"},
		}))
	}
	items, err := m.ListItems(types.ItemFilter{
		SortBy:    types.SortByCreatedAt,
		SortOrder: types.SortDesc,
	})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestSortDescByType(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.AddItem(&types.Item{Type: types.ItemTypeLogin, Name: "L", Fields: map[string]string{types.FieldPassword: "p"}}))
	require.NoError(t, m.AddItem(&types.Item{Type: types.ItemTypeSecureNote, Name: "N"}))

	items, err := m.ListItems(types.ItemFilter{
		SortBy:    types.SortByType,
		SortOrder: types.SortDesc,
	})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, types.ItemTypeSecureNote, items[0].Type) // "note" > "login" descending
}

func TestReadAndDecryptInvalidBase64Data(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Corrupt the data field to be invalid base64
	fp := filepath.Join(m.itemsPath, item.ID+".json")
	data, err := os.ReadFile(fp)
	require.NoError(t, err)

	var enc types.EncryptedItem
	require.NoError(t, json.Unmarshal(data, &enc))
	enc.Data = "not-valid-base64!!!"
	tampered, err := json.Marshal(enc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(fp, tampered, 0600))

	_, err = m.GetItem(item.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode item data")
}

func TestReadAndDecryptVaultKeyErrorDuringRead(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)

	locked := false
	vkFn := func() ([]byte, error) {
		if locked {
			return nil, errors.New("vault is locked")
		}
		return vaultKey, nil
	}
	m := NewManager(dir, vkFn, nil)

	// Add item successfully
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Now "lock" the vault — read should fail at vault key step
	locked = true
	_, err = m.GetItem(item.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get vault key")
}

func TestUpdateItemVaultKeyError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)

	locked := false
	vkFn := func() ([]byte, error) {
		if locked {
			return nil, errors.New("vault is locked")
		}
		return vaultKey, nil
	}
	m := NewManager(dir, vkFn, nil)

	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Lock, then try to update
	locked = true
	err = m.UpdateItem(item.ID, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get vault key")
}

func TestDeleteItemRemovesVersionFile(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	// Create a versions file
	versionsPath := filepath.Join(m.itemsPath, item.ID+".versions.json")
	require.NoError(t, os.WriteFile(versionsPath, []byte(`[{"version": 1}]`), 0600))

	// Delete should remove both files
	require.NoError(t, m.DeleteItem(item.ID))

	_, err := os.Stat(filepath.Join(m.itemsPath, item.ID+".json"))
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(versionsPath)
	assert.True(t, os.IsNotExist(err))
}

func TestAddItemWithExistingID(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	item.ID = "preset-id-123"
	require.NoError(t, m.AddItem(item))
	assert.Equal(t, "preset-id-123", item.ID)
	assert.Equal(t, 1, item.Version)

	got, err := m.GetItem("preset-id-123")
	require.NoError(t, err)
	assert.Equal(t, "GitHub", got.Name)
}

func TestUpdateItemInvalidType(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	item.Type = "invalid-type"
	err := m.UpdateItem(item.ID, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid item type")
}

// --- Feature 1: Recently Used Tracking tests ---

func TestGetItemSetsLastAccessedAt(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	before := time.Now().UTC()
	got, err := m.GetItem(item.ID)
	require.NoError(t, err)
	after := time.Now().UTC()

	assert.False(t, got.LastAccessedAt.IsZero(), "LastAccessedAt should be set")
	assert.True(t, !got.LastAccessedAt.Before(before), "LastAccessedAt should be >= before")
	assert.True(t, !got.LastAccessedAt.After(after), "LastAccessedAt should be <= after")
}

func TestGetItemLastAccessedAtPersists(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))

	_, err := m.GetItem(item.ID)
	require.NoError(t, err)

	// Read again — should have the last accessed timestamp persisted
	got2, err := m.GetItem(item.ID)
	require.NoError(t, err)
	assert.False(t, got2.LastAccessedAt.IsZero())
}

func TestGetItemDoesNotIncrementVersion(t *testing.T) {
	m := testManager(t)
	item := sampleLogin()
	require.NoError(t, m.AddItem(item))
	assert.Equal(t, 1, item.Version)

	got, err := m.GetItem(item.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.Version, "version should not change on access")
}

func TestSortByLastAccessed(t *testing.T) {
	m := testManager(t)

	// Add items and access them in a known order
	names := []string{"First", "Second", "Third"}
	ids := make([]string, 3)
	for i, name := range names {
		item := &types.Item{
			Type:   types.ItemTypeLogin,
			Name:   name,
			Fields: map[string]string{types.FieldPassword: "p"},
		}
		require.NoError(t, m.AddItem(item))
		ids[i] = item.ID
	}

	// Access in reverse order: Third, Second, First
	for i := len(ids) - 1; i >= 0; i-- {
		_, err := m.GetItem(ids[i])
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // ensure distinct timestamps
	}

	// Sort ascending by last accessed: Third (earliest access), Second, First (latest access)
	items, err := m.ListItems(types.ItemFilter{
		SortBy:    types.SortByLastAccessed,
		SortOrder: types.SortAsc,
	})
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, "Third", items[0].Name)
	assert.Equal(t, "First", items[2].Name)
}

func TestSortByLastAccessedDesc(t *testing.T) {
	m := testManager(t)

	for _, name := range []string{"A", "B"} {
		item := &types.Item{
			Type:   types.ItemTypeLogin,
			Name:   name,
			Fields: map[string]string{types.FieldPassword: "p"},
		}
		require.NoError(t, m.AddItem(item))
	}

	items, err := m.ListItems(types.ItemFilter{
		SortBy:    types.SortByLastAccessed,
		SortOrder: types.SortDesc,
	})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

// --- Feature 2: Passkey ItemType tests ---

func TestAddPasskeyItem(t *testing.T) {
	m := testManager(t)
	item := &types.Item{
		Type: types.ItemTypePasskey,
		Name: "My WebAuthn Key",
		Fields: map[string]string{
			types.FieldCredentialID:    "cred-abc123",
			types.FieldPasskeyPublicKey: "pk-xyz",
			types.FieldRelyingPartyID:  "example.com",
			types.FieldUserHandle:      "user-handle-1",
			types.FieldSignCount:       "5",
		},
	}
	require.NoError(t, m.AddItem(item))
	assert.NotEmpty(t, item.ID)

	got, err := m.GetItem(item.ID)
	require.NoError(t, err)
	assert.Equal(t, types.ItemTypePasskey, got.Type)
	assert.Equal(t, "cred-abc123", got.Fields[types.FieldCredentialID])
	assert.Equal(t, "example.com", got.Fields[types.FieldRelyingPartyID])
}

func TestPasskeyTypeIsValid(t *testing.T) {
	assert.True(t, types.ValidItemTypes[types.ItemTypePasskey])
}

// --- Feature 4: Auto-Categorization tests ---

func TestAutoCategorizeLogin(t *testing.T) {
	fields := map[string]string{
		types.FieldURL:      "https://example.com",
		types.FieldUsername: "user",
		types.FieldPassword: "pass",
	}
	assert.Equal(t, types.ItemTypeLogin, AutoCategorize(fields))
}

func TestAutoCategorizeAPIKey(t *testing.T) {
	assert.Equal(t, types.ItemTypeAPIKey, AutoCategorize(map[string]string{
		types.FieldAPIKey: "key-123",
	}))
	assert.Equal(t, types.ItemTypeAPIKey, AutoCategorize(map[string]string{
		types.FieldAPISecret: "secret-456",
	}))
}

func TestAutoCategorizeSSHKey(t *testing.T) {
	assert.Equal(t, types.ItemTypeSSHKey, AutoCategorize(map[string]string{
		types.FieldPrivateKey: "-----BEGIN RSA PRIVATE KEY-----",
	}))
	assert.Equal(t, types.ItemTypeSSHKey, AutoCategorize(map[string]string{
		types.FieldPublicKey: "ssh-rsa AAAA...",
	}))
}

func TestAutoCategorizeCredCard(t *testing.T) {
	assert.Equal(t, types.ItemTypeCreditCard, AutoCategorize(map[string]string{
		types.FieldCardNumber: "4111111111111111",
	}))
}

func TestAutoCategorizeIdentity(t *testing.T) {
	assert.Equal(t, types.ItemTypeIdentity, AutoCategorize(map[string]string{
		types.FieldFirstName: "John",
		types.FieldLastName:  "Doe",
	}))
}

func TestAutoCategorizePasskey(t *testing.T) {
	assert.Equal(t, types.ItemTypePasskey, AutoCategorize(map[string]string{
		types.FieldCredentialID:   "cred-123",
		types.FieldRelyingPartyID: "example.com",
	}))
}

func TestAutoCategorizeCustom(t *testing.T) {
	assert.Equal(t, types.ItemTypeCustom, AutoCategorize(map[string]string{
		"some_random_field": "value",
	}))
	assert.Equal(t, types.ItemTypeCustom, AutoCategorize(nil))
	assert.Equal(t, types.ItemTypeCustom, AutoCategorize(map[string]string{}))
}

func TestAddItemAutoCategorizesWhenTypeEmpty(t *testing.T) {
	m := testManager(t)
	item := &types.Item{
		Name: "Auto Login",
		Fields: map[string]string{
			types.FieldURL:      "https://github.com",
			types.FieldUsername: "user",
			types.FieldPassword: "pass",
		},
	}
	require.NoError(t, m.AddItem(item))
	assert.Equal(t, types.ItemTypeLogin, item.Type)
}

func TestAddItemPreservesExplicitType(t *testing.T) {
	m := testManager(t)
	item := &types.Item{
		Name: "My Note",
		Type: types.ItemTypeSecureNote,
	}
	require.NoError(t, m.AddItem(item))
	assert.Equal(t, types.ItemTypeSecureNote, item.Type)
}
