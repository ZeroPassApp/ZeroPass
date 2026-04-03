//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import "unsafe"

func capiFromResult(r C.ZPResult) goResult {
	defer ZPFreeResult(r)
	out := goResult{code: int(r.code)}
	if r.data != nil {
		out.data = C.GoString(r.data)
	}
	if r.error != nil {
		out.err = C.GoString(r.error)
	}
	return out
}

func withCStr(s string, fn func(*C.char) C.ZPResult) goResult {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return capiFromResult(fn(cs))
}

func withCStr2(a, b string, fn func(*C.char, *C.char) C.ZPResult) goResult {
	ca := C.CString(a)
	cb := C.CString(b)
	defer C.free(unsafe.Pointer(ca))
	defer C.free(unsafe.Pointer(cb))
	return capiFromResult(fn(ca, cb))
}

func zpCCreateVault(vaultPath, masterPassword string) goResult {
	cPath := C.CString(vaultPath)
	cPw := C.CString(masterPassword)
	defer C.free(unsafe.Pointer(cPath))
	defer C.free(unsafe.Pointer(cPw))
	return capiFromResult(ZPCreateVault(cPath, cPw))
}

func zpCOpenVault(vaultPath string) goResult {
	return withCStr(vaultPath, func(p *C.char) C.ZPResult { return ZPOpenVault(p) })
}

func zpCUnlock(handle int64, masterPassword string) goResult {
	cPw := C.CString(masterPassword)
	defer C.free(unsafe.Pointer(cPw))
	return capiFromResult(ZPUnlock(C.long(handle), cPw))
}

func zpCUnlockWithKey(handle int64, vaultKeyBase64 string) goResult {
	cKey := C.CString(vaultKeyBase64)
	defer C.free(unsafe.Pointer(cKey))
	return capiFromResult(ZPUnlockWithKey(C.long(handle), cKey))
}

func zpCLock(handle int64) goResult { return capiFromResult(ZPLock(C.long(handle))) }

func zpCCloseVault(handle int64) goResult { return capiFromResult(ZPCloseVault(C.long(handle))) }

func zpCIsLocked(handle int64) bool { return ZPIsLocked(C.long(handle)) == 1 }

func zpCListItems(handle int64, filterJSON string) goResult {
	if filterJSON == "" {
		return capiFromResult(ZPListItems(C.long(handle), nil))
	}
	cFilter := C.CString(filterJSON)
	defer C.free(unsafe.Pointer(cFilter))
	return capiFromResult(ZPListItems(C.long(handle), cFilter))
}

func zpCSearch(handle int64, query string) goResult {
	return withCStr(query, func(q *C.char) C.ZPResult { return ZPSearch(C.long(handle), q) })
}

func zpCGetItem(handle int64, id string) goResult {
	return withCStr(id, func(iid *C.char) C.ZPResult { return ZPGetItem(C.long(handle), iid) })
}

func zpCCreateItem(handle int64, itemJSON string) goResult {
	return withCStr(itemJSON, func(p *C.char) C.ZPResult { return ZPCreateItem(C.long(handle), p) })
}

func zpCUpdateItem(handle int64, id, itemJSON string) goResult {
	cID := C.CString(id)
	cPayload := C.CString(itemJSON)
	defer C.free(unsafe.Pointer(cID))
	defer C.free(unsafe.Pointer(cPayload))
	return capiFromResult(ZPUpdateItem(C.long(handle), cID, cPayload))
}

func zpCDeleteItem(handle int64, id string) goResult {
	return withCStr(id, func(iid *C.char) C.ZPResult { return ZPDeleteItem(C.long(handle), iid) })
}

func zpCGetVaultKey(handle int64) goResult { return capiFromResult(ZPGetVaultKey(C.long(handle))) }

func zpCGetVersionHistory(handle int64, id string) goResult {
	return withCStr(id, func(iid *C.char) C.ZPResult { return ZPGetVersionHistory(C.long(handle), iid) })
}

func zpCRestoreVersion(handle int64, id string, version int32) goResult {
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	return capiFromResult(ZPRestoreVersion(C.long(handle), cID, C.int(version)))
}

func zpCChangeMasterPassword(handle int64, oldPw, newPw string) goResult {
	return withCStr2(oldPw, newPw, func(o, n *C.char) C.ZPResult { return ZPChangeMasterPassword(C.long(handle), o, n) })
}

func zpCSyncSetup(handle int64, configJSON string) goResult {
	return withCStr(configJSON, func(cfg *C.char) C.ZPResult { return ZPSyncSetup(C.long(handle), cfg) })
}

func zpCGeneratePassword(length int32, optionsJSON string) goResult {
	if optionsJSON == "" {
		return capiFromResult(ZPGeneratePassword(C.int(length), nil))
	}
	return withCStr(optionsJSON, func(opts *C.char) C.ZPResult { return ZPGeneratePassword(C.int(length), opts) })
}

func zpCGeneratePassphrase(words int32, separator string) goResult {
	if separator == "" {
		return capiFromResult(ZPGeneratePassphrase(C.int(words), nil))
	}
	return withCStr(separator, func(sep *C.char) C.ZPResult { return ZPGeneratePassphrase(C.int(words), sep) })
}

func zpCScorePassword(password string) goResult {
	return withCStr(password, func(pw *C.char) C.ZPResult { return ZPScorePassword(pw) })
}
