package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	syncserver "github.com/zeropass/zeropass/core/sync/server"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

type smokeHarness struct {
	t      *testing.T
	apiKey string
	server *httptest.Server
	store  *SQLiteStorage
}

func newSmokeHarness(t *testing.T, dbPath, apiKey string) *smokeHarness {
	t.Helper()
	store, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.CreateTables())

	mux := http.NewServeMux()
	syncserver.NewHandler(store, apiKey).RegisterRoutes(mux)

	return &smokeHarness{t: t, apiKey: apiKey, server: httptest.NewServer(mux), store: store}
}

func (h *smokeHarness) Close() {
	h.server.Close()
	require.NoError(h.t, h.store.Close())
}

func (h *smokeHarness) request(method, path string, body any, withAuth bool) *http.Response {
	h.t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(h.t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, h.server.URL+path, reader)
	require.NoError(h.t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if withAuth && h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(h.t, err)
	return resp
}

func (h *smokeHarness) requestInvalidJSON(path, payload string) *http.Response {
	h.t.Helper()
	req, err := http.NewRequest(http.MethodPost, h.server.URL+path, bytes.NewBufferString(payload))
	require.NoError(h.t, err)
	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(h.t, err)
	return resp
}

func decodeResponse[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var out T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func TestSmoke_RestartPersistenceAndConflict(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "smoke.db")
	h := newSmokeHarness(t, dbPath, "smoke-key")

	resp := h.request(http.MethodGet, "/sync/pull?device_id=peer&since=0", nil, false)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	resp = h.requestInvalidJSON("/sync/push", "{bad json")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	resp = h.request(http.MethodPost, "/devices/register", protocol.DeviceInfo{DeviceID: "dev-a", DeviceName: "Smoke A"}, true)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	pushResp := decodeResponse[protocol.PushResponse](t, h.request(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "dev-a",
		Items: []protocol.SyncItem{{
			ItemID:    "smoke-item",
			Version:   1,
			DeviceID:  "dev-a",
			Payload:   "payload-v1",
			Timestamp: 1000,
			Checksum:  "c1",
		}},
	}, true))
	assert.Equal(t, []string{"smoke-item"}, pushResp.Accepted)
	h.Close()

	h = newSmokeHarness(t, dbPath, "smoke-key")
	defer h.Close()

	resp = h.request(http.MethodGet, "/health", nil, false)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	pullResp := decodeResponse[protocol.PullResponse](t, h.request(http.MethodGet, "/sync/pull?device_id=dev-b&since=0", nil, true))
	require.Len(t, pullResp.Items, 1)
	assert.Equal(t, "smoke-item", pullResp.Items[0].ItemID)
	assert.Equal(t, "payload-v1", pullResp.Items[0].Payload)

	conflictResp := decodeResponse[protocol.PushResponse](t, h.request(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "dev-b",
		Items: []protocol.SyncItem{{
			ItemID:    "smoke-item",
			Version:   1,
			DeviceID:  "dev-b",
			Payload:   "older-payload",
			Timestamp: 500,
			Checksum:  "c-old",
		}},
	}, true))
	assert.Empty(t, conflictResp.Accepted)
	require.Len(t, conflictResp.Conflicts, 1)
	assert.Equal(t, "smoke-item", conflictResp.Conflicts[0].ItemID)
}

func TestSmoke_NoAuthPreviewMode(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "preview.db")
	h := newSmokeHarness(t, dbPath, "")
	defer h.Close()

	pushResp := decodeResponse[protocol.PushResponse](t, h.request(http.MethodPost, "/sync/push", protocol.PushRequest{
		DeviceID: "preview-device",
		Items: []protocol.SyncItem{{
			ItemID:    "preview-item",
			Version:   1,
			DeviceID:  "preview-device",
			Payload:   "preview-payload",
			Timestamp: 1234,
			Checksum:  "preview-checksum",
		}},
	}, false))
	assert.Equal(t, []string{"preview-item"}, pushResp.Accepted)

	pullResp := decodeResponse[protocol.PullResponse](t, h.request(http.MethodGet, "/sync/pull?device_id=preview-peer&since=0", nil, false))
	require.Len(t, pullResp.Items, 1)
	assert.Equal(t, "preview-item", pullResp.Items[0].ItemID)
	assert.Equal(t, "preview-payload", pullResp.Items[0].Payload)
}