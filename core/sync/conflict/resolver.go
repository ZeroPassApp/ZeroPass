// Package conflict provides deterministic conflict resolution for the sync engine.
package conflict

import (
	"sync"

	"github.com/zeropass/zeropass/core/sync/protocol"
)

// Resolver determines which SyncItem wins when two devices modify the same item.
type Resolver struct {
	mu      sync.Mutex
	history []ConflictRecord
}

// ConflictRecord stores the result of a conflict resolution.
type ConflictRecord struct {
	ItemID string
	Winner protocol.SyncItem
	Loser  protocol.SyncItem
	Reason string
}

// NewResolver creates a new conflict resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve returns the winning SyncItem using a deterministic strategy:
//  1. Last-write-wins: higher timestamp wins.
//  2. If timestamps are equal, higher version number wins.
//  3. If still tied, the remote (server) item wins for determinism.
//
// The losing item is appended to conflict history.
func (r *Resolver) Resolve(local, remote protocol.SyncItem) protocol.SyncItem {
	winner, loser, reason := resolve(local, remote)

	r.mu.Lock()
	r.history = append(r.history, ConflictRecord{
		ItemID: local.ItemID,
		Winner: winner,
		Loser:  loser,
		Reason: reason,
	})
	r.mu.Unlock()

	return winner
}

// resolve performs the comparison without side effects.
func resolve(local, remote protocol.SyncItem) (winner, loser protocol.SyncItem, reason string) {
	switch {
	case local.Timestamp > remote.Timestamp:
		return local, remote, "local has later timestamp"
	case remote.Timestamp > local.Timestamp:
		return remote, local, "remote has later timestamp"
	// Timestamps equal — compare versions.
	case local.Version > remote.Version:
		return local, remote, "local has higher version"
	case remote.Version > local.Version:
		return remote, local, "remote has higher version"
	default:
		// Fully tied — server (remote) wins for determinism.
		return remote, local, "tie broken: server wins"
	}
}

// History returns a copy of all recorded conflict resolutions.
func (r *Resolver) History() []ConflictRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ConflictRecord, len(r.history))
	copy(out, r.history)
	return out
}

// ClearHistory removes all conflict history entries.
func (r *Resolver) ClearHistory() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history = nil
}
