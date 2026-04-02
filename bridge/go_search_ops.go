//go:build cgo

package main

import "fmt"

func zpSearch(handle int64, query string) goResult {
	if query == "" {
		return okJSONGo([]string{})
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
	if s.idx == nil {
		return errorResultGo(fmt.Errorf("search index unavailable"))
	}
	ids, err := s.idx.Search(query)
	if err != nil {
		return errorResultGo(err)
	}
	return okJSONGo(ids)
}
