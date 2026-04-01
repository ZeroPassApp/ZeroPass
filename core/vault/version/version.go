// Package version provides item version history management.
package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeropass/zeropass/core/vault/types"
)

// ItemVersion represents a historical snapshot of an item.
type ItemVersion struct {
	Version   int        `json:"version"`
	Item      types.Item `json:"item"`
	SavedAt   time.Time  `json:"saved_at"`
}

// Manager manages version history for vault items.
type Manager struct {
	itemsPath   string
	maxVersions int
}

// NewManager creates a new version manager.
func NewManager(itemsPath string, maxVersions int) *Manager {
	if maxVersions <= 0 {
		maxVersions = 10
	}
	return &Manager{
		itemsPath:   itemsPath,
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

	history, _ := m.loadHistory(item.ID) // ignore error if file doesn't exist

	v := ItemVersion{
		Version: item.Version,
		Item:    *item,
		SavedAt: time.Now().UTC(),
	}
	history = append(history, v)

	// Trim to max.
	if len(history) > m.maxVersions {
		history = history[len(history)-m.maxVersions:]
	}

	return m.saveHistory(item.ID, history)
}

// GetVersionHistory returns the version history for an item.
func (m *Manager) GetVersionHistory(id string) ([]ItemVersion, error) {
	if id == "" {
		return nil, errors.New("item ID must not be empty")
	}
	return m.loadHistory(id)
}

// RestoreVersion returns the item data for a specific version.
func (m *Manager) RestoreVersion(id string, version int) (*types.Item, error) {
	if id == "" {
		return nil, errors.New("item ID must not be empty")
	}

	history, err := m.loadHistory(id)
	if err != nil {
		return nil, err
	}

	for _, v := range history {
		if v.Version == version {
			restored := v.Item
			return &restored, nil
		}
	}
	return nil, fmt.Errorf("version %d not found for item %s", version, id)
}

// DeleteHistory removes all version history for an item.
func (m *Manager) DeleteHistory(id string) error {
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

func (m *Manager) loadHistory(id string) ([]ItemVersion, error) {
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

	var history []ItemVersion
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("unmarshal version history: %w", err)
	}
	return history, nil
}

func (m *Manager) saveHistory(id string, history []ItemVersion) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal version history: %w", err)
	}

	if err := os.MkdirAll(m.itemsPath, 0700); err != nil {
		return fmt.Errorf("ensure items dir: %w", err)
	}

	fp := m.versionFilePath(id)
	if err := os.WriteFile(fp, data, 0600); err != nil {
		return fmt.Errorf("write version history: %w", err)
	}
	return nil
}
