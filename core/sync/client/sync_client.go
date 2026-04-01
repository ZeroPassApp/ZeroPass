// Package client provides a sync client for the ZeroPass sync protocol.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/zeropass/zeropass/core/sync/conflict"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// SyncClient communicates with a ZeroPass sync server.
type SyncClient struct {
	mu            sync.Mutex
	serverURL     string
	deviceID      string
	apiKey        string
	httpClient    *http.Client
	lastSyncTime  int64
	resolver      *conflict.Resolver
}

// Option configures a SyncClient.
type Option func(*SyncClient)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(sc *SyncClient) { sc.httpClient = c }
}

// WithAPIKey sets the API key for server authentication.
func WithAPIKey(key string) Option {
	return func(sc *SyncClient) { sc.apiKey = key }
}

// WithLastSyncTime initialises the last-sync timestamp (e.g. from persistence).
func WithLastSyncTime(ts int64) Option {
	return func(sc *SyncClient) { sc.lastSyncTime = ts }
}

// NewSyncClient creates a SyncClient for the given server and device.
func NewSyncClient(serverURL, deviceID string, opts ...Option) *SyncClient {
	sc := &SyncClient{
		serverURL: serverURL,
		deviceID:  deviceID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		resolver: conflict.NewResolver(),
	}
	for _, o := range opts {
		o(sc)
	}
	return sc
}

// LastSyncTime returns the timestamp of the last successful sync.
func (sc *SyncClient) LastSyncTime() int64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.lastSyncTime
}

// ConflictHistory returns the resolver's conflict history.
func (sc *SyncClient) ConflictHistory() []conflict.ConflictRecord {
	return sc.resolver.History()
}

// Register registers this device with the sync server.
func (sc *SyncClient) Register(deviceName string) error {
	info := protocol.DeviceInfo{
		DeviceID:   sc.deviceID,
		DeviceName: deviceName,
	}
	body, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal device info: %w", err)
	}

	resp, err := sc.doRequest(http.MethodPost, "/devices/register", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("register device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return sc.readError(resp)
	}
	return nil
}

// Pull fetches items changed since the last sync timestamp.
func (sc *SyncClient) Pull() ([]protocol.SyncItem, error) {
	sc.mu.Lock()
	since := sc.lastSyncTime
	sc.mu.Unlock()

	params := url.Values{}
	params.Set("device_id", sc.deviceID)
	params.Set("since", strconv.FormatInt(since, 10))

	resp, err := sc.doRequest(http.MethodGet, "/sync/pull?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("pull: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, sc.readError(resp)
	}

	var pr protocol.PullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode pull response: %w", err)
	}

	return pr.Items, nil
}

// Push sends local changes to the server.
func (sc *SyncClient) Push(items []protocol.SyncItem) (*protocol.PushResponse, error) {
	req := protocol.PushRequest{
		DeviceID: sc.deviceID,
		Items:    items,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal push request: %w", err)
	}

	resp, err := sc.doRequest(http.MethodPost, "/sync/push", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("push: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, sc.readError(resp)
	}

	var pr protocol.PushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode push response: %w", err)
	}

	return &pr, nil
}

// Sync performs a full sync cycle: pull, resolve conflicts, push.
// localItems are the locally-changed items that need to be pushed.
func (sc *SyncClient) Sync(localItems []protocol.SyncItem) (*protocol.SyncResult, error) {
	// 1. Pull remote changes.
	remoteItems, err := sc.Pull()
	if err != nil {
		return nil, fmt.Errorf("sync pull: %w", err)
	}

	// 2. Build remote index for conflict detection.
	remoteByID := make(map[string]protocol.SyncItem, len(remoteItems))
	for _, ri := range remoteItems {
		remoteByID[ri.ItemID] = ri
	}

	// 3. Resolve conflicts and build the push list.
	var toPush []protocol.SyncItem
	conflictCount := 0
	for _, local := range localItems {
		if remote, exists := remoteByID[local.ItemID]; exists {
			winner := sc.resolver.Resolve(local, remote)
			if winner.DeviceID == local.DeviceID {
				// Local wins — push it.
				toPush = append(toPush, local)
			}
			conflictCount++
			// Remove from remote set so we don't double-count.
			delete(remoteByID, local.ItemID)
		} else {
			// No conflict — push local item.
			toPush = append(toPush, local)
		}
	}

	// 4. Push resolved items.
	pushed := 0
	if len(toPush) > 0 {
		pushResp, err := sc.Push(toPush)
		if err != nil {
			return nil, fmt.Errorf("sync push: %w", err)
		}
		pushed = len(pushResp.Accepted)
		conflictCount += len(pushResp.Conflicts)
	}

	// 5. Update last sync time.
	sc.mu.Lock()
	sc.lastSyncTime = time.Now().UnixNano()
	sc.mu.Unlock()

	return &protocol.SyncResult{
		Pulled:    len(remoteItems),
		Pushed:    pushed,
		Conflicts: conflictCount,
	}, nil
}

// doRequest builds and executes an HTTP request against the sync server.
func (sc *SyncClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	u := sc.serverURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if sc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	}
	return sc.httpClient.Do(req)
}

// readError reads an error response body and returns it as an error.
func (sc *SyncClient) readError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	var errResp protocol.ErrorResponse
	if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
		return fmt.Errorf("server error %d: %s", resp.StatusCode, errResp.Error)
	}
	return fmt.Errorf("server error %d: %s", resp.StatusCode, string(data))
}
