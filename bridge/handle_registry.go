//go:build cgo

package main

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/zeropass/zeropass/core/sync/client"
	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/version"
)

type vaultSession struct {
	mu     sync.Mutex
	closed atomic.Bool
	vault  *store.Vault
	mgr    *item.Manager
	idx    *index.Index
	verMgr *version.Manager

	lockFile *os.File

	syncCfg *syncConfig
	syncCli *client.SyncClient
}

func ensureOpenLocked(s *vaultSession) error {
	if s == nil || s.closed.Load() || s.vault == nil {
		return errSessionClosed
	}
	return nil
}

var (
	hMu     sync.RWMutex
	handles = make(map[int64]*vaultSession)
	nextID  int64
)

func registerSession(s *vaultSession) int64 {
	id := atomic.AddInt64(&nextID, 1)
	hMu.Lock()
	handles[id] = s
	hMu.Unlock()
	return id
}

func getSession(id int64) (*vaultSession, error) {
	hMu.RLock()
	s := handles[id]
	hMu.RUnlock()
	if s == nil {
		return nil, errHandleNotFound
	}
	if s.closed.Load() {
		return nil, errSessionClosed
	}
	return s, nil
}

func removeSession(id int64) {
	hMu.Lock()
	delete(handles, id)
	hMu.Unlock()
}
