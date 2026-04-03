//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"fmt"
)

//export ZPRegenerateRecovery
func ZPRegenerateRecovery(handle C.long) (res C.ZPResult) {
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

	if s.vault.IsLocked() {
		return errorResult(fmt.Errorf("vault is locked"))
	}
	mnemonic, err := s.vault.RegenerateRecovery()
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"mnemonic": mnemonic})
}
