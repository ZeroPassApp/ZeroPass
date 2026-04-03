// Package store provides vault lifecycle management including creation,
// opening, locking, and unlocking of vaults.
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zeropass/zeropass/core/crypto"
	aesgcm "github.com/zeropass/zeropass/core/crypto/cipher"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/crypto/kdf"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/index"
)

// Default paths and settings.
const (
	DefaultVaultDir  = ".zeropass"
	DefaultVaultName = "default"
	VaultMetaFile    = "vault.json"
	ItemsDir         = "items"
	IndexFile        = "index.db"
)

// VaultConfig holds vault-level configuration.
type VaultConfig struct {
	AutoLockTimeout   time.Duration `json:"auto_lock_timeout"`   // 0 = never
	ClipboardClearSec int           `json:"clipboard_clear_sec"` // seconds
	MaxVersions       int           `json:"max_versions"`        // per item
}

// DefaultConfig returns the default vault configuration.
func DefaultConfig() VaultConfig {
	return VaultConfig{
		AutoLockTimeout:   15 * time.Minute,
		ClipboardClearSec: 30,
		MaxVersions:       10,
	}
}

// VaultMetadata is persisted to vault.json.
type VaultMetadata struct {
	Salt                 string      `json:"salt"`                      // hex-encoded
	EncryptedVaultKey    string      `json:"encrypted_vault_key"`       // base64
	EncryptedRecoveryKey string      `json:"encrypted_recovery_key"`    // base64
	VaultKeyCheck        string      `json:"vault_key_check,omitempty"` // base64(AEAD(vaultKey, "zeropass:vault-key-check:v1"))
	CreatedAt            time.Time   `json:"created_at"`
	Config               VaultConfig `json:"config"`

	extra map[string]json.RawMessage `json:"-"`
}

func (m *VaultMetadata) UnmarshalJSON(data []byte) error {
	type alias VaultMetadata
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = VaultMetadata(a)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	delete(raw, "salt")
	delete(raw, "encrypted_vault_key")
	delete(raw, "encrypted_recovery_key")
	delete(raw, "vault_key_check")
	delete(raw, "created_at")
	delete(raw, "config")

	if len(raw) > 0 {
		m.extra = raw
	}
	return nil
}

func (m VaultMetadata) MarshalJSON() ([]byte, error) {
	type alias VaultMetadata

	b, err := json.Marshal(alias(m))
	if err != nil {
		return nil, err
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	for k, v := range m.extra {
		if _, ok := out[k]; ok {
			continue
		}
		out[k] = v
	}

	return json.Marshal(out)
}

// CreateResult is returned from Create with secrets the user must save.
type CreateResult struct {
	Mnemonic string // BIP-39 recovery mnemonic — show once
}

// Vault represents an open vault instance.
type Vault struct {
	mu       sync.RWMutex
	path     string // root directory of this vault
	meta     *VaultMetadata
	vaultKey []byte // nil when locked
	locked   bool

	// auto-lock
	lastAccess        time.Time
	autoTimer         *time.Timer
	autoLockDisabled  bool // runtime-only (not persisted)
}

const vaultKeyCheckMagic = "zeropass:vault-key-check:v1"

func makeVaultKeyCheck(vaultKey []byte) (string, error) {
	if len(vaultKey) != key.VaultKeySize {
		return "", fmt.Errorf("vault key must be %d bytes", key.VaultKeySize)
	}
	c := aesgcm.New()
	ct, err := c.Encrypt([]byte(vaultKeyCheckMagic), vaultKey)
	if err != nil {
		return "", err
	}
	return encoding.Base64StdEncode(ct), nil
}

func verifyVaultKeyCheck(check string, vaultKey []byte) error {
	if check == "" {
		return errors.New("vault key check is missing")
	}
	ct, err := encoding.Base64StdDecode(check)
	if err != nil {
		return fmt.Errorf("decode vault key check: %w", err)
	}
	c := aesgcm.New()
	pt, err := c.Decrypt(ct, vaultKey)
	if err != nil {
		return errors.New("invalid vault key")
	}
	if !bytes.Equal(pt, []byte(vaultKeyCheckMagic)) {
		return errors.New("invalid vault key")
	}
	return nil
}

// ensureVaultKeyCheckLocked ensures VaultMetadata.VaultKeyCheck exists and is valid.
// Caller must hold v.mu.
func (v *Vault) ensureVaultKeyCheckLocked(vaultKey []byte) {
	if v.meta == nil {
		return
	}
	if v.meta.VaultKeyCheck != "" {
		if err := verifyVaultKeyCheck(v.meta.VaultKeyCheck, vaultKey); err == nil {
			return
		}
	}

	check, err := makeVaultKeyCheck(vaultKey)
	if err != nil {
		return
	}

	metaCopy := *v.meta
	metaCopy.VaultKeyCheck = check
	if err := writeMetadata(v.path, &metaCopy); err != nil {
		return
	}
	v.meta.VaultKeyCheck = check
}

// Create creates a new vault on disk and returns the recovery mnemonic.
func Create(masterPassword string, vaultPath string, cfg VaultConfig) (*Vault, *CreateResult, error) {
	if masterPassword == "" {
		return nil, nil, errors.New("master password must not be empty")
	}

	// Ensure directory structure.
	itemsPath := filepath.Join(vaultPath, ItemsDir)
	if err := os.MkdirAll(itemsPath, 0700); err != nil {
		return nil, nil, fmt.Errorf("create vault dirs: %w", err)
	}

	// Derive master key.
	k := kdf.NewDefault()
	mk, err := key.DeriveNewMasterKey([]byte(masterPassword), k)
	if err != nil {
		return nil, nil, fmt.Errorf("derive master key: %w", err)
	}
	defer mk.Zero()

	// Generate vault key.
	vaultKey, err := key.GenerateVaultKey()
	if err != nil {
		return nil, nil, fmt.Errorf("generate vault key: %w", err)
	}

	// Encrypt vault key with master key.
	encVK, err := key.EncryptVaultKey(vaultKey, mk.Key())
	if err != nil {
		crypto.ZeroBytes(vaultKey)
		return nil, nil, fmt.Errorf("encrypt vault key: %w", err)
	}

	// Generate recovery mnemonic and encrypt vault key with it.
	recMgr := key.NewRecoveryKeyManager()
	mnemonic, err := recMgr.GenerateMnemonic()
	if err != nil {
		crypto.ZeroBytes(vaultKey)
		return nil, nil, fmt.Errorf("generate mnemonic: %w", err)
	}

	encRK, err := key.EncryptVaultKeyWithRecovery(vaultKey, mnemonic)
	if err != nil {
		crypto.ZeroBytes(vaultKey)
		return nil, nil, fmt.Errorf("encrypt vault key with recovery: %w", err)
	}

	check, err := makeVaultKeyCheck(vaultKey)
	if err != nil {
		crypto.ZeroBytes(vaultKey)
		return nil, nil, fmt.Errorf("vault key check: %w", err)
	}

	// Build metadata.
	meta := &VaultMetadata{
		Salt:                 encoding.HexEncode(mk.Salt()),
		EncryptedVaultKey:    encoding.Base64StdEncode(encVK.Ciphertext),
		EncryptedRecoveryKey: encoding.Base64StdEncode(encRK.Ciphertext),
		VaultKeyCheck:        check,
		CreatedAt:            time.Now().UTC(),
		Config:               cfg,
	}

	// Write vault.json.
	if err := writeMetadata(vaultPath, meta); err != nil {
		crypto.ZeroBytes(vaultKey)
		return nil, nil, err
	}

	v := &Vault{
		path:       vaultPath,
		meta:       meta,
		vaultKey:   vaultKey,
		locked:     false,
		lastAccess: time.Now(),
	}
	v.startAutoLock()

	return v, &CreateResult{Mnemonic: mnemonic}, nil
}

// Open loads vault metadata from disk. The vault starts locked.
func Open(vaultPath string) (*Vault, error) {
	meta, err := readMetadata(vaultPath)
	if err != nil {
		return nil, err
	}
	return &Vault{
		path:   vaultPath,
		meta:   meta,
		locked: true,
	}, nil
}

// Unlock decrypts the vault key using the master password.
func (v *Vault) Unlock(masterPassword string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.locked {
		return nil // already unlocked
	}

	salt, err := encoding.HexDecode(v.meta.Salt)
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	k := kdf.NewDefault()
	mk, err := key.DeriveMasterKeyWithSalt([]byte(masterPassword), salt, k)
	if err != nil {
		return fmt.Errorf("derive master key: %w", err)
	}
	defer mk.Zero()

	encBytes, err := encoding.Base64StdDecode(v.meta.EncryptedVaultKey)
	if err != nil {
		return fmt.Errorf("decode encrypted vault key: %w", err)
	}

	vaultKey, err := key.DecryptVaultKey(&key.EncryptedVaultKey{Ciphertext: encBytes}, mk.Key())
	if err != nil {
		return fmt.Errorf("unlock vault: %w", err)
	}

	v.vaultKey = vaultKey
	v.locked = false
	v.lastAccess = time.Now()
	v.startAutoLock()
	v.ensureVaultKeyCheckLocked(vaultKey)

	// Decrypt index file at rest (best-effort).
	_ = index.DecryptIndexFile(v.IndexPath(), v.vaultKey)

	return nil
}

// UnlockWithRecovery decrypts the vault key using a recovery mnemonic.
func (v *Vault) UnlockWithRecovery(mnemonic string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.locked {
		return nil
	}

	encBytes, err := encoding.Base64StdDecode(v.meta.EncryptedRecoveryKey)
	if err != nil {
		return fmt.Errorf("decode encrypted recovery key: %w", err)
	}

	vaultKey, err := key.DecryptVaultKeyWithRecovery(&key.EncryptedVaultKey{Ciphertext: encBytes}, mnemonic)
	if err != nil {
		return fmt.Errorf("unlock vault with recovery: %w", err)
	}

	v.vaultKey = vaultKey
	v.locked = false
	v.lastAccess = time.Now()
	v.startAutoLock()
	v.ensureVaultKeyCheckLocked(vaultKey)

	// Decrypt index file at rest (best-effort).
	_ = index.DecryptIndexFile(v.IndexPath(), v.vaultKey)

	return nil
}

// UnlockWithKey unlocks the vault using a raw vault key (e.g. TouchID flow).
// The provided key must be validated against VaultMetadata.VaultKeyCheck.
func (v *Vault) UnlockWithKey(vaultKey []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.locked {
		return nil
	}
	if len(vaultKey) != key.VaultKeySize {
		return fmt.Errorf("vault key must be %d bytes", key.VaultKeySize)
	}
	if v.meta.VaultKeyCheck == "" {
		return errors.New("vault key unlock not available: unlock with password once to upgrade vault metadata")
	}

	vk := make([]byte, len(vaultKey))
	copy(vk, vaultKey)

	if err := verifyVaultKeyCheck(v.meta.VaultKeyCheck, vk); err != nil {
		crypto.ZeroBytes(vk)
		return fmt.Errorf("unlock vault with key: %w", err)
	}

	// Decrypt index file at rest (best-effort).
	_ = index.DecryptIndexFile(v.IndexPath(), vk)

	v.vaultKey = vk
	v.locked = false
	v.lastAccess = time.Now()
	v.startAutoLock()
	return nil
}

// ChangeMasterPassword re-encrypts the vault key with a new master password.
func (v *Vault) ChangeMasterPassword(oldPassword, newPassword string) error {
	if newPassword == "" {
		return errors.New("new password must not be empty")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	salt, err := encoding.HexDecode(v.meta.Salt)
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	k := kdf.NewDefault()
	mkOld, err := key.DeriveMasterKeyWithSalt([]byte(oldPassword), salt, k)
	if err != nil {
		return fmt.Errorf("derive master key: %w", err)
	}
	defer mkOld.Zero()

	encBytes, err := encoding.Base64StdDecode(v.meta.EncryptedVaultKey)
	if err != nil {
		return fmt.Errorf("decode encrypted vault key: %w", err)
	}

	vk, err := key.DecryptVaultKey(&key.EncryptedVaultKey{Ciphertext: encBytes}, mkOld.Key())
	if err != nil {
		return fmt.Errorf("verify old password: %w", err)
	}
	defer crypto.ZeroBytes(vk)

	mkNew, err := key.DeriveNewMasterKey([]byte(newPassword), k)
	if err != nil {
		return fmt.Errorf("derive new master key: %w", err)
	}
	defer mkNew.Zero()

	encVK, err := key.EncryptVaultKey(vk, mkNew.Key())
	if err != nil {
		return fmt.Errorf("encrypt vault key: %w", err)
	}

	newSalt := encoding.HexEncode(mkNew.Salt())
	newEnc := encoding.Base64StdEncode(encVK.Ciphertext)

	metaCopy := *v.meta
	metaCopy.Salt = newSalt
	metaCopy.EncryptedVaultKey = newEnc

	if err := writeMetadata(v.path, &metaCopy); err != nil {
		return fmt.Errorf("persist metadata: %w", err)
	}

	v.meta.Salt = newSalt
	v.meta.EncryptedVaultKey = newEnc
	return nil
}

// Lock zeroes all key material from memory.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.lockInternal()
}

func (v *Vault) lockInternal() {
	// Encrypt the index file before zeroing key material.
	if v.vaultKey != nil {
		if err := index.EncryptIndexFile(v.IndexPath(), v.vaultKey); err != nil {
			// Fail-closed: index is derived data; don't risk leaving plaintext at rest.
			_ = os.Remove(v.IndexPath())
			_ = os.Remove(v.IndexPath() + "-wal")
			_ = os.Remove(v.IndexPath() + "-shm")
			_ = os.Remove(v.IndexPath() + ".enc")
			_ = os.Remove(v.IndexPath() + ".enc.tmp")
		}
		crypto.ZeroBytes(v.vaultKey)
		v.vaultKey = nil
	}
	v.locked = true
	if v.autoTimer != nil {
		v.autoTimer.Stop()
		v.autoTimer = nil
	}
}

// IsLocked reports whether the vault is currently locked.
func (v *Vault) IsLocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.locked
}

// VaultKey returns the current vault key. Caller must not retain.
func (v *Vault) VaultKey() ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.locked || v.vaultKey == nil {
		return nil, errors.New("vault is locked")
	}
	v.lastAccess = time.Now()
	return v.vaultKey, nil
}

// Path returns the vault root directory.
func (v *Vault) Path() string {
	return v.path
}

// ItemsPath returns the path to the items directory.
func (v *Vault) ItemsPath() string {
	return filepath.Join(v.path, ItemsDir)
}

// IndexPath returns the path to the index database.
func (v *Vault) IndexPath() string {
	return filepath.Join(v.path, IndexFile)
}

// Metadata returns the vault metadata.
func (v *Vault) Metadata() *VaultMetadata {
	return v.meta
}

// Config returns the vault configuration.
func (v *Vault) Config() VaultConfig {
	return v.meta.Config
}

// DisableAutoLock disables the vault's built-in auto-lock timer for this process.
// It does not persist the change to disk.
func (v *Vault) DisableAutoLock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.autoLockDisabled = true
	if v.autoTimer != nil {
		v.autoTimer.Stop()
		v.autoTimer = nil
	}
}

// Touch refreshes the last-access time (resets auto-lock timer).
func (v *Vault) Touch() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.lastAccess = time.Now()
}

// startAutoLock starts the auto-lock timer if configured.
func (v *Vault) startAutoLock() {
	if v.autoLockDisabled || v.meta.Config.AutoLockTimeout <= 0 {
		if v.autoTimer != nil {
			v.autoTimer.Stop()
			v.autoTimer = nil
		}
		return
	}
	if v.autoTimer != nil {
		v.autoTimer.Stop()
	}
	v.autoTimer = time.AfterFunc(v.meta.Config.AutoLockTimeout, func() {
		v.Lock()
	})
}

// --- file I/O helpers ---

func writeMetadata(vaultPath string, meta *VaultMetadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	fp := filepath.Join(vaultPath, VaultMetaFile)
	dir := filepath.Dir(fp)

	f, err := os.CreateTemp(dir, VaultMetaFile+".*.tmp")
	if err != nil {
		return fmt.Errorf("create tmp metadata: %w", err)
	}
	tmpPath := f.Name()

	written := 0
	for written < len(data) {
		n, err := f.Write(data[written:])
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("write tmp metadata: %w", err)
		}
		if n == 0 {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("short write tmp metadata: wrote %d of %d", written, len(data))
		}
		written += n
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sync tmp metadata: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close tmp metadata: %w", err)
	}

	if err := atomicReplaceFile(tmpPath, fp); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("finalize metadata: %w", err)
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

func readMetadata(vaultPath string) (*VaultMetadata, error) {
	fp := filepath.Join(vaultPath, VaultMetaFile)
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var meta VaultMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return &meta, nil
}

// RegenerateRecovery generates a new recovery mnemonic and re-encrypts the vault key
// with the new recovery key. Returns the new mnemonic. The vault must be unlocked.
func (v *Vault) RegenerateRecovery() (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.locked {
		return "", errors.New("vault is locked")
	}

	mnemonic, encRK, err := key.RegenerateRecoveryKey(v.vaultKey)
	if err != nil {
		return "", fmt.Errorf("regenerate recovery key: %w", err)
	}

	v.meta.EncryptedRecoveryKey = encoding.Base64StdEncode(encRK.Ciphertext)

	if err := writeMetadata(v.path, v.meta); err != nil {
		return "", fmt.Errorf("persist metadata: %w", err)
	}

	return mnemonic, nil
}
