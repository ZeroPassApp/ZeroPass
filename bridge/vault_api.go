//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/sync/client"
	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/version"
)

func openIndexBestEffort(path string) (*index.Index, error) {
	idx, err := index.Open(path)
	if err == nil {
		return idx, nil
	}
	// Index is derived data. If the file is corrupted (often due to power loss or
	// abrupt termination), delete and recreate.
	msg := err.Error()
	if strings.Contains(msg, "file is not a database") || strings.Contains(msg, "malformed") {
		_ = os.Remove(path)
		return index.Open(path)
	}
	return nil, err
}

func ensureManagersUnlocked(s *vaultSession) error {
	if err := ensureOpenLocked(s); err != nil {
		return err
	}
	if s.mgr != nil {
		return nil
	}
	if s.vault == nil {
		return fmt.Errorf("vault missing")
	}
	if s.vault.IsLocked() {
		return fmt.Errorf("vault is locked")
	}
	idx, err := openIndexBestEffort(s.vault.IndexPath())
	if err != nil {
		// Degrade gracefully: allow CRUD even if the index can't be opened.
		s.idx = nil
		s.mgr = item.NewManager(s.vault.ItemsPath(), s.vault.VaultKey, nil)
		s.verMgr = version.NewManager(s.vault.ItemsPath(), s.vault.VaultKey, s.vault.Config().MaxVersions)
		return nil
	}
	s.idx = idx
	s.mgr = item.NewManager(s.vault.ItemsPath(), s.vault.VaultKey, idx)
	s.verMgr = version.NewManager(s.vault.ItemsPath(), s.vault.VaultKey, s.vault.Config().MaxVersions)
	if err := s.idx.EnsureCurrentPolicy(s.mgr.AllItems); err != nil {
		_ = s.idx.Close()
		s.idx = nil
		s.mgr = nil
		s.verMgr = nil
		return fmt.Errorf("initialize search index: %w", err)
	}
	return nil
}

func closeIndex(s *vaultSession) {
	if s.idx != nil {
		_ = s.idx.Close()
		s.idx = nil
	}
	s.mgr = nil
}

//export ZPCreateVault
func ZPCreateVault(path *C.char, masterPassword *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	vaultPath := goString(path)
	pw := goString(masterPassword)

	lockFile, err := acquireVaultLock(vaultPath, true)
	if err != nil {
		return errorResult(err)
	}
	registered := false
	defer func() {
		if !registered {
			releaseVaultLock(lockFile)
		}
	}()

	cfg := store.DefaultConfig()
	v, cr, err := store.Create(pw, vaultPath, cfg)
	if err != nil {
		return errorResult(err)
	}
	v.DisableAutoLock()

	idx, _ := openIndexBestEffort(v.IndexPath())
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)
	verMgr := version.NewManager(v.ItemsPath(), v.VaultKey, v.Config().MaxVersions)

	sc, scfg := loadSyncClient(vaultPath)

	s := &vaultSession{vault: v, mgr: mgr, idx: idx, verMgr: verMgr, lockFile: lockFile, syncCli: sc, syncCfg: scfg}
	h := registerSession(s)
	registered = true

	return okJSON(map[string]any{"handle": h, "mnemonic": cr.Mnemonic})
}

//export ZPOpenVault
func ZPOpenVault(path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	vaultPath := goString(path)
	lockFile, err := acquireVaultLock(vaultPath, false)
	if err != nil {
		return errorResult(err)
	}
	registered := false
	defer func() {
		if !registered {
			releaseVaultLock(lockFile)
		}
	}()

	v, err := store.Open(vaultPath)
	if err != nil {
		return errorResult(err)
	}
	// Swift is responsible for auto-locking; disable Go-side timer.
	v.DisableAutoLock()

	sc, scfg := loadSyncClient(vaultPath)
	_ = sc

	s := &vaultSession{vault: v, lockFile: lockFile, syncCli: sc, syncCfg: scfg}
	h := registerSession(s)
	registered = true
	return okJSON(map[string]any{"handle": h})
}

//export ZPUnlock
func ZPUnlock(handle C.long, masterPassword *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	pw := goString(masterPassword)
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}

	if err := s.vault.Unlock(pw); err != nil {
		return errorResult(err)
	}
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

//export ZPUnlockWithRecovery
func ZPUnlockWithRecovery(handle C.long, mnemonic *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	m := goString(mnemonic)
	if m == "" {
		return errorResult(fmt.Errorf("mnemonic must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newMnemonic, err := s.vault.UnlockWithRecovery(m)
	if err != nil {
		return errorResult(err)
	}
	if err := ensureManagersUnlocked(s); err != nil {
		// Still return the mnemonic so the client can show it
		return okJSON(map[string]any{"newMnemonic": newMnemonic, "warning": err.Error()})
	}
	return okJSON(map[string]any{"newMnemonic": newMnemonic})
}

//export ZPUnlockWithKey
func ZPUnlockWithKey(handle C.long, vaultKeyBase64 *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	keyB64 := goString(vaultKeyBase64)
	if keyB64 == "" {
		return errorResult(fmt.Errorf("vault key must not be empty"))
	}
	vk, err := encoding.Base64StdDecode(keyB64)
	if err != nil {
		return errorResult(fmt.Errorf("decode vault key: %w", err))
	}
	defer crypto.ZeroBytes(vk)

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}

	if err := s.vault.UnlockWithKey(vk); err != nil {
		return errorResult(err)
	}
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

//export ZPLock
func ZPLock(handle C.long) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}

	// Close the index DB before encrypting the file.
	closeIndex(s)
	s.vault.Lock()
	return okNoData()
}

//export ZPCloseVault
func ZPCloseVault(handle C.long) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.Load() {
		return okNoData()
	}
	s.closed.Store(true)

	closeIndex(s)
	s.vault.Lock()
	releaseVaultLock(s.lockFile)
	s.vault = nil
	removeSession(h)
	return okNoData()
}

//export ZPIsLocked
func ZPIsLocked(handle C.long) (ret C.int) {
	defer func() {
		if recover() != nil {
			ret = 1
		}
	}()
	id := int64(handle)
	s, err := getSession(id)
	if err != nil || s == nil || s.vault == nil {
		return 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return 1
	}
	if s.vault.IsLocked() {
		return 1
	}
	return 0
}

//export ZPChangeMasterPassword
func ZPChangeMasterPassword(handle C.long, oldPassword *C.char, newPassword *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	oldPw := goString(oldPassword)
	newPw := goString(newPassword)

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}

	if err := s.vault.ChangeMasterPassword(oldPw, newPw); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

// --- sync config persistence (phase 1) ---

type syncConfig struct {
	ServerURL    string `json:"server_url"`
	DeviceID     string `json:"device_id"`
	APIKey       string `json:"api_key,omitempty"`
	LastSyncTime int64  `json:"last_sync_time,omitempty"`
}

const syncConfigFileName = "sync.json"

func loadSyncClient(vaultPath string) (*client.SyncClient, *syncConfig) {
	fp := filepath.Join(vaultPath, syncConfigFileName)
	b, err := os.ReadFile(fp)
	if err != nil {
		return nil, nil
	}
	var cfg syncConfig
	if json.Unmarshal(b, &cfg) != nil {
		return nil, nil
	}
	if cfg.ServerURL == "" || cfg.DeviceID == "" {
		return nil, &cfg
	}
	opts := []client.Option{client.WithLastSyncTime(cfg.LastSyncTime)}
	cfg.APIKey = ""
	return client.NewSyncClient(cfg.ServerURL, cfg.DeviceID, opts...), &cfg
}
