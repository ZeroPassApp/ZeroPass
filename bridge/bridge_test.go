//go:build ignore

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/zeropass/zeropass/core/vault/types"
)

func cstr(s string) *C.char {
	cs := C.CString(s)
	return cs
}

func freeCStr(p *C.char) {
	if p == nil {
		return
	}
	C.free(unsafe.Pointer(p))
}

func readResult(r C.ZPResult) (code int, data string, errStr string) {
	defer ZPFreeResult(r)
	code = int(r.code)
	if r.data != nil {
		data = C.GoString(r.data)
	}
	if r.error != nil {
		errStr = C.GoString(r.error)
	}
	return
}

func requireOK(t *testing.T, r C.ZPResult) string {
	t.Helper()
	code, data, errStr := readResult(r)
	if code != 0 {
		t.Fatalf("expected ok, got code=%d err=%q data=%q", code, errStr, data)
	}
	if errStr != "" {
		t.Fatalf("expected no error string, got %q", errStr)
	}
	return data
}

func requireErr(t *testing.T, r C.ZPResult) (code int, errStr string) {
	t.Helper()
	code, _, errStr = readResult(r)
	if code == 0 {
		t.Fatalf("expected error, got ok")
	}
	if errStr == "" {
		t.Fatalf("expected error string")
	}
	return
}

func TestBridgeCreateCRUDSearchLockClose(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "hunter2"

	cPath := cstr(vaultPath)
	cPw := cstr(pw)
	defer freeCStr(cPath)
	defer freeCStr(cPw)

	createData := requireOK(t, ZPCreateVault(cPath, cPw))
	var created struct {
		Handle   int64  `json:"handle"`
		Mnemonic string `json:"mnemonic"`
	}
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}
	if created.Handle == 0 {
		t.Fatalf("expected non-zero handle")
	}
	if created.Mnemonic == "" {
		t.Fatalf("expected recovery mnemonic")
	}

	// Empty list
	listData := requireOK(t, ZPListItems(C.long(created.Handle), nil))
	var listed []*types.Item
	if err := json.Unmarshal([]byte(listData), &listed); err != nil {
		t.Fatalf("parse list JSON: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected 0 items, got %d", len(listed))
	}

	// Create item
	newItem := types.Item{Type: types.ItemTypeLogin, Name: "GitHub", Fields: map[string]string{"username": "devuser", "password": "secret"}, Tags: []string{"work"}, Favorite: true}
	b, _ := json.Marshal(newItem)
	cItem := cstr(string(b))
	defer freeCStr(cItem)

	createdItemJSON := requireOK(t, ZPCreateItem(C.long(created.Handle), cItem))
	var createdItem types.Item
	if err := json.Unmarshal([]byte(createdItemJSON), &createdItem); err != nil {
		t.Fatalf("parse created item JSON: %v", err)
	}
	if createdItem.ID == "" {
		t.Fatalf("expected item ID")
	}

	// Search by name
	cQuery := cstr("GitHub")
	defer freeCStr(cQuery)
	searchJSON := requireOK(t, ZPSearch(C.long(created.Handle), cQuery))
	var ids []string
	if err := json.Unmarshal([]byte(searchJSON), &ids); err != nil {
		t.Fatalf("parse search JSON: %v", err)
	}
	if len(ids) != 1 || ids[0] != createdItem.ID {
		t.Fatalf("expected search to return created item ID, got %v", ids)
	}

	// Update item
	createdItem.Name = "GitHub Updated"
	ub, _ := json.Marshal(createdItem)
	cID := cstr(createdItem.ID)
	cUpd := cstr(string(ub))
	defer freeCStr(cID)
	defer freeCStr(cUpd)
	requireOK(t, ZPUpdateItem(C.long(created.Handle), cID, cUpd))

	// Version history should have at least one entry
	histJSON := requireOK(t, ZPGetVersionHistory(C.long(created.Handle), cID))
	var hist []any
	if err := json.Unmarshal([]byte(histJSON), &hist); err != nil {
		t.Fatalf("parse version history JSON: %v", err)
	}
	if len(hist) == 0 {
		t.Fatalf("expected non-empty version history")
	}

	// Lock should block item operations
	requireOK(t, ZPLock(C.long(created.Handle)))
	if ZPIsLocked(C.long(created.Handle)) != 1 {
		t.Fatalf("expected locked")
	}
	_, _ = requireErr(t, ZPGetItem(C.long(created.Handle), cID))

	// Unlock and delete
	requireOK(t, ZPUnlock(C.long(created.Handle), cPw))
	requireOK(t, ZPDeleteItem(C.long(created.Handle), cID))

	requireOK(t, ZPCloseVault(C.long(created.Handle)))
}

func TestBridgeAdvisoryLock(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "pw"

	cPath := cstr(vaultPath)
	cPw := cstr(pw)
	defer freeCStr(cPath)
	defer freeCStr(cPw)

	createData := requireOK(t, ZPCreateVault(cPath, cPw))
	var created struct {
		Handle int64 `json:"handle"`
	}
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}

	// Second open should fail due to lock held by first handle.
	code, errStr := requireErr(t, ZPOpenVault(cPath))
	if code == 0 || errStr == "" {
		t.Fatalf("expected lock error")
	}

	requireOK(t, ZPCloseVault(C.long(created.Handle)))
}
