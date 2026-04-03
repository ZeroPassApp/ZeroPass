//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	zpencoding "github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/sync/client"
	"github.com/zeropass/zeropass/core/sync/protocol"
	"github.com/zeropass/zeropass/core/vault/types"
)

func syncConfigPath(vaultPath string) string {
	return filepath.Join(vaultPath, syncConfigFileName)
}

func persistSyncConfig(vaultPath string, cfg *syncConfig) error {
	if cfg == nil {
		return nil
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomicBytes(syncConfigPath(vaultPath), 0600, b)
}

func ensureSyncClient(s *vaultSession) error {
	if s.syncCli != nil {
		return nil
	}
	// Lazy load from disk.
	sc, cfg := loadSyncClient(s.vault.Path())
	s.syncCli = sc
	s.syncCfg = cfg
	if s.syncCli == nil {
		return fmt.Errorf("sync not configured")
	}
	return nil
}

//export ZPSyncSetup
func ZPSyncSetup(handle C.long, configJSON *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	payload := goString(configJSON)
	if payload == "" {
		return errorResult(fmt.Errorf("config JSON must not be empty"))
	}

	var cfg syncConfig
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		return errorResult(fmt.Errorf("parse config JSON: %w", err))
	}
	if cfg.ServerURL == "" || cfg.DeviceID == "" {
		return errorResult(fmt.Errorf("server_url and device_id are required"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}

	opts := []client.Option{client.WithLastSyncTime(cfg.LastSyncTime)}
	if cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKey(cfg.APIKey))
	}
	s.syncCli = client.NewSyncClient(cfg.ServerURL, cfg.DeviceID, opts...)
	s.syncCfg = &cfg

	if err := persistSyncConfig(s.vault.Path(), &cfg); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

//export ZPSyncRegister
func ZPSyncRegister(handle C.long, deviceName *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	name := goString(deviceName)
	if name == "" {
		return errorResult(fmt.Errorf("device name must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureOpenLocked(s); err != nil {
		return errorResult(err)
	}
	if err := ensureSyncClient(s); err != nil {
		return errorResult(err)
	}

	if err := s.syncCli.Register(name); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

//export ZPSyncPull
func ZPSyncPull(handle C.long) (res C.ZPResult) {
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
	if err := ensureSyncClient(s); err != nil {
		return errorResult(err)
	}

	items, err := s.syncCli.Pull()
	if err != nil {
		return errorResult(err)
	}
	return okJSON(items)
}

func collectLocalSyncItems(vaultPath string, since int64, deviceID string) ([]protocol.SyncItem, error) {
	itemsDir := filepath.Join(vaultPath, "items")
	entries, err := os.ReadDir(itemsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []protocol.SyncItem
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".versions.json") {
			continue
		}
		fp := filepath.Join(itemsDir, name)
		info, err := os.Stat(fp)
		if err != nil {
			continue
		}
		ts := info.ModTime().UnixNano()
		if since > 0 && ts <= since {
			continue
		}
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		var enc types.EncryptedItem
		if json.Unmarshal(data, &enc) != nil {
			continue
		}
		if enc.ID == "" {
			continue
		}
		chk := enc.Checksum
		if chk == "" {
			ct, err := zpencoding.Base64StdDecode(enc.Data)
			if err != nil {
				return nil, fmt.Errorf("decode item payload for checksum (%s): %w", fp, err)
			}
			sum := sha256.Sum256(ct)
			chk = fmt.Sprintf("%x", sum)
		}
		out = append(out, protocol.SyncItem{
			ItemID:    enc.ID,
			Version:   enc.Version,
			DeviceID:  deviceID,
			Payload:   enc.Data,
			Timestamp: ts,
			Checksum:  chk,
			Deleted:   false,
		})
	}
	return out, nil
}

//export ZPSyncPush
func ZPSyncPush(handle C.long) (res C.ZPResult) {
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
	if err := ensureSyncClient(s); err != nil {
		return errorResult(err)
	}
	if s.syncCfg == nil {
		return errorResult(fmt.Errorf("sync not configured"))
	}

	local, err := collectLocalSyncItems(s.vault.Path(), s.syncCfg.LastSyncTime, s.syncCfg.DeviceID)
	if err != nil {
		return errorResult(err)
	}

	resp, err := s.syncCli.Push(local)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(resp)
}

//export ZPSyncFull
func ZPSyncFull(handle C.long) (res C.ZPResult) {
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
	if err := ensureSyncClient(s); err != nil {
		return errorResult(err)
	}
	if s.syncCfg == nil {
		return errorResult(fmt.Errorf("sync not configured"))
	}

	local, err := collectLocalSyncItems(s.vault.Path(), s.syncCfg.LastSyncTime, s.syncCfg.DeviceID)
	if err != nil {
		return errorResult(err)
	}

	result, err := s.syncCli.Sync(local)
	if err != nil {
		return errorResult(err)
	}

	// Persist updated last-sync time.
	if s.syncCfg != nil {
		s.syncCfg.LastSyncTime = s.syncCli.LastSyncTime()
		_ = persistSyncConfig(s.vault.Path(), s.syncCfg)
	}

	return okJSON(map[string]any{
		"result":    result,
		"conflicts": s.syncCli.ConflictHistory(),
		"synced_at": time.Now().UTC(),
	})
}
