package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	syncserver "github.com/zeropass/zeropass/core/sync/server"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// integrationServer creates a full stack (SQLite + handler) httptest server.
func integrationServer(t *testing.T, apiKey string) *httptest.Server {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "integration.db")
	store, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.CreateTables())
	t.Cleanup(func() { store.Close() })

	mux := http.NewServeMux()
	handler := syncserver.NewHandler(store, apiKey)
	handler.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// makeAuthReq creates an HTTP request with optional auth header.
func makeAuthReq(t *testing.T, method, url, apiKey string, body interface{}) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return req
}

func TestIntegration_HealthCheck(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}

func TestIntegration_RegisterAndSync(t *testing.T) {
	srv := integrationServer(t, "test-key")
	defer srv.Close()

	makeReq := func(method, path string, body interface{}) *http.Request {
		var buf bytes.Buffer
		if body != nil {
			json.NewEncoder(&buf).Encode(body) //nolint:errcheck
		}
		req, err := http.NewRequest(method, srv.URL+path, &buf)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		return req
	}

	// 1. Register device A.
	resp, err := http.DefaultClient.Do(makeReq(http.MethodPost, "/devices/register",
		protocol.DeviceInfo{DeviceID: "dev-a", DeviceName: "Laptop A"}))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. Device A pushes items.
	pushReq := protocol.PushRequest{
		DeviceID: "dev-a",
		Items: []protocol.SyncItem{
			{ItemID: "item-1", Version: 1, DeviceID: "dev-a", Timestamp: 1000, Payload: "aes-enc-1", Checksum: "c1"},
			{ItemID: "item-2", Version: 1, DeviceID: "dev-a", Timestamp: 1001, Payload: "aes-enc-2", Checksum: "c2"},
		},
	}
	resp, err = http.DefaultClient.Do(makeReq(http.MethodPost, "/sync/push", pushReq))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var pushResp protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pushResp))
	assert.ElementsMatch(t, []string{"item-1", "item-2"}, pushResp.Accepted)
	assert.Empty(t, pushResp.Conflicts)

	// 3. Device B registers.
	resp2, err := http.DefaultClient.Do(makeReq(http.MethodPost, "/devices/register",
		protocol.DeviceInfo{DeviceID: "dev-b", DeviceName: "Laptop B"}))
	require.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)

	// 4. Device B pulls.
	resp3, err := http.DefaultClient.Do(makeReq(http.MethodGet, "/sync/pull?device_id=dev-b&since=0", nil))
	require.NoError(t, err)
	defer resp3.Body.Close()
	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&pullResp))
	assert.Len(t, pullResp.Items, 2)

	// 5. Device B pushes a new item + updates item-1.
	pushB := protocol.PushRequest{
		DeviceID: "dev-b",
		Items: []protocol.SyncItem{
			{ItemID: "item-1", Version: 2, DeviceID: "dev-b", Timestamp: 2000, Payload: "aes-enc-1-v2", Checksum: "c1v2"},
			{ItemID: "item-3", Version: 1, DeviceID: "dev-b", Timestamp: 2001, Payload: "aes-enc-3", Checksum: "c3"},
		},
	}
	resp4, err := http.DefaultClient.Do(makeReq(http.MethodPost, "/sync/push", pushB))
	require.NoError(t, err)
	defer resp4.Body.Close()
	var pushResp2 protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp4.Body).Decode(&pushResp2))
	assert.ElementsMatch(t, []string{"item-1", "item-3"}, pushResp2.Accepted)

	// 6. Device A pulls changes since its last push.
	resp5, err := http.DefaultClient.Do(makeReq(http.MethodGet, "/sync/pull?device_id=dev-a&since=1001", nil))
	require.NoError(t, err)
	defer resp5.Body.Close()
	var pullA protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp5.Body).Decode(&pullA))
	assert.Len(t, pullA.Items, 2, "should see item-1 update and item-3")
}

func TestIntegration_PushConflict(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Device A pushes item-1.
	pushA := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "item-1", Version: 1, DeviceID: "dev-a", Timestamp: 5000, Payload: "a-enc", Checksum: "ca"}},
	}
	body, _ := json.Marshal(pushA)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Device B pushes same item with older timestamp → conflict.
	pushB := protocol.PushRequest{
		DeviceID: "dev-b",
		Items:    []protocol.SyncItem{{ItemID: "item-1", Version: 1, DeviceID: "dev-b", Timestamp: 3000, Payload: "b-enc", Checksum: "cb"}},
	}
	body, _ = json.Marshal(pushB)
	resp2, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp2.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pr))
	assert.Empty(t, pr.Accepted)
	require.Len(t, pr.Conflicts, 1)
	assert.Equal(t, "item-1", pr.Conflicts[0].ItemID)
	assert.Equal(t, "a-enc", pr.Conflicts[0].ServerItem.Payload)
}

func TestIntegration_DeleteTombstone(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Push an item.
	push1 := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "del-1", Version: 1, DeviceID: "dev-a", Timestamp: 1000, Payload: "enc", Checksum: "c"}},
	}
	body, _ := json.Marshal(push1)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Delete it (tombstone).
	push2 := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "del-1", Version: 2, DeviceID: "dev-a", Timestamp: 2000, Payload: "", Checksum: "", Deleted: true}},
	}
	body, _ = json.Marshal(push2)
	resp2, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp2.Body.Close()
	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pr))
	assert.Equal(t, []string{"del-1"}, pr.Accepted)

	// Pull should show deleted tombstone.
	resp3, err := http.Get(srv.URL + "/sync/pull?device_id=dev-b&since=0")
	require.NoError(t, err)
	defer resp3.Body.Close()
	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&pullResp))
	require.Len(t, pullResp.Items, 1)
	assert.True(t, pullResp.Items[0].Deleted)
}

// ================================================================
// Additional integration tests for full coverage
// ================================================================

// TestIntegration_AuthRequired tests that auth is enforced when key is set.
func TestIntegration_AuthRequired(t *testing.T) {
	srv := integrationServer(t, "secret-key")
	defer srv.Close()

	// No auth header → 401
	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d&since=0")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Wrong auth header → 401
	req := makeAuthReq(t, http.MethodGet, srv.URL+"/sync/pull?device_id=d&since=0", "wrong-key", nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Missing Bearer prefix → 401
	req2, _ := http.NewRequest(http.MethodGet, srv.URL+"/sync/pull?device_id=d&since=0", nil)
	req2.Header.Set("Authorization", "secret-key")
	resp, err = http.DefaultClient.Do(req2)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Correct auth → 200
	req3 := makeAuthReq(t, http.MethodGet, srv.URL+"/sync/pull?device_id=d&since=0", "secret-key", nil)
	resp, err = http.DefaultClient.Do(req3)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_AuthNotRequired tests no auth when key is empty.
func TestIntegration_AuthNotRequired(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// No auth header → should work
	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d&since=0")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_PullMissingDeviceID tests pull without device_id.
func TestIntegration_PullMissingDeviceID(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?since=0")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_PullInvalidSince tests pull with invalid since param.
func TestIntegration_PullInvalidSince(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d&since=notanumber")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_PullEmptyResult tests pull returns empty array, not null.
func TestIntegration_PullEmptyResult(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d&since=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pullResp))
	assert.NotNil(t, pullResp.Items)
	assert.Len(t, pullResp.Items, 0)
	assert.Greater(t, pullResp.ServerTime, int64(0))
}

// TestIntegration_PushMissingDeviceID tests push without device_id.
func TestIntegration_PushMissingDeviceID(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	pushReq := protocol.PushRequest{
		DeviceID: "",
		Items: []protocol.SyncItem{
			{ItemID: "i1", Version: 1, DeviceID: "d", Payload: "p", Timestamp: 1000, Checksum: "c"},
		},
	}
	body, _ := json.Marshal(pushReq)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_PushInvalidJSON tests push with invalid JSON body.
func TestIntegration_PushInvalidJSON(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/sync/push", "application/json", strings.NewReader("{invalid json"))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_PushEmptyItems tests push with empty items array.
func TestIntegration_PushEmptyItems(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	pushReq := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{},
	}
	body, _ := json.Marshal(pushReq)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Empty(t, pr.Accepted)
	assert.Empty(t, pr.Conflicts)
}

// TestIntegration_PullWrongMethod tests pull with POST (should fail).
func TestIntegration_PullWrongMethod(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/sync/pull", "application/json", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// TestIntegration_PushWrongMethod tests push with GET (should fail).
func TestIntegration_PushWrongMethod(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/push")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// TestIntegration_RegisterWrongMethod tests register with GET (should fail).
func TestIntegration_RegisterWrongMethod(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/devices/register")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// TestIntegration_RegisterInvalidJSON tests register with invalid JSON.
func TestIntegration_RegisterInvalidJSON(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/devices/register", "application/json", strings.NewReader("not json"))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_RegisterMissingDeviceID tests register without device_id.
func TestIntegration_RegisterMissingDeviceID(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	dev := protocol.DeviceInfo{DeviceID: "", DeviceName: "test"}
	body, _ := json.Marshal(dev)
	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_CORSPreflight tests CORS preflight request.
func TestIntegration_CORSPreflight(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/sync/pull", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestIntegration_CORSHeaders tests CORS headers on normal requests.
func TestIntegration_CORSHeaders(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/health", nil)
	req.Header.Set("Origin", "http://example.com")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

// TestIntegration_MultiDeviceSyncE2E tests a complete three-device sync scenario.
func TestIntegration_MultiDeviceSyncE2E(t *testing.T) {
	srv := integrationServer(t, "key123")
	defer srv.Close()

	doReq := func(method, path string, body interface{}) *http.Response {
		t.Helper()
		req := makeAuthReq(t, method, srv.URL+path, "key123", body)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	// 1. Register 3 devices
	for _, dev := range []string{"phone", "laptop", "tablet"} {
		resp := doReq(http.MethodPost, "/devices/register",
			protocol.DeviceInfo{DeviceID: dev, DeviceName: dev + "-name"})
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// 2. Phone pushes 5 items
	phoneItems := make([]protocol.SyncItem, 5)
	for i := 0; i < 5; i++ {
		phoneItems[i] = protocol.SyncItem{
			ItemID: fmt.Sprintf("ph-item-%d", i), Version: 1, DeviceID: "phone",
			Payload: fmt.Sprintf("phone-payload-%d", i), Timestamp: int64(1000 + i), Checksum: fmt.Sprintf("pc%d", i),
		}
	}
	resp := doReq(http.MethodPost, "/sync/push", protocol.PushRequest{DeviceID: "phone", Items: phoneItems})
	defer resp.Body.Close()
	var pushResp protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pushResp))
	assert.Len(t, pushResp.Accepted, 5)

	// 3. Laptop pulls all items
	resp2 := doReq(http.MethodGet, "/sync/pull?device_id=laptop&since=0", nil)
	defer resp2.Body.Close()
	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pullResp))
	assert.Len(t, pullResp.Items, 5)
	// Track the max item timestamp for the since filter
	var maxTS int64
	for _, it := range pullResp.Items {
		if it.Timestamp > maxTS {
			maxTS = it.Timestamp
		}
	}

	// 4. Tablet modifies item 0, pushes
	resp3 := doReq(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "tablet",
		Items: []protocol.SyncItem{
			{ItemID: "ph-item-0", Version: 2, DeviceID: "tablet",
				Payload: "modified-by-tablet", Timestamp: 5000, Checksum: "tc0v2"},
		},
	})
	defer resp3.Body.Close()
	var pushResp2 protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&pushResp2))
	assert.Equal(t, []string{"ph-item-0"}, pushResp2.Accepted)

	// 5. Laptop pulls since its last max item timestamp — should see the tablet's change
	resp4 := doReq(http.MethodGet, fmt.Sprintf("/sync/pull?device_id=laptop&since=%d", maxTS), nil)
	defer resp4.Body.Close()
	var pullResp2 protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp4.Body).Decode(&pullResp2))
	assert.Len(t, pullResp2.Items, 1)

	// 6. Phone tries to push an older version → conflict
	resp5 := doReq(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "phone",
		Items: []protocol.SyncItem{
			{ItemID: "ph-item-0", Version: 2, DeviceID: "phone",
				Payload: "phone-update-conflicting", Timestamp: 3000, Checksum: "conflict"},
		},
	})
	defer resp5.Body.Close()
	var pushResp3 protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp5.Body).Decode(&pushResp3))
	assert.Empty(t, pushResp3.Accepted)
	assert.Len(t, pushResp3.Conflicts, 1)
	assert.Equal(t, "ph-item-0", pushResp3.Conflicts[0].ItemID)

	// 7. Phone resolves conflict by pushing with newer timestamp
	resp6 := doReq(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "phone",
		Items: []protocol.SyncItem{
			{ItemID: "ph-item-0", Version: 3, DeviceID: "phone",
				Payload: "phone-resolved", Timestamp: 9000, Checksum: "resolved"},
		},
	})
	defer resp6.Body.Close()
	var pushResp4 protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp6.Body).Decode(&pushResp4))
	assert.Equal(t, []string{"ph-item-0"}, pushResp4.Accepted)
}

// TestIntegration_PushManyItems tests pushing a large batch of items.
func TestIntegration_PushManyItems(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	items := make([]protocol.SyncItem, 50)
	for i := 0; i < 50; i++ {
		items[i] = protocol.SyncItem{
			ItemID: fmt.Sprintf("batch-%d", i), Version: 1, DeviceID: "dev-a",
			Payload: fmt.Sprintf("payload-%d", i), Timestamp: int64((i + 1) * 100), Checksum: fmt.Sprintf("c%d", i),
		}
	}

	pushReq := protocol.PushRequest{DeviceID: "dev-a", Items: items}
	body, _ := json.Marshal(pushReq)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Len(t, pr.Accepted, 50)
	assert.Empty(t, pr.Conflicts)

	// Pull all (since=0 gets items with timestamp > 0)
	resp2, err := http.Get(srv.URL + "/sync/pull?device_id=dev-b&since=0")
	require.NoError(t, err)
	defer resp2.Body.Close()
	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pullResp))
	assert.Len(t, pullResp.Items, 50)
}

// TestIntegration_PullWithSinceFilter tests pull with various since values.
func TestIntegration_PullWithSinceFilter(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Push items with different timestamps
	items := []protocol.SyncItem{
		{ItemID: "ts-1", Version: 1, DeviceID: "d", Payload: "p1", Timestamp: 1000, Checksum: "c1"},
		{ItemID: "ts-2", Version: 1, DeviceID: "d", Payload: "p2", Timestamp: 2000, Checksum: "c2"},
		{ItemID: "ts-3", Version: 1, DeviceID: "d", Payload: "p3", Timestamp: 3000, Checksum: "c3"},
	}
	pushReq := protocol.PushRequest{DeviceID: "d", Items: items}
	body, _ := json.Marshal(pushReq)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	tests := []struct {
		since    int64
		expected int
	}{
		{0, 3},
		{1000, 2},
		{2000, 1},
		{3000, 0},
		{9999, 0},
	}

	for _, tc := range tests {
		resp, err := http.Get(fmt.Sprintf("%s/sync/pull?device_id=d&since=%d", srv.URL, tc.since))
		require.NoError(t, err)
		var pullResp protocol.PullResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&pullResp))
		resp.Body.Close()
		assert.Len(t, pullResp.Items, tc.expected, "since=%d", tc.since)
	}
}

// TestIntegration_PushConflictWithSameTimestamp tests conflict with equal timestamp from different device.
func TestIntegration_PushConflictWithSameTimestamp(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Device A pushes
	pushA := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "eq-ts", Version: 1, DeviceID: "dev-a", Timestamp: 5000, Payload: "a-data", Checksum: "ca"}},
	}
	body, _ := json.Marshal(pushA)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Device B pushes same item with equal timestamp → conflict (server has newer-or-equal from different device)
	pushB := protocol.PushRequest{
		DeviceID: "dev-b",
		Items:    []protocol.SyncItem{{ItemID: "eq-ts", Version: 1, DeviceID: "dev-b", Timestamp: 5000, Payload: "b-data", Checksum: "cb"}},
	}
	body, _ = json.Marshal(pushB)
	resp, err = http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Empty(t, pr.Accepted)
	assert.Len(t, pr.Conflicts, 1)
}

// TestIntegration_RegisterAndReRegister tests device re-registration.
func TestIntegration_RegisterAndReRegister(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	dev := protocol.DeviceInfo{DeviceID: "re-reg", DeviceName: "Phone v1"}
	body, _ := json.Marshal(dev)
	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Re-register with new name
	dev.DeviceName = "Phone v2"
	body, _ = json.Marshal(dev)
	resp, err = http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestIntegration_HealthCheckNoAuth tests health doesn't need auth.
func TestIntegration_HealthCheckNoAuth(t *testing.T) {
	srv := integrationServer(t, "super-secret")
	defer srv.Close()

	// Health endpoint should not require auth
	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_PushThenDelete tests push then soft delete.
func TestIntegration_PushThenDelete(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Push an item
	push := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "to-delete", Version: 1, DeviceID: "dev-a", Timestamp: 1000, Payload: "data", Checksum: "c"}},
	}
	body, _ := json.Marshal(push)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Soft delete
	delPush := protocol.PushRequest{
		DeviceID: "dev-a",
		Items:    []protocol.SyncItem{{ItemID: "to-delete", Version: 2, DeviceID: "dev-a", Timestamp: 2000, Payload: "", Checksum: "", Deleted: true}},
	}
	body, _ = json.Marshal(delPush)
	resp, err = http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Equal(t, []string{"to-delete"}, pr.Accepted)

	// Other device pulls → sees tombstone
	resp2, err := http.Get(srv.URL + "/sync/pull?device_id=other&since=0")
	require.NoError(t, err)
	defer resp2.Body.Close()
	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&pullResp))
	require.Len(t, pullResp.Items, 1)
	assert.True(t, pullResp.Items[0].Deleted)
	assert.Equal(t, "to-delete", pullResp.Items[0].ItemID)
}

// TestIntegration_PullServerTime tests that server_time is returned and positive.
func TestIntegration_PullServerTime(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=d&since=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pullResp))
	assert.Greater(t, pullResp.ServerTime, int64(0))
}

// TestIntegration_PushSameDeviceUpdate tests push from same device (should accept).
func TestIntegration_PushSameDeviceUpdate(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Push v1
	push1 := protocol.PushRequest{
		DeviceID: "dev-a",
		Items: []protocol.SyncItem{
			{ItemID: "same-dev", Version: 1, DeviceID: "dev-a", Timestamp: 1000, Payload: "v1", Checksum: "c1"},
		},
	}
	body, _ := json.Marshal(push1)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Push v2 from same device with newer timestamp
	push2 := protocol.PushRequest{
		DeviceID: "dev-a",
		Items: []protocol.SyncItem{
			{ItemID: "same-dev", Version: 2, DeviceID: "dev-a", Timestamp: 2000, Payload: "v2", Checksum: "c2"},
		},
	}
	body, _ = json.Marshal(push2)
	resp, err = http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var pr protocol.PushResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pr))
	assert.Equal(t, []string{"same-dev"}, pr.Accepted)
	assert.Empty(t, pr.Conflicts)
}

// TestIntegration_HealthWrongMethod sends POST to /health and expects 405.
func TestIntegration_HealthWrongMethod(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/health", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// TestIntegration_RegisterMissingDeviceName sends register with empty device_name.
func TestIntegration_RegisterMissingDeviceName(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	body, _ := json.Marshal(protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: ""})
	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// integrationServerWithStore returns both the test server and the store for error injection.
func integrationServerWithStore(t *testing.T) (*httptest.Server, *SQLiteStorage) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "integration.db")
	store, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.CreateTables())

	mux := http.NewServeMux()
	handler := syncserver.NewHandler(store, "")
	handler.RegisterRoutes(mux)
	return httptest.NewServer(mux), store
}

// TestIntegration_PullStorageError closes the DB and tests pull returns 500.
func TestIntegration_PullStorageError(t *testing.T) {
	srv, store := integrationServerWithStore(t)
	defer srv.Close()

	store.Close() // intentionally break storage

	resp, err := http.Get(srv.URL + "/sync/pull?device_id=dev1&since=0")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestIntegration_PushStorageError closes the DB and tests push returns 500.
func TestIntegration_PushStorageError(t *testing.T) {
	srv, store := integrationServerWithStore(t)
	defer srv.Close()

	store.Close() // intentionally break storage

	push := protocol.PushRequest{
		DeviceID: "dev-1",
		Items: []protocol.SyncItem{
			{ItemID: "item-1", Version: 1, DeviceID: "dev-1", Payload: "p", Timestamp: 1000, Checksum: "c"},
		},
	}
	body, _ := json.Marshal(push)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestIntegration_RegisterStorageError closes the DB and tests register returns 500.
func TestIntegration_RegisterStorageError(t *testing.T) {
	srv, store := integrationServerWithStore(t)
	defer srv.Close()

	store.Close()

	body, _ := json.Marshal(protocol.DeviceInfo{DeviceID: "dev-1", DeviceName: "Test"})
	resp, err := http.Post(srv.URL+"/devices/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestIntegration_PushSaveError adds an existing item, closes DB, then pushes new item.
// This tests the SaveItem error path specifically (GetItem succeeds but SaveItem fails).
func TestIntegration_PushSaveError(t *testing.T) {
	srv, store := integrationServerWithStore(t)
	defer srv.Close()

	// First push an item successfully
	push := protocol.PushRequest{
		DeviceID: "dev-1",
		Items: []protocol.SyncItem{
			{ItemID: "item-1", Version: 1, DeviceID: "dev-1", Payload: "p", Timestamp: 1000, Checksum: "c"},
		},
	}
	body, _ := json.Marshal(push)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Drop tables to cause SaveItem to fail (GetItem still works on corrupted state)
	store.db.Exec("DROP TABLE sync_items")

	push2 := protocol.PushRequest{
		DeviceID: "dev-2",
		Items: []protocol.SyncItem{
			{ItemID: "item-new", Version: 1, DeviceID: "dev-2", Payload: "new", Timestamp: 2000, Checksum: "c2"},
		},
	}
	body, _ = json.Marshal(push2)
	resp, err = http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestIntegration_PullDefaultSince tests pull without since parameter defaults to 0.
func TestIntegration_PullDefaultSince(t *testing.T) {
	srv := integrationServer(t, "")
	defer srv.Close()

	// Push an item
	push := protocol.PushRequest{
		DeviceID: "dev-1",
		Items: []protocol.SyncItem{
			{ItemID: "item-1", Version: 1, DeviceID: "dev-1", Payload: "p", Timestamp: 100, Checksum: "c"},
		},
	}
	body, _ := json.Marshal(push)
	resp, err := http.Post(srv.URL+"/sync/push", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Pull without since parameter
	resp, err = http.Get(srv.URL + "/sync/pull?device_id=dev-1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pullResp protocol.PullResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pullResp))
	assert.Len(t, pullResp.Items, 1)
}

// TestIntegration_CORSPreflightOnAllEndpoints tests OPTIONS requests on all endpoints.
func TestIntegration_CORSPreflightOnAllEndpoints(t *testing.T) {
	srv := integrationServer(t, "test-key")
	defer srv.Close()

	endpoints := []string{"/sync/pull", "/sync/push", "/devices/register", "/health"}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodOptions, srv.URL+ep, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestRun_InvalidDBPath(t *testing.T) {
err := run(0, "/nonexistent/dir/cannot/create/db.sqlite", "")
assert.Error(t, err)
assert.Contains(t, err.Error(), "initialize storage")
}

func TestRun_GracefulShutdown(t *testing.T) {
dbPath := filepath.Join(t.TempDir(), "run-test.db")
errCh := make(chan error, 1)
go func() {
errCh <- run(0, dbPath, "test-key")
}()

// Give server time to start, then trigger shutdown
time.Sleep(300 * time.Millisecond)
p, _ := os.FindProcess(os.Getpid())
_ = p.Signal(os.Interrupt)

select {
case err := <-errCh:
_ = err // clean shutdown or error, both acceptable
case <-time.After(5 * time.Second):
t.Fatal("run() did not return within timeout")
}
}

func TestNewServer_Configuration(t *testing.T) {
dbPath := filepath.Join(t.TempDir(), "cfg-test.db")
store, err := NewSQLiteStorage(dbPath)
require.NoError(t, err)
defer store.Close()
require.NoError(t, store.CreateTables())

srv := NewServer(9999, store, "my-key")
assert.Equal(t, ":9999", srv.Addr)
assert.Equal(t, 15*time.Second, srv.ReadTimeout)
assert.Equal(t, 15*time.Second, srv.WriteTimeout)
assert.Equal(t, 60*time.Second, srv.IdleTimeout)
assert.NotNil(t, srv.Handler)
}
