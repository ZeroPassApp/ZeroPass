package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// Compile-time interface compliance: BlobStorage must satisfy StorageWithTables.
var _ StorageWithTables = (*BlobStorage)(nil)

// mockStorage is an in-memory StorageWithTables for testing blob wrapping.
type mockStorage struct {
	mu      sync.Mutex
	items   map[string]protocol.SyncItem
	devices map[string]protocol.DeviceInfo
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		items:   make(map[string]protocol.SyncItem),
		devices: make(map[string]protocol.DeviceInfo),
	}
}

func (m *mockStorage) SaveItem(item protocol.SyncItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[item.ItemID] = item
	return nil
}

func (m *mockStorage) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []protocol.SyncItem
	for _, it := range m.items {
		if it.Timestamp > since {
			out = append(out, it)
		}
	}
	return out, nil
}

func (m *mockStorage) GetItem(itemID string) (*protocol.SyncItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[itemID]
	if !ok {
		return nil, nil
	}
	return &it, nil
}

func (m *mockStorage) RegisterDevice(device protocol.DeviceInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[device.DeviceID] = device
	return nil
}

func (m *mockStorage) UpdateDeviceSync(deviceID string, timestamp int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.devices[deviceID]; ok {
		d.LastSyncAt = timestamp
		m.devices[deviceID] = d
	}
	return nil
}

func (m *mockStorage) CreateTables() error { return nil }
func (m *mockStorage) Close() error        { return nil }

// fakeS3Server creates an httptest.Server simulating S3-compatible PUT/GET.
func fakeS3Server(t *testing.T) (*httptest.Server, map[string][]byte) {
	t.Helper()
	objects := make(map[string][]byte)
	mu := sync.Mutex{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/")
		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "read error", http.StatusInternalServerError)
				return
			}
			objects[key] = data
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			data, ok := objects[key]
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(data) //nolint:errcheck
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	t.Cleanup(srv.Close)
	return srv, objects
}

func TestBlobStorage_SaveAndGetItem(t *testing.T) {
	s3Srv, objects := fakeS3Server(t)
	base := newMockStorage()

	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint:  s3Srv.URL,
		Bucket:    "test-bucket",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})

	item := protocol.SyncItem{
		ItemID:    "item-1",
		Version:   1,
		DeviceID:  "dev-a",
		Payload:   "my-encrypted-data",
		Timestamp: 1000,
		Checksum:  "abc",
	}

	require.NoError(t, blob.SaveItem(item))

	// Verify payload was uploaded to S3.
	s3Key := "test-bucket/items/item-1/1"
	assert.Equal(t, []byte("my-encrypted-data"), objects[s3Key])

	// Verify DB stores an s3:// reference.
	dbItem, err := base.GetItem("item-1")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(dbItem.Payload, blobRefPrefix))

	// GetItem should resolve the reference back to the original payload.
	got, err := blob.GetItem("item-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "my-encrypted-data", got.Payload)
	assert.Equal(t, "item-1", got.ItemID)
	assert.Equal(t, 1, got.Version)
}

func TestBlobStorage_GetItemsSince(t *testing.T) {
	s3Srv, _ := fakeS3Server(t)
	base := newMockStorage()

	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: s3Srv.URL,
		Bucket:   "test-bucket",
	})

	for _, item := range []protocol.SyncItem{
		{ItemID: "i1", Version: 1, DeviceID: "d", Payload: "data-1", Timestamp: 1000, Checksum: "c1"},
		{ItemID: "i2", Version: 1, DeviceID: "d", Payload: "data-2", Timestamp: 2000, Checksum: "c2"},
	} {
		require.NoError(t, blob.SaveItem(item))
	}

	// GetItemsSince should resolve all s3:// references.
	items, err := blob.GetItemsSince(0)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	payloads := map[string]string{}
	for _, it := range items {
		payloads[it.ItemID] = it.Payload
	}
	assert.Equal(t, "data-1", payloads["i1"])
	assert.Equal(t, "data-2", payloads["i2"])
}

func TestBlobStorage_EmptyPayload(t *testing.T) {
	s3Srv, objects := fakeS3Server(t)
	base := newMockStorage()

	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: s3Srv.URL,
		Bucket:   "test-bucket",
	})

	item := protocol.SyncItem{
		ItemID:    "empty",
		Version:   1,
		DeviceID:  "d",
		Payload:   "",
		Timestamp: 1000,
		Checksum:  "",
	}
	require.NoError(t, blob.SaveItem(item))

	// Empty payload should not be uploaded to S3.
	assert.Empty(t, objects)

	got, err := blob.GetItem("empty")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "", got.Payload)
}

func TestBlobStorage_NonS3Payload(t *testing.T) {
	base := newMockStorage()
	// No S3 server needed — payload doesn't start with "s3://".
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "test-bucket",
	})

	// Insert directly into base storage (bypass blob upload).
	base.mu.Lock()
	base.items["direct"] = protocol.SyncItem{
		ItemID:    "direct",
		Version:   1,
		DeviceID:  "d",
		Payload:   "not-an-s3-ref",
		Timestamp: 1000,
		Checksum:  "c",
	}
	base.mu.Unlock()

	got, err := blob.GetItem("direct")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "not-an-s3-ref", got.Payload)
}

func TestBlobStorage_GetItem_NotFound(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "test-bucket",
	})

	got, err := blob.GetItem("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestBlobStorage_Delegates(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "test-bucket",
	})

	// RegisterDevice delegates to base.
	require.NoError(t, blob.RegisterDevice(protocol.DeviceInfo{DeviceID: "d1", DeviceName: "Test"}))
	assert.Contains(t, base.devices, "d1")

	// UpdateDeviceSync delegates to base.
	require.NoError(t, blob.UpdateDeviceSync("d1", 5000))

	// CreateTables delegates to base.
	require.NoError(t, blob.CreateTables())

	// Close delegates to base.
	require.NoError(t, blob.Close())
}

func TestBlobStorage_DefaultRegion(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "test-bucket",
	})
	assert.Equal(t, "us-east-1", blob.config.Region)
}

func TestBlobStorage_CustomRegion(t *testing.T) {
	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: "http://unused:9999",
		Bucket:   "test-bucket",
		Region:   "eu-west-1",
	})
	assert.Equal(t, "eu-west-1", blob.config.Region)
}

func TestBlobStorage_BuildURL_WithScheme(t *testing.T) {
	blob := &BlobStorage{
		config: BlobStorageConfig{
			Endpoint: "http://localhost:9000",
			Bucket:   "mybucket",
		},
	}
	u := blob.buildURL("items/test/1")
	assert.Equal(t, "http://localhost:9000/mybucket/items/test/1", u)
}

func TestBlobStorage_BuildURL_WithoutScheme(t *testing.T) {
	blob := &BlobStorage{
		config: BlobStorageConfig{
			Endpoint: "localhost:9000",
			Bucket:   "mybucket",
		},
	}
	u := blob.buildURL("items/test/1")
	assert.Equal(t, "http://localhost:9000/mybucket/items/test/1", u)
}

func TestBlobStorage_BuildURL_SSL(t *testing.T) {
	blob := &BlobStorage{
		config: BlobStorageConfig{
			Endpoint: "s3.amazonaws.com",
			Bucket:   "mybucket",
			UseSSL:   true,
		},
	}
	u := blob.buildURL("items/test/1")
	assert.Equal(t, "https://s3.amazonaws.com/mybucket/items/test/1", u)
}

func TestBlobStorage_S3Auth(t *testing.T) {
	var gotAuth string
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer authSrv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint:  authSrv.URL,
		Bucket:    "b",
		AccessKey: "myaccess",
		SecretKey: "mysecret",
	})

	item := protocol.SyncItem{
		ItemID: "auth-item", Version: 1, DeviceID: "d",
		Payload: "test", Timestamp: 1000, Checksum: "c",
	}
	require.NoError(t, blob.SaveItem(item))

	// Verify basic auth was sent.
	assert.Contains(t, gotAuth, "Basic ")
}

func TestBlobStorage_PutObjectError(t *testing.T) {
	// Server that always returns 500.
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer errSrv.Close()

	base := newMockStorage()
	blob := NewBlobStorage(base, BlobStorageConfig{
		Endpoint: errSrv.URL,
		Bucket:   "b",
	})

	item := protocol.SyncItem{
		ItemID: "fail", Version: 1, DeviceID: "d",
		Payload: "data", Timestamp: 1000, Checksum: "c",
	}
	err := blob.SaveItem(item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload blob")
}

func TestBlobStorage_GetObjectError(t *testing.T) {
	// Server accepts PUT but returns 404 for GET.
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

	item := protocol.SyncItem{
		ItemID: "lost", Version: 1, DeviceID: "d",
		Payload: "data", Timestamp: 1000, Checksum: "c",
	}
	require.NoError(t, blob.SaveItem(item))

	_, err := blob.GetItem("lost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download blob")
}
