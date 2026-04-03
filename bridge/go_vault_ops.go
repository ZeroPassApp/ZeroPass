//go:build cgo

package main

import (
	"fmt"

	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/version"
)

func zpCreateVault(vaultPath, masterPassword string) goResult {
	if vaultPath == "" {
		return errorResultGo(fmt.Errorf("vault path must not be empty"))
	}
	if masterPassword == "" {
		return errorResultGo(fmt.Errorf("master password must not be empty"))
	}

	lockFile, err := acquireVaultLock(vaultPath, true)
	if err != nil {
		return errorResultGo(err)
	}
	registered := false
	defer func() {
		if !registered {
			releaseVaultLock(lockFile)
		}
	}()

	cfg := store.DefaultConfig()
	v, cr, err := store.Create(masterPassword, vaultPath, cfg)
	if err != nil {
		return errorResultGo(err)
	}
	v.DisableAutoLock()

	idx, _ := openIndexBestEffort(v.IndexPath())
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)
	verMgr := version.NewManager(v.ItemsPath(), v.VaultKey, v.Config().MaxVersions)

	sc, scfg := loadSyncClient(vaultPath)

	s := &vaultSession{vault: v, mgr: mgr, idx: idx, verMgr: verMgr, lockFile: lockFile, syncCli: sc, syncCfg: scfg}
	h := registerSession(s)
	registered = true

	return okJSONGo(map[string]any{"handle": h, "mnemonic": cr.Mnemonic})
}

func zpOpenVault(vaultPath string) goResult {
	if vaultPath == "" {
		return errorResultGo(fmt.Errorf("vault path must not be empty"))
	}
	lockFile, err := acquireVaultLock(vaultPath, false)
	if err != nil {
		return errorResultGo(err)
	}
	registered := false
	defer func() {
		if !registered {
			releaseVaultLock(lockFile)
		}
	}()

	v, err := store.Open(vaultPath)
	if err != nil {
		return errorResultGo(err)
	}
	v.DisableAutoLock()

	sc, scfg := loadSyncClient(vaultPath)

	s := &vaultSession{vault: v, lockFile: lockFile, syncCli: sc, syncCfg: scfg}
	h := registerSession(s)
	registered = true
	return okJSONGo(map[string]any{"handle": h})
}

func zpUnlock(handle int64, masterPassword string) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResultGo(err)
	}

	if err := s.vault.Unlock(masterPassword); err != nil {
		return errorResultGo(err)
	}
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}
	return okNoDataGo()
}

func zpLock(handle int64) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResultGo(err)
	}

	closeIndex(s)
	s.vault.Lock()
	return okNoDataGo()
}

func zpCloseVault(handle int64) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.Load() {
		return okNoDataGo()
	}
	s.closed.Store(true)

	closeIndex(s)
	s.vault.Lock()
	releaseVaultLock(s.lockFile)
	s.vault = nil
	removeSession(handle)
	return okNoDataGo()
}

func zpIsLocked(handle int64) bool {
	s, err := getSession(handle)
	if err != nil || s == nil || s.vault == nil {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return true
	}
	return s.vault.IsLocked()
}

func zpChangeMasterPassword(handle int64, oldPw, newPw string) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResultGo(err)
	}

	if err := s.vault.ChangeMasterPassword(oldPw, newPw); err != nil {
		return errorResultGo(err)
	}
	return okNoDataGo()
}

func zpRebuildIndex(handle int64) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}
	if s.idx == nil {
		return errorResultGo(fmt.Errorf("search index unavailable"))
	}

	items, err := s.mgr.AllItems()
	if err != nil {
		return errorResultGo(err)
	}
	if err := s.idx.RebuildIndex(items); err != nil {
		return errorResultGo(err)
	}
	return okNoDataGo()
}

// ensure we keep index imported when only referenced through openIndexBestEffort
var _ = index.Open
