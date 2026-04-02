//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"github.com/zeropass/zeropass/core/vault/health"
)

//export ZPAnalyzeHealth
func ZPAnalyzeHealth(handle C.long) (res C.ZPResult) {
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

	items, err := s.mgr.AllItems()
	if err != nil {
		return errorResult(err)
	}

	a := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := a.Analyze(items)
	return okJSON(report)
}
