// Package server provides HTTP handlers for the ZeroPass sync server.
package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

// Storage is the interface the sync handler uses to persist data.
type Storage interface {
	SaveItem(item protocol.SyncItem) error
	GetItemsSince(since int64) ([]protocol.SyncItem, error)
	GetItem(itemID string) (*protocol.SyncItem, error)
	RegisterDevice(device protocol.DeviceInfo) error
	UpdateDeviceSync(deviceID string, timestamp int64) error
}

// Handler holds the sync server's HTTP handler state.
type Handler struct {
	store  Storage
	apiKey string // empty = no auth required
}

// NewHandler creates a new sync handler backed by the given storage.
func NewHandler(store Storage, apiKey string) *Handler {
	return &Handler{store: store, apiKey: apiKey}
}

// RegisterRoutes adds all sync endpoints to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/sync/pull", h.withCORS(h.withAuth(h.handlePull)))
	mux.HandleFunc("/sync/push", h.withCORS(h.withAuth(h.handlePush)))
	mux.HandleFunc("/devices/register", h.withCORS(h.withAuth(h.handleRegister)))
	mux.HandleFunc("/health", h.withCORS(h.handleHealth))
}

// handlePull serves GET /sync/pull?since=<ts>&device_id=<id>.
func (h *Handler) handlePull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		h.writeError(w, http.StatusBadRequest, "missing device_id")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var since int64
	if sinceStr != "" {
		var err error
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid since parameter")
			return
		}
	}

	items, err := h.store.GetItemsSince(since)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to fetch items")
		return
	}
	if items == nil {
		items = []protocol.SyncItem{}
	}

	_ = h.store.UpdateDeviceSync(deviceID, time.Now().UnixNano())

	h.writeJSON(w, http.StatusOK, protocol.PullResponse{
		Items:      items,
		ServerTime: time.Now().UnixNano(),
	})
}

// handlePush serves POST /sync/push.
func (h *Handler) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req protocol.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DeviceID == "" {
		h.writeError(w, http.StatusBadRequest, "missing device_id")
		return
	}

	var accepted []string
	var conflicts []protocol.ConflictInfo

	for _, item := range req.Items {
		existing, err := h.store.GetItem(item.ItemID)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "storage error")
			return
		}

		if existing != nil && existing.Timestamp >= item.Timestamp && existing.DeviceID != item.DeviceID {
			// Conflict: server has a newer-or-equal version from another device.
			conflicts = append(conflicts, protocol.ConflictInfo{
				ItemID:     item.ItemID,
				ServerItem: *existing,
				Message:    "server has newer or equal version",
			})
			continue
		}

		item.DeviceID = req.DeviceID
		if err := h.store.SaveItem(item); err != nil {
			h.writeError(w, http.StatusInternalServerError, "failed to save item")
			return
		}
		accepted = append(accepted, item.ItemID)
	}

	if accepted == nil {
		accepted = []string{}
	}
	if conflicts == nil {
		conflicts = []protocol.ConflictInfo{}
	}

	_ = h.store.UpdateDeviceSync(req.DeviceID, time.Now().UnixNano())

	h.writeJSON(w, http.StatusOK, protocol.PushResponse{
		Accepted:   accepted,
		Conflicts:  conflicts,
		ServerTime: time.Now().UnixNano(),
	})
}

// handleRegister serves POST /devices/register.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var info protocol.DeviceInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if info.DeviceID == "" || info.DeviceName == "" {
		h.writeError(w, http.StatusBadRequest, "device_id and device_name are required")
		return
	}

	if err := h.store.RegisterDevice(info); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to register device")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// handleHealth serves GET /health.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// withAuth wraps a handler with Bearer token authentication.
func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.apiKey == "" {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !crypto.SecureCompareStrings(auth, "Bearer "+h.apiKey) {
			h.writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

// withCORS handles OPTIONS requests without advertising permissive browser CORS
// defaults. Native clients do not require CORS, and wildcard origins are unsafe
// as a default posture.
func (h *Handler) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// writeJSON encodes v as JSON and writes it to w.
func (h *Handler) writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// writeError writes a JSON error response.
func (h *Handler) writeError(w http.ResponseWriter, code int, msg string) {
	h.writeJSON(w, code, protocol.ErrorResponse{Error: msg})
}
