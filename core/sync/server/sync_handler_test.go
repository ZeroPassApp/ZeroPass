package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// memStore is a minimal in-memory Storage for testing the handler.
type memStore struct {
	items   map[string]protocol.SyncItem
	devices map[string]protocol.DeviceInfo
}

func newMemStore() *memStore {
	return &memStore{
		items:   make(map[string]protocol.SyncItem),
		devices: make(map[string]protocol.DeviceInfo),
	}
}

func (m *memStore) SaveItem(item protocol.SyncItem) error {
	m.items[item.ItemID] = item
	return nil
}

func (m *memStore) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	var out []protocol.SyncItem
	for _, it := range m.items {
		if it.Timestamp > since {
			out = append(out, it)
		}
	}
	return out, nil
}

func (m *memStore) GetItem(itemID string) (*protocol.SyncItem, error) {
	it, ok := m.items[itemID]
	if !ok {
		return nil, nil
	}
	return &it, nil
}

func (m *memStore) RegisterDevice(device protocol.DeviceInfo) error {
	m.devices[device.DeviceID] = device
	return nil
}

func (m *memStore) UpdateDeviceSync(deviceID string, timestamp int64) error {
	if d, ok := m.devices[deviceID]; ok {
		d.LastSyncAt = timestamp
		m.devices[deviceID] = d
	}
	return nil
}

func newTestServer(apiKey string) (*httptest.Server, *memStore) {
	store := newMemStore()
	handler := NewHandler(store, apiKey)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return httptest.NewServer(mux), store
}

func TestHealth(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}

func TestRegisterDevice(t *testing.T) {
	srv, store := newTestServer("")
	defer srv.Close()

	info := protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: "Laptop"}
	body, _ := json.Marshal(info)

	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, store.devices, "dev-1")
}

func TestRegisterDevice_MissingFields(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	info := protocol.DeviceInfo{DeviceID: "dev-1"}
	body, _ := json.Marshal(info)

	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPull_Empty(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=dev-1&since=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pr protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Empty(t, pr.Items)
	assert.Greater(t, pr.ServerTime, int64(0))
}

func TestPull_MissingDeviceID(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?since=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPull_WithItems(t *testing.T) {
	srv, store := newTestServer("")
	defer srv.Close()

	store.items["i1"] = protocol.SyncItem{ItemID: "i1", Version: 1, Timestamp: 1000, Payload: "enc"}
	store.items["i2"] = protocol.SyncItem{ItemID: "i2", Version: 1, Timestamp: 2000, Payload: "enc2"}

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=dev-1&since=500")
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Len(t, pr.Items, 2)
}

func TestPull_SinceFilters(t *testing.T) {
	srv, store := newTestServer("")
	defer srv.Close()

	store.items["i1"] = protocol.SyncItem{ItemID: "i1", Version: 1, Timestamp: 1000, Payload: "enc"}
	store.items["i2"] = protocol.SyncItem{ItemID: "i2", Version: 1, Timestamp: 3000, Payload: "enc2"}

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=dev-1&since=2000")
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Len(t, pr.Items, 1)
	assert.Equal(t, "i2", pr.Items[0].ItemID)
}

func TestPush_Success(t *testing.T) {
	srv, store := newTestServer("")
	defer srv.Close()

	req := protocol.PushRequest{
		DeviceID: "dev-1",
		Items: []protocol.SyncItem{
			{ItemID: "i1", Version: 1, Timestamp: 1000, Payload: "enc", Checksum: "abc"},
		},
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Equal(t, []string{"i1"}, pr.Accepted)
	assert.Empty(t, pr.Conflicts)

	assert.Contains(t, store.items, "i1")
}

func TestPush_Conflict(t *testing.T) {
	srv, store := newTestServer("")
	defer srv.Close()

	// Pre-seed server with a newer item from another device.
	store.items["i1"] = protocol.SyncItem{
		ItemID: "i1", Version: 2, DeviceID: "dev-2",
		Timestamp: 5000, Payload: "server-enc", Checksum: "xyz",
	}

	req := protocol.PushRequest{
		DeviceID: "dev-1",
		Items: []protocol.SyncItem{
			{ItemID: "i1", Version: 1, Timestamp: 3000, Payload: "client-enc", Checksum: "abc"},
		},
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Empty(t, pr.Accepted)
	require.Len(t, pr.Conflicts, 1)
	assert.Equal(t, "i1", pr.Conflicts[0].ItemID)
}

func TestPush_MissingDeviceID(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	req := protocol.PushRequest{Items: []protocol.SyncItem{{ItemID: "i1"}}}
	body, _ := json.Marshal(req)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuth_Required(t *testing.T) {
	srv, _ := newTestServer("my-secret")
	defer srv.Close()

	// No auth header.
	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d1&since=0")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Wrong key.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/sync/pull?device_id=d1&since=0", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Correct key.
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/sync/pull?device_id=d1&since=0", nil)
	req.Header.Set("Authorization", "Bearer my-secret")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuth_NotRequired(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d1&since=0")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCORS_Preflight(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/sync/pull", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestMethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	// POST to /health should fail.
	resp, err := http.Post(srv.URL+"/health", "application/json", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	// GET to /sync/push should fail.
	resp, err = http.Get(srv.URL + "/sync/push")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// E2E: Two clients syncing through the handler.
func TestE2E_TwoClients(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	// Client A pushes an item.
	pushA := protocol.PushRequest{
		DeviceID: "dev-a",
		Items: []protocol.SyncItem{
			{ItemID: "shared-1", Version: 1, DeviceID: "dev-a", Timestamp: 1000, Payload: "a-enc", Checksum: "a1"},
		},
	}
	body, _ := json.Marshal(pushA)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Client B pulls and sees the item.
	resp, err = http.Get(srv.URL + "/sync/pull?device_id=dev-b&since=0")
	require.NoError(t, err)
	defer resp.Body.Close()
	var pullB protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pullB))
	require.Len(t, pullB.Items, 1)
	assert.Equal(t, "shared-1", pullB.Items[0].ItemID)
	assert.Equal(t, "a-enc", pullB.Items[0].Payload)

	// Client B modifies and pushes back with newer timestamp.
	pushB := protocol.PushRequest{
		DeviceID: "dev-b",
		Items: []protocol.SyncItem{
			{ItemID: "shared-1", Version: 2, DeviceID: "dev-b", Timestamp: 2000, Payload: "b-enc", Checksum: "b1"},
		},
	}
	body, _ = json.Marshal(pushB)
	resp2, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp2.Body.Close()
	var pushResp protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pushResp))
	assert.Equal(t, []string{"shared-1"}, pushResp.Accepted)

	// Client A pulls and sees B's update.
	resp3, err := http.Get(srv.URL + "/sync/pull?device_id=dev-a&since=1500")
	require.NoError(t, err)
	defer resp3.Body.Close()
	var pullA protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&pullA))
	require.Len(t, pullA.Items, 1)
	assert.Equal(t, "b-enc", pullA.Items[0].Payload)
}

func TestPull_InvalidSince(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d1&since=notanumber")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPush_InvalidBody(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader([]byte("not json")))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegister_InvalidBody(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader([]byte("{bad")))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegister_MethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/devices/register")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestPull_MethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/sync/pull", "application/json", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHealth_MethodNotAllowed_PUT(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/health", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestPush_EmptyItems(t *testing.T) {
	srv, _ := newTestServer("")
	defer srv.Close()

	req := protocol.PushRequest{DeviceID: "d1", Items: []protocol.SyncItem{}}
	body, _ := json.Marshal(req)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Empty(t, pr.Accepted)
	assert.Empty(t, pr.Conflicts)
}
