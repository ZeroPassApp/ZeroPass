package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropass/zeropass/core/vault/health"
	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/types"
)

const testMasterPassword = "SuperStrongP@ssword123!XYZ"

// strongInitPassword is a password that passes the zxcvbn strength check (score >= 3)
const strongInitPassword = "Xk9#mL2pQ7!nR4wF6bJ8"

// createTestVault creates a temporary vault for testing and returns the vault, path, and cleanup func.
func createTestVault(t *testing.T) (*store.Vault, string, func()) {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test-vault")

	v, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	require.NotEmpty(t, result.Mnemonic)

	return v, vaultPath, func() {
		v.Lock()
	}
}

// createTestManager creates an item manager and index for a test vault.
func createTestManager(t *testing.T, v *store.Vault) (*item.Manager, *index.Index) {
	t.Helper()
	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)
	t.Cleanup(func() { idx.Close() })
	return mgr, idx
}

// TestGetVaultPath tests vault path resolution.
func TestGetVaultPath(t *testing.T) {
	// Default path
	flagVaultPath = ""
	path := getVaultPath()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	expected := filepath.Join(home, ".zeropass", "vaults", "default")
	assert.Equal(t, expected, path)

	// Custom path
	flagVaultPath = "/custom/path"
	path = getVaultPath()
	assert.Equal(t, "/custom/path", path)

	// Reset
	flagVaultPath = ""
}

// TestParseTags tests tag parsing from comma-separated strings.
func TestParseTags(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"tag1", []string{"tag1"}},
		{"tag1,tag2", []string{"tag1", "tag2"}},
		{"tag1, tag2, tag3", []string{"tag1", "tag2", "tag3"}},
		{" tag1 , , tag2 ", []string{"tag1", "tag2"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestItemPreview tests the preview generation for items.
func TestItemPreview(t *testing.T) {
	tests := []struct {
		name     string
		item     *types.Item
		expected string
	}{
		{
			name: "login with URL",
			item: &types.Item{
				Type:   types.ItemTypeLogin,
				Fields: map[string]string{types.FieldURL: "https://example.com"},
			},
			expected: "https://example.com",
		},
		{
			name: "login with username only",
			item: &types.Item{
				Type:   types.ItemTypeLogin,
				Fields: map[string]string{types.FieldUsername: "user@test.com"},
			},
			expected: "user@test.com",
		},
		{
			name: "credit card",
			item: &types.Item{
				Type:   types.ItemTypeCreditCard,
				Fields: map[string]string{types.FieldCardNumber: "4111111111111111"},
			},
			expected: "****1111",
		},
		{
			name: "identity",
			item: &types.Item{
				Type:   types.ItemTypeIdentity,
				Fields: map[string]string{types.FieldFirstName: "John", types.FieldLastName: "Doe"},
			},
			expected: "John Doe",
		},
		{
			name: "secure note",
			item: &types.Item{
				Type:  types.ItemTypeSecureNote,
				Notes: "This is a short note",
			},
			expected: "This is a short note",
		},
		{
			name: "secure note long",
			item: &types.Item{
				Type:  types.ItemTypeSecureNote,
				Notes: "This is a very long note that exceeds the maximum preview length limit",
			},
			expected: "This is a very long note that exceeds ...",
		},
		{
			name: "api key",
			item: &types.Item{
				Type:   types.ItemTypeAPIKey,
				Fields: map[string]string{types.FieldAPIKey: "sk-1234567890abcdef"},
			},
			expected: "sk-12345…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := itemPreview(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDefaultSecretField tests default secret field selection.
func TestDefaultSecretField(t *testing.T) {
	tests := []struct {
		name     string
		item     *types.Item
		expected string
	}{
		{
			name: "login returns password",
			item: &types.Item{
				Type:   types.ItemTypeLogin,
				Fields: map[string]string{types.FieldPassword: "secret123"},
			},
			expected: "secret123",
		},
		{
			name: "apikey returns secret",
			item: &types.Item{
				Type:   types.ItemTypeAPIKey,
				Fields: map[string]string{types.FieldAPISecret: "mysecret"},
			},
			expected: "mysecret",
		},
		{
			name: "apikey returns key when no secret",
			item: &types.Item{
				Type:   types.ItemTypeAPIKey,
				Fields: map[string]string{types.FieldAPIKey: "mykey"},
			},
			expected: "mykey",
		},
		{
			name: "credit card returns card number",
			item: &types.Item{
				Type:   types.ItemTypeCreditCard,
				Fields: map[string]string{types.FieldCardNumber: "4111111111111111"},
			},
			expected: "4111111111111111",
		},
		{
			name: "note returns notes",
			item: &types.Item{
				Type:  types.ItemTypeSecureNote,
				Notes: "my secret note",
			},
			expected: "my secret note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultSecretField(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestItemToOutputMap tests JSON output map generation.
func TestItemToOutputMap(t *testing.T) {
	itm := &types.Item{
		ID:   "test-id",
		Name: "Test Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "user",
			types.FieldPassword: "secret",
		},
	}

	// Without secrets
	out := itemToOutputMap(itm, false)
	assert.Equal(t, "test-id", out["id"])
	assert.Equal(t, "Test Item", out["name"])
	fields := out["fields"].(map[string]string)
	assert.Equal(t, "user", fields[types.FieldUsername])
	assert.Equal(t, "********", fields[types.FieldPassword])

	// With secrets
	out = itemToOutputMap(itm, true)
	fields = out["fields"].(map[string]string)
	assert.Equal(t, "secret", fields[types.FieldPassword])
}

// TestFormatOutputJSON tests JSON output formatting.
func TestFormatOutputJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origOutput := flagOutput
	flagOutput = "json"
	defer func() {
		flagOutput = origOutput
		os.Stdout = old
	}()

	formatOutput(map[string]string{"key": "value"})

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

// TestScoreDisplay tests score display formatting.
func TestScoreDisplay(t *testing.T) {
	assert.Contains(t, scoreDisplay(100), "Excellent")
	assert.Contains(t, scoreDisplay(90), "Excellent")
	assert.Contains(t, scoreDisplay(80), "Good")
	assert.Contains(t, scoreDisplay(50), "Fair")
	assert.Contains(t, scoreDisplay(30), "Poor")
	assert.Contains(t, scoreDisplay(0), "Poor")
}

// TestSeverityIcon tests severity icon rendering.
func TestSeverityIcon(t *testing.T) {
	assert.Equal(t, "🔴", severityIcon(health.SeverityCritical))
	assert.Equal(t, "🟡", severityIcon(health.SeverityWarning))
	assert.Equal(t, "🔵", severityIcon(health.SeverityInfo))
	assert.Equal(t, "⚪", severityIcon("unknown"))
}

// TestParseEnvFile tests .env file parsing.
func TestParseEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := `# Database config
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD=zp://my-db/password
API_KEY="zp://api-service/api_key"
SIMPLE='plain-value'
QUOTED="hello world"

# Empty line above is intentional
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	require.NoError(t, err)

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.True(t, hasRefs)
	assert.Equal(t, "localhost", vars["DB_HOST"])
	assert.Equal(t, "5432", vars["DB_PORT"])
	assert.Equal(t, "zp://my-db/password", vars["DB_PASSWORD"])
	assert.Equal(t, "zp://api-service/api_key", vars["API_KEY"])
	assert.Equal(t, "plain-value", vars["SIMPLE"])
	assert.Equal(t, "hello world", vars["QUOTED"])
}

// TestParseEnvFileNoRefs tests .env file without zp:// references.
func TestParseEnvFileNoRefs(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := `KEY=value
OTHER=stuff
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	require.NoError(t, err)

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.False(t, hasRefs)
	assert.Equal(t, "value", vars["KEY"])
}

// TestParseEnvFileInvalid tests invalid .env file parsing.
func TestParseEnvFileInvalid(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := `VALID=ok
INVALID_LINE
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	require.NoError(t, err)

	_, _, err = ParseEnvFile(envPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

// TestParseEnvFileNotFound tests missing .env file.
func TestParseEnvFileNotFound(t *testing.T) {
	_, _, err := ParseEnvFile("/nonexistent/.env")
	assert.Error(t, err)
}

// TestUnquoteEnvValue tests quote removal.
func TestUnquoteEnvValue(t *testing.T) {
	assert.Equal(t, "hello", unquoteEnvValue(`"hello"`))
	assert.Equal(t, "world", unquoteEnvValue(`'world'`))
	assert.Equal(t, "no-quotes", unquoteEnvValue("no-quotes"))
	assert.Equal(t, "", unquoteEnvValue(`""`))
	assert.Equal(t, "", unquoteEnvValue(`''`))
	assert.Equal(t, "a", unquoteEnvValue("a"))
}

// TestVaultCreateAndOpen tests vault creation, locking, and reopening.
func TestVaultCreateAndOpen(t *testing.T) {
	v, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	// Vault should start unlocked
	assert.False(t, v.IsLocked())

	// Lock
	v.Lock()
	assert.True(t, v.IsLocked())

	// Re-open and unlock
	v2, err := store.Open(vaultPath)
	require.NoError(t, err)
	assert.True(t, v2.IsLocked())

	err = v2.Unlock(testMasterPassword)
	require.NoError(t, err)
	assert.False(t, v2.IsLocked())
	v2.Lock()
}

// TestItemCRUD tests full create/read/update/delete cycle.
func TestItemCRUD(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add
	itm := &types.Item{
		Name: "Test Login",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "testuser",
			types.FieldPassword: "testpass",
			types.FieldURL:      "https://test.com",
		},
		Tags: []string{"test", "dev"},
	}

	err := mgr.AddItem(itm)
	require.NoError(t, err)
	assert.NotEmpty(t, itm.ID)
	assert.Equal(t, 1, itm.Version)

	// Get
	retrieved, err := mgr.GetItem(itm.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Login", retrieved.Name)
	assert.Equal(t, "testuser", retrieved.Fields[types.FieldUsername])
	assert.Equal(t, "testpass", retrieved.Fields[types.FieldPassword])

	// Update
	retrieved.Name = "Updated Login"
	err = mgr.UpdateItem(retrieved.ID, retrieved)
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.Version)

	updated, err := mgr.GetItem(itm.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Login", updated.Name)

	// List
	items, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Delete
	err = mgr.DeleteItem(itm.ID)
	require.NoError(t, err)

	items, err = mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

// TestItemListFilter tests listing with filters.
func TestItemListFilter(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add various items
	items := []*types.Item{
		{
			Name:   "Login 1",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldUsername: "user1"},
			Tags:   []string{"prod"},
		},
		{
			Name:     "Login 2",
			Type:     types.ItemTypeLogin,
			Fields:   map[string]string{types.FieldUsername: "user2"},
			Tags:     []string{"dev"},
			Favorite: true,
		},
		{
			Name:   "API Key",
			Type:   types.ItemTypeAPIKey,
			Fields: map[string]string{types.FieldAPIKey: "key123"},
			Tags:   []string{"prod"},
		},
	}

	for _, itm := range items {
		require.NoError(t, mgr.AddItem(itm))
	}

	// Filter by type
	logins, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeLogin})
	require.NoError(t, err)
	assert.Len(t, logins, 2)

	apikeys, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeAPIKey})
	require.NoError(t, err)
	assert.Len(t, apikeys, 1)

	// Filter by tag
	prodItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"prod"}})
	require.NoError(t, err)
	assert.Len(t, prodItems, 2)

	// Filter by favorite
	fav := true
	favItems, err := mgr.ListItems(types.ItemFilter{Favorite: &fav})
	require.NoError(t, err)
	assert.Len(t, favItems, 1)
	assert.Equal(t, "Login 2", favItems[0].Name)
}

// TestSearchIndex tests search functionality.
func TestSearchIndex(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add items
	itm1 := &types.Item{
		Name:   "GitHub Login",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "devuser", types.FieldURL: "https://github.com"},
	}
	itm2 := &types.Item{
		Name:   "Gmail Account",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user@gmail.com"},
	}

	require.NoError(t, mgr.AddItem(itm1))
	require.NoError(t, mgr.AddItem(itm2))

	// Search
	ids, err := idx.Search("github")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, itm1.ID, ids[0])

	ids, err = idx.Search("gmail")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, itm2.ID, ids[0])
}

// TestFindItemByQuery tests item lookup by query.
func TestFindItemByQuery(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	itm := &types.Item{
		Name:   "My Special Item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user1"},
	}
	require.NoError(t, mgr.AddItem(itm))

	// Find by ID
	found, err := findItemByQuery(mgr, idx, itm.ID)
	require.NoError(t, err)
	assert.Equal(t, itm.ID, found.ID)

	// Not found
	_, err = findItemByQuery(mgr, idx, "nonexistent-xyz-12345")
	assert.Error(t, err)
}

// TestHealthAnalysis tests the health analysis integration.
func TestHealthAnalysis(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add items with different password qualities
	weakItem := &types.Item{
		Name:   "Weak Pass Site",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldPassword: "password123"},
	}
	strongItem := &types.Item{
		Name:   "Strong Pass Site",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldPassword: "X9#kL2$mP7@nQ4&wR6!"},
	}

	require.NoError(t, mgr.AddItem(weakItem))
	require.NoError(t, mgr.AddItem(strongItem))

	items, err := mgr.AllItems()
	require.NoError(t, err)

	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(items)

	assert.Equal(t, 2, report.TotalItems)
	assert.Equal(t, 2, report.LoginItems)
	assert.GreaterOrEqual(t, report.WeakCount, 1) // password123 should be weak
}

// TestExportImportIntegration tests the full export/import cycle.
func TestExportImportIntegration(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add items
	for _, name := range []string{"Item A", "Item B", "Item C"} {
		itm := &types.Item{
			Name:   name,
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldUsername: "user-" + name, types.FieldPassword: "pass-" + name},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	// Export as CSV
	items, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, items, 3)

	// Convert to plain items for export
	plainItems := make([]types.Item, len(items))
	for i, itm := range items {
		plainItems[i] = *itm
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(plainItems)
	require.NoError(t, err)

	// Verify the exported JSON is valid
	var imported []types.Item
	err = json.NewDecoder(&buf).Decode(&imported)
	require.NoError(t, err)
	assert.Len(t, imported, 3)
}

// TestPrintItemsTableEmpty tests empty table rendering.
func TestPrintItemsTableEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "text"

	printItemsTable(nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	assert.Contains(t, buf.String(), "No items found")
}

// TestPrintItemsTableJSON tests JSON table output.
func TestPrintItemsTableJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	items := []*types.Item{
		{
			ID:     "id-1",
			Name:   "Test",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldURL: "https://test.com"},
		},
	}
	printItemsTable(items)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result []map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "id-1", result[0]["id"])
	assert.Equal(t, "Test", result[0]["name"])
	assert.Equal(t, "https://test.com", result[0]["url"])
}

// TestRootCommandHelp tests that the root command produces help output.
func TestRootCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := *rootCmd // shallow copy to avoid state leakage
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "ZeroPass")
}

// TestRootCommandVersion tests version output.
func TestRootCommandVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := *rootCmd // shallow copy
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), version)
}

// TestGenerateCommandDirect tests password generation logic directly.
func TestGenerateCommandDirect(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	// Reset flags
	genLength = 32
	genNoSymbols = false
	genNoDigits = false
	genNoUppercase = false
	genPassphrase = false
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	pw := result["password"].(string)
	assert.Len(t, pw, 32)
}

// TestGeneratePassphraseDirect tests passphrase generation directly.
func TestGeneratePassphraseDirect(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genPassphrase = true
	genWords = 4
	genSeparator = "-"
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput
	genPassphrase = false

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	pw := result["password"].(string)
	words := strings.Split(pw, "-")
	assert.Len(t, words, 4)
}

// TestIntegrationFullWorkflow tests the complete workflow:
// init vault → add items → search → list → export
func TestIntegrationFullWorkflow(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Step 1: Add various items
	loginItem := &types.Item{
		Name: "GitHub",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "dev@example.com",
			types.FieldPassword: "gh-secret-123",
			types.FieldURL:      "https://github.com",
		},
		Tags:     []string{"dev", "env:production"},
		Favorite: true,
	}
	require.NoError(t, mgr.AddItem(loginItem))

	apiItem := &types.Item{
		Name: "Stripe API",
		Type: types.ItemTypeAPIKey,
		Fields: map[string]string{
			types.FieldAPIKey:    "sk_test_12345",
			types.FieldAPISecret: "whsec_67890",
		},
		Tags: []string{"billing", "env:production"},
	}
	require.NoError(t, mgr.AddItem(apiItem))

	noteItem := &types.Item{
		Name:  "Server Notes",
		Type:  types.ItemTypeSecureNote,
		Notes: "Production server is at 10.0.0.1",
		Tags:  []string{"infra"},
	}
	require.NoError(t, mgr.AddItem(noteItem))

	// Step 2: Search
	ids, err := idx.Search("github")
	require.NoError(t, err)
	assert.Len(t, ids, 1)

	found, err := mgr.GetItem(ids[0])
	require.NoError(t, err)
	assert.Equal(t, "GitHub", found.Name)
	assert.Equal(t, "dev@example.com", found.Fields[types.FieldUsername])

	// Step 3: List all
	all, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Step 4: List with filter
	fav := true
	favItems, err := mgr.ListItems(types.ItemFilter{Favorite: &fav})
	require.NoError(t, err)
	assert.Len(t, favItems, 1)
	assert.Equal(t, "GitHub", favItems[0].Name)

	// Step 5: Filter by env tag
	prodItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:production"}})
	require.NoError(t, err)
	assert.Len(t, prodItems, 2)

	// Step 6: Update item
	loginItem.Fields[types.FieldPassword] = "new-password-456"
	err = mgr.UpdateItem(loginItem.ID, loginItem)
	require.NoError(t, err)
	assert.Equal(t, 2, loginItem.Version)

	updated, err := mgr.GetItem(loginItem.ID)
	require.NoError(t, err)
	assert.Equal(t, "new-password-456", updated.Fields[types.FieldPassword])

	// Step 7: Health check
	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(all)
	assert.Equal(t, 3, report.TotalItems)

	// Step 8: Delete
	err = mgr.DeleteItem(noteItem.ID)
	require.NoError(t, err)

	remaining, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, remaining, 2)
}

// TestRecoveryMnemonicValidation tests mnemonic validation.
func TestRecoveryMnemonicValidation(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test-vault")

	_, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	// Open fresh and try recovery
	v2, err := store.Open(vaultPath)
	require.NoError(t, err)

	err = v2.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	assert.False(t, v2.IsLocked())
	v2.Lock()

	// Invalid mnemonic should fail
	v3, err := store.Open(vaultPath)
	require.NoError(t, err)

	err = v3.UnlockWithRecovery("invalid mnemonic phrase that does not work at all")
	assert.Error(t, err)
}

// TestPrintItemTextNoSecrets verifies secrets are masked.
func TestPrintItemTextNoSecrets(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origOutput := flagOutput
	flagOutput = "text"

	itm := &types.Item{
		ID:   "test-123",
		Name: "Secret Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "visible-user",
			types.FieldPassword: "hidden-password",
		},
	}
	printItemText(itm, false)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "visible-user")
	assert.Contains(t, output, "********")
	assert.NotContains(t, output, "hidden-password")
}

// TestPrintItemTextWithSecrets verifies secrets are shown when requested.
func TestPrintItemTextWithSecrets(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	origOutput := flagOutput
	flagOutput = "text"

	itm := &types.Item{
		ID:   "test-123",
		Name: "Secret Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldPassword: "shown-password",
		},
	}
	printItemText(itm, true)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "shown-password")
}

// TestEnvTagPrefix tests environment tag format.
func TestEnvTagPrefix(t *testing.T) {
	assert.Equal(t, "env:", envTagPrefix)
	assert.True(t, strings.HasPrefix("env:production", envTagPrefix))
	assert.Equal(t, "production", strings.TrimPrefix("env:production", envTagPrefix))
}

// ---------- Additional helpers.go tests ----------

// TestGetVaultPath_CustomFlag tests vault path with custom flag.
func TestGetVaultPath_CustomFlag(t *testing.T) {
	orig := flagVaultPath
	defer func() { flagVaultPath = orig }()

	flagVaultPath = "/my/custom/vault"
	assert.Equal(t, "/my/custom/vault", getVaultPath())
}

// TestGetVaultPath_EmptyDefault tests default vault path contains expected components.
func TestGetVaultPath_EmptyDefault(t *testing.T) {
	orig := flagVaultPath
	defer func() { flagVaultPath = orig }()

	flagVaultPath = ""
	p := getVaultPath()
	assert.Contains(t, p, store.DefaultVaultDir)
	assert.Contains(t, p, "vaults")
	assert.Contains(t, p, store.DefaultVaultName)
}

// TestZeroString verifies best-effort zeroing (no crash).
func TestZeroString(t *testing.T) {
	s := "secret-data"
	assert.NotPanics(t, func() { zeroString(s) })
}

// TestParseTags_ExtraSpaces tests various whitespace combinations.
func TestParseTags_ExtraSpaces(t *testing.T) {
	result := parseTags("  a , b , c  ")
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

// TestParseTags_AllEmpty tests all-empty-after-split.
func TestParseTags_AllEmpty(t *testing.T) {
	result := parseTags(",,,")
	assert.Empty(t, result)
}

// TestParseTags_SingleTag tests a single tag without commas.
func TestParseTags_SingleTag(t *testing.T) {
	result := parseTags("production")
	assert.Equal(t, []string{"production"}, result)
}

// TestItemToOutputMap_AllSecretFieldsMasked tests all secret field types are masked.
func TestItemToOutputMap_AllSecretFieldsMasked(t *testing.T) {
	itm := &types.Item{
		ID:   "all-secrets",
		Name: "All Secrets",
		Type: types.ItemTypeCustom,
		Fields: map[string]string{
			types.FieldPassword:   "pw",
			types.FieldAPISecret:  "secret",
			types.FieldPrivateKey: "priv-key",
			types.FieldCVV:        "123",
			types.FieldPassphrase: "my-phrase",
			types.FieldUsername:   "visible-user",
		},
	}

	out := itemToOutputMap(itm, false)
	fields := out["fields"].(map[string]string)
	assert.Equal(t, "********", fields[types.FieldPassword])
	assert.Equal(t, "********", fields[types.FieldAPISecret])
	assert.Equal(t, "********", fields[types.FieldPrivateKey])
	assert.Equal(t, "********", fields[types.FieldCVV])
	assert.Equal(t, "********", fields[types.FieldPassphrase])
	assert.Equal(t, "visible-user", fields[types.FieldUsername])
}

// TestItemToOutputMap_AllFieldsVisible tests all fields visible with showSecret=true.
func TestItemToOutputMap_AllFieldsVisible(t *testing.T) {
	itm := &types.Item{
		ID:   "vis",
		Name: "Visible",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldPassword: "secret123",
			types.FieldUsername: "user",
		},
		CustomFields: map[string]string{"custom_key": "custom_val"},
		Notes:        "my notes",
		Tags:         []string{"a", "b"},
		Favorite:     true,
		Version:      3,
	}

	out := itemToOutputMap(itm, true)
	fields := out["fields"].(map[string]string)
	assert.Equal(t, "secret123", fields[types.FieldPassword])
	assert.Equal(t, "user", fields[types.FieldUsername])
	assert.Equal(t, "Visible", out["name"])
	assert.Equal(t, "my notes", out["notes"])
	assert.Equal(t, true, out["favorite"])
	assert.Equal(t, 3, out["version"])
	cfs := out["custom_fields"].(map[string]string)
	assert.Equal(t, "custom_val", cfs["custom_key"])
}

// TestItemPreview_SSHKey tests SSH key fingerprint preview.
func TestItemPreview_SSHKey(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeSSHKey,
		Fields: map[string]string{types.FieldFingerprint: "SHA256:abc123def"},
	}
	assert.Equal(t, "SHA256:abc123def", itemPreview(itm))
}

// TestItemPreview_EmptyFields tests items with no fields produce empty preview.
func TestItemPreview_EmptyFields(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{},
	}
	assert.Equal(t, "", itemPreview(itm))
}

// TestItemPreview_ShortAPIKey tests short API key (<=8 chars).
func TestItemPreview_ShortAPIKey(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeAPIKey,
		Fields: map[string]string{types.FieldAPIKey: "abc"},
	}
	assert.Equal(t, "abc", itemPreview(itm))
}

// TestItemPreview_ShortCardNumber tests credit card number <4 digits.
func TestItemPreview_ShortCardNumber(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeCreditCard,
		Fields: map[string]string{types.FieldCardNumber: "12"},
	}
	assert.Equal(t, "", itemPreview(itm))
}

// TestItemPreview_IdentityFirstNameOnly tests identity with only first name.
func TestItemPreview_IdentityFirstNameOnly(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeIdentity,
		Fields: map[string]string{types.FieldFirstName: "Alice"},
	}
	assert.Equal(t, "Alice", itemPreview(itm))
}

// TestItemPreview_IdentityLastNameOnly tests identity with only last name.
func TestItemPreview_IdentityLastNameOnly(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeIdentity,
		Fields: map[string]string{types.FieldLastName: "Smith"},
	}
	assert.Equal(t, "Smith", itemPreview(itm))
}

// TestItemPreview_EmptyNote tests note with empty notes field.
func TestItemPreview_EmptyNote(t *testing.T) {
	itm := &types.Item{
		Type:  types.ItemTypeSecureNote,
		Notes: "",
	}
	assert.Equal(t, "", itemPreview(itm))
}

// TestItemPreview_CustomType tests custom type returns empty preview.
func TestItemPreview_CustomType(t *testing.T) {
	itm := &types.Item{
		Type: types.ItemTypeCustom,
	}
	assert.Equal(t, "", itemPreview(itm))
}

// TestDefaultSecretField_SSHKey tests SSH key returns private key.
func TestDefaultSecretField_SSHKey(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeSSHKey,
		Fields: map[string]string{types.FieldPrivateKey: "-----BEGIN RSA PRIVATE KEY-----"},
	}
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----", defaultSecretField(itm))
}

// TestDefaultSecretField_CustomWithPassword tests custom type with password field.
func TestDefaultSecretField_CustomWithPassword(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeCustom,
		Fields: map[string]string{types.FieldPassword: "fallback-pw"},
	}
	assert.Equal(t, "fallback-pw", defaultSecretField(itm))
}

// TestDefaultSecretField_IdentityNoSecret tests identity type returns empty.
func TestDefaultSecretField_IdentityNoSecret(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeIdentity,
		Fields: map[string]string{types.FieldFirstName: "John"},
	}
	assert.Equal(t, "", defaultSecretField(itm))
}

// TestDefaultSecretField_EmptyLogin tests login with no password.
func TestDefaultSecretField_EmptyLogin(t *testing.T) {
	itm := &types.Item{
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user"},
	}
	assert.Equal(t, "", defaultSecretField(itm))
}

// ---------- Additional format/display tests ----------

// TestPrintItemJSON tests printItem in JSON mode.
func TestPrintItemJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	itm := &types.Item{
		ID:   "json-test",
		Name: "JSON Test",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "user1",
			types.FieldPassword: "pass1",
		},
	}
	printItem(itm, false)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "json-test", result["id"])
	assert.Equal(t, "JSON Test", result["name"])
	fields := result["fields"].(map[string]interface{})
	assert.Equal(t, "********", fields[types.FieldPassword])
}

// TestPrintItemText_WithTags tests text output includes tags.
func TestPrintItemText_WithTags(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "text"

	itm := &types.Item{
		ID:       "tag-test",
		Name:     "Tagged Item",
		Type:     types.ItemTypeLogin,
		Fields:   map[string]string{types.FieldUsername: "user"},
		Tags:     []string{"prod", "critical"},
		Favorite: true,
		Notes:    "important notes",
	}
	printItemText(itm, true)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Tagged Item")
	assert.Contains(t, output, "prod, critical")
	assert.Contains(t, output, "★")
	assert.Contains(t, output, "important notes")
	assert.Contains(t, output, "tag-test")
}

// TestPrintItemText_CustomFields tests custom fields are displayed.
func TestPrintItemText_CustomFields(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	itm := &types.Item{
		ID:           "cf-test",
		Name:         "Custom Fields Item",
		Type:         types.ItemTypeCustom,
		Fields:       map[string]string{},
		CustomFields: map[string]string{"env": "staging", "region": "us-east-1"},
	}
	printItemText(itm, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	assert.Contains(t, output, "Custom Fields Item")
	// Custom fields should be present
	assert.Contains(t, output, "staging")
	assert.Contains(t, output, "us-east-1")
}

// TestFormatOutputText tests non-JSON format (falls back to JSON currently).
func TestFormatOutputText(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "text"

	formatOutput(map[string]string{"hello": "world"})

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	// Even in text mode, formatOutput uses JSON
	var result map[string]string
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "world", result["hello"])
}

// TestPrintItemsTable_MultipleItems tests table output with multiple items.
func TestPrintItemsTable_MultipleItems(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "text"

	items := []*types.Item{
		{
			ID:     "id-1",
			Name:   "Login One",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldURL: "https://one.com", types.FieldUsername: "user1"},
		},
		{
			ID:       "id-2",
			Name:     "API Key Two",
			Type:     types.ItemTypeAPIKey,
			Fields:   map[string]string{types.FieldAPIKey: "key-12345678901234"},
			Favorite: true,
		},
		{
			ID:   "id-3",
			Name: "My Note",
			Type: types.ItemTypeSecureNote,
			Notes: "Short note",
		},
	}
	printItemsTable(items)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Login One")
	assert.Contains(t, output, "API Key Two")
	assert.Contains(t, output, "My Note")
	assert.Contains(t, output, "Total: 3 items")
	assert.Contains(t, output, "★")
}

// TestPrintItemsTable_LongName tests truncation of long names.
func TestPrintItemsTable_LongName(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "text"

	items := []*types.Item{
		{
			ID:     "id-long",
			Name:   "This is a very long item name that exceeds the maximum allowed display width for names in the table",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{},
		},
	}
	printItemsTable(items)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	assert.Contains(t, output, "…")
	assert.Contains(t, output, "Total: 1 items")
}

// TestPrintItemsTable_JSONFormat tests JSON table output with multiple items.
func TestPrintItemsTable_JSONFormat(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	items := []*types.Item{
		{
			ID:       "id-1",
			Name:     "Item One",
			Type:     types.ItemTypeLogin,
			Fields:   map[string]string{types.FieldURL: "https://one.com", types.FieldUsername: "user1"},
			Tags:     []string{"dev"},
			Favorite: true,
		},
		{
			ID:     "id-2",
			Name:   "Item Two",
			Type:   types.ItemTypeAPIKey,
			Fields: map[string]string{types.FieldAPIKey: "key123"},
		},
	}
	printItemsTable(items)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result []map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "id-1", result[0]["id"])
	assert.Equal(t, "https://one.com", result[0]["url"])
	assert.Equal(t, "user1", result[0]["username"])
	assert.Equal(t, true, result[0]["favorite"])
}

// ---------- .env / run.go tests ----------

// TestParseEnvFile_OnlyComments tests file with only comments.
func TestParseEnvFile_OnlyComments(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `# This is a comment
# Another comment
# And another one`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.False(t, hasRefs)
	assert.Empty(t, vars)
}

// TestParseEnvFile_EmptyFile tests an empty file.
func TestParseEnvFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte(""), 0600))

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.False(t, hasRefs)
	assert.Empty(t, vars)
}

// TestParseEnvFile_MixedRefsAndLiterals tests mixed zp:// refs and literals.
func TestParseEnvFile_MixedRefsAndLiterals(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD=zp://db-cred/password
API_URL=https://api.example.com
API_KEY=zp://api-cred/api_key
REDIS_URL="redis://localhost:6379"
DEBUG=true`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.True(t, hasRefs)
	assert.Equal(t, "localhost", vars["DB_HOST"])
	assert.Equal(t, "5432", vars["DB_PORT"])
	assert.Equal(t, "zp://db-cred/password", vars["DB_PASSWORD"])
	assert.Equal(t, "https://api.example.com", vars["API_URL"])
	assert.Equal(t, "zp://api-cred/api_key", vars["API_KEY"])
	assert.Equal(t, "redis://localhost:6379", vars["REDIS_URL"])
	assert.Equal(t, "true", vars["DEBUG"])
}

// TestParseEnvFile_ValuesWithEquals tests values containing '=' sign.
func TestParseEnvFile_ValuesWithEquals(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `CONNECTION=postgres://user:pass@host/db?sslmode=require
ENCODED=base64==value`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.False(t, hasRefs)
	assert.Equal(t, "postgres://user:pass@host/db?sslmode=require", vars["CONNECTION"])
	assert.Equal(t, "base64==value", vars["ENCODED"])
}

// TestParseEnvFile_EmptyValue tests keys with empty values.
func TestParseEnvFile_EmptyValue(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `EMPTY_KEY=
ANOTHER=value`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, _, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "", vars["EMPTY_KEY"])
	assert.Equal(t, "value", vars["ANOTHER"])
}

// TestParseEnvFile_WhitespaceAroundEquals tests whitespace around '='.
func TestParseEnvFile_WhitespaceAroundEquals(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `  KEY1  =  value1  
KEY2=value2`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, _, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "value1", vars["KEY1"])
	assert.Equal(t, "value2", vars["KEY2"])
}

// TestParseEnvFile_MultipleZPRefs tests multiple zp:// references.
func TestParseEnvFile_MultipleZPRefs(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `SECRET1=zp://item1/password
SECRET2=zp://item2/api_key
SECRET3=zp://item3/private_key`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.True(t, hasRefs)
	assert.Len(t, vars, 3)
	assert.Equal(t, "zp://item1/password", vars["SECRET1"])
	assert.Equal(t, "zp://item2/api_key", vars["SECRET2"])
	assert.Equal(t, "zp://item3/private_key", vars["SECRET3"])
}

// TestUnquoteEnvValue_Mismatched tests mismatched quotes stay as-is.
func TestUnquoteEnvValue_Mismatched(t *testing.T) {
	assert.Equal(t, `"hello'`, unquoteEnvValue(`"hello'`))
	assert.Equal(t, `'hello"`, unquoteEnvValue(`'hello"`))
}

// TestUnquoteEnvValue_NestedQuotes tests nested quotes.
func TestUnquoteEnvValue_NestedQuotes(t *testing.T) {
	assert.Equal(t, `he said "hi"`, unquoteEnvValue(`'he said "hi"'`))
	assert.Equal(t, `it's fine`, unquoteEnvValue(`"it's fine"`))
}

// TestUnquoteEnvValue_Empty tests empty string.
func TestUnquoteEnvValue_Empty(t *testing.T) {
	assert.Equal(t, "", unquoteEnvValue(""))
}

// TestResolveZPReference_NonReference tests non-zp:// values pass through.
func TestResolveZPReference_NonReference(t *testing.T) {
	result, err := resolveZPReference("plain-value", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "plain-value", result)
}

// TestResolveZPReference_InvalidFormat tests invalid zp:// format.
func TestResolveZPReference_InvalidFormat(t *testing.T) {
	_, err := resolveZPReference("zp://no-field-specified", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid zp:// reference")
}

// TestResolveZPReference_Vault tests resolving against a real vault.
func TestResolveZPReference_Vault(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	itm := &types.Item{
		Name: "my-db",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldPassword: "db-secret-123",
			types.FieldUsername: "dbuser",
		},
		Notes: "Database credentials for prod",
	}
	require.NoError(t, mgr.AddItem(itm))

	// Resolve password field
	result, err := resolveZPReference("zp://my-db/password", mgr, idx)
	require.NoError(t, err)
	assert.Equal(t, "db-secret-123", result)

	// Resolve username field
	result, err = resolveZPReference("zp://my-db/username", mgr, idx)
	require.NoError(t, err)
	assert.Equal(t, "dbuser", result)

	// Resolve notes field
	result, err = resolveZPReference("zp://my-db/notes", mgr, idx)
	require.NoError(t, err)
	assert.Equal(t, "Database credentials for prod", result)
}

// TestResolveZPReference_MissingItem tests resolving a non-existent item.
func TestResolveZPReference_MissingItem(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	_, err := resolveZPReference("zp://nonexistent-item/password", mgr, idx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolve item")
}

// TestResolveZPReference_MissingField tests resolving a non-existent field.
func TestResolveZPReference_MissingField(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	itm := &types.Item{
		Name:   "test-item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user"},
	}
	require.NoError(t, mgr.AddItem(itm))

	_, err := resolveZPReference("zp://test-item/nonexistent_field", mgr, idx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field")
	assert.Contains(t, err.Error(), "not found")
}

// TestResolveZPReference_CustomField tests resolving a custom field.
func TestResolveZPReference_CustomField(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	itm := &types.Item{
		Name:         "custom-item",
		Type:         types.ItemTypeCustom,
		Fields:       map[string]string{},
		CustomFields: map[string]string{"my_custom": "custom-value"},
	}
	require.NoError(t, mgr.AddItem(itm))

	result, err := resolveZPReference("zp://custom-item/my_custom", mgr, idx)
	require.NoError(t, err)
	assert.Equal(t, "custom-value", result)
}

// ---------- env.go tests ----------

// TestRunEnvCreate_ValidName tests env create with a valid name.
func TestRunEnvCreate_ValidName(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "json"
	defer func() { flagOutput = origOutput }()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runEnvCreate(envCreateCmd, []string{"production"})

	w.Close()
	os.Stdout = old

	require.NoError(t, err)
	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "created", result["status"])
	assert.Equal(t, "production", result["environment"])
	assert.Equal(t, "env:production", result["tag"])
}

// TestRunEnvCreate_InvalidName tests env create with invalid names.
func TestRunEnvCreate_InvalidName(t *testing.T) {
	tests := []string{"has space", "has\ttab", "has,comma", "has:colon", "has\nnewline"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			err := runEnvCreate(envCreateCmd, []string{name})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot contain")
		})
	}
}

// ---------- generate.go tests ----------

// TestGenerateNoSymbols tests password generation without symbols.
func TestGenerateNoSymbols(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genLength = 16
	genNoSymbols = true
	genNoDigits = false
	genNoUppercase = false
	genPassphrase = false
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput
	genNoSymbols = false

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	pw := result["password"].(string)
	assert.Len(t, pw, 16)
	// No symbols expected
	for _, c := range pw {
		assert.False(t, strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", c),
			"found unexpected symbol: %c", c)
	}
}

// TestGenerateNoDigits tests password generation without digits.
func TestGenerateNoDigits(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genLength = 20
	genNoSymbols = false
	genNoDigits = true
	genNoUppercase = false
	genPassphrase = false
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput
	genNoDigits = false

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	pw := result["password"].(string)
	assert.Len(t, pw, 20)
	for _, c := range pw {
		assert.False(t, c >= '0' && c <= '9', "found unexpected digit: %c", c)
	}
}

// TestGenerateNoUppercase tests password generation without uppercase.
func TestGenerateNoUppercase(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genLength = 24
	genNoSymbols = false
	genNoDigits = false
	genNoUppercase = true
	genPassphrase = false
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput
	genNoUppercase = false

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	pw := result["password"].(string)
	assert.Len(t, pw, 24)
	for _, c := range pw {
		assert.False(t, c >= 'A' && c <= 'Z', "found unexpected uppercase: %c", c)
	}
}

// TestGeneratePassphrase_CustomSeparator tests passphrase with custom separator.
func TestGeneratePassphrase_CustomSeparator(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genPassphrase = true
	genWords = 5
	genSeparator = "_"
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput
	genPassphrase = false
	genSeparator = "-"

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	pw := result["password"].(string)
	words := strings.Split(pw, "_")
	assert.Len(t, words, 5)
}

// TestGenerateJSON_StrengthInfo tests that JSON output includes strength info.
func TestGenerateJSON_StrengthInfo(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	origOutput := flagOutput
	flagOutput = "json"

	genLength = 32
	genNoSymbols = false
	genNoDigits = false
	genNoUppercase = false
	genPassphrase = false
	genCopy = false

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	flagOutput = origOutput

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	strength, ok := result["strength"].(map[string]interface{})
	require.True(t, ok)
	assert.NotNil(t, strength["score"])
	assert.NotNil(t, strength["feedback"])
}

// ---------- add.go field building tests ----------

// TestBuildIdentityFields tests building identity item fields.
func TestBuildIdentityFields(t *testing.T) {
	// Save and restore globals
	origFirst := addFirstName
	origLast := addLastName
	origEmail := addEmail
	origPhone := addPhone
	origAddr := addAddress
	defer func() {
		addFirstName = origFirst
		addLastName = origLast
		addEmail = origEmail
		addPhone = origPhone
		addAddress = origAddr
	}()

	addFirstName = "Jane"
	addLastName = "Doe"
	addEmail = "jane@example.com"
	addPhone = "+1-555-1234"
	addAddress = "123 Main St"

	fields := make(map[string]string)
	buildIdentityFields(fields)

	assert.Equal(t, "Jane", fields[types.FieldFirstName])
	assert.Equal(t, "Doe", fields[types.FieldLastName])
	assert.Equal(t, "jane@example.com", fields[types.FieldEmail])
	assert.Equal(t, "+1-555-1234", fields[types.FieldPhone])
	assert.Equal(t, "123 Main St", fields[types.FieldAddress])
}

// TestBuildIdentityFields_Empty tests building identity with no fields set.
func TestBuildIdentityFields_Empty(t *testing.T) {
	origFirst := addFirstName
	origLast := addLastName
	origEmail := addEmail
	origPhone := addPhone
	origAddr := addAddress
	defer func() {
		addFirstName = origFirst
		addLastName = origLast
		addEmail = origEmail
		addPhone = origPhone
		addAddress = origAddr
	}()

	addFirstName = ""
	addLastName = ""
	addEmail = ""
	addPhone = ""
	addAddress = ""

	fields := make(map[string]string)
	buildIdentityFields(fields)

	assert.Empty(t, fields)
}

// TestBuildAPIKeyFields tests building API key fields from flags.
func TestBuildAPIKeyFields_WithFlags(t *testing.T) {
	origKey := addAPIKey
	origSecret := addAPISecret
	origURL := addURL
	defer func() {
		addAPIKey = origKey
		addAPISecret = origSecret
		addURL = origURL
	}()

	addAPIKey = "ak_test_key"
	addAPISecret = "as_test_secret"
	addURL = "https://api.example.com"

	fields := make(map[string]string)
	buildAPIKeyFields(fields)

	assert.Equal(t, "ak_test_key", fields[types.FieldAPIKey])
	assert.Equal(t, "as_test_secret", fields[types.FieldAPISecret])
	assert.Equal(t, "https://api.example.com", fields[types.FieldEndpoint])
}

// TestBuildCreditCardFields tests building credit card fields from flags.
func TestBuildCreditCardFields_WithFlags(t *testing.T) {
	origNum := addCardNumber
	origHolder := addCardHolder
	origExpiry := addCardExpiry
	origCVV := addCardCVV
	defer func() {
		addCardNumber = origNum
		addCardHolder = origHolder
		addCardExpiry = origExpiry
		addCardCVV = origCVV
	}()

	addCardNumber = "4111111111111111"
	addCardHolder = "John Doe"
	addCardExpiry = "12/25"
	addCardCVV = "123"

	fields := make(map[string]string)
	buildCreditCardFields(fields)

	assert.Equal(t, "4111111111111111", fields[types.FieldCardNumber])
	assert.Equal(t, "John Doe", fields[types.FieldCardHolder])
	assert.Equal(t, "12/25", fields[types.FieldExpiry])
	assert.Equal(t, "123", fields[types.FieldCVV])
}

// TestBuildSSHKeyFields_WithFiles tests building SSH key fields from files.
func TestBuildSSHKeyFields_WithFiles(t *testing.T) {
	dir := t.TempDir()
	privPath := filepath.Join(dir, "id_rsa")
	pubPath := filepath.Join(dir, "id_rsa.pub")

	privContent := "-----BEGIN RSA PRIVATE KEY-----\nMIIEo...\n-----END RSA PRIVATE KEY-----\n"
	pubContent := "ssh-rsa AAAAB3... user@host\n"

	require.NoError(t, os.WriteFile(privPath, []byte(privContent), 0600))
	require.NoError(t, os.WriteFile(pubPath, []byte(pubContent), 0644))

	origPriv := addPrivateKeyFile
	origPub := addPublicKeyFile
	defer func() {
		addPrivateKeyFile = origPriv
		addPublicKeyFile = origPub
	}()

	addPrivateKeyFile = privPath
	addPublicKeyFile = pubPath

	fields := make(map[string]string)
	err := buildSSHKeyFields(fields)
	require.NoError(t, err)

	assert.Equal(t, privContent, fields[types.FieldPrivateKey])
	assert.Equal(t, pubContent, fields[types.FieldPublicKey])
}

// TestBuildSSHKeyFields_MissingFile tests SSH key with non-existent file.
func TestBuildSSHKeyFields_MissingFile(t *testing.T) {
	origPriv := addPrivateKeyFile
	defer func() { addPrivateKeyFile = origPriv }()

	addPrivateKeyFile = "/nonexistent/id_rsa"
	addPublicKeyFile = ""

	fields := make(map[string]string)
	err := buildSSHKeyFields(fields)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read private key file")
}

// TestBuildSSHKeyFields_Empty tests SSH key with no files set.
func TestBuildSSHKeyFields_Empty(t *testing.T) {
	origPriv := addPrivateKeyFile
	origPub := addPublicKeyFile
	defer func() {
		addPrivateKeyFile = origPriv
		addPublicKeyFile = origPub
	}()

	addPrivateKeyFile = ""
	addPublicKeyFile = ""

	fields := make(map[string]string)
	err := buildSSHKeyFields(fields)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

// ---------- score/health display tests ----------

// TestScoreDisplay_Boundaries tests exact boundaries for score display.
func TestScoreDisplay_Boundaries(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "Excellent"},
		{90, "Excellent"},
		{89, "Good"},
		{70, "Good"},
		{69, "Fair"},
		{50, "Fair"},
		{49, "Poor"},
		{0, "Poor"},
	}
	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.expected, " ", ""), func(t *testing.T) {
			result := scoreDisplay(tt.score)
			assert.Contains(t, result, tt.expected)
		})
	}
}

// ---------- Vault workflow tests ----------

// TestItemCRUD_AllTypes tests CRUD for all item types.
func TestItemCRUD_AllTypes(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	items := []*types.Item{
		{
			Name:   "Login",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldUsername: "u", types.FieldPassword: "p", types.FieldURL: "https://x.com"},
		},
		{
			Name:   "API",
			Type:   types.ItemTypeAPIKey,
			Fields: map[string]string{types.FieldAPIKey: "k", types.FieldAPISecret: "s"},
		},
		{
			Name:   "SSH",
			Type:   types.ItemTypeSSHKey,
			Fields: map[string]string{types.FieldPrivateKey: "priv", types.FieldPublicKey: "pub"},
		},
		{
			Name:  "Note",
			Type:  types.ItemTypeSecureNote,
			Notes: "my secure note",
		},
		{
			Name: "Card",
			Type: types.ItemTypeCreditCard,
			Fields: map[string]string{
				types.FieldCardNumber: "4111111111111111",
				types.FieldCardHolder: "Test User",
				types.FieldExpiry:     "12/25",
				types.FieldCVV:        "123",
			},
		},
		{
			Name: "Identity",
			Type: types.ItemTypeIdentity,
			Fields: map[string]string{
				types.FieldFirstName: "Alice",
				types.FieldLastName:  "Wonderland",
				types.FieldEmail:     "alice@example.com",
			},
		},
		{
			Name:         "Custom",
			Type:         types.ItemTypeCustom,
			Fields:       map[string]string{},
			CustomFields: map[string]string{"env": "prod"},
		},
	}

	for _, itm := range items {
		require.NoError(t, mgr.AddItem(itm), "add %s", itm.Name)
		assert.NotEmpty(t, itm.ID, "ID for %s", itm.Name)
	}

	// Verify all items exist
	all, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, all, 7)

	// Delete all
	for _, itm := range items {
		require.NoError(t, mgr.DeleteItem(itm.ID), "delete %s", itm.Name)
	}

	all, err = mgr.AllItems()
	require.NoError(t, err)
	assert.Empty(t, all)
}

// TestFindItemByQuery_ByName tests finding items by name.
func TestFindItemByQuery_ByName(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	itm := &types.Item{
		Name:   "Unique Name Item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "u"},
	}
	require.NoError(t, mgr.AddItem(itm))

	// Search by name via index
	found, err := findItemByQuery(mgr, idx, "Unique Name Item")
	require.NoError(t, err)
	assert.Equal(t, itm.ID, found.ID)
}

// ---------- Recovery tests ----------

// TestRecoveryMnemonic_FullCycle tests create → lock → recover cycle.
func TestRecoveryMnemonic_FullCycle(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test-vault")

	v, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	require.NotEmpty(t, result.Mnemonic)

	// Add an item while unlocked
	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)

	itm := &types.Item{
		Name:   "Pre-Recovery Item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user1"},
	}
	require.NoError(t, mgr.AddItem(itm))
	idx.Close()
	v.Lock()

	// Recover with mnemonic
	v2, err := store.Open(vaultPath)
	require.NoError(t, err)
	err = v2.UnlockWithRecovery(result.Mnemonic)
	require.NoError(t, err)
	assert.False(t, v2.IsLocked())

	// Verify the item is still accessible
	idx2, err := index.Open(v2.IndexPath())
	require.NoError(t, err)
	defer idx2.Close()
	mgr2 := item.NewManager(v2.ItemsPath(), v2.VaultKey, idx2)

	retrieved, err := mgr2.GetItem(itm.ID)
	require.NoError(t, err)
	assert.Equal(t, "Pre-Recovery Item", retrieved.Name)
	v2.Lock()
}

// TestOpenVault_NonExistent tests opening a non-existent vault.
func TestOpenVault_NonExistent(t *testing.T) {
	_, err := openVault("/nonexistent/path/vault")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open vault")
}

// ---------- health.go display tests ----------

// TestHealthAnalysis_NoItems tests health analysis with empty vault.
func TestHealthAnalysis_NoItems(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	items, err := mgr.AllItems()
	require.NoError(t, err)

	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(items)
	assert.Equal(t, 0, report.TotalItems)
	assert.Equal(t, 0, report.LoginItems)
	assert.Equal(t, 0, report.WeakCount)
	assert.Equal(t, 0, report.ReusedCount)
}

// TestHealthAnalysis_ReusedPasswords tests detection of reused passwords.
func TestHealthAnalysis_ReusedPasswords(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add two items with the same password
	for _, name := range []string{"Site A", "Site B"} {
		itm := &types.Item{
			Name: name,
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldPassword: "SameP@ssw0rd!Reused!!",
			},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	items, err := mgr.AllItems()
	require.NoError(t, err)

	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(items)
	assert.Equal(t, 2, report.TotalItems)
	assert.GreaterOrEqual(t, report.ReusedCount, 1)
}

// ---------- Full integration extended tests ----------

// TestIntegrationWorkflow_EditAndGet tests the edit→get cycle.
func TestIntegrationWorkflow_EditAndGet(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add item
	itm := &types.Item{
		Name: "Edit Target",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "original-user",
			types.FieldPassword: "original-pass",
			types.FieldURL:      "https://original.com",
		},
		Tags: []string{"v1"},
	}
	require.NoError(t, mgr.AddItem(itm))

	// Find it
	found, err := findItemByQuery(mgr, idx, itm.ID)
	require.NoError(t, err)
	assert.Equal(t, "Edit Target", found.Name)

	// Edit fields
	found.Fields[types.FieldPassword] = "new-password-strong!"
	found.Fields[types.FieldURL] = "https://updated.com"
	found.Name = "Updated Edit Target"
	found.Tags = []string{"v2"}
	err = mgr.UpdateItem(found.ID, found)
	require.NoError(t, err)
	assert.Equal(t, 2, found.Version)

	// Verify
	retrieved, err := mgr.GetItem(found.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Edit Target", retrieved.Name)
	assert.Equal(t, "new-password-strong!", retrieved.Fields[types.FieldPassword])
	assert.Equal(t, "https://updated.com", retrieved.Fields[types.FieldURL])
	assert.Equal(t, []string{"v2"}, retrieved.Tags)
}

// TestIntegrationWorkflow_SearchByType tests searching and filtering by type.
func TestIntegrationWorkflow_SearchByType(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add mixed types
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Login1", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "u1"},
	}))
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "API1", Type: types.ItemTypeAPIKey,
		Fields: map[string]string{types.FieldAPIKey: "k1"},
	}))
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Note1", Type: types.ItemTypeSecureNote,
		Notes: "important note",
	}))

	// Filter by type
	logins, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeLogin})
	require.NoError(t, err)
	assert.Len(t, logins, 1)
	assert.Equal(t, "Login1", logins[0].Name)

	apis, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeAPIKey})
	require.NoError(t, err)
	assert.Len(t, apis, 1)

	notes, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeSecureNote})
	require.NoError(t, err)
	assert.Len(t, notes, 1)
}

// TestIntegrationWorkflow_Sorting tests different sort orders.
func TestIntegrationWorkflow_Sorting(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Zebra", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "z"},
	}))
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Alpha", Type: types.ItemTypeAPIKey,
		Fields: map[string]string{types.FieldAPIKey: "a"},
	}))
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Middle", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "m"},
	}))

	// Sort by name
	items, err := mgr.ListItems(types.ItemFilter{SortBy: types.SortByName})
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, "Alpha", items[0].Name)
	assert.Equal(t, "Middle", items[1].Name)
	assert.Equal(t, "Zebra", items[2].Name)

	// Sort by type
	items, err = mgr.ListItems(types.ItemFilter{SortBy: types.SortByType})
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

// TestIntegrationWorkflow_ExportImportJSON tests JSON export and reimport cycle.
func TestIntegrationWorkflow_ExportImportJSON(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add items
	for i := 0; i < 5; i++ {
		itm := &types.Item{
			Name: fmt.Sprintf("Item-%d", i),
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: fmt.Sprintf("user%d", i),
				types.FieldPassword: fmt.Sprintf("pass%d", i),
			},
			Tags: []string{fmt.Sprintf("tag%d", i)},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	// Export
	items, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, items, 5)

	// Export to JSON
	plainItems := make([]types.Item, len(items))
	for i, itm := range items {
		plainItems[i] = *itm
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(plainItems)
	require.NoError(t, err)

	// Decode and verify
	var imported []types.Item
	err = json.NewDecoder(&buf).Decode(&imported)
	require.NoError(t, err)
	assert.Len(t, imported, 5)

	for i, imp := range imported {
		assert.Equal(t, fmt.Sprintf("Item-%d", i), imp.Name)
		assert.Equal(t, fmt.Sprintf("user%d", i), imp.Fields[types.FieldUsername])
	}
}

// TestRootCommand_UnknownFlag tests that unknown flags produce an error.
func TestRootCommand_UnknownFlag(t *testing.T) {
	cmd := *rootCmd
	cmd.SetArgs([]string{"--unknown-flag"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestRootCommand_OutputFlag tests setting output flag.
func TestRootCommand_OutputFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := *rootCmd
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--output", "json", "--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

// TestVersion_Constant tests the version constant.
func TestVersion_Constant(t *testing.T) {
	assert.NotEmpty(t, version)
	assert.Contains(t, version, "0.1")
}

// ---------- buildLoginFields tests ----------

// TestBuildLoginFields_WithFlags tests building login fields when flags are pre-set.
func TestBuildLoginFields_WithFlags(t *testing.T) {
	origUser := addUsername
	origPW := addPassword
	origURL := addURL
	origGenPW := addGeneratePassword
	defer func() {
		addUsername = origUser
		addPassword = origPW
		addURL = origURL
		addGeneratePassword = origGenPW
	}()

	addUsername = "testuser"
	addPassword = "testpass123"
	addURL = "https://example.com"
	addGeneratePassword = false

	fields := make(map[string]string)
	err := buildLoginFields(addCmd, fields)
	require.NoError(t, err)

	assert.Equal(t, "testuser", fields[types.FieldUsername])
	assert.Equal(t, "testpass123", fields[types.FieldPassword])
	assert.Equal(t, "https://example.com", fields[types.FieldURL])
}

// TestBuildLoginFields_GeneratePassword tests auto-generate password.
func TestBuildLoginFields_GeneratePassword(t *testing.T) {
	origUser := addUsername
	origPW := addPassword
	origURL := addURL
	origGenPW := addGeneratePassword
	defer func() {
		addUsername = origUser
		addPassword = origPW
		addURL = origURL
		addGeneratePassword = origGenPW
	}()

	addUsername = "genuser"
	addPassword = ""
	addURL = ""
	addGeneratePassword = true

	fields := make(map[string]string)
	err := buildLoginFields(addCmd, fields)
	require.NoError(t, err)

	assert.Equal(t, "genuser", fields[types.FieldUsername])
	assert.NotEmpty(t, fields[types.FieldPassword])
	assert.Len(t, fields[types.FieldPassword], 32)
}

// TestBuildLoginFields_NoURL tests building login fields without URL.
func TestBuildLoginFields_NoURL(t *testing.T) {
	origUser := addUsername
	origPW := addPassword
	origURL := addURL
	origGenPW := addGeneratePassword
	defer func() {
		addUsername = origUser
		addPassword = origPW
		addURL = origURL
		addGeneratePassword = origGenPW
	}()

	addUsername = "user"
	addPassword = "pass"
	addURL = ""
	addGeneratePassword = false

	fields := make(map[string]string)
	err := buildLoginFields(addCmd, fields)
	require.NoError(t, err)

	assert.Equal(t, "user", fields[types.FieldUsername])
	assert.Equal(t, "pass", fields[types.FieldPassword])
	_, hasURL := fields[types.FieldURL]
	assert.False(t, hasURL)
}

// ---------- addImportedItems tests ----------

// TestAddImportedItems_Empty tests importing zero items.
func TestAddImportedItems_Empty(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)
	imported := []types.Item{}

	err := addImportedItems(addCmd, mgr, imported)
	require.NoError(t, err)
}

// TestAddImportedItems_Success tests importing valid items.
func TestAddImportedItems_Success(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	origOutput := flagOutput
	flagOutput = "json"
	defer func() { flagOutput = origOutput }()

	imported := []types.Item{
		{
			Name: "Imported Login 1",
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: "iuser1",
				types.FieldPassword: "ipass1",
			},
		},
		{
			Name: "Imported Login 2",
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: "iuser2",
				types.FieldPassword: "ipass2",
			},
		},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := addImportedItems(addCmd, mgr, imported)

	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	// Verify items were added
	all, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

// ---------- runRecovery tests ----------

// TestRunRecovery_ValidateValid tests validating a syntactically valid mnemonic.
func TestRunRecovery_ValidateValid(t *testing.T) {
	origOutput := flagOutput
	origValidate := recoveryValidate
	defer func() {
		flagOutput = origOutput
		recoveryValidate = origValidate
	}()

	// Create a vault to get a valid mnemonic
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "rv-vault")
	_, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	flagOutput = "json"
	recoveryValidate = result.Mnemonic

	output := captureStdout(t, func() {
		err := runRecovery(recoveryCmd, nil)
		require.NoError(t, err)
	})

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &res))
	assert.Equal(t, true, res["valid"])
}

// TestRunRecovery_ValidateInvalid tests validating an invalid mnemonic.
func TestRunRecovery_ValidateInvalid(t *testing.T) {
	origOutput := flagOutput
	origValidate := recoveryValidate
	defer func() {
		flagOutput = origOutput
		recoveryValidate = origValidate
	}()

	flagOutput = "json"
	recoveryValidate = "these words are not a valid mnemonic phrase at all nope"

	output := captureStdout(t, func() {
		err := runRecovery(recoveryCmd, nil)
		require.NoError(t, err)
	})

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &res))
	assert.Equal(t, false, res["valid"])
}

// TestRunRecovery_ValidateText tests mnemonic validation with text output.
func TestRunRecovery_ValidateText(t *testing.T) {
	origOutput := flagOutput
	origValidate := recoveryValidate
	defer func() {
		flagOutput = origOutput
		recoveryValidate = origValidate
	}()

	flagOutput = "text"
	recoveryValidate = "invalid mnemonic"

	// Validation output goes to stderr, just verify no error
	err := runRecovery(recoveryCmd, nil)
	assert.NoError(t, err)
}

// ---------- runEnvCreate text output test ----------

// TestRunEnvCreate_TextOutput tests env create with text output.
func TestRunEnvCreate_TextOutput(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "text"
	defer func() { flagOutput = origOutput }()

	// Text output goes to stderr via cmd.ErrOrStderr(); just verify no error
	err := runEnvCreate(envCreateCmd, []string{"staging"})
	require.NoError(t, err)
}

// ---------- Execute function test ----------

// TestExecute_Help tests the Execute function with help flag.
func TestExecute_Help(t *testing.T) {
	// Save and restore os.Args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"zeropass", "--help"}

	// Execute should not error on --help
	// Note: Execute prints to stdout, which is fine
	err := Execute()
	assert.NoError(t, err)
}

// ---------- newItemManager test ----------

// TestNewItemManager tests creating an item manager from a vault.
func TestNewItemManager(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx, err := newItemManager(v)
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.NotNil(t, idx)
	idx.Close()
}

// ---------- runGenerate text mode ----------

// TestRunGenerate_TextMode tests generate in text mode.
func TestRunGenerate_TextMode(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "text"
	defer func() { flagOutput = origOutput }()

	genLength = 16
	genNoSymbols = false
	genNoDigits = false
	genNoUppercase = false
	genPassphrase = false
	genCopy = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	// Text mode prints the password on stdout
	assert.NotEmpty(t, output)
	// Password should be on the first line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines[0], 16)
}

// TestRunGenerate_PassphraseTextMode tests passphrase generation in text mode.
func TestRunGenerate_PassphraseTextMode(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "text"
	defer func() { flagOutput = origOutput }()

	genPassphrase = true
	genWords = 4
	genSeparator = "-"
	genCopy = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGenerate(generateCmd, nil)

	w.Close()
	os.Stdout = old
	genPassphrase = false

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	// First line should be the passphrase
	words := strings.Split(lines[0], "-")
	assert.Len(t, words, 4)
}

// ---------- buildCreditCardFields full coverage ----------

// TestBuildCreditCardFields_PartialFlags tests credit card with partial flags.
func TestBuildCreditCardFields_PartialFlags(t *testing.T) {
	origNum := addCardNumber
	origHolder := addCardHolder
	origExpiry := addCardExpiry
	origCVV := addCardCVV
	defer func() {
		addCardNumber = origNum
		addCardHolder = origHolder
		addCardExpiry = origExpiry
		addCardCVV = origCVV
	}()

	// Only set number and holder
	addCardNumber = "5500000000000004"
	addCardHolder = "Jane Smith"
	addCardExpiry = ""
	addCardCVV = ""

	fields := make(map[string]string)
	buildCreditCardFields(fields)

	assert.Equal(t, "5500000000000004", fields[types.FieldCardNumber])
	assert.Equal(t, "Jane Smith", fields[types.FieldCardHolder])
	_, hasExpiry := fields[types.FieldExpiry]
	assert.False(t, hasExpiry)
	_, hasCVV := fields[types.FieldCVV]
	assert.False(t, hasCVV)
}

// ---------- buildAPIKeyFields edge cases ----------

// TestBuildAPIKeyFields_NoURL tests API key without URL.
func TestBuildAPIKeyFields_NoURL(t *testing.T) {
	origKey := addAPIKey
	origSecret := addAPISecret
	origURL := addURL
	defer func() {
		addAPIKey = origKey
		addAPISecret = origSecret
		addURL = origURL
	}()

	addAPIKey = "mykey"
	addAPISecret = "mysecret"
	addURL = ""

	fields := make(map[string]string)
	buildAPIKeyFields(fields)

	assert.Equal(t, "mykey", fields[types.FieldAPIKey])
	assert.Equal(t, "mysecret", fields[types.FieldAPISecret])
	_, hasEndpoint := fields[types.FieldEndpoint]
	assert.False(t, hasEndpoint)
}

// ---------- addImportedItems text mode ----------

// TestAddImportedItems_TextMode tests importing items with text output.
func TestAddImportedItems_TextMode(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	origOutput := flagOutput
	flagOutput = "text"
	defer func() { flagOutput = origOutput }()

	imported := []types.Item{
		{
			Name: "Text Import 1",
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: "tuser",
				types.FieldPassword: "tpass",
			},
		},
	}

	err := addImportedItems(addCmd, mgr, imported)
	require.NoError(t, err)

	all, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

// ---------- recovery text mode tests ----------

// TestRunRecovery_ValidateValidTextMode tests validate with text output.
func TestRunRecovery_ValidateValidTextMode(t *testing.T) {
	origOutput := flagOutput
	origValidate := recoveryValidate
	defer func() {
		flagOutput = origOutput
		recoveryValidate = origValidate
	}()

	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "rv-vault")
	_, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	flagOutput = "text"
	recoveryValidate = result.Mnemonic

	// Text validation output goes to stderr, just verify no error
	err = runRecovery(recoveryCmd, nil)
	require.NoError(t, err)
}

// ---------- printItemsTable edge cases ----------

// TestPrintItemsTable_WithUsername tests JSON table includes username.
func TestPrintItemsTable_WithUsername(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "json"
	defer func() { flagOutput = origOutput }()

	items := []*types.Item{
		{
			ID:     "with-user",
			Name:   "User Item",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldUsername: "myuser"},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printItemsTable(items)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 1)
	assert.Equal(t, "myuser", result[0]["username"])
}

// ---------- Execute error path ----------

// TestExecute_UnknownCommand tests Execute with an unknown command.
func TestExecute_UnknownCommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"zeropass", "nonexistent-command"}

	// Capture stderr
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	err := Execute()

	w.Close()
	os.Stderr = oldStderr

	assert.Error(t, err)
}

// ================================================================
// STDIN-PIPED TESTS — test run* functions via piped stdin
// ================================================================

// pipeStdin redirects os.Stdin to a pipe containing the given text.
// Returns a cleanup function. Use for functions that read stdin ONCE.
func pipeStdin(t *testing.T, input string) func() {
	t.Helper()
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r
	_, err = fmt.Fprintln(w, input)
	require.NoError(t, err)
	w.Close()
	return func() { os.Stdin = origStdin; r.Close() }
}

// pipeStdinLines redirects os.Stdin to a pipe and writes lines one at a time
// with delays, so each bufio.NewScanner sees exactly one line.
// Returns a cleanup function.
func pipeStdinLines(t *testing.T, lines ...string) func() {
	t.Helper()
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r
	go func() {
		for _, line := range lines {
			fmt.Fprintln(w, line)
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(50 * time.Millisecond)
		w.Close()
	}()
	return func() {
		os.Stdin = origStdin
		r.Close()
	}
}

// suppressStderr hides stderr output during test. Returns cleanup.
func suppressStderr(t *testing.T) func() {
	t.Helper()
	old := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w
	return func() {
		w.Close()
		os.Stderr = old
	}
}

// captureStdoutResult captures stdout output into a string.
func captureStdoutResult(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// createVaultWithItems creates a test vault, adds some items, and locks it.
// Returns vaultPath.
func createVaultWithItems(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "pipe-vault")

	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)
	defer idx.Close()

	vk, err := v.VaultKey()
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), func() ([]byte, error) { return vk, nil }, idx)

	// Add a login item
	loginItem := &types.Item{
		Name:     "Test Login",
		Type:     types.ItemTypeLogin,
		Fields:   map[string]string{types.FieldUsername: "user@test.com", types.FieldPassword: "Passw0rd!", types.FieldURL: "https://test.com"},
		Tags:     []string{"env:production", "web"},
		Favorite: true,
	}
	require.NoError(t, mgr.AddItem(loginItem))

	// Add an API key item
	apiItem := &types.Item{
		Name:   "Test API",
		Type:   types.ItemTypeAPIKey,
		Fields: map[string]string{types.FieldAPIKey: "ak_123", types.FieldAPISecret: "secret_456"},
		Tags:   []string{"env:production"},
	}
	require.NoError(t, mgr.AddItem(apiItem))

	// Add a note item
	noteItem := &types.Item{
		Name:  "Test Note",
		Type:  types.ItemTypeSecureNote,
		Notes: "This is a secret note.",
	}
	require.NoError(t, mgr.AddItem(noteItem))

	v.Lock()
	return vaultPath
}

// ---------- promptPassword ----------

func TestPromptPassword_NonTerminal(t *testing.T) {
	cleanup := pipeStdin(t, "MySecret123")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	pw, err := promptPassword("Password: ")
	require.NoError(t, err)
	assert.Equal(t, "MySecret123", pw)
}

func TestPromptPassword_EmptyInput(t *testing.T) {
	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close() // Empty input
	defer func() { os.Stdin = origStdin }()

	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	_, err := promptPassword("Password: ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input available")
}

// ---------- promptLine ----------

func TestPromptLine_NonTerminal(t *testing.T) {
	cleanup := pipeStdin(t, "hello world")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	result := promptLine("Enter: ")
	assert.Equal(t, "hello world", result)
}

func TestPromptLine_EmptyPipe(t *testing.T) {
	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close()
	defer func() { os.Stdin = origStdin }()

	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	result := promptLine("Enter: ")
	assert.Equal(t, "", result)
}

// ---------- promptConfirmPassword ----------

func TestPromptConfirmPassword_Match(t *testing.T) {
	cleanup := pipeStdinLines(t, "MyPa$$w0rd!", "MyPa$$w0rd!")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	pw, err := promptConfirmPassword()
	require.NoError(t, err)
	assert.Equal(t, "MyPa$$w0rd!", pw)
}

func TestPromptConfirmPassword_Mismatch(t *testing.T) {
	cleanup := pipeStdinLines(t, "password1", "password2")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	_, err := promptConfirmPassword()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "passwords do not match")
}

// ---------- promptYesNo ----------

func TestPromptYesNo_Yes(t *testing.T) {
	cleanup := pipeStdin(t, "y")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	assert.True(t, promptYesNo("Continue?"))
}

func TestPromptYesNo_YesFull(t *testing.T) {
	cleanup := pipeStdin(t, "yes")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	assert.True(t, promptYesNo("Continue?"))
}

func TestPromptYesNo_No(t *testing.T) {
	cleanup := pipeStdin(t, "n")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	assert.False(t, promptYesNo("Continue?"))
}

// ---------- openAndUnlockVault ----------

func TestOpenAndUnlockVault_Success(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	defer func() { flagVaultPath = origPath }()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	v, err := openAndUnlockVault()
	require.NoError(t, err)
	assert.NotNil(t, v)
	v.Lock()
}

func TestOpenAndUnlockVault_WrongPassword(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	defer func() { flagVaultPath = origPath }()

	cleanup := pipeStdin(t, "wrong-password")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	_, err := openAndUnlockVault()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unlock vault")
}

func TestOpenAndUnlockVault_NonexistentVault(t *testing.T) {
	origPath := flagVaultPath
	flagVaultPath = filepath.Join(t.TempDir(), "nonexistent-vault")
	defer func() { flagVaultPath = origPath }()

	_, err := openAndUnlockVault()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open vault")
}

// ---------- openVault ----------

func TestOpenVault_Success(t *testing.T) {
	vaultPath := createVaultWithItems(t)
	v, err := openVault(vaultPath)
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestOpenVault_Nonexistent(t *testing.T) {
	_, err := openVault(filepath.Join(t.TempDir(), "no-vault"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open vault")
}

// ---------- runLock ----------

func TestRunLock_JSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	output := captureStdoutResult(t, func() {
		err := runLock(lockCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "locked", result["status"])
}

func TestRunLock_Text(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	err := runLock(lockCmd, nil)
	require.NoError(t, err)
}

func TestRunLock_NonexistentVault(t *testing.T) {
	origPath := flagVaultPath
	flagVaultPath = filepath.Join(t.TempDir(), "no-vault")
	defer func() { flagVaultPath = origPath }()

	err := runLock(lockCmd, nil)
	assert.Error(t, err)
}

// ---------- runUnlock ----------

func TestRunUnlock_WithPassword(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origRecovery := unlockRecovery
	flagVaultPath = vaultPath
	flagOutput = "json"
	unlockRecovery = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		unlockRecovery = origRecovery
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runUnlock(unlockCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "unlocked", result["status"])
}

func TestRunUnlock_WrongPassword(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origRecovery := unlockRecovery
	flagVaultPath = vaultPath
	unlockRecovery = false
	defer func() {
		flagVaultPath = origPath
		unlockRecovery = origRecovery
	}()

	cleanup := pipeStdin(t, "wrong")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runUnlock(unlockCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect master password")
}

func TestRunUnlock_WithRecovery(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "rec-vault")
	_, result, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	origPath := flagVaultPath
	origOutput := flagOutput
	origRecovery := unlockRecovery
	flagVaultPath = vaultPath
	flagOutput = "json"
	unlockRecovery = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		unlockRecovery = origRecovery
	}()

	cleanup := pipeStdin(t, result.Mnemonic)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runUnlock(unlockCmd, nil)
		require.NoError(t, err)
	})

	var res map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &res))
	assert.Equal(t, "unlocked", res["status"])
}

func TestRunUnlock_NonexistentVault(t *testing.T) {
	origPath := flagVaultPath
	flagVaultPath = filepath.Join(t.TempDir(), "no-vault")
	defer func() { flagVaultPath = origPath }()

	err := runUnlock(unlockCmd, nil)
	assert.Error(t, err)
}

// ---------- runList ----------

func TestRunList_JSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListType := listType
	origListTag := listTag
	origListFavorite := listFavorite
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listType = ""
	listTag = ""
	listFavorite = false
	listSort = "name"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listType = origListType
		listTag = origListTag
		listFavorite = origListFavorite
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 3) // 3 items in our test vault
}

func TestRunList_FilterByType(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListType := listType
	origListTag := listTag
	origListFavorite := listFavorite
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listType = "login"
	listTag = ""
	listFavorite = false
	listSort = "name"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listType = origListType
		listTag = origListTag
		listFavorite = origListFavorite
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 1)
}

func TestRunList_FilterByTag(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListType := listType
	origListTag := listTag
	origListFavorite := listFavorite
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listType = ""
	listTag = "env:production"
	listFavorite = false
	listSort = "name"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listType = origListType
		listTag = origListTag
		listFavorite = origListFavorite
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 2) // login and API key both have env:production
}

func TestRunList_FilterFavorite(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListType := listType
	origListTag := listTag
	origListFavorite := listFavorite
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listType = ""
	listTag = ""
	listFavorite = true
	listSort = "name"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listType = origListType
		listTag = origListTag
		listFavorite = origListFavorite
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 1) // Only login is favorite
}

func TestRunList_SortByDate(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listSort = "date"
	listType = ""
	listTag = ""
	listFavorite = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 3)
}

func TestRunList_SortByType(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = true
	listSort = "type"
	listType = ""
	listTag = ""
	listFavorite = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 3)
}

func TestRunList_TextMode(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origListJSON := listJSON
	origListSort := listSort
	flagVaultPath = vaultPath
	flagOutput = "text"
	listJSON = false
	listSort = "name"
	listType = ""
	listTag = ""
	listFavorite = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		listJSON = origListJSON
		listSort = origListSort
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runList(listCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Test Login")
	assert.Contains(t, output, "Test API")
	assert.Contains(t, output, "Test Note")
}

// ---------- runGet ----------

func TestRunGet_JSONOutput(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origGetJSON := getJSON
	origGetCopy := getCopy
	origGetField := getField
	flagVaultPath = vaultPath
	flagOutput = "text"
	getJSON = true
	getCopy = false
	getField = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		getJSON = origGetJSON
		getCopy = origGetCopy
		getField = origGetField
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runGet(getCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "Test Login", result["name"])
	assert.Equal(t, "login", result["type"])
}

func TestRunGet_SpecificField(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origGetJSON := getJSON
	origGetCopy := getCopy
	origGetField := getField
	flagVaultPath = vaultPath
	flagOutput = "text"
	getJSON = false
	getCopy = false
	getField = "username"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		getJSON = origGetJSON
		getCopy = origGetCopy
		getField = origGetField
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runGet(getCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	assert.Contains(t, strings.TrimSpace(output), "user@test.com")
}

func TestRunGet_SpecificFieldJSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origGetJSON := getJSON
	origGetCopy := getCopy
	origGetField := getField
	flagVaultPath = vaultPath
	flagOutput = "json"
	getJSON = false
	getCopy = false
	getField = "password"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		getJSON = origGetJSON
		getCopy = origGetCopy
		getField = origGetField
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runGet(getCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "Passw0rd!", result["value"])
}

func TestRunGet_FieldNotFound(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origGetField := getField
	origGetCopy := getCopy
	flagVaultPath = vaultPath
	getField = "nonexistent"
	getCopy = false
	defer func() {
		flagVaultPath = origPath
		getField = origGetField
		getCopy = origGetCopy
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runGet(getCmd, []string{"Test Login"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field \"nonexistent\" not found")
}

func TestRunGet_ItemNotFound(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origGetField := getField
	origGetCopy := getCopy
	flagVaultPath = vaultPath
	getField = ""
	getCopy = false
	defer func() {
		flagVaultPath = origPath
		getField = origGetField
		getCopy = origGetCopy
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runGet(getCmd, []string{"Nonexistent Item"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no items matching")
}

func TestRunGet_TextOutput(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origGetJSON := getJSON
	origGetCopy := getCopy
	origGetField := getField
	flagVaultPath = vaultPath
	flagOutput = "text"
	getJSON = false
	getCopy = false
	getField = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		getJSON = origGetJSON
		getCopy = origGetCopy
		getField = origGetField
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runGet(getCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Test Login")
}

// ---------- runSearch ----------

func TestRunSearch_FindsItems(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origSearchType := searchType
	origSearchTag := searchTag
	flagVaultPath = vaultPath
	flagOutput = "json"
	searchType = ""
	searchTag = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		searchType = origSearchType
		searchTag = origSearchTag
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runSearch(searchCmd, []string{"Test"})
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestRunSearch_WithTypeFilter(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origSearchType := searchType
	origSearchTag := searchTag
	flagVaultPath = vaultPath
	flagOutput = "json"
	searchType = "apikey"
	searchTag = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		searchType = origSearchType
		searchTag = origSearchTag
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runSearch(searchCmd, []string{"Test"})
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	for _, item := range result {
		assert.Equal(t, "apikey", item["type"])
	}
}

func TestRunSearch_WithTagFilter(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origSearchType := searchType
	origSearchTag := searchTag
	flagVaultPath = vaultPath
	flagOutput = "json"
	searchType = ""
	searchTag = "web"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		searchType = origSearchType
		searchTag = origSearchTag
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runSearch(searchCmd, []string{"Test"})
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	for _, item := range result {
		tags, ok := item["tags"].([]interface{})
		if ok {
			found := false
			for _, t := range tags {
				if t == "web" {
					found = true
				}
			}
			assert.True(t, found)
		}
	}
}

// ---------- runHealth ----------

func TestRunHealth_JSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runHealth(healthCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result, "overall_score")
	assert.Contains(t, result, "total_items")
}

func TestRunHealth_Text(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runHealth(healthCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "PASSWORD HEALTH REPORT")
	assert.Contains(t, output, "Overall Score")
}

// ---------- runExport ----------

func TestRunExport_JSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origFormat := exportFormat
	origFile := exportOutputFile
	flagVaultPath = vaultPath
	flagOutput = "text"
	exportFormat = "json"
	exportOutputFile = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		exportFormat = origFormat
		exportOutputFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runExport(exportCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Len(t, result, 3)
}

func TestRunExport_CSV(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origFormat := exportFormat
	origFile := exportOutputFile
	flagVaultPath = vaultPath
	flagOutput = "text"
	exportFormat = "csv"
	exportOutputFile = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		exportFormat = origFormat
		exportOutputFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runExport(exportCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Test Login")
}

func TestRunExport_ToFile(t *testing.T) {
	vaultPath := createVaultWithItems(t)
	outFile := filepath.Join(t.TempDir(), "export.json")

	origPath := flagVaultPath
	origOutput := flagOutput
	origFormat := exportFormat
	origFile := exportOutputFile
	flagVaultPath = vaultPath
	flagOutput = "text"
	exportFormat = "json"
	exportOutputFile = outFile
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		exportFormat = origFormat
		exportOutputFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runExport(exportCmd, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(outFile)
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))
	assert.Len(t, result, 3)
}

func TestRunExport_Encrypted(t *testing.T) {
	vaultPath := createVaultWithItems(t)
	outFile := filepath.Join(t.TempDir(), "export.enc")

	origPath := flagVaultPath
	origOutput := flagOutput
	origFormat := exportFormat
	origFile := exportOutputFile
	flagVaultPath = vaultPath
	flagOutput = "text"
	exportFormat = "encrypted"
	exportOutputFile = outFile
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		exportFormat = origFormat
		exportOutputFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runExport(exportCmd, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)
}

func TestRunExport_UnknownFormat(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origFormat := exportFormat
	origFile := exportOutputFile
	flagVaultPath = vaultPath
	exportFormat = "xml"
	exportOutputFile = ""
	defer func() {
		flagVaultPath = origPath
		exportFormat = origFormat
		exportOutputFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runExport(exportCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown export format")
}

// ---------- runEdit ----------

func TestRunEdit_UpdateName(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origEditName := editName
	origEditUsername := editUsername
	origEditPassword := editPassword
	origEditURL := editURL
	origEditTags := editTags
	origEditFavorite := editFavorite
	origEditNotes := editNotes
	flagVaultPath = vaultPath
	flagOutput = "json"
	editName = "Updated Login"
	editUsername = ""
	editPassword = ""
	editURL = ""
	editTags = ""
	editFavorite = ""
	editNotes = ""
	// Clear all edit flags
	editAPIKey = ""
	editAPISecret = ""
	editCardNumber = ""
	editCardHolder = ""
	editCardExpiry = ""
	editCardCVV = ""
	editFirstName = ""
	editLastName = ""
	editEmail = ""
	editPhone = ""
	editAddress = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		editName = origEditName
		editUsername = origEditUsername
		editPassword = origEditPassword
		editURL = origEditURL
		editTags = origEditTags
		editFavorite = origEditFavorite
		editNotes = origEditNotes
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEdit(editCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "Updated Login", result["name"])
}

func TestRunEdit_UpdateMultipleFields(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	editName = ""
	editUsername = "newuser"
	editPassword = "NewP@ss123!"
	editURL = "https://updated.com"
	editTags = "tag1,tag2"
	editFavorite = "false"
	editNotes = "updated notes"
	editAPIKey = ""
	editAPISecret = ""
	editCardNumber = ""
	editCardHolder = ""
	editCardExpiry = ""
	editCardCVV = ""
	editFirstName = ""
	editLastName = ""
	editEmail = ""
	editPhone = ""
	editAddress = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		editName = ""
		editUsername = ""
		editPassword = ""
		editURL = ""
		editTags = ""
		editFavorite = ""
		editNotes = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEdit(editCmd, []string{"Test Login"})
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	fields := result["fields"].(map[string]interface{})
	assert.Equal(t, "newuser", fields["username"])
}

func TestRunEdit_NoFieldsSpecified(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	editName = ""
	editUsername = ""
	editPassword = ""
	editURL = ""
	editTags = ""
	editFavorite = ""
	editNotes = ""
	editAPIKey = ""
	editAPISecret = ""
	editCardNumber = ""
	editCardHolder = ""
	editCardExpiry = ""
	editCardCVV = ""
	editFirstName = ""
	editLastName = ""
	editEmail = ""
	editPhone = ""
	editAddress = ""
	defer func() { flagVaultPath = origPath }()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runEdit(editCmd, []string{"Test Login"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields specified")
}

func TestRunEdit_ItemNotFound(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	editName = "something"
	defer func() {
		flagVaultPath = origPath
		editName = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runEdit(editCmd, []string{"Nonexistent"})
	assert.Error(t, err)
}

// ---------- runDelete ----------

func TestRunDelete_Force(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origForce := deleteForce
	flagVaultPath = vaultPath
	flagOutput = "json"
	deleteForce = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		deleteForce = origForce
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runDelete(deleteCmd, []string{"Test Note"})
		require.NoError(t, err)
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "Test Note", result["name"])
}

func TestRunDelete_TextMode(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origForce := deleteForce
	flagVaultPath = vaultPath
	flagOutput = "text"
	deleteForce = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		deleteForce = origForce
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runDelete(deleteCmd, []string{"Test Note"})
	require.NoError(t, err)
}

func TestRunDelete_ItemNotFound(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origForce := deleteForce
	flagVaultPath = vaultPath
	deleteForce = true
	defer func() {
		flagVaultPath = origPath
		deleteForce = origForce
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runDelete(deleteCmd, []string{"Nonexistent"})
	assert.Error(t, err)
}

// ---------- runAdd ----------

func TestRunAdd_LoginType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "login"
	addName = "My Login"
	addUsername = "user1"
	addPassword = "Pass123!"
	addURL = "https://example.com"
	addTags = "work"
	addFavorite = false
	addGeneratePassword = false
	addNotes = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addUsername = ""
		addPassword = ""
		addURL = ""
		addTags = ""
		addFavorite = false
		addGeneratePassword = false
		addNotes = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "My Login", result["name"])
	assert.Equal(t, "login", result["type"])
}

func TestRunAdd_IdentityType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "identity"
	addName = "My Identity"
	addFirstName = "John"
	addLastName = "Doe"
	addEmail = "john@example.com"
	addPhone = "555-1234"
	addAddress = "123 Main St"
	addTags = ""
	addFavorite = true
	addNotes = ""
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addFirstName = ""
		addLastName = ""
		addEmail = ""
		addPhone = ""
		addAddress = ""
		addFavorite = false
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "identity", result["type"])
}

func TestRunAdd_NoteType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "note"
	addName = "My Note"
	addNotes = "Secret content here"
	addTags = ""
	addFavorite = false
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addNotes = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "note", result["type"])
}

func TestRunAdd_CreditCardType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "creditcard"
	addName = "My Card"
	addCardNumber = "4111111111111111"
	addCardHolder = "John Doe"
	addCardExpiry = "12/25"
	addCardCVV = "123"
	addTags = ""
	addFavorite = false
	addNotes = ""
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addCardNumber = ""
		addCardHolder = ""
		addCardExpiry = ""
		addCardCVV = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "creditcard", result["type"])
}

func TestRunAdd_APIKeyType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "apikey"
	addName = "My API Key"
	addAPIKey = "key_abc"
	addAPISecret = "secret_xyz"
	addURL = "https://api.example.com"
	addTags = ""
	addFavorite = false
	addNotes = ""
	addUsername = ""
	addPassword = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addAPIKey = ""
		addAPISecret = ""
		addURL = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "apikey", result["type"])
}

func TestRunAdd_SSHKeyType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	// Create temp SSH key files
	privKeyFile := filepath.Join(dir, "id_rsa")
	pubKeyFile := filepath.Join(dir, "id_rsa.pub")
	os.WriteFile(privKeyFile, []byte("-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"), 0600)
	os.WriteFile(pubKeyFile, []byte("ssh-rsa AAAA... test@host"), 0644)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "sshkey"
	addName = "My SSH Key"
	addPrivateKeyFile = privKeyFile
	addPublicKeyFile = pubKeyFile
	addTags = ""
	addFavorite = false
	addNotes = ""
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addPrivateKeyFile = ""
		addPublicKeyFile = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "sshkey", result["type"])
}

func TestRunAdd_CustomType(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "custom"
	addName = "Custom Item"
	addNotes = "custom notes"
	addTags = "misc"
	addFavorite = false
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addNotes = ""
		addTags = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "custom", result["type"])
}

func TestRunAdd_InvalidType(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	addType = "invalid"
	addName = "Blah"
	defer func() {
		flagVaultPath = origPath
		addType = "login"
		addName = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runAdd(addCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid item type")
}

func TestRunAdd_GeneratePassword(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "login"
	addName = "Auto Gen Login"
	addUsername = "autouser"
	addGeneratePassword = true
	addPassword = ""
	addURL = ""
	addTags = ""
	addFavorite = false
	addNotes = ""
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addUsername = ""
		addGeneratePassword = false
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	fields := result["fields"].(map[string]interface{})
	assert.NotEmpty(t, fields["password"])
}

func TestRunAdd_TextMode(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "add-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	addType = "note"
	addName = "Text Note"
	addNotes = "note content"
	addTags = ""
	addFavorite = false
	addUsername = ""
	addPassword = ""
	addURL = ""
	addGeneratePassword = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		addType = "login"
		addName = ""
		addNotes = ""
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Text Note")
}

// ---------- runEnvList ----------

func TestRunEnvList_JSON(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEnvList(envListCmd, nil)
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.GreaterOrEqual(t, len(result), 1) // At least "production" env
}

func TestRunEnvList_Text(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEnvList(envListCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "production")
}

func TestRunEnvList_NoEnvironments(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "empty-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEnvList(envListCmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "No environments found")
}

// ---------- runEnvUse ----------

func TestRunEnvUse_WithItems(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEnvUse(envUseCmd, []string{"production"})
		require.NoError(t, err)
	})

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestRunEnvUse_NoItems(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runEnvUse(envUseCmd, []string{"nonexistent-env"})
	require.NoError(t, err)
}

func TestRunEnvUse_TextMode(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runEnvUse(envUseCmd, []string{"production"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Test Login")
}

// ---------- runInit ----------

func TestRunInit_JSON(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "new-vault")

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	// runInit needs: 2x promptPassword (confirm) + promptYesNo
	cleanup := pipeStdinLines(t, strongInitPassword, strongInitPassword, "y")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runInit(initCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "created", result["status"])
}

func TestRunInit_TextMode(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "new-vault")

	origPath := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "text"
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	cleanup := pipeStdinLines(t, strongInitPassword, strongInitPassword, "y")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runInit(initCmd, nil)
	require.NoError(t, err)
}

func TestRunInit_AlreadyExists(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	defer func() { flagVaultPath = origPath }()

	err := runInit(initCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault already exists")
}

func TestRunInit_WeakPassword(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "new-vault")

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	defer func() { flagVaultPath = origPath }()

	// Short password
	cleanup := pipeStdinLines(t, "short", "short")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runInit(initCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least")
}

func TestRunInit_PasswordMismatch(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "new-vault")

	origPath := flagVaultPath
	flagVaultPath = vaultPath
	defer func() { flagVaultPath = origPath }()

	cleanup := pipeStdinLines(t, testMasterPassword, "DifferentPassword123!")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runInit(initCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "passwords do not match")
}

// ---------- runRun (env runner) ----------

func TestRunRun_NoCommand(t *testing.T) {
	err := runRun(runCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no command specified")
}

func TestRunRun_NoEnvFile(t *testing.T) {
	origEnvFile := runEnvFile
	runEnvFile = filepath.Join(t.TempDir(), "nonexistent.env")
	defer func() { runEnvFile = origEnvFile }()

	err := runRun(runCmd, []string{"echo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse env file")
}

func TestRunRun_NoRefsCommandNotFound(t *testing.T) {
	// .env file without zp:// refs: vault opening is skipped
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("FOO=bar\nBAZ=qux\n"), 0o644)

	origEnvFile := runEnvFile
	runEnvFile = envPath
	defer func() { runEnvFile = origEnvFile }()

	err := runRun(runCmd, []string{"__nonexistent_command_abc123__"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not found")
}

func TestRunRun_WithZPRefsNeedVault(t *testing.T) {
	// .env file WITH zp:// refs: needs vault but none available
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("SECRET=zp://myitem/password\n"), 0o644)

	origEnvFile := runEnvFile
	origVault := flagVaultPath
	runEnvFile = envPath
	flagVaultPath = filepath.Join(dir, "nonexistent.vault")
	defer func() {
		runEnvFile = origEnvFile
		flagVaultPath = origVault
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()

	err := runRun(runCmd, []string{"echo"})
	assert.Error(t, err) // fails to open vault
}

func TestRunRun_WithZPRefsVaultResolveFails(t *testing.T) {
	// Create vault, provide env file with zp:// ref to nonexistent item
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("DB_PASS=zp://nonexistent-item/password\n"), 0o644)

	origVault := flagVaultPath
	origEnvFile := runEnvFile
	flagVaultPath = vaultPath
	runEnvFile = envPath
	defer func() {
		flagVaultPath = origVault
		runEnvFile = origEnvFile
	}()

	cleanup := pipeStdinLines(t, strongInitPassword, strongInitPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	// Init vault
	origOutput := flagOutput
	flagOutput = "json"
	defer func() { flagOutput = origOutput }()
	err := runInit(initCmd, nil)
	require.NoError(t, err)

	// Now try runRun — needs to unlock vault, then resolve zp:// ref
	cleanup2 := pipeStdin(t, strongInitPassword)
	defer cleanup2()

	err = runRun(runCmd, []string{"echo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolve")
}

// ---------- runRecovery via piped stdin ----------

func TestRunRecovery_ShowRecovery_NoVault(t *testing.T) {
	dir := t.TempDir()
	origVault := flagVaultPath
	flagVaultPath = filepath.Join(dir, "nonexistent.vault")
	defer func() { flagVaultPath = origVault }()

	origValidate := recoveryValidate
	recoveryValidate = ""
	defer func() { recoveryValidate = origValidate }()

	err := runRecovery(recoveryCmd, nil)
	assert.Error(t, err) // vault doesn't exist
}

func TestRunRecovery_ShowRecovery_EmptyMnemonic(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	origVault := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origVault
		flagOutput = origOutput
	}()

	// Init vault
	cleanup := pipeStdinLines(t, strongInitPassword, strongInitPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runInit(initCmd, nil)
	require.NoError(t, err)

	origValidate := recoveryValidate
	recoveryValidate = ""
	defer func() { recoveryValidate = origValidate }()

	// Pipe empty line for "Enter recovery mnemonic to test: "
	cleanup2 := pipeStdin(t, "")
	defer cleanup2()

	err = runRecovery(recoveryCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no mnemonic provided")
}

func TestRunRecovery_ShowRecovery_InvalidMnemonic(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.vault")

	origVault := flagVaultPath
	origOutput := flagOutput
	flagVaultPath = vaultPath
	flagOutput = "json"
	defer func() {
		flagVaultPath = origVault
		flagOutput = origOutput
	}()

	// Init vault
	cleanup := pipeStdinLines(t, strongInitPassword, strongInitPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runInit(initCmd, nil)
	require.NoError(t, err)

	origValidate := recoveryValidate
	recoveryValidate = ""
	defer func() { recoveryValidate = origValidate }()

	// Pipe invalid mnemonic
	cleanup2 := pipeStdin(t, "not a valid mnemonic phrase")
	defer cleanup2()

	err = runRecovery(recoveryCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mnemonic")
}

// ---------- buildLoginFields additional tests ----------

func TestBuildLoginFields_PromptedPassword(t *testing.T) {
	origUser := addUsername
	origPw := addPassword
	origGen := addGeneratePassword
	origURL := addURL
	addUsername = ""
	addPassword = ""
	addGeneratePassword = false
	addURL = "https://example.com"
	defer func() {
		addUsername = origUser
		addPassword = origPw
		addGeneratePassword = origGen
		addURL = origURL
	}()

	// Pipe username and password
	cleanup := pipeStdinLines(t, "prompteduser", "promptedpass")
	defer cleanup()

	fields := make(map[string]string)
	err := buildLoginFields(addCmd, fields)
	require.NoError(t, err)
	assert.Equal(t, "prompteduser", fields["username"])
	assert.Equal(t, "promptedpass", fields["password"])
	assert.Equal(t, "https://example.com", fields["url"])
}

func TestBuildLoginFields_GeneratePasswordFlag(t *testing.T) {
	origUser := addUsername
	origPw := addPassword
	origGen := addGeneratePassword
	addUsername = "testuser"
	addPassword = ""
	addGeneratePassword = true
	defer func() {
		addUsername = origUser
		addPassword = origPw
		addGeneratePassword = origGen
	}()

	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	fields := make(map[string]string)
	err := buildLoginFields(addCmd, fields)
	require.NoError(t, err)
	assert.Equal(t, "testuser", fields["username"])
	assert.NotEmpty(t, fields["password"])
	assert.Len(t, fields["password"], 32) // 32-char generated password
}

// ---------- additional buildAPIKeyFields test ----------

func TestBuildAPIKeyFields_WithURLEndpoint(t *testing.T) {
	origKey := addAPIKey
	origSecret := addAPISecret
	origURL := addURL
	addAPIKey = "my-api-key"
	addAPISecret = "my-secret"
	addURL = "https://api.example.com"
	defer func() {
		addAPIKey = origKey
		addAPISecret = origSecret
		addURL = origURL
	}()

	fields := make(map[string]string)
	buildAPIKeyFields(fields)
	assert.Equal(t, "my-api-key", fields["api_key"])
	assert.Equal(t, "my-secret", fields["api_secret"])
	assert.Equal(t, "https://api.example.com", fields["endpoint"])
}

// ---------- additional runExport tests ----------

func TestRunExport_ToStdout(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origVault := flagVaultPath
	origFormat := exportFormat
	origFile := exportOutputFile
	origOutput := flagOutput
	flagVaultPath = vaultPath
	exportFormat = "json"
	exportOutputFile = "" // stdout
	flagOutput = "json"
	defer func() {
		flagVaultPath = origVault
		exportFormat = origFormat
		exportOutputFile = origFile
		flagOutput = origOutput
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	out := captureStdoutResult(t, func() {
		err := runExport(exportCmd, nil)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "Test Login")
}

// ---------- additional getVaultPath test ----------

func TestGetVaultPath_FlagSet(t *testing.T) {
	origVault := flagVaultPath
	flagVaultPath = "/custom/test/path.vault"
	defer func() { flagVaultPath = origVault }()

	path := getVaultPath()
	assert.Equal(t, "/custom/test/path.vault", path)
}

func TestGetVaultPath_DefaultPath(t *testing.T) {
	origVault := flagVaultPath
	flagVaultPath = ""
	defer func() { flagVaultPath = origVault }()

	path := getVaultPath()
	assert.Contains(t, path, "vaults")
	assert.Contains(t, path, "default")
}

// ---------- additional runGenerate test ----------

func TestRunGenerate_PassphraseMode(t *testing.T) {
	origGen := genPassphrase
	origWords := genWords
	origSep := genSeparator
	origCopy := genCopy
	origOutput := flagOutput
	genPassphrase = true
	genWords = 5
	genSeparator = "."
	genCopy = false
	flagOutput = "json"
	defer func() {
		genPassphrase = origGen
		genWords = origWords
		genSeparator = origSep
		genCopy = origCopy
		flagOutput = origOutput
	}()

	out := captureStdoutResult(t, func() {
		err := runGenerate(generateCmd, nil)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "password")
	// Passphrase uses dots as separator
	assert.Contains(t, out, ".")
}

func TestRunGenerate_RandomNoSymbolsNoDigits(t *testing.T) {
	origLen := genLength
	origNoSym := genNoSymbols
	origNoDig := genNoDigits
	origNoUp := genNoUppercase
	origCopy := genCopy
	origOutput := flagOutput
	genLength = 20
	genNoSymbols = true
	genNoDigits = true
	genNoUppercase = false
	genCopy = false
	flagOutput = "json"
	defer func() {
		genLength = origLen
		genNoSymbols = origNoSym
		genNoDigits = origNoDig
		genNoUppercase = origNoUp
		genCopy = origCopy
		flagOutput = origOutput
	}()

	out := captureStdoutResult(t, func() {
		err := runGenerate(generateCmd, nil)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "password")
}

// ---------- findItemByQuery multi-match selection ----------

func TestFindItemByQuery_MultiMatchSelect(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add two items with similar names
	item1 := &types.Item{
		Name:   "Duplicate",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user1", types.FieldPassword: "pass1"},
	}
	require.NoError(t, mgr.AddItem(item1))

	item2 := &types.Item{
		Name:   "Duplicate",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user2", types.FieldPassword: "pass2"},
	}
	require.NoError(t, mgr.AddItem(item2))

	// Pipe "1" to select first item
	cleanup2 := pipeStdin(t, "1")
	defer cleanup2()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	found, err := findItemByQuery(mgr, idx, "Duplicate")
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Duplicate", found.Name)
}

func TestFindItemByQuery_MultiMatchInvalidSelection(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add two items with similar names
	item1 := &types.Item{
		Name:   "Same Name",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user1", types.FieldPassword: "pass1"},
	}
	require.NoError(t, mgr.AddItem(item1))

	item2 := &types.Item{
		Name:   "Same Name",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "user2", types.FieldPassword: "pass2"},
	}
	require.NoError(t, mgr.AddItem(item2))

	// Pipe "99" (invalid selection)
	cleanup2 := pipeStdin(t, "99")
	defer cleanup2()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	_, err := findItemByQuery(mgr, idx, "Same Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid selection")
}

func TestFindItemByQuery_MultiMatchNonNumeric(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	item1 := &types.Item{
		Name:   "Dup Item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "u1", types.FieldPassword: "p1"},
	}
	require.NoError(t, mgr.AddItem(item1))

	item2 := &types.Item{
		Name:   "Dup Item",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{types.FieldUsername: "u2", types.FieldPassword: "p2"},
	}
	require.NoError(t, mgr.AddItem(item2))

	// Pipe non-numeric input
	cleanup2 := pipeStdin(t, "abc")
	defer cleanup2()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	_, err := findItemByQuery(mgr, idx, "Dup Item")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid selection")
}

// ---------- E2E piped: Add -> Get -> Edit -> Delete ----------

func TestPipedE2E_AddGetEditDelete(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "e2e-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	origPath := flagVaultPath
	origOutput := flagOutput
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
	}()

	// Step 1: Add
	flagVaultPath = vaultPath
	flagOutput = "json"
	addType = "login"
	addName = "E2E Item"
	addUsername = "e2e_user"
	addPassword = "E2eP@ss!"
	addURL = "https://e2e.test"
	addTags = "e2e"
	addFavorite = false
	addGeneratePassword = false
	addNotes = ""

	cleanStdin := pipeStdin(t, testMasterPassword)
	cleanStderr := suppressStderr(t)

	output := captureStdoutResult(t, func() {
		err := runAdd(addCmd, nil)
		require.NoError(t, err)
	})

	cleanStdin()
	cleanStderr()

	var addResult map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &addResult))
	itemID := addResult["id"].(string)

	// Reset add flags
	addType = "login"
	addName = ""
	addUsername = ""
	addPassword = ""
	addURL = ""
	addTags = ""

	// Step 2: Get by ID
	getJSON = true
	getCopy = false
	getField = ""

	cleanStdin = pipeStdin(t, testMasterPassword)
	cleanStderr = suppressStderr(t)

	output = captureStdoutResult(t, func() {
		err := runGet(getCmd, []string{itemID})
		require.NoError(t, err)
	})

	cleanStdin()
	cleanStderr()
	getJSON = false

	var getResult map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &getResult))
	assert.Equal(t, "E2E Item", getResult["name"])

	// Step 3: Edit
	editName = "E2E Updated"
	editUsername = ""
	editPassword = ""
	editURL = ""
	editTags = ""
	editFavorite = ""
	editNotes = ""
	editAPIKey = ""
	editAPISecret = ""
	editCardNumber = ""
	editCardHolder = ""
	editCardExpiry = ""
	editCardCVV = ""
	editFirstName = ""
	editLastName = ""
	editEmail = ""
	editPhone = ""
	editAddress = ""

	cleanStdin = pipeStdin(t, testMasterPassword)
	cleanStderr = suppressStderr(t)

	output = captureStdoutResult(t, func() {
		err := runEdit(editCmd, []string{itemID})
		require.NoError(t, err)
	})

	cleanStdin()
	cleanStderr()
	editName = ""

	var editResult map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &editResult))
	assert.Equal(t, "E2E Updated", editResult["name"])

	// Step 4: Delete
	deleteForce = true

	cleanStdin = pipeStdin(t, testMasterPassword)
	cleanStderr = suppressStderr(t)

	output = captureStdoutResult(t, func() {
		err := runDelete(deleteCmd, []string{itemID})
		require.NoError(t, err)
	})

	cleanStdin()
	cleanStderr()
	deleteForce = false

	var delResult map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &delResult))
	assert.Equal(t, "deleted", delResult["status"])

	// Step 5: Verify deletion
	cleanStdin = pipeStdin(t, testMasterPassword)
	cleanStderr = suppressStderr(t)

	err = runGet(getCmd, []string{itemID})

	cleanStdin()
	cleanStderr()

	assert.Error(t, err) // Item should be gone
}

// ---------- runImport via piped stdin ----------

func TestRunImport_UnknownSource(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origFrom := importFrom
	origFile := importFile
	flagVaultPath = vaultPath
	importFrom = "unknown"
	importFile = filepath.Join(t.TempDir(), "dummy.csv")
	os.WriteFile(importFile, []byte("name,url,username,password\n"), 0644)
	defer func() {
		flagVaultPath = origPath
		importFrom = origFrom
		importFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runImport(importCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown import source")
}

func TestRunImport_FileNotFound(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origFrom := importFrom
	origFile := importFile
	flagVaultPath = vaultPath
	importFrom = "csv"
	importFile = filepath.Join(t.TempDir(), "nonexistent.csv")
	defer func() {
		flagVaultPath = origPath
		importFrom = origFrom
		importFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runImport(importCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open import file")
}

func TestRunImport_CSVSuccess(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "import-vault")
	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)
	v.Lock()

	// Create a CSV file
	csvContent := "name,url,username,password,notes\nCSV Item,https://csv.com,csvuser,csvpass,csv notes\n"
	csvFile := filepath.Join(dir, "import.csv")
	os.WriteFile(csvFile, []byte(csvContent), 0644)

	origPath := flagVaultPath
	origOutput := flagOutput
	origFrom := importFrom
	origFile := importFile
	flagVaultPath = vaultPath
	flagOutput = "json"
	importFrom = "csv"
	importFile = csvFile
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		importFrom = origFrom
		importFile = origFile
	}()

	cleanup := pipeStdin(t, testMasterPassword)
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runImport(importCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, float64(1), result["imported"])
}

// ---------- runDelete with confirmation prompt ----------

func TestRunDelete_WithConfirmation(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origForce := deleteForce
	flagVaultPath = vaultPath
	flagOutput = "json"
	deleteForce = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		deleteForce = origForce
	}()

	// Need: 1x promptPassword (openAndUnlockVault) + 1x promptLine (promptYesNo)
	cleanup := pipeStdinLines(t, testMasterPassword, "y")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	output := captureStdoutResult(t, func() {
		err := runDelete(deleteCmd, []string{"Test Note"})
		require.NoError(t, err)
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "deleted", result["status"])
}

func TestRunDelete_CancelConfirmation(t *testing.T) {
	vaultPath := createVaultWithItems(t)

	origPath := flagVaultPath
	origOutput := flagOutput
	origForce := deleteForce
	flagVaultPath = vaultPath
	flagOutput = "text"
	deleteForce = false
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		deleteForce = origForce
	}()

	cleanup := pipeStdinLines(t, testMasterPassword, "n")
	defer cleanup()
	cleanStderr := suppressStderr(t)
	defer cleanStderr()

	err := runDelete(deleteCmd, []string{"Test Note"})
	require.NoError(t, err) // Should succeed (just cancels)
}
