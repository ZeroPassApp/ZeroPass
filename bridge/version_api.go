//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"fmt"
)

//export ZPGetVersionHistory
func ZPGetVersionHistory(handle C.long, itemID *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	id := goString(itemID)
	if id == "" {
		return errorResult(fmt.Errorf("item ID must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}
	if s.verMgr == nil {
		return errorResult(fmt.Errorf("version manager unavailable"))
	}

	history, err := s.verMgr.GetVersionHistory(id)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(history)
}

//export ZPRestoreVersion
func ZPRestoreVersion(handle C.long, itemID *C.char, version C.int) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	id := goString(itemID)
	if id == "" {
		return errorResult(fmt.Errorf("item ID must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}
	if s.verMgr == nil {
		return errorResult(fmt.Errorf("version manager unavailable"))
	}

	cur, err := s.mgr.GetItem(id)
	if err != nil {
		return errorResult(err)
	}
	if err := s.verMgr.SaveVersion(cur); err != nil {
		return errorResult(fmt.Errorf("snapshot current version: %w", err))
	}

	restored, err := s.verMgr.RestoreVersion(id, int(version))
	if err != nil {
		return errorResult(err)
	}
	restored.ID = id
	restored.CreatedAt = cur.CreatedAt
	restored.Version = cur.Version

	if err := s.mgr.UpdateItem(id, restored); err != nil {
		return errorResult(err)
	}
	return okJSON(restored)
}
