//go:build cgo

package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/zeropass/zeropass/core/vault/types"
)

func TestBridgeCAPI_CreateCRUDSearchLockClose(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "hunter2"

	createData := requireOKGo(t, zpCCreateVault(vaultPath, pw))
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

	listData := requireOKGo(t, zpCListItems(created.Handle, ""))
	var listed []*types.Item
	if err := json.Unmarshal([]byte(listData), &listed); err != nil {
		t.Fatalf("parse list JSON: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected 0 items, got %d", len(listed))
	}

	newItem := types.Item{Type: types.ItemTypeLogin, Name: "GitHub", Fields: map[string]string{"username": "devuser", "password": "secret"}, Tags: []string{"work"}, Favorite: true}
	b, _ := json.Marshal(newItem)
	createdItemJSON := requireOKGo(t, zpCCreateItem(created.Handle, string(b)))
	var createdItem types.Item
	if err := json.Unmarshal([]byte(createdItemJSON), &createdItem); err != nil {
		t.Fatalf("parse created item JSON: %v", err)
	}
	if createdItem.ID == "" {
		t.Fatalf("expected item ID")
	}

	searchJSON := requireOKGo(t, zpCSearch(created.Handle, "GitHub"))
	var ids []string
	if err := json.Unmarshal([]byte(searchJSON), &ids); err != nil {
		t.Fatalf("parse search JSON: %v", err)
	}
	if len(ids) != 1 || ids[0] != createdItem.ID {
		t.Fatalf("expected search to return created item ID, got %v", ids)
	}

	createdItem.Name = "GitHub Updated"
	ub, _ := json.Marshal(createdItem)
	requireOKGo(t, zpCUpdateItem(created.Handle, createdItem.ID, string(ub)))

	histJSON := requireOKGo(t, zpCGetVersionHistory(created.Handle, createdItem.ID))
	var hist []any
	if err := json.Unmarshal([]byte(histJSON), &hist); err != nil {
		t.Fatalf("parse version history JSON: %v", err)
	}
	if len(hist) == 0 {
		t.Fatalf("expected non-empty version history")
	}

	requireOKGo(t, zpCLock(created.Handle))
	if !zpCIsLocked(created.Handle) {
		t.Fatalf("expected locked")
	}
	_, _ = requireErrGo(t, zpCGetItem(created.Handle, createdItem.ID))

	requireOKGo(t, zpCUnlock(created.Handle, pw))
	requireOKGo(t, zpCDeleteItem(created.Handle, createdItem.ID))
	requireOKGo(t, zpCCloseVault(created.Handle))
}

func TestBridgeCAPI_AdvisoryLock(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "pw"

	createData := requireOKGo(t, zpCCreateVault(vaultPath, pw))
	var created struct{ Handle int64 `json:"handle"` }
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}

	_, _ = requireErrGo(t, zpCOpenVault(vaultPath))
	requireOKGo(t, zpCCloseVault(created.Handle))
}
