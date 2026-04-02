//go:build cgo

package main

import "fmt"

func zpGetVersionHistory(handle int64, id string) goResult {
	if id == "" {
		return errorResultGo(fmt.Errorf("item ID must not be empty"))
	}
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	h, err := s.verMgr.GetVersionHistory(id)
	if err != nil {
		return errorResultGo(err)
	}
	return okJSONGo(h)
}
