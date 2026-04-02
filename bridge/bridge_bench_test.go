//go:build cgo

package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func BenchmarkUnlockWithKey(b *testing.B) {
	vaultPath := filepath.Join(b.TempDir(), "vault")
	pw := "bench-master-password"

	createData := requireOKGo(b, zpCCreateVault(vaultPath, pw))
	var created struct{ Handle int64 `json:"handle"` }
	if err := json.Unmarshal([]byte(createData), &created); err != nil {
		b.Fatalf("parse create JSON: %v", err)
	}

	requireOKGo(b, zpCUnlock(created.Handle, pw))

	keyJSON := requireOKGo(b, zpCGetVaultKey(created.Handle))
	var keyResp struct{ VaultKeyBase64 string `json:"vaultKeyBase64"` }
	if err := json.Unmarshal([]byte(keyJSON), &keyResp); err != nil {
		b.Fatalf("parse vault key JSON: %v", err)
	}
	if keyResp.VaultKeyBase64 == "" {
		b.Fatalf("expected non-empty vault key")
	}

	requireOKGo(b, zpCLock(created.Handle))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requireOKGo(b, zpCUnlockWithKey(created.Handle, keyResp.VaultKeyBase64))
		requireOKGo(b, zpCLock(created.Handle))
	}

	b.StopTimer()
	requireOKGo(b, zpCCloseVault(created.Handle))
}
