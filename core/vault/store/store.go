// Package store provides vault lifecycle management including creation,
// opening, locking, and unlocking of vaults.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/crypto/kdf"
	"github.com/zeropass/zeropass/core/crypto/key"
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
	Salt                 string     `json:"salt"`                   // hex-encoded
	EncryptedVaultKey    string     `json:"encrypted_vault_key"`    // base64
	EncryptedRecoveryKey string     `json:"encrypted_recovery_key"` // base64
	CreatedAt            time.Time  `json:"created_at"`
	Config               VaultConfig `json:"config"`
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
	lastAccess time.Time
	autoTimer  *time.Timer
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

	// Build metadata.
	meta := &VaultMetadata{
		Salt:                 encoding.HexEncode(mk.Salt()),
		EncryptedVaultKey:    encoding.Base64StdEncode(encVK.Ciphertext),
		EncryptedRecoveryKey: encoding.Base64StdEncode(encRK.Ciphertext),
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
	return nil
}

// Lock zeroes all key material from memory.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.lockInternal()
}

func (v *Vault) lockInternal() {
	if v.vaultKey != nil {
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
	v.mu.RLock()
	defer v.mu.RUnlock()
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

// Touch refreshes the last-access time (resets auto-lock timer).
func (v *Vault) Touch() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.lastAccess = time.Now()
}

// startAutoLock starts the auto-lock timer if configured.
func (v *Vault) startAutoLock() {
	if v.meta.Config.AutoLockTimeout <= 0 {
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
	if err := os.WriteFile(fp, data, 0600); err != nil {
		return fmt.Errorf("write metadata: %w", err)
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
