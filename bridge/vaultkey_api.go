//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"fmt"

	"github.com/zeropass/zeropass/core/crypto/encoding"
)

//export ZPGetVaultKey
func ZPGetVaultKey(handle C.long) (res C.ZPResult) {
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
	if s.vault == nil || s.vault.IsLocked() {
		return errorResult(fmt.Errorf("vault is locked"))
	}

	vk, err := s.vault.VaultKey()
	if err != nil {
		return errorResult(err)
	}

	return okJSON(map[string]any{"vaultKeyBase64": encoding.Base64StdEncode(vk)})
}
