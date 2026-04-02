package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// errMockStorage is a mock that returns errors for specific methods.
type errMockStorage struct {
	mockStorage
	getItemsSinceErr error
}

func (e *errMockStorage) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	if e.getItemsSinceErr != nil {
		return nil, e.getItemsSinceErr
	}
	return e.mockStorage.GetItemsSince(since)
}

// TestBlobStorage_GetItemsSince_ResolveError tests GetItemsSince when resolve fails.
func TestBlobStorage_GetItemsSince_ResolveError(t *testing.T) {
	// Server accepts PUT but returns 404 for GET (resolve will fail)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
	})

	// Save an item (uploads to S3, stores s3:// ref in base)
	require.NoError(t, blob.SaveItem(protocol.SyncItem{
		ItemID: "fail-resolve", Version: 1, DeviceID: "d",
		Payload: "secret-data", Timestamp: 1000, Checksum: "c",
	}))

	// GetItemsSince should fail when trying to resolve the s3:// reference
	_, err := blob.GetItemsSince(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download blob")
}

// TestBlobStorage_GetItemsSince_BaseError tests GetItemsSince when base storage returns error.
func TestBlobStorage_GetItemsSince_BaseError(t *testing.T) {
	base := &errMockStorage{
		mockStorage:      *newMockStorage(),
		getItemsSinceErr: fmt.Errorf("db connection lost"),
	}
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "b",
	})

	_, err := blob.GetItemsSince(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db connection lost")
}

// TestBlobStorage_ResolvePayload_InvalidRef tests resolvePayload with malformed s3:// ref.
func TestBlobStorage_ResolvePayload_InvalidRef(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "b",
	})

	// Manually insert item with invalid s3:// ref (no "/" separator after bucket)
	base.mu.Lock()
	base.items["bad-ref"] = protocol.SyncItem{
		ItemID:    "bad-ref",
		Version:   1,
		DeviceID:  "d",
		Payload:   "s3://nobucketseparator",
		Timestamp: 1000,
		Checksum:  "c",
	}
	base.mu.Unlock()

	// GetItem should fail with invalid blob reference
	_, err := blob.GetItem("bad-ref")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid blob reference")
}

// TestBlobStorage_GetItemsSince_InvalidRef tests GetItemsSince with malformed s3:// ref.
func TestBlobStorage_GetItemsSince_InvalidRef(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "b",
	})

	// Manually insert item with invalid s3:// ref
	base.mu.Lock()
	base.items["bad-ref-since"] = protocol.SyncItem{
		ItemID:    "bad-ref-since",
		Version:   1,
		DeviceID:  "d",
		Payload:   "s3://nobucketseparator",
		Timestamp: 1000,
		Checksum:  "c",
	}
	base.mu.Unlock()

	_, err := blob.GetItemsSince(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid blob reference")
}

// TestBlobStorage_PutObject_NoAuth tests PUT without credentials.
func TestBlobStorage_PutObject_NoAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
		// No AccessKey / SecretKey
	})

	require.NoError(t, blob.SaveItem(protocol.SyncItem{
		ItemID: "noauth", Version: 1, DeviceID: "d",
		Payload: "test", Timestamp: 1000, Checksum: "c",
	}))

	// No auth header should be sent
	assert.Empty(t, gotAuth)
}

// TestBlobStorage_BuildURL_HTTPS tests buildURL with https scheme.
func TestBlobStorage_BuildURL_HTTPS(t *testing.T) {
	blob := &BlobStorage{
		config: BlobStorageConfig{
			Endpoint: "https://s3.example.com",
			Bucket:   "bucket",
		},
	}
	u := blob.buildURL("key/path")
	assert.Equal(t, "https://s3.example.com/bucket/key/path", u)
}

// TestBlobStorage_GetObject_NoAuth tests GET without credentials (no auth header).
func TestBlobStorage_GetObject_NoAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("payload-data")) //nolint:errcheck
		}
	}))
	defer srv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
		// No AccessKey / SecretKey
	})

	// Save item to create s3:// reference
	require.NoError(t, blob.SaveItem(protocol.SyncItem{
		ItemID: "noauth-get", Version: 1, DeviceID: "d",
		Payload: "test", Timestamp: 1000, Checksum: "c",
	}))

	// Reset to capture GET auth header
	gotAuth = ""
	got, err := blob.GetItem("noauth-get")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "payload-data", got.Payload)
	assert.Empty(t, gotAuth, "GET should not send auth when no credentials configured")
}

// TestBlobStorage_PutObject_ClientDoError tests putObject when client.Do fails (server closed).
func TestBlobStorage_PutObject_ClientDoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// Close the server before making the request
	srv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
	})

	err := blob.SaveItem(protocol.SyncItem{
		ItemID: "client-err", Version: 1, DeviceID: "d",
		Payload: "data", Timestamp: 1000, Checksum: "c",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload blob")
}

// TestBlobStorage_GetObject_ClientDoError tests getObject when client.Do fails (server closed).
func TestBlobStorage_GetObject_ClientDoError(t *testing.T) {
	// Use a working server for PUT, then close before GET
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
	})

	// Save succeeds
	require.NoError(t, blob.SaveItem(protocol.SyncItem{
		ItemID: "get-err", Version: 1, DeviceID: "d",
		Payload: "data", Timestamp: 1000, Checksum: "c",
	}))

	// Close server so GET fails
	srv.Close()

	_, err := blob.GetItem("get-err")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download blob")
}

// TestBlobStorage_PutObject_Non200Status tests putObject with a non-OK/Created status.
func TestBlobStorage_PutObject_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: srv.URL,
		Bucket:   "b",
	})

	err := blob.SaveItem(protocol.SyncItem{
		ItemID: "forbidden", Version: 1, DeviceID: "d",
		Payload: "data", Timestamp: 1000, Checksum: "c",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload blob")
}

// TestBlobStorage_SaveItem_DeletedItem tests SaveItem with a deleted item (empty payload).
func TestBlobStorage_SaveItem_DeletedItem(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "b",
	})

	err := blob.SaveItem(protocol.SyncItem{
		ItemID:    "del-item",
		Version:   2,
		DeviceID:  "d",
		Payload:   "", // empty = deleted
		Timestamp: 2000,
		Checksum:  "c",
		Deleted:   true,
	})
	require.NoError(t, err)

	got, err := blob.GetItem("del-item")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "", got.Payload)
	assert.True(t, got.Deleted)
}
