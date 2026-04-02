//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const (
	zpCodeOK         = 0
	zpCodeLocked     = 1
	zpCodeNotFound   = 2
	zpCodeAuthFailed = 3
	zpCodeInternal   = 4
	zpCodeBusy       = 5
)

var (
	errHandleNotFound = errors.New("handle not found")
	errSessionClosed  = errors.New("session closed")
	errVaultBusy      = errors.New("vault is already open")
)

func recoverToResult(out *C.ZPResult) {
	if r := recover(); r != nil {
		msg := fmt.Sprintf("panic: %v", r)
		// Keep stack traces out of the C boundary by default; they can be logged on the Swift side.
		_ = debug.Stack()
		*out = toCResult(goResult{code: zpCodeInternal, err: msg})
	}
}

func goString(p *C.char) string {
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

func codeForError(err error) int {
	if err == nil {
		return zpCodeOK
	}
	if errors.Is(err, errHandleNotFound) || errors.Is(err, errSessionClosed) {
		return zpCodeNotFound
	}
	if errors.Is(err, os.ErrNotExist) {
		return zpCodeNotFound
	}
	if errors.Is(err, errVaultBusy) {
		return zpCodeBusy
	}
	msg := err.Error()
	if strings.Contains(msg, "vault is locked") {
		return zpCodeLocked
	}
	if strings.Contains(msg, "invalid master password") ||
		strings.Contains(msg, "verify old password") ||
		strings.Contains(msg, "unlock vault") ||
		strings.Contains(msg, "invalid vault key") ||
		strings.Contains(msg, "server error 401") ||
		strings.Contains(msg, "unauthorized") {
		return zpCodeAuthFailed
	}
	return zpCodeInternal
}

func toCResult(r goResult) C.ZPResult {
	var data *C.char
	var errStr *C.char
	if r.data != "" {
		data = C.CString(r.data)
	}
	if r.err != "" {
		errStr = C.CString(r.err)
	}
	return C.ZPResult{data: data, error: errStr, code: C.int(r.code)}
}

func okNoData() C.ZPResult {
	return toCResult(okNoDataGo())
}

func okJSON(v any) C.ZPResult {
	return toCResult(okJSONGo(v))
}

func errorResult(err error) C.ZPResult {
	return toCResult(errorResultGo(err))
}
