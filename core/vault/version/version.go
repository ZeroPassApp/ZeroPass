// Package version provides item version history management.
package version

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/crypto/cipher"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/types"
)

// ItemVersion represents a historical snapshot of an item.
//
// NOTE: This is the decrypted API surface. On-disk persistence is encrypted.
type ItemVersion struct {
	Version int        `json:"version"`
	Item    types.Item `json:"item"`
	SavedAt time.Time  `json:"saved_at"`
}

type persistedVersion struct {
	Version  int       `json:"version"`
	SavedAt  time.Time `json:"saved_at"`
	Data     string    `json:"data"`     // base64(AEAD(JSON(item)))
	Checksum string    `json:"checksum"` // sha256(ciphertext) hex
}

// Manager manages version history for vault items.
type Manager struct {
	itemsPath   string
	vaultKey    func() ([]byte, error)
	maxVersions int
}

// NewManager creates a new version manager.
// vaultKeyFn must error if the vault is locked.
func NewManager(itemsPath string, vaultKeyFn func() ([]byte, error), maxVersions int) *Manager {
	if maxVersions <= 0 {
		maxVersions = 10
	}
	return &Manager{
		itemsPath:   itemsPath,
		vaultKey:    vaultKeyFn,
		maxVersions: maxVersions,
	}
}

// SaveVersion appends the current item state to the version history.
func (m *Manager) SaveVersion(item *types.Item) error {
	if item == nil {
		return errors.New("item must not be nil")
	}
	if item.ID == "" {
		return errors.New("item ID must not be empty")
	}
	if err := validateItemID(item.ID); err != nil {
		return err
	}
	if m.vaultKey == nil {
		return errors.New("vault key provider is missing")
	}

	history, err := m.loadPersistedHistory(item.ID)
	if err != nil {
		return err
	}

	pv, err := m.encryptItem(item)
	if err != nil {
		return err
	}
	history = append(history, pv)

	// Trim to max.
	if len(history) > m.maxVersions {
		history = history[len(history)-m.maxVersions:]
	}

	return m.savePersistedHistory(item.ID, history)
}

// GetVersionHistory returns the version history for an item.
func (m *Manager) GetVersionHistory(id string) ([]ItemVersion, error) {
	if id == "" {
		return nil, errors.New("item ID must not be empty")
	}
	if err := validateItemID(id); err != nil {
		return nil, err
	}
	if m.vaultKey == nil {
		return nil, errors.New("vault key provider is missing")
	}

	persisted, err := m.loadPersistedHistory(id)
	if err != nil {
		return nil, err
	}
	if persisted == nil {
		return nil, nil
	}

	out := make([]ItemVersion, 0, len(persisted))
	for _, pv := range persisted {
		itm, err := m.decryptItem(id, pv)
		if err != nil {
			return nil, err
		}
		out = append(out, ItemVersion{Version: pv.Version, Item: *itm, SavedAt: pv.SavedAt})
	}
	return out, nil
}

// RestoreVersion returns the item data for a specific version.
func (m *Manager) RestoreVersion(id string, version int) (*types.Item, error) {
	if id == "" {
		return nil, errors.New("item ID must not be empty")
	}
	if err := validateItemID(id); err != nil {
		return nil, err
	}
	if m.vaultKey == nil {
		return nil, errors.New("vault key provider is missing")
	}

	history, err := m.loadPersistedHistory(id)
	if err != nil {
		return nil, err
	}

	for _, pv := range history {
		if pv.Version == version {
			return m.decryptItem(id, pv)
		}
	}
	return nil, fmt.Errorf("version %d not found for item %s", version, id)
}

// DeleteHistory removes all version history for an item.
func (m *Manager) DeleteHistory(id string) error {
	if id == "" {
		return errors.New("item ID must not be empty")
	}
	if err := validateItemID(id); err != nil {
		return err
	}
	fp := m.versionFilePath(id)
	err := os.Remove(fp)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete history: %w", err)
	}
	return nil
}

// --- persistence ---

func (m *Manager) versionFilePath(id string) string {
	return filepath.Join(m.itemsPath, id+".versions.json")
}

func (m *Manager) loadPersistedHistory(id string) ([]persistedVersion, error) {
	fp := m.versionFilePath(id)
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read version history: %w", err)
	}

	// Skip empty files.
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}

	// New encrypted format.
	var persisted []persistedVersion
	if err := json.Unmarshal(data, &persisted); err == nil {
		if len(persisted) == 0 || persisted[0].Data != "" {
			return persisted, nil
		}
	}

	// Legacy plaintext format: [{version, saved_at, item}] — migrate in-place.
	var legacy []ItemVersion
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("unmarshal version history: %w", err)
	}

	migrated := make([]persistedVersion, 0, len(legacy))
	for _, v := range legacy {
		itm := v.Item
		itm.ID = id
		itm.Version = v.Version
		pv, err := m.encryptItem(&itm)
		if err != nil {
			return nil, err
		}
		pv.SavedAt = v.SavedAt
		migrated = append(migrated, pv)
	}
	if err := m.savePersistedHistory(id, migrated); err != nil {
		return nil, err
	}
	return migrated, nil
}

func (m *Manager) savePersistedHistory(id string, history []persistedVersion) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal version history: %w", err)
	}

	if err := os.MkdirAll(m.itemsPath, 0700); err != nil {
		return fmt.Errorf("ensure items dir: %w", err)
	}

	fp := m.versionFilePath(id)
	dir := filepath.Dir(fp)
	base := filepath.Base(fp)

	f, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return fmt.Errorf("create tmp version history: %w", err)
	}
	tmpPath := f.Name()

	written := 0
	for written < len(data) {
		n, err := f.Write(data[written:])
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("write tmp version history: %w", err)
		}
		if n == 0 {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("short write tmp version history: wrote %d of %d", written, len(data))
		}
		written += n
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sync tmp version history: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close tmp version history: %w", err)
	}

	if err := atomicReplaceFile(tmpPath, fp); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("finalize version history: %w", err)
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

func (m *Manager) encryptItem(item *types.Item) (persistedVersion, error) {
	vk, err := m.vaultKey()
	if err != nil {
		return persistedVersion{}, fmt.Errorf("get vault key: %w", err)
	}

	itemKey, err := key.DeriveItemKey(vk, item.ID)
	if err != nil {
		return persistedVersion{}, fmt.Errorf("derive item key: %w", err)
	}
	defer crypto.ZeroBytes(itemKey)

	plaintext, err := json.Marshal(item)
	if err != nil {
		return persistedVersion{}, fmt.Errorf("marshal item: %w", err)
	}

	c := cipher.New()
	ciphertext, err := c.Encrypt(plaintext, itemKey)
	if err != nil {
		return persistedVersion{}, fmt.Errorf("encrypt item: %w", err)
	}

	return persistedVersion{
		Version:  item.Version,
		SavedAt:  time.Now().UTC(),
		Data:     encoding.Base64StdEncode(ciphertext),
		Checksum: fmt.Sprintf("%x", sha256.Sum256(ciphertext)),
	}, nil
}

func validateItemID(id string) error {
	return types.ValidateItemID(id)
}

func atomicReplaceFile(from, to string) error {
	if err := os.Rename(from, to); err == nil {
		return nil
	} else if runtime.GOOS == "windows" {
		_ = os.Remove(to)
		return os.Rename(from, to)
	} else {
		return err
	}
}

func (m *Manager) decryptItem(id string, pv persistedVersion) (*types.Item, error) {
	ciphertext, err := encoding.Base64StdDecode(pv.Data)
	if err != nil {
		return nil, fmt.Errorf("decode version data: %w", err)
	}
	if pv.Checksum != "" {
		checksum := fmt.Sprintf("%x", sha256.Sum256(ciphertext))
		if checksum != pv.Checksum {
			return nil, errors.New("version checksum mismatch — data may be corrupted")
		}
	}

	vk, err := m.vaultKey()
	if err != nil {
		return nil, fmt.Errorf("get vault key: %w", err)
	}

	itemKey, err := key.DeriveItemKey(vk, id)
	if err != nil {
		return nil, fmt.Errorf("derive item key: %w", err)
	}
	defer crypto.ZeroBytes(itemKey)

	c := cipher.New()
	plaintext, err := c.Decrypt(ciphertext, itemKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt item: %w", err)
	}

	var item types.Item
	if err := json.Unmarshal(plaintext, &item); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}
	return &item, nil
}
