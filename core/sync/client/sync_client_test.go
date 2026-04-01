package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// mockServer creates an httptest.Server that simulates the sync server.
func mockServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Device registration.
	mux.HandleFunc("/devices/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var info protocol.DeviceInfo
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			writeJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "bad request"})
			return
		}
		if info.DeviceID == "" || info.DeviceName == "" {
			writeJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "missing fields"})
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	// Pull.
	mux.HandleFunc("/sync/pull", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := protocol.PullResponse{
			Items: []protocol.SyncItem{
				{
					ItemID:    "item-remote-1",
					Version:   1,
					DeviceID:  "device-other",
					Payload:   "encrypted-remote",
					Timestamp: 5000,
					Checksum:  "abc",
					Deleted:   false,
				},
			},
			ServerTime: time.Now().UnixNano(),
		}
		writeJSON(w, http.StatusOK, resp)
	})

	// Push.
	mux.HandleFunc("/sync/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req protocol.PushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "bad body"})
			return
		}
		accepted := make([]string, 0, len(req.Items))
		for _, it := range req.Items {
			accepted = append(accepted, it.ItemID)
		}
		writeJSON(w, http.StatusOK, protocol.PushResponse{
			Accepted:   accepted,
			Conflicts:  nil,
			ServerTime: time.Now().UnixNano(),
		})
	})

	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func TestRegister_Success(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	err := c.Register("My Laptop")
	require.NoError(t, err)
}

func TestRegister_EmptyName(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	err := c.Register("")
	require.Error(t, err)
}

func TestPull_ReturnsItems(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	items, err := c.Pull()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "item-remote-1", items[0].ItemID)
	assert.Equal(t, "encrypted-remote", items[0].Payload)
}

func TestPush_AcceptsItems(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	resp, err := c.Push([]protocol.SyncItem{
		{ItemID: "local-1", Version: 1, DeviceID: "dev-1", Payload: "enc", Timestamp: 1000, Checksum: "x"},
		{ItemID: "local-2", Version: 1, DeviceID: "dev-1", Payload: "enc", Timestamp: 1000, Checksum: "y"},
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"local-1", "local-2"}, resp.Accepted)
	assert.Empty(t, resp.Conflicts)
}

func TestSync_FullCycle(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")

	localItems := []protocol.SyncItem{
		{ItemID: "local-only", Version: 1, DeviceID: "dev-1", Payload: "enc", Timestamp: 9000, Checksum: "z"},
	}

	result, err := c.Sync(localItems)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Pulled)
	assert.Equal(t, 1, result.Pushed)
	assert.Equal(t, 0, result.Conflicts)
	assert.Greater(t, c.LastSyncTime(), int64(0))
}

func TestSync_WithConflict_LocalWins(t *testing.T) {
	// Custom server that returns an item with the same ID as local but older timestamp.
	mux := http.NewServeMux()
	mux.HandleFunc("/sync/pull", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, protocol.PullResponse{
			Items: []protocol.SyncItem{
				{ItemID: "shared-1", Version: 1, DeviceID: "dev-2", Payload: "remote-enc", Timestamp: 1000, Checksum: "r"},
			},
			ServerTime: time.Now().UnixNano(),
		})
	})
	mux.HandleFunc("/sync/push", func(w http.ResponseWriter, r *http.Request) {
		var req protocol.PushRequest
		json.NewDecoder(r.Body).Decode(&req) //nolint:errcheck
		accepted := make([]string, 0, len(req.Items))
		for _, it := range req.Items {
			accepted = append(accepted, it.ItemID)
		}
		writeJSON(w, http.StatusOK, protocol.PushResponse{
			Accepted:   accepted,
			ServerTime: time.Now().UnixNano(),
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	local := []protocol.SyncItem{
		{ItemID: "shared-1", Version: 2, DeviceID: "dev-1", Payload: "local-enc", Timestamp: 5000, Checksum: "l"},
	}

	result, err := c.Sync(local)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Conflicts)
	assert.Equal(t, 1, result.Pushed, "local winner should be pushed")
}

func TestSync_WithConflict_RemoteWins(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sync/pull", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, protocol.PullResponse{
			Items: []protocol.SyncItem{
				{ItemID: "shared-1", Version: 2, DeviceID: "dev-2", Payload: "remote-enc", Timestamp: 9000, Checksum: "r"},
			},
			ServerTime: time.Now().UnixNano(),
		})
	})
	mux.HandleFunc("/sync/push", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, protocol.PushResponse{ServerTime: time.Now().UnixNano()})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	local := []protocol.SyncItem{
		{ItemID: "shared-1", Version: 1, DeviceID: "dev-1", Payload: "local-enc", Timestamp: 1000, Checksum: "l"},
	}

	result, err := c.Sync(local)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Conflicts)
	assert.Equal(t, 0, result.Pushed, "remote winner means nothing to push")
}

func TestAPIKeyHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, protocol.PullResponse{ServerTime: time.Now().UnixNano()})
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1", WithAPIKey("secret-key"))
	_, err := c.Pull()
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-key", gotAuth)
}

func TestServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "db down"})
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	_, err := c.Pull()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewSyncClient("http://localhost", "dev-1", WithHTTPClient(custom))
	assert.Equal(t, custom, c.httpClient)
}

func TestWithLastSyncTime(t *testing.T) {
	c := NewSyncClient("http://localhost", "dev-1", WithLastSyncTime(42))
	assert.Equal(t, int64(42), c.LastSyncTime())
}

func TestConflictHistory(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	assert.Empty(t, c.ConflictHistory())
}

func TestPush_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("plain error"))
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	_, err := c.Push([]protocol.SyncItem{{ItemID: "i1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestRegister_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "fail"})
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	err := c.Register("test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fail")
}

func TestPull_InvalidServerURL(t *testing.T) {
	c := NewSyncClient("http://[::1]:namedport", "dev-1")
	_, err := c.Pull()
	require.Error(t, err)
}

func TestPush_InvalidServerURL(t *testing.T) {
	c := NewSyncClient("http://[::1]:namedport", "dev-1")
	_, err := c.Push([]protocol.SyncItem{{ItemID: "i1"}})
	require.Error(t, err)
}

func TestRegister_InvalidServerURL(t *testing.T) {
	c := NewSyncClient("http://[::1]:namedport", "dev-1")
	err := c.Register("test")
	require.Error(t, err)
}

func TestPull_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json{{{"))
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	_, err := c.Pull()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode pull response")
}

func TestPush_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer srv.Close()

	c := NewSyncClient(srv.URL, "dev-1")
	_, err := c.Push([]protocol.SyncItem{{ItemID: "i1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode push response")
}
