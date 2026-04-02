package main

import (
"os"
"path/filepath"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// ───────── run() error paths ─────────

func TestRun_UnsupportedStorageType(t *testing.T) {
err := run(serverConfig{StorageType: "redis"})
require.Error(t, err)
assert.Contains(t, err.Error(), "unsupported storage type: redis")
}

func TestRun_PostgresWithoutURL(t *testing.T) {
err := run(serverConfig{StorageType: "postgres", PostgresURL: ""})
require.Error(t, err)
assert.Contains(t, err.Error(), "--postgres-url required")
}

func TestRun_PostgresWithBadURL(t *testing.T) {
err := run(serverConfig{
StorageType: "postgres",
PostgresURL: "host=127.0.0.1 port=1 dbname=x connect_timeout=1 sslmode=disable",
})
require.Error(t, err)
assert.Contains(t, err.Error(), "initialize storage")
}

func TestRun_EmptyStorageTypeDefaultsToSQLite(t *testing.T) {
dir := t.TempDir()

// Empty StorageType should default to "sqlite", not "unsupported".
// Use a bad DB path so sqlite init fails (proves it chose sqlite path).
badDir := filepath.Join(dir, "readonly")
require.NoError(t, os.MkdirAll(badDir, 0500))
badDB := filepath.Join(badDir, "sub", "test.db")

err := run(serverConfig{
StorageType: "",
DBPath:      badDB,
})
require.Error(t, err)
assert.NotContains(t, err.Error(), "unsupported storage type")
assert.Contains(t, err.Error(), "initialize storage")
}

func TestRun_SQLiteCreateTablesSuccess(t *testing.T) {
dir := t.TempDir()
dbPath := filepath.Join(dir, "tables-test.db")

// Test NewServer with different storage backends directly (not run()).
s, err := NewSQLiteStorage(dbPath)
require.NoError(t, err)
require.NoError(t, s.CreateTables())

blobStore := NewBlobStorage(s, BlobStorageConfig{
Endpoint: "http://localhost:9999",
Bucket:   "test",
})
srv := NewServer(0, blobStore, "test-key")
require.NotNil(t, srv)
assert.NotNil(t, srv.Handler)

s.Close()
}

func TestNewServer_DifferentPorts(t *testing.T) {
s := tempDB(t)

tests := []struct {
port int
addr string
}{
{8080, ":8080"},
{443, ":443"},
{0, ":0"},
}

for _, tt := range tests {
srv := NewServer(tt.port, s, "")
assert.Equal(t, tt.addr, srv.Addr)
}
}

func TestNewServer_WithBlobStorage(t *testing.T) {
s := tempDB(t)
blob := NewBlobStorage(s, BlobStorageConfig{
Endpoint: "http://localhost:9999",
Bucket:   "zeropass",
})
srv := NewServer(8443, blob, "my-api-key")
require.NotNil(t, srv)
assert.Equal(t, ":8443", srv.Addr)
}
