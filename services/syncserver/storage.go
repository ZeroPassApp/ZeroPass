package main

import (
	"database/sql"
	"fmt"

	"github.com/zeropass/zeropass/core/sync/protocol"

	_ "modernc.org/sqlite"
)

// SQLiteStorage implements the sync server Storage interface using SQLite.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage opens (or creates) a SQLite database at the given path.
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	return &SQLiteStorage{db: db}, nil
}

// Close closes the underlying database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// CreateTables initialises the database schema.
func (s *SQLiteStorage) CreateTables() error {
	const schema = `
	CREATE TABLE IF NOT EXISTS sync_items (
		item_id   TEXT PRIMARY KEY,
		version   INTEGER NOT NULL DEFAULT 1,
		device_id TEXT    NOT NULL,
		payload   TEXT    NOT NULL,
		timestamp INTEGER NOT NULL,
		checksum  TEXT    NOT NULL DEFAULT '',
		deleted   INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_sync_items_timestamp ON sync_items(timestamp);
	CREATE INDEX IF NOT EXISTS idx_sync_items_device    ON sync_items(device_id);

	CREATE TABLE IF NOT EXISTS devices (
		device_id   TEXT PRIMARY KEY,
		device_name TEXT NOT NULL,
		last_sync_at INTEGER NOT NULL DEFAULT 0
	);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}

// SaveItem inserts or replaces a sync item.
func (s *SQLiteStorage) SaveItem(item protocol.SyncItem) error {
	const q = `
	INSERT INTO sync_items (item_id, version, device_id, payload, timestamp, checksum, deleted)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(item_id) DO UPDATE SET
		version   = excluded.version,
		device_id = excluded.device_id,
		payload   = excluded.payload,
		timestamp = excluded.timestamp,
		checksum  = excluded.checksum,
		deleted   = excluded.deleted
	`
	deleted := 0
	if item.Deleted {
		deleted = 1
	}
	_, err := s.db.Exec(q, item.ItemID, item.Version, item.DeviceID, item.Payload, item.Timestamp, item.Checksum, deleted)
	if err != nil {
		return fmt.Errorf("save item %s: %w", item.ItemID, err)
	}
	return nil
}

// GetItemsSince returns all items with a timestamp strictly greater than since.
func (s *SQLiteStorage) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	const q = `SELECT item_id, version, device_id, payload, timestamp, checksum, deleted
	           FROM sync_items WHERE timestamp > ? ORDER BY timestamp ASC`

	rows, err := s.db.Query(q, since)
	if err != nil {
		return nil, fmt.Errorf("query items since %d: %w", since, err)
	}
	defer rows.Close()

	var items []protocol.SyncItem
	for rows.Next() {
		var it protocol.SyncItem
		var deleted int
		if err := rows.Scan(&it.ItemID, &it.Version, &it.DeviceID, &it.Payload, &it.Timestamp, &it.Checksum, &deleted); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		it.Deleted = deleted != 0
		items = append(items, it)
	}
	return items, rows.Err()
}

// GetItem returns a single item by ID, or nil if not found.
func (s *SQLiteStorage) GetItem(itemID string) (*protocol.SyncItem, error) {
	const q = `SELECT item_id, version, device_id, payload, timestamp, checksum, deleted
	           FROM sync_items WHERE item_id = ?`

	var it protocol.SyncItem
	var deleted int
	err := s.db.QueryRow(q, itemID).Scan(&it.ItemID, &it.Version, &it.DeviceID, &it.Payload, &it.Timestamp, &it.Checksum, &deleted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get item %s: %w", itemID, err)
	}
	it.Deleted = deleted != 0
	return &it, nil
}

// RegisterDevice inserts or updates a device record.
func (s *SQLiteStorage) RegisterDevice(device protocol.DeviceInfo) error {
	const q = `
	INSERT INTO devices (device_id, device_name, last_sync_at)
	VALUES (?, ?, ?)
	ON CONFLICT(device_id) DO UPDATE SET
		device_name  = excluded.device_name,
		last_sync_at = excluded.last_sync_at
	`
	_, err := s.db.Exec(q, device.DeviceID, device.DeviceName, device.LastSyncAt)
	if err != nil {
		return fmt.Errorf("register device %s: %w", device.DeviceID, err)
	}
	return nil
}

// UpdateDeviceSync updates the last-sync timestamp for a device.
func (s *SQLiteStorage) UpdateDeviceSync(deviceID string, timestamp int64) error {
	const q = `UPDATE devices SET last_sync_at = ? WHERE device_id = ?`
	_, err := s.db.Exec(q, timestamp, deviceID)
	if err != nil {
		return fmt.Errorf("update device sync %s: %w", deviceID, err)
	}
	return nil
}
