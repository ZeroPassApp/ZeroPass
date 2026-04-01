package conflict

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/sync/protocol"
)

func item(id string, version int, ts int64, device string, deleted bool) protocol.SyncItem {
	return protocol.SyncItem{
		ItemID:    id,
		Version:   version,
		DeviceID:  device,
		Payload:   "enc-payload",
		Timestamp: ts,
		Checksum:  "abc123",
		Deleted:   deleted,
	}
}

func TestResolve_RemoteLaterTimestamp(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 1000, "d1", false)
	remote := item("i1", 1, 2000, "d2", false)

	winner := r.Resolve(local, remote)
	assert.Equal(t, remote.DeviceID, winner.DeviceID)
	assert.Equal(t, int64(2000), winner.Timestamp)
}

func TestResolve_LocalLaterTimestamp(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 3000, "d1", false)
	remote := item("i1", 1, 2000, "d2", false)

	winner := r.Resolve(local, remote)
	assert.Equal(t, local.DeviceID, winner.DeviceID)
	assert.Equal(t, int64(3000), winner.Timestamp)
}

func TestResolve_SameTimestamp_HigherVersionWins(t *testing.T) {
	r := NewResolver()
	local := item("i1", 3, 1000, "d1", false)
	remote := item("i1", 2, 1000, "d2", false)

	winner := r.Resolve(local, remote)
	assert.Equal(t, local.DeviceID, winner.DeviceID)
	assert.Equal(t, 3, winner.Version)
}

func TestResolve_SameTimestamp_RemoteHigherVersion(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 1000, "d1", false)
	remote := item("i1", 5, 1000, "d2", false)

	winner := r.Resolve(local, remote)
	assert.Equal(t, remote.DeviceID, winner.DeviceID)
	assert.Equal(t, 5, winner.Version)
}

func TestResolve_TiedTimestampAndVersion_ServerWins(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 1000, "d1", false)
	remote := item("i1", 1, 1000, "d2", false)

	winner := r.Resolve(local, remote)
	assert.Equal(t, remote.DeviceID, winner.DeviceID, "server (remote) should win ties")
}

func TestResolve_Tombstone_LaterDeleteWins(t *testing.T) {
	r := NewResolver()
	local := item("i1", 2, 1000, "d1", false)
	remote := item("i1", 3, 2000, "d2", true)

	winner := r.Resolve(local, remote)
	assert.True(t, winner.Deleted, "deleted item with later timestamp should win")
	assert.Equal(t, remote.DeviceID, winner.DeviceID)
}

func TestResolve_Tombstone_EarlierDeleteLoses(t *testing.T) {
	r := NewResolver()
	local := item("i1", 2, 3000, "d1", false)
	remote := item("i1", 3, 2000, "d2", true)

	winner := r.Resolve(local, remote)
	assert.False(t, winner.Deleted, "non-deleted with later timestamp should win")
	assert.Equal(t, local.DeviceID, winner.DeviceID)
}

func TestHistory_RecordsConflicts(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 1000, "d1", false)
	remote := item("i1", 1, 2000, "d2", false)

	r.Resolve(local, remote)

	history := r.History()
	require.Len(t, history, 1)
	assert.Equal(t, "i1", history[0].ItemID)
	assert.Equal(t, remote.DeviceID, history[0].Winner.DeviceID)
	assert.Equal(t, local.DeviceID, history[0].Loser.DeviceID)
	assert.Equal(t, "remote has later timestamp", history[0].Reason)
}

func TestHistory_MultipleConflicts(t *testing.T) {
	r := NewResolver()
	r.Resolve(item("i1", 1, 1000, "d1", false), item("i1", 1, 2000, "d2", false))
	r.Resolve(item("i2", 1, 3000, "d1", false), item("i2", 1, 1000, "d2", false))
	r.Resolve(item("i3", 1, 1000, "d1", false), item("i3", 1, 1000, "d2", false))

	history := r.History()
	require.Len(t, history, 3)
	assert.Equal(t, "remote has later timestamp", history[0].Reason)
	assert.Equal(t, "local has later timestamp", history[1].Reason)
	assert.Equal(t, "tie broken: server wins", history[2].Reason)
}

func TestClearHistory(t *testing.T) {
	r := NewResolver()
	r.Resolve(item("i1", 1, 1000, "d1", false), item("i1", 1, 2000, "d2", false))
	require.Len(t, r.History(), 1)

	r.ClearHistory()
	assert.Empty(t, r.History())
}

func TestResolve_EqualTimestamp_EqualVersion_DifferentPayloads(t *testing.T) {
	r := NewResolver()
	local := item("i1", 1, 1000, "d1", false)
	local.Payload = "payload-a"
	remote := item("i1", 1, 1000, "d2", false)
	remote.Payload = "payload-b"

	winner := r.Resolve(local, remote)
	assert.Equal(t, remote.DeviceID, winner.DeviceID, "server wins deterministic tie")
	assert.Equal(t, "payload-b", winner.Payload)
}
