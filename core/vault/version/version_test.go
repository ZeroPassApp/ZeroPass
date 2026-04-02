package version

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/types"
)

func testVaultKeyFn() func() ([]byte, error) {
	vk := bytes.Repeat([]byte{0x11}, key.VaultKeySize)
	return func() ([]byte, error) { return vk, nil }
}

func testVersionManager(t *testing.T) *Manager {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	return NewManager(dir, testVaultKeyFn(), 10)
}

func sampleItem(version int) *types.Item {
	return &types.Item{
		ID:      "item-123",
		Type:    types.ItemTypeLogin,
		Name:    "GitHub",
		Version: version,
		Fields:  map[string]string{"password": "pass-v" + string(rune('0'+version))},
	}
}

func TestSaveAndGetHistory(t *testing.T) {
	m := testVersionManager(t)

	item := sampleItem(1)
	require.NoError(t, m.SaveVersion(item))

	item.Version = 2
	item.Fields["password"] = "pass-v2"
	require.NoError(t, m.SaveVersion(item))

	history, err := m.GetVersionHistory("item-123")
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Equal(t, 1, history[0].Version)
	assert.Equal(t, 2, history[1].Version)
	assert.Equal(t, "pass-v2", history[1].Item.Fields["password"])
	assert.False(t, history[0].SavedAt.IsZero())
}

func TestMaxVersionsTrimming(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 3) // keep only 3 versions

	for i := 1; i <= 5; i++ {
		item := &types.Item{
			ID:      "item-x",
			Type:    types.ItemTypeLogin,
			Name:    "test",
			Version: i,
		}
		require.NoError(t, m.SaveVersion(item))
	}

	history, err := m.GetVersionHistory("item-x")
	require.NoError(t, err)
	require.Len(t, history, 3)
	// Oldest kept = version 3
	assert.Equal(t, 3, history[0].Version)
	assert.Equal(t, 5, history[2].Version)
}

func TestRestoreVersion(t *testing.T) {
	m := testVersionManager(t)

	for i := 1; i <= 3; i++ {
		item := &types.Item{
			ID:      "item-r",
			Type:    types.ItemTypeLogin,
			Name:    "test",
			Version: i,
			Fields:  map[string]string{"value": string(rune('a' + i - 1))},
		}
		require.NoError(t, m.SaveVersion(item))
	}

	restored, err := m.RestoreVersion("item-r", 2)
	require.NoError(t, err)
	assert.Equal(t, 2, restored.Version)
	assert.Equal(t, "b", restored.Fields["value"])
}

func TestRestoreVersionNotFound(t *testing.T) {
	m := testVersionManager(t)
	item := sampleItem(1)
	require.NoError(t, m.SaveVersion(item))

	_, err := m.RestoreVersion("item-123", 99)
	assert.Error(t, err)
}

func TestDeleteHistory(t *testing.T) {
	m := testVersionManager(t)
	item := sampleItem(1)
	require.NoError(t, m.SaveVersion(item))

	require.NoError(t, m.DeleteHistory("item-123"))

	history, err := m.GetVersionHistory("item-123")
	require.NoError(t, err)
	assert.Empty(t, history)
}

func TestDeleteHistoryNonExistent(t *testing.T) {
	m := testVersionManager(t)
	err := m.DeleteHistory("nope")
	assert.NoError(t, err)
}

func TestGetHistoryEmptyID(t *testing.T) {
	m := testVersionManager(t)
	_, err := m.GetVersionHistory("")
	assert.Error(t, err)
}

func TestSaveVersionNil(t *testing.T) {
	m := testVersionManager(t)
	err := m.SaveVersion(nil)
	assert.Error(t, err)
}

func TestSaveVersionEmptyID(t *testing.T) {
	m := testVersionManager(t)
	err := m.SaveVersion(&types.Item{})
	assert.Error(t, err)
}

func TestRestoreVersionEmptyID(t *testing.T) {
	m := testVersionManager(t)
	_, err := m.RestoreVersion("", 1)
	assert.Error(t, err)
}

func TestGetHistoryNoFile(t *testing.T) {
	m := testVersionManager(t)
	history, err := m.GetVersionHistory("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, history)
}

func TestDefaultMaxVersions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 0) // should default to 10
	assert.Equal(t, 10, m.maxVersions)
}

func TestNegativeMaxVersions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), -5) // should default to 10
	assert.Equal(t, 10, m.maxVersions)
}

func TestLoadHistoryCorruptedJSON(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	// Write corrupted JSON to the version file
	fp := filepath.Join(dir, "item-corrupt.versions.json")
	require.NoError(t, os.WriteFile(fp, []byte("{invalid json"), 0600))

	_, err := m.GetVersionHistory("item-corrupt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal version history")
}

func TestLoadHistoryEmptyFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	// Write empty file
	fp := filepath.Join(dir, "item-empty.versions.json")
	require.NoError(t, os.WriteFile(fp, []byte(""), 0600))

	history, err := m.GetVersionHistory("item-empty")
	require.NoError(t, err)
	assert.Nil(t, history)
}

func TestLoadHistoryWhitespaceFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	// Write whitespace-only file
	fp := filepath.Join(dir, "item-ws.versions.json")
	require.NoError(t, os.WriteFile(fp, []byte("   \n  \t  "), 0600))

	history, err := m.GetVersionHistory("item-ws")
	require.NoError(t, err)
	assert.Nil(t, history)
}

func TestSaveVersionCreatesDir(t *testing.T) {
	// Use a path that doesn't exist yet — saveHistory should create it
	dir := filepath.Join(t.TempDir(), "nonexistent", "items")
	m := NewManager(dir, testVaultKeyFn(), 10)

	item := &types.Item{
		ID:      "item-new",
		Type:    types.ItemTypeLogin,
		Name:    "Test",
		Version: 1,
	}
	require.NoError(t, m.SaveVersion(item))

	history, err := m.GetVersionHistory("item-new")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestRestoreVersionNoHistory(t *testing.T) {
	m := testVersionManager(t)
	// No versions saved for this item
	_, err := m.RestoreVersion("no-history-item", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRestoreVersionFromCorruptedFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	fp := filepath.Join(dir, "item-bad.versions.json")
	require.NoError(t, os.WriteFile(fp, []byte("not json at all"), 0600))

	_, err := m.RestoreVersion("item-bad", 1)
	assert.Error(t, err)
}

func TestSaveVersionAppendsToExisting(t *testing.T) {
	m := testVersionManager(t)
	for i := 1; i <= 5; i++ {
		item := &types.Item{
			ID:      "item-append",
			Type:    types.ItemTypeLogin,
			Name:    "Appended",
			Version: i,
			Fields:  map[string]string{"v": fmt.Sprintf("%d", i)},
		}
		require.NoError(t, m.SaveVersion(item))
	}

	history, err := m.GetVersionHistory("item-append")
	require.NoError(t, err)
	assert.Len(t, history, 5)
	assert.Equal(t, "1", history[0].Item.Fields["v"])
	assert.Equal(t, "5", history[4].Item.Fields["v"])
}

func TestMaxVersionsExactlyAtLimit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 3)

	for i := 1; i <= 3; i++ {
		require.NoError(t, m.SaveVersion(&types.Item{
			ID: "item-limit", Type: types.ItemTypeLogin, Name: "test", Version: i,
		}))
	}

	history, err := m.GetVersionHistory("item-limit")
	require.NoError(t, err)
	assert.Len(t, history, 3)
	assert.Equal(t, 1, history[0].Version)
	assert.Equal(t, 3, history[2].Version)
}

func TestDeleteHistoryThenRestore(t *testing.T) {
	m := testVersionManager(t)
	item := &types.Item{
		ID: "item-del", Type: types.ItemTypeLogin, Name: "test", Version: 1,
	}
	require.NoError(t, m.SaveVersion(item))
	require.NoError(t, m.DeleteHistory("item-del"))

	// Restore should fail — history was deleted
	_, err := m.RestoreVersion("item-del", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestVersionFilePath(t *testing.T) {
	m := NewManager("/tmp/items", testVaultKeyFn(), 10)
	assert.Equal(t, "/tmp/items/my-id.versions.json", m.versionFilePath("my-id"))
}

func TestSaveVersionOverwriteOldVersions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 2) // keep only 2

	for i := 1; i <= 5; i++ {
		item := &types.Item{
			ID:     "overflow",
			Name:   fmt.Sprintf("v%d", i),
			Fields: map[string]string{"v": fmt.Sprintf("%d", i)},
		}
		require.NoError(t, m.SaveVersion(item))
	}

	// Should only have the latest 2
	versions, err := m.GetVersionHistory("overflow")
	require.NoError(t, err)
	assert.Len(t, versions, 2)
}

func TestRestoreToLatestVersion(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	for i := 1; i <= 3; i++ {
		require.NoError(t, m.SaveVersion(&types.Item{
			ID:      "restore-latest",
			Name:    fmt.Sprintf("v%d", i),
			Version: i,
		}))
	}

	// Restore version 3
	item, err := m.RestoreVersion("restore-latest", 3)
	require.NoError(t, err)
	assert.Equal(t, "v3", item.Name)
}

func TestRestoreVersionOutOfBounds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	require.NoError(t, m.SaveVersion(&types.Item{
		ID:      "bounded",
		Name:    "v1",
		Version: 1,
	}))

	_, err := m.RestoreVersion("bounded", 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteHistoryThenGet(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	require.NoError(t, m.SaveVersion(&types.Item{
		ID:   "delete-then-get",
		Name: "v1",
	}))
	require.NoError(t, m.DeleteHistory("delete-then-get"))

	versions, err := m.GetVersionHistory("delete-then-get")
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestGetVersionHistoryUnreadableFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	// Create a version file that's a directory (causes read error, not IsNotExist)
	versionFile := filepath.Join(dir, "unreadable.versions.json")
	require.NoError(t, os.MkdirAll(versionFile, 0700))

	_, err := m.GetVersionHistory("unreadable")
	assert.Error(t, err)
}

func TestSaveVersionToUnwritableDir(t *testing.T) {
	// Use a path that definitely can't be written to
	m := NewManager("/nonexistent/deeply/nested/path/items", testVaultKeyFn(), 10)
	err := m.SaveVersion(&types.Item{
		ID:   "cant-write",
		Name: "v1",
	})
	assert.Error(t, err)
}

func TestDeleteHistoryNonExistentFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "items")
	require.NoError(t, os.MkdirAll(dir, 0700))
	m := NewManager(dir, testVaultKeyFn(), 10)

	// Should not error for non-existent history
	err := m.DeleteHistory("does-not-exist")
	assert.NoError(t, err)
}
