// Package protocol defines shared types for the ZeroPass sync protocol.
package protocol

// SyncItem represents a single encrypted item for sync transfer.
// The Payload is always an opaque encrypted blob — the server never sees plaintext.
type SyncItem struct {
	ItemID    string `json:"item_id"`
	Version   int    `json:"version"`
	DeviceID  string `json:"device_id"`
	Payload   string `json:"payload"`   // Base64 encrypted blob
	Timestamp int64  `json:"timestamp"` // Unix timestamp (nanoseconds)
	Checksum  string `json:"checksum"`  // SHA-256 hex of payload
	Deleted   bool   `json:"deleted"`   // Tombstone for deletes
}

// PullRequest is sent by a client to fetch items changed since a given time.
type PullRequest struct {
	DeviceID string `json:"device_id"`
	Since    int64  `json:"since"` // Last sync timestamp (Unix nanoseconds)
}

// PullResponse contains items changed since the requested timestamp.
type PullResponse struct {
	Items      []SyncItem `json:"items"`
	ServerTime int64      `json:"server_time"`
}

// PushRequest is sent by a client to upload local changes.
type PushRequest struct {
	DeviceID string     `json:"device_id"`
	Items    []SyncItem `json:"items"`
}

// PushResponse reports which items were accepted and which conflicted.
type PushResponse struct {
	Accepted   []string       `json:"accepted"`
	Conflicts  []ConflictInfo `json:"conflicts"`
	ServerTime int64          `json:"server_time"`
}

// ConflictInfo describes a conflict for a single item.
type ConflictInfo struct {
	ItemID     string   `json:"item_id"`
	ServerItem SyncItem `json:"server_item"`
	Message    string   `json:"message"`
}

// DeviceInfo represents a registered sync device.
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	LastSyncAt int64  `json:"last_sync_at"`
}

// SyncResult is the aggregate result of a full sync cycle (pull + push).
type SyncResult struct {
	Pulled    int `json:"pulled"`
	Pushed    int `json:"pushed"`
	Conflicts int `json:"conflicts"`
}

// ErrorResponse is the standard error body returned by the sync server.
type ErrorResponse struct {
	Error string `json:"error"`
}
