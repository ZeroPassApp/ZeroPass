package main

import (
	"database/sql"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// newMockPostgres returns a PostgresStorage backed by sqlmock.
func newMockPostgres(t *testing.T) (*PostgresStorage, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &PostgresStorage{db: db}, mock
}

// ───────── Close ─────────

func TestPostgresClose(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectClose()
	err := s.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ───────── CreateTables ─────────

func TestPostgresCreateTables_Success(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS sync_items").
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := s.CreateTables()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresCreateTables_Error(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS sync_items").
		WillReturnError(fmt.Errorf("schema error"))
	err := s.CreateTables()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create tables")
}

// ───────── SaveItem ─────────

func TestPostgresSaveItem_Success(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO sync_items").
		WithArgs("item-1", 1, "dev-a", "payload", int64(1000), "checksum", false).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := s.SaveItem(protocol.SyncItem{
		ItemID: "item-1", Version: 1, DeviceID: "dev-a",
		Payload: "payload", Timestamp: 1000, Checksum: "checksum",
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSaveItem_Error(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO sync_items").
		WithArgs("fail", 1, "d", "p", int64(0), "c", false).
		WillReturnError(fmt.Errorf("insert error"))
	err := s.SaveItem(protocol.SyncItem{
		ItemID: "fail", Version: 1, DeviceID: "d",
		Payload: "p", Checksum: "c",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save item fail")
}

func TestPostgresSaveItem_Deleted(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO sync_items").
		WithArgs("del-1", 2, "dev-b", "", int64(2000), "", true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := s.SaveItem(protocol.SyncItem{
		ItemID: "del-1", Version: 2, DeviceID: "dev-b",
		Timestamp: 2000, Deleted: true,
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ───────── GetItemsSince ─────────

func TestPostgresGetItemsSince_Success(t *testing.T) {
	s, mock := newMockPostgres(t)
	rows := sqlmock.NewRows([]string{"item_id", "version", "device_id", "payload", "timestamp", "checksum", "deleted"}).
		AddRow("i1", 1, "d", "p1", int64(1000), "c1", false).
		AddRow("i2", 2, "d", "p2", int64(2000), "c2", true)
	mock.ExpectQuery("SELECT item_id, version, device_id, payload, timestamp, checksum, deleted").
		WithArgs(int64(0)).
		WillReturnRows(rows)

	items, err := s.GetItemsSince(0)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "i1", items[0].ItemID)
	assert.False(t, items[0].Deleted)
	assert.Equal(t, "i2", items[1].ItemID)
	assert.True(t, items[1].Deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetItemsSince_Empty(t *testing.T) {
	s, mock := newMockPostgres(t)
	rows := sqlmock.NewRows([]string{"item_id", "version", "device_id", "payload", "timestamp", "checksum", "deleted"})
	mock.ExpectQuery("SELECT item_id, version, device_id, payload, timestamp, checksum, deleted").
		WithArgs(int64(9999)).
		WillReturnRows(rows)

	items, err := s.GetItemsSince(9999)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetItemsSince_QueryError(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectQuery("SELECT item_id").
		WithArgs(int64(0)).
		WillReturnError(fmt.Errorf("query error"))

	items, err := s.GetItemsSince(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query items since 0")
	assert.Nil(t, items)
}

func TestPostgresGetItemsSince_ScanError(t *testing.T) {
	s, mock := newMockPostgres(t)
	// Return a row with wrong column count to trigger scan error
	rows := sqlmock.NewRows([]string{"item_id", "version", "device_id", "payload", "timestamp", "checksum", "deleted"}).
		AddRow("i1", "not-an-int", "d", "p", int64(0), "c", false)
	mock.ExpectQuery("SELECT item_id").
		WithArgs(int64(0)).
		WillReturnRows(rows)

	items, err := s.GetItemsSince(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scan item")
	assert.Nil(t, items)
}

// ───────── GetItem ─────────

func TestPostgresGetItem_Found(t *testing.T) {
	s, mock := newMockPostgres(t)
	rows := sqlmock.NewRows([]string{"item_id", "version", "device_id", "payload", "timestamp", "checksum", "deleted"}).
		AddRow("item-1", 3, "dev-x", "secret", int64(5000), "hash", false)
	mock.ExpectQuery("SELECT item_id, version, device_id, payload, timestamp, checksum, deleted").
		WithArgs("item-1").
		WillReturnRows(rows)

	item, err := s.GetItem("item-1")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "item-1", item.ItemID)
	assert.Equal(t, 3, item.Version)
	assert.Equal(t, "secret", item.Payload)
	assert.False(t, item.Deleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetItem_NotFound(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectQuery("SELECT item_id").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	item, err := s.GetItem("missing")
	require.NoError(t, err)
	assert.Nil(t, item)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetItem_Error(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectQuery("SELECT item_id").
		WithArgs("err-item").
		WillReturnError(fmt.Errorf("db error"))

	item, err := s.GetItem("err-item")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get item err-item")
	assert.Nil(t, item)
}

func TestPostgresGetItem_Deleted(t *testing.T) {
	s, mock := newMockPostgres(t)
	rows := sqlmock.NewRows([]string{"item_id", "version", "device_id", "payload", "timestamp", "checksum", "deleted"}).
		AddRow("del-1", 1, "d", "", int64(1000), "", true)
	mock.ExpectQuery("SELECT item_id").
		WithArgs("del-1").
		WillReturnRows(rows)

	item, err := s.GetItem("del-1")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.True(t, item.Deleted)
}

// ───────── RegisterDevice ─────────

func TestPostgresRegisterDevice_Success(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO devices").
		WithArgs("dev-1", "Laptop", int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.RegisterDevice(protocol.DeviceInfo{
		DeviceID: "dev-1", DeviceName: "Laptop", LastSyncAt: 0,
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRegisterDevice_Error(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO devices").
		WithArgs("dev-err", "X", int64(0)).
		WillReturnError(fmt.Errorf("constraint error"))

	err := s.RegisterDevice(protocol.DeviceInfo{
		DeviceID: "dev-err", DeviceName: "X",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register device dev-err")
}

func TestPostgresRegisterDevice_Upsert(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("INSERT INTO devices").
		WithArgs("dev-1", "Updated", int64(5000)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.RegisterDevice(protocol.DeviceInfo{
		DeviceID: "dev-1", DeviceName: "Updated", LastSyncAt: 5000,
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ───────── UpdateDeviceSync ─────────

func TestPostgresUpdateDeviceSync_Success(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("UPDATE devices SET last_sync_at").
		WithArgs(int64(9999), "dev-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.UpdateDeviceSync("dev-1", 9999)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUpdateDeviceSync_Error(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("UPDATE devices").
		WithArgs(int64(100), "dev-err").
		WillReturnError(fmt.Errorf("update error"))

	err := s.UpdateDeviceSync("dev-err", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update device sync dev-err")
}

func TestPostgresUpdateDeviceSync_NonExistent(t *testing.T) {
	s, mock := newMockPostgres(t)
	mock.ExpectExec("UPDATE devices SET last_sync_at").
		WithArgs(int64(100), "nonexistent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := s.UpdateDeviceSync("nonexistent", 100)
	assert.NoError(t, err) // 0 rows affected is not an error
}
