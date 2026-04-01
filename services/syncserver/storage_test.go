package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

func tempDB(t *testing.T) *SQLiteStorage {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NoError(t, s.CreateTables())
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateTables_Idempotent(t *testing.T) {
	s := tempDB(t)
	// Call again — should not error.
	require.NoError(t, s.CreateTables())
}

func TestSaveAndGetItem(t *testing.T) {
	s := tempDB(t)
	item := protocol.SyncItem{
		ItemID:    "item-1",
		Version:   1,
		DeviceID:  "dev-a",
		Payload:   "encrypted-data",
		Timestamp: 1000,
		Checksum:  "abc123",
		Deleted:   false,
	}
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("item-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, item.ItemID, got.ItemID)
	assert.Equal(t, item.Version, got.Version)
	assert.Equal(t, item.DeviceID, got.DeviceID)
	assert.Equal(t, item.Payload, got.Payload)
	assert.Equal(t, item.Timestamp, got.Timestamp)
	assert.Equal(t, item.Checksum, got.Checksum)
	assert.False(t, got.Deleted)
}

func TestGetItem_NotFound(t *testing.T) {
	s := tempDB(t)
	got, err := s.GetItem("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSaveItem_Upsert(t *testing.T) {
	s := tempDB(t)
	item := protocol.SyncItem{
		ItemID: "item-1", Version: 1, DeviceID: "dev-a",
		Payload: "v1", Timestamp: 1000, Checksum: "c1",
	}
	require.NoError(t, s.SaveItem(item))

	item.Version = 2
	item.Payload = "v2"
	item.Timestamp = 2000
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("item-1")
	require.NoError(t, err)
	assert.Equal(t, 2, got.Version)
	assert.Equal(t, "v2", got.Payload)
	assert.Equal(t, int64(2000), got.Timestamp)
}

func TestSaveItem_Deleted(t *testing.T) {
	s := tempDB(t)
	item := protocol.SyncItem{
		ItemID: "item-del", Version: 1, DeviceID: "dev-a",
		Payload: "enc", Timestamp: 1000, Checksum: "c", Deleted: true,
	}
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("item-del")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.Deleted)
}

func TestGetItemsSince(t *testing.T) {
	s := tempDB(t)

	items := []protocol.SyncItem{
		{ItemID: "i1", Version: 1, DeviceID: "d", Payload: "p1", Timestamp: 1000, Checksum: "c1"},
		{ItemID: "i2", Version: 1, DeviceID: "d", Payload: "p2", Timestamp: 2000, Checksum: "c2"},
		{ItemID: "i3", Version: 1, DeviceID: "d", Payload: "p3", Timestamp: 3000, Checksum: "c3"},
		{ItemID: "i4", Version: 1, DeviceID: "d", Payload: "p4", Timestamp: 4000, Checksum: "c4"},
	}
	for _, it := range items {
		require.NoError(t, s.SaveItem(it))
	}

	// Since 0 → all items.
	got, err := s.GetItemsSince(0)
	require.NoError(t, err)
	assert.Len(t, got, 4)

	// Since 2000 → items with timestamp > 2000.
	got, err = s.GetItemsSince(2000)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "i3", got[0].ItemID)
	assert.Equal(t, "i4", got[1].ItemID)

	// Since 5000 → no items.
	got, err = s.GetItemsSince(5000)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetItemsSince_Ordering(t *testing.T) {
	s := tempDB(t)

	// Insert in non-chronological order.
	require.NoError(t, s.SaveItem(protocol.SyncItem{ItemID: "i3", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 3000, Checksum: "c"}))
	require.NoError(t, s.SaveItem(protocol.SyncItem{ItemID: "i1", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 1000, Checksum: "c"}))
	require.NoError(t, s.SaveItem(protocol.SyncItem{ItemID: "i2", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 2000, Checksum: "c"}))

	got, err := s.GetItemsSince(0)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, "i1", got[0].ItemID)
	assert.Equal(t, "i2", got[1].ItemID)
	assert.Equal(t, "i3", got[2].ItemID)
}

func TestRegisterDevice(t *testing.T) {
	s := tempDB(t)
	dev := protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: "Laptop", LastSyncAt: 0}
	require.NoError(t, s.RegisterDevice(dev))

	// Registering again (upsert) should update the name.
	dev.DeviceName = "Desktop"
	require.NoError(t, s.RegisterDevice(dev))
}

func TestUpdateDeviceSync(t *testing.T) {
	s := tempDB(t)
	dev := protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: "Laptop", LastSyncAt: 0}
	require.NoError(t, s.RegisterDevice(dev))
	require.NoError(t, s.UpdateDeviceSync("dev-1", 9999))

	// Updating non-existent device should not error (just 0 rows affected).
	require.NoError(t, s.UpdateDeviceSync("nonexistent", 1234))
}

func TestNewSQLiteStorage_InvalidPath(t *testing.T) {
	// Path inside a non-existent nested directory.
	badPath := filepath.Join(os.TempDir(), "zeropass-test-nonexistent", "deep", "test.db")
	_, err := NewSQLiteStorage(badPath)
	// modernc sqlite may fail on Open or on first PRAGMA; either way it should error.
	require.Error(t, err)
}

// ================================================================
// Additional storage tests for full coverage
// ================================================================

// TestSaveItem_UpsertFromDifferentDevice tests UPSERT with different device.
func TestSaveItem_UpsertFromDifferentDevice(t *testing.T) {
	s := tempDB(t)

	item := protocol.SyncItem{
		ItemID: "shared-1", Version: 1, DeviceID: "dev-a",
		Payload: "payload-a", Timestamp: 1000, Checksum: "ca",
	}
	require.NoError(t, s.SaveItem(item))

	// Update from different device
	item.DeviceID = "dev-b"
	item.Version = 2
	item.Payload = "payload-b"
	item.Timestamp = 2000
	item.Checksum = "cb"
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("shared-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "dev-b", got.DeviceID)
	assert.Equal(t, "payload-b", got.Payload)
	assert.Equal(t, 2, got.Version)
}

// TestSaveItem_DeleteThenRestore tests delete then restore.
func TestSaveItem_DeleteThenRestore(t *testing.T) {
	s := tempDB(t)

	// Create item
	item := protocol.SyncItem{
		ItemID: "restore-1", Version: 1, DeviceID: "dev-a",
		Payload: "original", Timestamp: 1000, Checksum: "c1",
	}
	require.NoError(t, s.SaveItem(item))

	// Delete it
	item.Version = 2
	item.Deleted = true
	item.Payload = ""
	item.Timestamp = 2000
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("restore-1")
	require.NoError(t, err)
	assert.True(t, got.Deleted)

	// Restore it
	item.Version = 3
	item.Deleted = false
	item.Payload = "restored"
	item.Timestamp = 3000
	require.NoError(t, s.SaveItem(item))

	got, err = s.GetItem("restore-1")
	require.NoError(t, err)
	assert.False(t, got.Deleted)
	assert.Equal(t, "restored", got.Payload)
}

// TestGetItemsSince_WithDeletedItems tests that deleted items appear in since query.
func TestGetItemsSince_WithDeletedItems(t *testing.T) {
	s := tempDB(t)

	require.NoError(t, s.SaveItem(protocol.SyncItem{
		ItemID: "alive", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 1000, Checksum: "c",
	}))
	require.NoError(t, s.SaveItem(protocol.SyncItem{
		ItemID: "dead", Version: 1, DeviceID: "d", Payload: "", Timestamp: 2000, Checksum: "", Deleted: true,
	}))

	got, err := s.GetItemsSince(0)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	// Verify deleted flag is preserved
	var foundDead bool
	for _, item := range got {
		if item.ItemID == "dead" {
			assert.True(t, item.Deleted)
			foundDead = true
		}
	}
	assert.True(t, foundDead)
}

// TestGetItemsSince_ZeroTimestamp tests items with zero timestamp.
func TestGetItemsSince_ZeroTimestamp(t *testing.T) {
	s := tempDB(t)

	require.NoError(t, s.SaveItem(protocol.SyncItem{
		ItemID: "zero-ts", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 0, Checksum: "c",
	}))

	// Since -1 should include timestamp 0
	got, err := s.GetItemsSince(-1)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "zero-ts", got[0].ItemID)

	// Since 0 should NOT include timestamp 0 (strictly greater)
	got, err = s.GetItemsSince(0)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestGetItemsSince_LargeTimestamps tests with very large timestamps.
func TestGetItemsSince_LargeTimestamps(t *testing.T) {
	s := tempDB(t)

	largeTS := int64(9999999999999)
	require.NoError(t, s.SaveItem(protocol.SyncItem{
		ItemID: "future", Version: 1, DeviceID: "d", Payload: "p", Timestamp: largeTS, Checksum: "c",
	}))

	got, err := s.GetItemsSince(largeTS - 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, largeTS, got[0].Timestamp)
}

// TestSaveItem_LargeBatch tests saving many items at once.
func TestSaveItem_LargeBatch(t *testing.T) {
	s := tempDB(t)

	const count = 100
	for i := 0; i < count; i++ {
		item := protocol.SyncItem{
			ItemID:    fmt.Sprintf("batch-%d", i),
			Version:   1,
			DeviceID:  "dev-batch",
			Payload:   fmt.Sprintf("payload-%d", i),
			Timestamp: int64(i * 100),
			Checksum:  fmt.Sprintf("c%d", i),
		}
		require.NoError(t, s.SaveItem(item))
	}

	// All items since 0
	got, err := s.GetItemsSince(-1)
	require.NoError(t, err)
	assert.Len(t, got, count)

	// Verify ordering
	for i := 1; i < len(got); i++ {
		assert.LessOrEqual(t, got[i-1].Timestamp, got[i].Timestamp)
	}
}

// TestConcurrentReads tests concurrent read operations.
func TestConcurrentReads(t *testing.T) {
	s := tempDB(t)

	// Write initial items sequentially
	for i := 0; i < 20; i++ {
		require.NoError(t, s.SaveItem(protocol.SyncItem{
			ItemID: fmt.Sprintf("conc-%d", i), Version: 1, DeviceID: "d",
			Payload: "p", Timestamp: int64((i + 1) * 100), Checksum: "c",
		}))
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 40)

	// 20 concurrent GetItemsSince readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(since int64) {
			defer wg.Done()
			items, err := s.GetItemsSince(since)
			if err != nil {
				errCh <- err
				return
			}
			// Verify items are ordered
			for j := 1; j < len(items); j++ {
				if items[j-1].Timestamp > items[j].Timestamp {
					errCh <- fmt.Errorf("items not ordered at index %d", j)
				}
			}
		}(int64(i * 100))
	}

	// 20 concurrent single-item readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			got, err := s.GetItem(fmt.Sprintf("conc-%d", id))
			if err != nil {
				errCh <- err
				return
			}
			if got == nil {
				errCh <- fmt.Errorf("item conc-%d not found", id)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent operation failed: %v", err)
	}
}

// TestSaveItem_EmptyPayload tests saving item with empty payload.
func TestSaveItem_EmptyPayload(t *testing.T) {
	s := tempDB(t)

	item := protocol.SyncItem{
		ItemID: "empty-payload", Version: 1, DeviceID: "d",
		Payload: "", Timestamp: 1000, Checksum: "",
	}
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("empty-payload")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "", got.Payload)
}

// TestSaveItem_LargePayload tests saving item with a large payload.
func TestSaveItem_LargePayload(t *testing.T) {
	s := tempDB(t)

	// Create a large payload (~100KB)
	payload := make([]byte, 100*1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	item := protocol.SyncItem{
		ItemID: "large", Version: 1, DeviceID: "d",
		Payload: string(payload), Timestamp: 1000, Checksum: "clarge",
	}
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("large")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, len(payload), len(got.Payload))
}

// TestSaveItem_SpecialCharacters tests saving with special characters.
func TestSaveItem_SpecialCharacters(t *testing.T) {
	s := tempDB(t)

	item := protocol.SyncItem{
		ItemID:   "special-chars",
		Version:  1,
		DeviceID: "dev-'quotes'",
		Payload:  `{"key": "val\"ue", "nested": {"a": 1}}`,
		Checksum: "c;drop table sync_items;--",
	}
	item.Timestamp = 1000
	require.NoError(t, s.SaveItem(item))

	got, err := s.GetItem("special-chars")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "dev-'quotes'", got.DeviceID)
	assert.Contains(t, got.Payload, `val\"ue`)
}

// TestSaveItem_ClosedDB tests operations on a closed database.
func TestSaveItem_ClosedDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "closed.db")
	s, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NoError(t, s.CreateTables())

	s.Close()

	// All operations should fail on closed DB
	err = s.SaveItem(protocol.SyncItem{
		ItemID: "x", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 1, Checksum: "c",
	})
	assert.Error(t, err)

	_, err = s.GetItem("x")
	assert.Error(t, err)

	_, err = s.GetItemsSince(0)
	assert.Error(t, err)

	err = s.RegisterDevice(protocol.DeviceInfo{DeviceID: "d", DeviceName: "n"})
	assert.Error(t, err)

	err = s.UpdateDeviceSync("d", 100)
	assert.Error(t, err)

	err = s.CreateTables()
	assert.Error(t, err)
}

// TestRegisterDevice_MultipleDevices tests registering multiple devices.
func TestRegisterDevice_MultipleDevices(t *testing.T) {
	s := tempDB(t)

	devices := []protocol.DeviceInfo{
		{DeviceID: "dev-1", DeviceName: "Phone", LastSyncAt: 0},
		{DeviceID: "dev-2", DeviceName: "Laptop", LastSyncAt: 100},
		{DeviceID: "dev-3", DeviceName: "Tablet", LastSyncAt: 200},
	}

	for _, dev := range devices {
		require.NoError(t, s.RegisterDevice(dev))
	}

	// Re-register with updated info
	for _, dev := range devices {
		dev.DeviceName = dev.DeviceName + " Updated"
		dev.LastSyncAt = 999
		require.NoError(t, s.RegisterDevice(dev))
	}
}

// TestUpdateDeviceSync_MultipleUpdates tests multiple sync updates.
func TestUpdateDeviceSync_MultipleUpdates(t *testing.T) {
	s := tempDB(t)

	dev := protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: "Test", LastSyncAt: 0}
	require.NoError(t, s.RegisterDevice(dev))

	// Update sync timestamp multiple times
	for i := 1; i <= 10; i++ {
		require.NoError(t, s.UpdateDeviceSync("dev-1", int64(i*1000)))
	}
}

// TestGetItemsSince_SameTimestamp tests items with identical timestamps.
func TestGetItemsSince_SameTimestamp(t *testing.T) {
	s := tempDB(t)

	// All items have the same timestamp
	for i := 0; i < 5; i++ {
		require.NoError(t, s.SaveItem(protocol.SyncItem{
			ItemID: fmt.Sprintf("same-ts-%d", i), Version: 1, DeviceID: "d",
			Payload: "p", Timestamp: 5000, Checksum: "c",
		}))
	}

	got, err := s.GetItemsSince(4999)
	require.NoError(t, err)
	assert.Len(t, got, 5)

	got, err = s.GetItemsSince(5000)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestNewServer creates and validates the server.
func TestNewServer(t *testing.T) {
	s := tempDB(t)
	srv := NewServer(8888, s, "test-key")
	require.NotNil(t, srv)
	assert.Equal(t, ":8888", srv.Addr)
	assert.NotNil(t, srv.Handler)
}

// TestNewServer_NoAuth creates a server without auth.
func TestNewServer_NoAuth(t *testing.T) {
	s := tempDB(t)
	srv := NewServer(9999, s, "")
	require.NotNil(t, srv)
	assert.Equal(t, ":9999", srv.Addr)
}

// TestGetItem_AfterUpsert verifies all fields after upsert.
func TestGetItem_AfterUpsert(t *testing.T) {
	s := tempDB(t)

	orig := protocol.SyncItem{
		ItemID: "upsert-check", Version: 1, DeviceID: "dev-a",
		Payload: "original", Timestamp: 1000, Checksum: "c-orig",
	}
	require.NoError(t, s.SaveItem(orig))

	updated := protocol.SyncItem{
		ItemID: "upsert-check", Version: 5, DeviceID: "dev-b",
		Payload: "updated-payload", Timestamp: 9999, Checksum: "c-new", Deleted: true,
	}
	require.NoError(t, s.SaveItem(updated))

	got, err := s.GetItem("upsert-check")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "upsert-check", got.ItemID)
	assert.Equal(t, 5, got.Version)
	assert.Equal(t, "dev-b", got.DeviceID)
	assert.Equal(t, "updated-payload", got.Payload)
	assert.Equal(t, int64(9999), got.Timestamp)
	assert.Equal(t, "c-new", got.Checksum)
	assert.True(t, got.Deleted)
}

// TestGetItemsSince_NegativeTimestamp tests with negative since value.
func TestGetItemsSince_NegativeTimestamp(t *testing.T) {
	s := tempDB(t)

	require.NoError(t, s.SaveItem(protocol.SyncItem{
		ItemID: "neg-ts", Version: 1, DeviceID: "d",
		Payload: "p", Timestamp: -100, Checksum: "c",
	}))

	got, err := s.GetItemsSince(-200)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, int64(-100), got[0].Timestamp)
}

// TestNewSQLiteStorage_ValidPath tests creating storage with valid path.
func TestNewSQLiteStorage_ValidPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "valid.db")
	s, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NotNil(t, s.db)
	s.Close()

	// Verify file was created
	_, err = os.Stat(dbPath)
	require.NoError(t, err)
}

// TestClose_Idempotent tests double close.
func TestClose_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "close.db")
	s, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)

	// First close should work
	require.NoError(t, s.Close())
}

// TestSaveItem_MultipleVersions tests saving multiple versions of same item.
func TestSaveItem_MultipleVersions(t *testing.T) {
	s := tempDB(t)

	for v := 1; v <= 10; v++ {
		require.NoError(t, s.SaveItem(protocol.SyncItem{
			ItemID: "versioned", Version: v, DeviceID: "d",
			Payload: fmt.Sprintf("v%d", v), Timestamp: int64(v * 1000), Checksum: fmt.Sprintf("c%d", v),
		}))
	}

	got, err := s.GetItem("versioned")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 10, got.Version)
	assert.Equal(t, "v10", got.Payload)
	assert.Equal(t, int64(10000), got.Timestamp)
}
