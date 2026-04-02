package main

import (
	"database/sql"
	"fmt"

	"github.com/zeropass/zeropass/core/sync/protocol"

	_ "github.com/lib/pq"
)

// PostgresStorage implements the sync server Storage interface using PostgreSQL.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage opens a PostgreSQL connection using the given connection string.
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

// Close closes the underlying database connection.
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// CreateTables initialises the PostgreSQL schema.
func (s *PostgresStorage) CreateTables() error {
	const schema = `
	CREATE TABLE IF NOT EXISTS sync_items (
		item_id   TEXT PRIMARY KEY,
		version   INTEGER NOT NULL DEFAULT 1,
		device_id TEXT NOT NULL,
		payload   TEXT NOT NULL,
		timestamp BIGINT NOT NULL,
		checksum  TEXT NOT NULL DEFAULT '',
		deleted   BOOLEAN NOT NULL DEFAULT FALSE
	);
	CREATE INDEX IF NOT EXISTS idx_sync_items_timestamp ON sync_items(timestamp);
	CREATE INDEX IF NOT EXISTS idx_sync_items_device ON sync_items(device_id);
	CREATE TABLE IF NOT EXISTS devices (
		device_id    TEXT PRIMARY KEY,
		device_name  TEXT NOT NULL,
		last_sync_at BIGINT NOT NULL DEFAULT 0
	);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}

// SaveItem inserts or updates a sync item using PostgreSQL ON CONFLICT syntax.
func (s *PostgresStorage) SaveItem(item protocol.SyncItem) error {
	const q = `
	INSERT INTO sync_items (item_id, version, device_id, payload, timestamp, checksum, deleted)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (item_id) DO UPDATE SET
		version   = EXCLUDED.version,
		device_id = EXCLUDED.device_id,
		payload   = EXCLUDED.payload,
		timestamp = EXCLUDED.timestamp,
		checksum  = EXCLUDED.checksum,
		deleted   = EXCLUDED.deleted
	`
	_, err := s.db.Exec(q, item.ItemID, item.Version, item.DeviceID, item.Payload, item.Timestamp, item.Checksum, item.Deleted)
	if err != nil {
		return fmt.Errorf("save item %s: %w", item.ItemID, err)
	}
	return nil
}

// GetItemsSince returns all items with a timestamp strictly greater than since.
func (s *PostgresStorage) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	const q = `SELECT item_id, version, device_id, payload, timestamp, checksum, deleted
	           FROM sync_items WHERE timestamp > $1 ORDER BY timestamp ASC`

	rows, err := s.db.Query(q, since)
	if err != nil {
		return nil, fmt.Errorf("query items since %d: %w", since, err)
	}
	defer rows.Close()

	var items []protocol.SyncItem
	for rows.Next() {
		var it protocol.SyncItem
		if err := rows.Scan(&it.ItemID, &it.Version, &it.DeviceID, &it.Payload, &it.Timestamp, &it.Checksum, &it.Deleted); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// GetItem returns a single item by ID, or nil if not found.
func (s *PostgresStorage) GetItem(itemID string) (*protocol.SyncItem, error) {
	const q = `SELECT item_id, version, device_id, payload, timestamp, checksum, deleted
	           FROM sync_items WHERE item_id = $1`

	var it protocol.SyncItem
	err := s.db.QueryRow(q, itemID).Scan(&it.ItemID, &it.Version, &it.DeviceID, &it.Payload, &it.Timestamp, &it.Checksum, &it.Deleted)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get item %s: %w", itemID, err)
	}
	return &it, nil
}

// RegisterDevice inserts or updates a device record.
func (s *PostgresStorage) RegisterDevice(device protocol.DeviceInfo) error {
	const q = `
	INSERT INTO devices (device_id, device_name, last_sync_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (device_id) DO UPDATE SET
		device_name  = EXCLUDED.device_name,
		last_sync_at = EXCLUDED.last_sync_at
	`
	_, err := s.db.Exec(q, device.DeviceID, device.DeviceName, device.LastSyncAt)
	if err != nil {
		return fmt.Errorf("register device %s: %w", device.DeviceID, err)
	}
	return nil
}

// UpdateDeviceSync updates the last-sync timestamp for a device.
func (s *PostgresStorage) UpdateDeviceSync(deviceID string, timestamp int64) error {
	const q = `UPDATE devices SET last_sync_at = $1 WHERE device_id = $2`
	_, err := s.db.Exec(q, timestamp, deviceID)
	if err != nil {
		return fmt.Errorf("update device sync %s: %w", deviceID, err)
	}
	return nil
}
