//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"fmt"
)

//export ZPSearch
func ZPSearch(handle C.long, query *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	q := goString(query)
	if q == "" {
		return okJSON([]string{})
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
	if s.idx == nil {
		return errorResult(fmt.Errorf("search index unavailable"))
	}

	ids, err := s.idx.Search(q)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(ids)
}

//export ZPRebuildIndex
func ZPRebuildIndex(handle C.long) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}
	if s.idx == nil {
		return errorResult(fmt.Errorf("search index unavailable"))
	}

	items, err := s.mgr.AllItems()
	if err != nil {
		return errorResult(err)
	}
	if err := s.idx.RebuildIndex(items); err != nil {
		return errorResult(err)
	}
	return okNoData()
}
