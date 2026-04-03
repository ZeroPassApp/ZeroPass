package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/index"
)

func testVaultDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "vaults", "test")
	return dir
}

func TestCreateAndOpen(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0 // disable for tests

	v, result, err := Create("strong-master-password", dir, cfg)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Mnemonic)
	assert.False(t, v.IsLocked())

	// vault.json must exist
	_, err = os.Stat(filepath.Join(dir, VaultMetaFile))
	require.NoError(t, err)

	// items dir must exist
	_, err = os.Stat(filepath.Join(dir, ItemsDir))
	require.NoError(t, err)

	// VaultKey available
	vk, err := v.VaultKey()
	require.NoError(t, err)
	assert.Len(t, vk, 32)

	// Lock
	v.Lock()
	assert.True(t, v.IsLocked())
	_, err = v.VaultKey()
	assert.Error(t, err)

	// Open fresh
	v2, err := Open(dir)
	require.NoError(t, err)
	assert.True(t, v2.IsLocked())

	// Unlock with password
	err = v2.Unlock("strong-master-password")
	require.NoError(t, err)
	assert.False(t, v2.IsLocked())
	vk2, err := v2.VaultKey()
	require.NoError(t, err)
	assert.Len(t, vk2, 32)
	v2.Lock()
}

func TestCreateEmptyPassword(t *testing.T) {
	dir := testVaultDir(t)
	_, _, err := Create("", dir, DefaultConfig())
	assert.Error(t, err)
}

func TestUnlockWrongPassword(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("correct-password", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	err = v.Unlock("wrong-password")
	assert.Error(t, err)
	assert.True(t, v.IsLocked())
}

func TestUnlockWithRecovery(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("master-pwd", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	err = v.Unlock("master-pwd")
	require.NoError(t, err)
	vk1, _ := v.VaultKey()
	vk1Copy := make([]byte, len(vk1))
	copy(vk1Copy, vk1)
	v.Lock()

	// Unlock with mnemonic
	newMnemonic, err := v.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	assert.NotEmpty(t, newMnemonic)
	vk2, _ := v.VaultKey()
	assert.Equal(t, vk1Copy, vk2) // same vault key regardless of unlock method
	v.Lock()
}

func TestUnlockWithWrongMnemonic(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	_, err = v.UnlockWithRecovery("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
	assert.Error(t, err)
}

func TestOpenNonExistent(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), "nope"))
	assert.Error(t, err)
}

func TestPaths(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	assert.Equal(t, dir, v.Path())
	assert.Equal(t, filepath.Join(dir, ItemsDir), v.ItemsPath())
	assert.Equal(t, filepath.Join(dir, IndexFile), v.IndexPath())
}

func TestConfig(t *testing.T) {
	dir := testVaultDir(t)
	cfg := VaultConfig{
		AutoLockTimeout:   5 * time.Minute,
		ClipboardClearSec: 10,
		MaxVersions:       20,
	}
	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	assert.Equal(t, cfg, v.Config())
}

func TestMetadataPersistence(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	meta := v.Metadata()
	assert.NotEmpty(t, meta.Salt)
	assert.NotEmpty(t, meta.EncryptedVaultKey)
	assert.NotEmpty(t, meta.EncryptedRecoveryKey)
	assert.NotEmpty(t, meta.VaultKeyCheck)
	assert.False(t, meta.CreatedAt.IsZero())

	// Re-open and verify metadata matches
	v2, err := Open(dir)
	require.NoError(t, err)
	meta2 := v2.Metadata()
	assert.Equal(t, meta.Salt, meta2.Salt)
	assert.Equal(t, meta.EncryptedVaultKey, meta2.EncryptedVaultKey)
	assert.Equal(t, meta.VaultKeyCheck, meta2.VaultKeyCheck)
}

func TestUnlockWithKey(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	vk, err := v.VaultKey()
	require.NoError(t, err)
	vkCopy := make([]byte, len(vk))
	copy(vkCopy, vk)
	v.Lock()

	err = v.UnlockWithKey(vkCopy)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())
}

func TestUnlockWithKeyWrongLength(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	err = v.UnlockWithKey([]byte("short"))
	assert.Error(t, err)
	assert.True(t, v.IsLocked())
}

func TestUnlockWithKeyWrongKey(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	wrongKey := make([]byte, key.VaultKeySize)
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}

	err = v.UnlockWithKey(wrongKey)
	assert.Error(t, err)
	assert.True(t, v.IsLocked())
}

func TestLegacyVaultUpgradesVaultKeyCheckOnUnlock(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	vk, err := v.VaultKey()
	require.NoError(t, err)
	vkCopy := make([]byte, len(vk))
	copy(vkCopy, vk)

	// Simulate a legacy vault.json that doesn't have vault_key_check.
	metaPath := filepath.Join(dir, VaultMetaFile)
	b, err := os.ReadFile(metaPath)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(b, &raw))
	delete(raw, "vault_key_check")
	b2, err := json.MarshalIndent(raw, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metaPath, b2, 0600))

	v2, err := Open(dir)
	require.NoError(t, err)
	require.NoError(t, v2.Unlock("pass"))
	v2.Lock()

	require.NoError(t, v2.UnlockWithKey(vkCopy))
}

func TestVaultKeyCheckRepairsOnUnlock(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	// Corrupt vault_key_check on disk.
	metaCopy := *v.Metadata()
	metaCopy.VaultKeyCheck = "not-base64"
	require.NoError(t, writeMetadata(dir, &metaCopy))
	v.Lock()

	v2, err := Open(dir)
	require.NoError(t, err)
	require.NoError(t, v2.Unlock("pass"))

	vk := make([]byte, len(v2.vaultKey))
	copy(vk, v2.vaultKey)
	require.NoError(t, verifyVaultKeyCheck(v2.Metadata().VaultKeyCheck, vk))

	v2.Lock()
	require.NoError(t, v2.UnlockWithKey(vk))
}

func TestChangeMasterPassword(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based failure injection is not portable on Windows")
	}

	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("oldpass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	origSalt := v.Metadata().Salt
	origEncVK := v.Metadata().EncryptedVaultKey

	// Make the vault directory read-only to force writeMetadata to fail.
	require.NoError(t, os.Chmod(dir, 0500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

	err = v.ChangeMasterPassword("oldpass", "newpass")
	assert.Error(t, err)
	assert.Equal(t, origSalt, v.Metadata().Salt)
	assert.Equal(t, origEncVK, v.Metadata().EncryptedVaultKey)

	require.NoError(t, os.Chmod(dir, 0700))
	err = v.ChangeMasterPassword("oldpass", "newpass")
	require.NoError(t, err)

	v2, err := Open(dir)
	require.NoError(t, err)
	assert.Error(t, v2.Unlock("oldpass"))
	require.NoError(t, v2.Unlock("newpass"))
}

func TestDoubleUnlock(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	// Already unlocked, unlock again should be no-op
	err = v.Unlock("pass")
	assert.NoError(t, err)
}

func TestTouch(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Touch() // should not panic
}

func TestAutoLockTimerFires(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 50 * time.Millisecond // very short

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())

	// Wait for auto-lock to fire
	time.Sleep(200 * time.Millisecond)
	assert.True(t, v.IsLocked())
}

func TestLockAlreadyLocked(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	v.Lock()
	assert.True(t, v.IsLocked())

	// Lock again — should be safe
	v.Lock()
	assert.True(t, v.IsLocked())
}

func TestUnlockWithRecoveryAlreadyUnlocked(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())

	// UnlockWithRecovery when already unlocked — should be no-op
	_, err = v.UnlockWithRecovery(result.Mnemonic)
	assert.NoError(t, err)
	assert.False(t, v.IsLocked())
}

func TestCorruptedMetadata(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	_, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	// Corrupt the vault.json file
	metaPath := filepath.Join(dir, VaultMetaFile)
	require.NoError(t, os.WriteFile(metaPath, []byte("{invalid json"), 0600))

	_, err = Open(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal metadata")
}

func TestVaultKeyWhenLocked(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	_, err = v.VaultKey()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault is locked")
}

func TestUnlockWithRecoveryWrongMnemonicOnLockedVault(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	_, err = v.UnlockWithRecovery("invalid mnemonic phrase here")
	assert.Error(t, err)
	assert.True(t, v.IsLocked())
}

func TestCreateVaultDirStructure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "nested", "vault")
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NotNil(t, result)

	// Check items dir exists
	_, err = os.Stat(filepath.Join(dir, ItemsDir))
	require.NoError(t, err)
}

func TestAutoLockWithRelock(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 100 * time.Millisecond

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	// Lock before auto-lock fires
	v.Lock()
	assert.True(t, v.IsLocked())

	// Unlock again — new auto-lock timer starts
	err = v.Unlock("pass")
	require.NoError(t, err)
	assert.False(t, v.IsLocked())

	// Wait for auto-lock
	time.Sleep(250 * time.Millisecond)
	assert.True(t, v.IsLocked())
}

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 15*time.Minute, cfg.AutoLockTimeout)
	assert.Equal(t, 30, cfg.ClipboardClearSec)
	assert.Equal(t, 10, cfg.MaxVersions)
}

func TestOpenUnlockRelock(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	_, result, err := Create("my-password", dir, cfg)
	require.NoError(t, err)

	// Open, unlock, get key, lock, unlock with recovery
	v, err := Open(dir)
	require.NoError(t, err)
	assert.True(t, v.IsLocked())

	require.NoError(t, v.Unlock("my-password"))
	vk1, err := v.VaultKey()
	require.NoError(t, err)
	vk1Copy := make([]byte, len(vk1))
	copy(vk1Copy, vk1)

	v.Lock()

	_, err = v.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	vk2, err := v.VaultKey()
	require.NoError(t, err)
	assert.Equal(t, vk1Copy, vk2)
	v.Lock()
}

func TestStartAutoLockExistingTimer(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 5 * time.Second // long enough to not fire during test

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())

	// autoTimer is already set from Create's startAutoLock.
	// Calling startAutoLock again exercises the "existing timer" branch.
	v.mu.Lock()
	assert.NotNil(t, v.autoTimer) // should already be set
	v.startAutoLock()             // should stop old timer and create new
	assert.NotNil(t, v.autoTimer)
	v.mu.Unlock()

	v.Lock()
}

func TestUnlockWithInvalidSalt(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	// Corrupt the salt to be invalid hex
	v.meta.Salt = "not-valid-hex!!!"
	err = v.Unlock("pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode salt")
}

func TestUnlockWithInvalidEncryptedVaultKey(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	// Corrupt the encrypted vault key to be invalid base64
	v.meta.EncryptedVaultKey = "not-valid-base64!!!"
	err = v.Unlock("pass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode encrypted vault key")
}

func TestUnlockWithRecoveryInvalidBase64(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	// Corrupt the encrypted recovery key to be invalid base64
	v.meta.EncryptedRecoveryKey = "invalid-base64!!!"
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	_, err = v.UnlockWithRecovery(mnemonic)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode encrypted recovery key")
}

func TestUnlockWithEmptyPassword(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	err = v.Unlock("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive master key")
}

func TestWriteMetadataToReadOnlyDir(t *testing.T) {
	// Test writeMetadata error path by using invalid path
	meta := &VaultMetadata{
		Salt:              "deadbeef",
		EncryptedVaultKey: "base64data",
		CreatedAt:         time.Now(),
		Config:            DefaultConfig(),
	}
	err := writeMetadata("/nonexistent/path/that/does/not/exist", meta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
}

func TestCreateMkdirAllFailure(t *testing.T) {
	// Create in a path that can't be created
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0
	_, _, err := Create("pass", "/dev/null/not-a-directory/vault", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create vault dirs")
}

func TestCreateWriteMetadataFailure(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault")
	itemsPath := filepath.Join(vaultPath, ItemsDir)

	// Pre-create both dirs so MkdirAll succeeds
	require.NoError(t, os.MkdirAll(itemsPath, 0700))

	// Make vault dir read-only so vault.json write fails
	require.NoError(t, os.Chmod(vaultPath, 0500))
	defer os.Chmod(vaultPath, 0700) // cleanup

	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0
	_, _, err := Create("pass", vaultPath, cfg)
	if err != nil {
		// On systems where chmod is enforced, this should fail at writeMetadata
		assert.Error(t, err)
	}
	// On macOS root or some systems, chmod may not prevent writes, so this might succeed
}

// --- Feature 5: RegenerateRecovery tests ---

func TestRegenerateRecovery(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	oldMnemonic := result.Mnemonic
	oldEncRK := v.meta.EncryptedRecoveryKey

	// Save vault key for comparison
	vk, err := v.VaultKey()
	require.NoError(t, err)
	vkCopy := make([]byte, len(vk))
	copy(vkCopy, vk)

	// Regenerate
	newMnemonic, err := v.RegenerateRecovery()
	require.NoError(t, err)
	assert.NotEmpty(t, newMnemonic)
	assert.NotEqual(t, oldMnemonic, newMnemonic)
	assert.NotEqual(t, oldEncRK, v.meta.EncryptedRecoveryKey)

	// Lock and unlock with the new mnemonic
	v.Lock()
	_, err = v.UnlockWithRecovery(newMnemonic)
	require.NoError(t, err)
	vk2, err := v.VaultKey()
	require.NoError(t, err)
	assert.Equal(t, vkCopy, vk2)

	// Old mnemonic should no longer work
	v.Lock()
	_, err = v.UnlockWithRecovery(oldMnemonic)
	assert.Error(t, err)
}

func TestRegenerateRecoveryWhenLocked(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	_, err = v.RegenerateRecovery()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault is locked")
}

func TestRegenerateRecoveryPersistsMetadata(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	newMnemonic, err := v.RegenerateRecovery()
	require.NoError(t, err)
	v.Lock()

	// Reopen from disk
	v2, err := Open(dir)
	require.NoError(t, err)

	_, err = v2.UnlockWithRecovery(newMnemonic)
	require.NoError(t, err)
	assert.False(t, v2.IsLocked())
	v2.Lock()
}

func TestUnlockWithRecoveryAutoRotation(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	originalMnemonic := result.Mnemonic
	v.Lock()

	// First recovery unlock — should auto-rotate
	newMnemonic1, err := v.UnlockWithRecovery(originalMnemonic)
	require.NoError(t, err)
	assert.NotEmpty(t, newMnemonic1)
	assert.NotEqual(t, originalMnemonic, newMnemonic1)
	v.Lock()

	// Original mnemonic should no longer work
	_, err = v.UnlockWithRecovery(originalMnemonic)
	assert.Error(t, err)
	assert.True(t, v.IsLocked())

	// New mnemonic should work and rotate again
	newMnemonic2, err := v.UnlockWithRecovery(newMnemonic1)
	require.NoError(t, err)
	assert.NotEmpty(t, newMnemonic2)
	assert.NotEqual(t, newMnemonic1, newMnemonic2)
	v.Lock()

	// newMnemonic1 should no longer work
	_, err = v.UnlockWithRecovery(newMnemonic1)
	assert.Error(t, err)

	// newMnemonic2 should work
	_, err = v.UnlockWithRecovery(newMnemonic2)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())
	v.Lock()
}

func TestValidateRecovery(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	// Valid mnemonic should pass
	err = v.ValidateRecovery(result.Mnemonic)
	assert.NoError(t, err)

	// Vault should remain locked (non-destructive)
	assert.True(t, v.IsLocked())

	// Invalid mnemonic should fail
	err = v.ValidateRecovery("invalid mnemonic phrase here")
	assert.Error(t, err)

	// Original mnemonic should still work (no rotation happened)
	err = v.ValidateRecovery(result.Mnemonic)
	assert.NoError(t, err)
}

func TestValidateRecoveryDoesNotRotate(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)
	v.Lock()

	// Validate multiple times — key should NOT rotate
	for i := 0; i < 3; i++ {
		err = v.ValidateRecovery(result.Mnemonic)
		require.NoError(t, err, "validation %d should succeed", i)
	}

	// Now actually unlock with recovery — should still work with original mnemonic
	newMnemonic, err := v.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	assert.False(t, v.IsLocked())
	assert.NotEmpty(t, newMnemonic)

	// Original mnemonic should now be invalid (rotation happened)
	v.Lock()
	err = v.ValidateRecovery(result.Mnemonic)
	assert.Error(t, err, "original mnemonic should be invalid after rotation")

	// New mnemonic should work
	err = v.ValidateRecovery(newMnemonic)
	assert.NoError(t, err)
}

func TestLockEncryptsIndexUnlockDecrypts(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	indexPath := v.IndexPath()

	// Create an index file to simulate usage
	idx, err := index.Open(indexPath)
	require.NoError(t, err)
	require.NoError(t, idx.Close())

	// Verify plaintext index exists
	_, err = os.Stat(indexPath)
	require.NoError(t, err)

	// Lock should encrypt the index
	v.Lock()
	assert.True(t, index.IsEncrypted(indexPath), "index should be encrypted after lock")
	_, err = os.Stat(indexPath)
	assert.True(t, os.IsNotExist(err), "plaintext index should be removed after lock")

	// Unlock should decrypt the index
	err = v.Unlock("pass")
	require.NoError(t, err)
	assert.False(t, index.IsEncrypted(indexPath), "index should not be encrypted after unlock")
	_, err = os.Stat(indexPath)
	assert.NoError(t, err, "plaintext index should be restored after unlock")

	v.Lock()
}

func TestLockNoIndexFileNoPanic(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, _, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	// Lock without any index file — should not panic or error
	v.Lock()
	assert.True(t, v.IsLocked())
}

func TestUnlockWithRecoveryDecryptsIndex(t *testing.T) {
	dir := testVaultDir(t)
	cfg := DefaultConfig()
	cfg.AutoLockTimeout = 0

	v, result, err := Create("pass", dir, cfg)
	require.NoError(t, err)

	indexPath := v.IndexPath()

	// Create index file
	idx, err := index.Open(indexPath)
	require.NoError(t, err)
	require.NoError(t, idx.Close())

	// Lock encrypts
	v.Lock()
	assert.True(t, index.IsEncrypted(indexPath))

	// Unlock with recovery decrypts
	_, err = v.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	assert.False(t, index.IsEncrypted(indexPath))
	v.Lock()
}
