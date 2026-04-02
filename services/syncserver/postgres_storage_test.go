package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Compile-time interface compliance: PostgresStorage must satisfy StorageWithTables.
var _ StorageWithTables = (*PostgresStorage)(nil)

func TestNewPostgresStorage_InvalidConnStr(t *testing.T) {
	// Connection to a port that should be refused immediately.
	_, err := NewPostgresStorage("host=127.0.0.1 port=1 dbname=nonexistent connect_timeout=1 sslmode=disable")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping postgres")
}

func TestNewPostgresStorage_MalformedConnStr(t *testing.T) {
	// Completely malformed DSN should still fail at Ping.
	_, err := NewPostgresStorage("not-a-valid-dsn")
	assert.Error(t, err)
}
