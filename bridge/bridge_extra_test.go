//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeropass/zeropass/core/vault/types"
)

func TestBridgeCAPI_VaultKeyUnlockWithKey_ChangePassword(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "hunter2"

	createData := requireOKGo(t, zpCCreateVault(vaultPath, pw))
	var created struct {
		Handle int64 `json:"handle"`
	}
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}

	keyJSON := requireOKGo(t, zpCGetVaultKey(created.Handle))
	var keyResp struct {
		VaultKeyBase64 string `json:"vaultKeyBase64"`
	}
	if err := json.Unmarshal([]byte(keyJSON), &keyResp); err != nil {
		t.Fatalf("parse vault key JSON: %v", err)
	}
	if keyResp.VaultKeyBase64 == "" {
		t.Fatalf("expected non-empty vaultKeyBase64")
	}

	requireOKGo(t, zpCLock(created.Handle))
	if !zpCIsLocked(created.Handle) {
		t.Fatalf("expected locked")
	}

	requireOKGo(t, zpCUnlockWithKey(created.Handle, keyResp.VaultKeyBase64))
	if zpCIsLocked(created.Handle) {
		t.Fatalf("expected unlocked")
	}

	newPw := "hunter2-new"
	requireOKGo(t, zpCChangeMasterPassword(created.Handle, pw, newPw))

	requireOKGo(t, zpCLock(created.Handle))
	requireOKGo(t, zpCUnlock(created.Handle, newPw))

	requireOKGo(t, zpCCloseVault(created.Handle))
}

func TestBridgeCAPI_RestoreVersion(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "pw"

	createData := requireOKGo(t, zpCCreateVault(vaultPath, pw))
	var created struct {
		Handle int64 `json:"handle"`
	}
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}

	newItem := types.Item{Type: types.ItemTypeLogin, Name: "GitHub", Fields: map[string]string{"username": "devuser", "password": "secret"}}
	b, _ := json.Marshal(newItem)
	createdItemJSON := requireOKGo(t, zpCCreateItem(created.Handle, string(b)))
	var createdItem types.Item
	if err := json.Unmarshal([]byte(createdItemJSON), &createdItem); err != nil {
		t.Fatalf("parse created item JSON: %v", err)
	}

	createdItem.Name = "GitHub Updated"
	ub, _ := json.Marshal(createdItem)
	requireOKGo(t, zpCUpdateItem(created.Handle, createdItem.ID, string(ub)))

	histJSON := requireOKGo(t, zpCGetVersionHistory(created.Handle, createdItem.ID))
	var hist []struct {
		Version int        `json:"version"`
		Item    types.Item `json:"item"`
	}
	if err := json.Unmarshal([]byte(histJSON), &hist); err != nil {
		t.Fatalf("parse version history JSON: %v", err)
	}
	if len(hist) == 0 {
		t.Fatalf("expected non-empty version history")
	}

	restJSON := requireOKGo(t, zpCRestoreVersion(created.Handle, createdItem.ID, int32(hist[0].Version)))
	var restored types.Item
	if err := json.Unmarshal([]byte(restJSON), &restored); err != nil {
		t.Fatalf("parse restored item JSON: %v", err)
	}
	if restored.ID != createdItem.ID {
		t.Fatalf("expected restored ID %q, got %q", createdItem.ID, restored.ID)
	}

	requireOKGo(t, zpCCloseVault(created.Handle))
}

func TestBridgeCAPI_SyncSetup_PersistsConfig(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault")
	pw := "pw"

	createData := requireOKGo(t, zpCCreateVault(vaultPath, pw))
	var created struct {
		Handle int64 `json:"handle"`
	}
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		t.Fatalf("parse create JSON: %v", err)
	}

	cfg := map[string]any{
		"server_url": "http://127.0.0.1:9999",
		"device_id":  "device-test",
		"api_key":    "",
	}
	cfgBytes, _ := json.Marshal(cfg)
	requireOKGo(t, zpCSyncSetup(created.Handle, string(cfgBytes)))

	b, err := os.ReadFile(filepath.Join(vaultPath, syncConfigFileName))
	if err != nil {
		t.Fatalf("read sync config: %v", err)
	}
	if !strings.Contains(string(b), "server_url") {
		t.Fatalf("expected persisted sync config to contain server_url, got: %s", string(b))
	}

	requireOKGo(t, zpCCloseVault(created.Handle))
}

func TestBridgeCAPI_CryptoAPI_Basics(t *testing.T) {
	// Generate password
	pwJSON := requireOKGo(t, zpCGeneratePassword(16, ""))
	var pwResp struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(pwJSON), &pwResp); err != nil {
		t.Fatalf("parse password JSON: %v", err)
	}
	if len(pwResp.Password) != 16 {
		t.Fatalf("expected password length 16, got %d", len(pwResp.Password))
	}

	// Generate passphrase with default separator
	ppJSON := requireOKGo(t, zpCGeneratePassphrase(4, ""))
	var ppResp struct {
		Passphrase string `json:"passphrase"`
	}
	if err := json.Unmarshal([]byte(ppJSON), &ppResp); err != nil {
		t.Fatalf("parse passphrase JSON: %v", err)
	}
	if !strings.Contains(ppResp.Passphrase, "-") {
		t.Fatalf("expected passphrase to contain '-', got %q", ppResp.Passphrase)
	}

	// Score password
	scoreJSON := requireOKGo(t, zpCScorePassword("correct horse battery staple"))
	var scoreResp struct {
		Score int `json:"Score"`
	}
	if err := json.Unmarshal([]byte(scoreJSON), &scoreResp); err != nil {
		t.Fatalf("parse score JSON: %v", err)
	}
	if scoreResp.Score <= 0 {
		t.Fatalf("expected score > 0, got %d", scoreResp.Score)
	}
}
