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
	"github.com/zeropass/zeropass/core/vault/importexport"
	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/types"
)

// captureStdout captures stdout output during fn execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// ---------- Full Integration Tests ----------

// TestIntegration_CompleteWorkflow tests the entire CRUD lifecycle
// init → add various types → search → list → edit → get → delete → export → import
func TestIntegration_CompleteWorkflow(t *testing.T) {
	v, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	// Verify vault path
	assert.DirExists(t, vaultPath)

	mgr, idx := createTestManager(t, v)

	// === Step 1: Add items of every type ===
	loginItem := &types.Item{
		Name: "GitHub Login",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "dev@github.com",
			types.FieldPassword: "ghp_SuperSecure123!",
			types.FieldURL:      "https://github.com",
		},
		Tags:     []string{"dev", "env:production"},
		Favorite: true,
	}
	require.NoError(t, mgr.AddItem(loginItem))
	assert.NotEmpty(t, loginItem.ID)
	assert.Equal(t, 1, loginItem.Version)

	apiItem := &types.Item{
		Name: "Stripe API Key",
		Type: types.ItemTypeAPIKey,
		Fields: map[string]string{
			types.FieldAPIKey:    "sk_live_abcdef123456",
			types.FieldAPISecret: "whsec_secret_xyz",
			types.FieldEndpoint:  "https://api.stripe.com",
		},
		Tags: []string{"billing", "env:production"},
	}
	require.NoError(t, mgr.AddItem(apiItem))

	sshItem := &types.Item{
		Name: "Deploy SSH Key",
		Type: types.ItemTypeSSHKey,
		Fields: map[string]string{
			types.FieldPrivateKey:  "-----BEGIN RSA KEY-----\nMIIE...\n-----END RSA KEY-----",
			types.FieldPublicKey:   "ssh-rsa AAAAB3... deploy@server",
			types.FieldFingerprint: "SHA256:abcdef123456",
		},
	}
	require.NoError(t, mgr.AddItem(sshItem))

	noteItem := &types.Item{
		Name:  "Server Documentation",
		Type:  types.ItemTypeSecureNote,
		Notes: "Production server: 10.0.0.1\nStaging: 10.0.0.2\nDB Port: 5432",
		Tags:  []string{"infra", "docs"},
	}
	require.NoError(t, mgr.AddItem(noteItem))

	cardItem := &types.Item{
		Name: "Company Credit Card",
		Type: types.ItemTypeCreditCard,
		Fields: map[string]string{
			types.FieldCardNumber: "4111222233334444",
			types.FieldCardHolder: "John Developer",
			types.FieldExpiry:     "12/26",
			types.FieldCVV:        "999",
		},
	}
	require.NoError(t, mgr.AddItem(cardItem))

	identityItem := &types.Item{
		Name: "Company Identity",
		Type: types.ItemTypeIdentity,
		Fields: map[string]string{
			types.FieldFirstName: "John",
			types.FieldLastName:  "Developer",
			types.FieldEmail:     "john@company.com",
			types.FieldPhone:     "+1-555-0100",
			types.FieldAddress:   "123 Tech Ave, SF, CA 94105",
		},
	}
	require.NoError(t, mgr.AddItem(identityItem))

	customItem := &types.Item{
		Name:         "Custom Config",
		Type:         types.ItemTypeCustom,
		Fields:       map[string]string{},
		CustomFields: map[string]string{"env": "production", "region": "us-east-1"},
		Notes:        "Custom config for deployment",
	}
	require.NoError(t, mgr.AddItem(customItem))

	// === Step 2: Verify all items exist ===
	allItems, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, allItems, 7)

	// === Step 3: Search functionality ===
	ids, err := idx.Search("github")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, loginItem.ID, ids[0])

	ids, err = idx.Search("stripe")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, apiItem.ID, ids[0])

	// === Step 4: List with filters ===
	// By type
	logins, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeLogin})
	require.NoError(t, err)
	assert.Len(t, logins, 1)

	apikeys, err := mgr.ListItems(types.ItemFilter{Type: types.ItemTypeAPIKey})
	require.NoError(t, err)
	assert.Len(t, apikeys, 1)

	// By tag
	prodItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:production"}})
	require.NoError(t, err)
	assert.Len(t, prodItems, 2) // login + api key

	// By favorite
	fav := true
	favItems, err := mgr.ListItems(types.ItemFilter{Favorite: &fav})
	require.NoError(t, err)
	assert.Len(t, favItems, 1)
	assert.Equal(t, "GitHub Login", favItems[0].Name)

	// === Step 5: Edit item ===
	loginItem.Fields[types.FieldPassword] = "new_super_secure_password_456!"
	loginItem.Fields[types.FieldURL] = "https://github.com/login"
	loginItem.Notes = "Updated login info"
	err = mgr.UpdateItem(loginItem.ID, loginItem)
	require.NoError(t, err)
	assert.Equal(t, 2, loginItem.Version)

	// === Step 6: Get and verify edit ===
	updated, err := mgr.GetItem(loginItem.ID)
	require.NoError(t, err)
	assert.Equal(t, "new_super_secure_password_456!", updated.Fields[types.FieldPassword])
	assert.Equal(t, "https://github.com/login", updated.Fields[types.FieldURL])
	assert.Equal(t, "Updated login info", updated.Notes)
	assert.Equal(t, 2, updated.Version)

	// === Step 7: Delete an item ===
	err = mgr.DeleteItem(noteItem.ID)
	require.NoError(t, err)

	remaining, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, remaining, 6)

	// Verify the deleted item is gone
	_, err = mgr.GetItem(noteItem.ID)
	assert.Error(t, err)

	// === Step 8: Export/Import JSON ===
	plainItems := make([]types.Item, len(remaining))
	for i, itm := range remaining {
		plainItems[i] = *itm
	}

	var exportBuf bytes.Buffer
	err = importexport.ExportJSON(plainItems, &exportBuf)
	require.NoError(t, err)
	assert.True(t, exportBuf.Len() > 0)

	// Verify it's valid JSON
	var importedItems []types.Item
	err = json.Unmarshal(exportBuf.Bytes(), &importedItems)
	require.NoError(t, err)
	assert.Len(t, importedItems, 6)

	// === Step 9: Export CSV ===
	var csvBuf bytes.Buffer
	err = importexport.ExportCSV(plainItems, &csvBuf)
	require.NoError(t, err)
	assert.Contains(t, csvBuf.String(), "GitHub Login")
}

// TestIntegration_HealthReport tests health analysis on vault with varied passwords.
func TestIntegration_HealthReport(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add items with various password qualities
	weakPassItems := []struct {
		name string
		pw   string
	}{
		{"Weak Site 1", "password"},
		{"Weak Site 2", "123456"},
		{"Weak Site 3", "qwerty"},
	}
	for _, wp := range weakPassItems {
		require.NoError(t, mgr.AddItem(&types.Item{
			Name: wp.name, Type: types.ItemTypeLogin,
			Fields: map[string]string{types.FieldPassword: wp.pw},
		}))
	}

	// Strong passwords
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Strong Site", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldPassword: "X9#kL2$mP7@nQ4&wR6!vT8"},
	}))

	// Reused password
	reusedPw := "Reused!P@ssw0rd#2024"
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Reused A", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldPassword: reusedPw},
	}))
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Reused B", Type: types.ItemTypeLogin,
		Fields: map[string]string{types.FieldPassword: reusedPw},
	}))

	// Non-login items (should not affect login stats)
	require.NoError(t, mgr.AddItem(&types.Item{
		Name: "Just a Note", Type: types.ItemTypeSecureNote,
		Notes: "no password here",
	}))

	items, err := mgr.AllItems()
	require.NoError(t, err)

	analyzer := health.NewAnalyzer(health.DefaultAnalyzerConfig())
	report := analyzer.Analyze(items)

	assert.Equal(t, 7, report.TotalItems)
	assert.Equal(t, 6, report.LoginItems)
	assert.GreaterOrEqual(t, report.WeakCount, 1)
	assert.GreaterOrEqual(t, report.ReusedCount, 1)

	// Verify findings exist
	assert.NotEmpty(t, report.Findings)
}

// TestIntegration_GenerateVariousConfigs tests generate with all possible configs.
func TestIntegration_GenerateVariousConfigs(t *testing.T) {
	configs := []struct {
		name       string
		length     int
		noSymbols  bool
		noDigits   bool
		noUpper    bool
		passphrase bool
		words      int
		sep        string
	}{
		{"default", 32, false, false, false, false, 0, ""},
		{"short", 8, false, false, false, false, 0, ""},
		{"long", 64, false, false, false, false, 0, ""},
		{"no-symbols", 16, true, false, false, false, 0, ""},
		{"no-digits", 16, false, true, false, false, 0, ""},
		{"no-upper", 16, false, false, true, false, 0, ""},
		{"lowercase-only", 16, true, true, true, false, 0, ""},
		{"passphrase-4", 0, false, false, false, true, 4, "-"},
		{"passphrase-6", 0, false, false, false, true, 6, "."},
		{"passphrase-8", 0, false, false, false, true, 8, " "},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			origOutput := flagOutput
			flagOutput = "json"
			defer func() { flagOutput = origOutput }()

			genLength = cfg.length
			genNoSymbols = cfg.noSymbols
			genNoDigits = cfg.noDigits
			genNoUppercase = cfg.noUpper
			genPassphrase = cfg.passphrase
			genWords = cfg.words
			genSeparator = cfg.sep
			genCopy = false

			output := captureStdout(t, func() {
				err := runGenerate(generateCmd, nil)
				require.NoError(t, err)
			})

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(output), &result))
			pw := result["password"].(string)
			assert.NotEmpty(t, pw)

			if cfg.passphrase {
				words := strings.Split(pw, cfg.sep)
				assert.Len(t, words, cfg.words)
			} else {
				assert.Len(t, pw, cfg.length)
			}

			// Verify strength info
			strength := result["strength"].(map[string]interface{})
			assert.NotNil(t, strength["score"])
			assert.NotNil(t, strength["feedback"])
		})
	}
}

// TestIntegration_PrintFunctions tests all print/display functions end-to-end.
func TestIntegration_PrintFunctions(t *testing.T) {
	now := time.Now()
	fullItem := &types.Item{
		ID:   "full-test-id",
		Name: "Full Test Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "testuser",
			types.FieldPassword: "secretpass",
			types.FieldURL:      "https://test.com",
		},
		CustomFields: map[string]string{
			"extra_field": "extra_value",
		},
		Notes:     "Some notes here",
		Tags:      []string{"dev", "test"},
		Favorite:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   3,
	}

	// Test text output with secrets hidden
	origOutput := flagOutput
	flagOutput = "text"

	output := captureStdout(t, func() {
		printItem(fullItem, false)
	})
	assert.Contains(t, output, "full-test-id")
	assert.Contains(t, output, "Full Test Item")
	assert.Contains(t, output, "login")
	assert.Contains(t, output, "testuser")
	assert.Contains(t, output, "********")     // password hidden
	assert.NotContains(t, output, "secretpass") // password not shown
	assert.Contains(t, output, "https://test.com")
	assert.Contains(t, output, "extra_value")
	assert.Contains(t, output, "Some notes here")
	assert.Contains(t, output, "dev, test")
	assert.Contains(t, output, "★")
	assert.Contains(t, output, "3")

	// Test text output with secrets shown
	output = captureStdout(t, func() {
		printItem(fullItem, true)
	})
	assert.Contains(t, output, "secretpass")

	// Test JSON output
	flagOutput = "json"
	output = captureStdout(t, func() {
		printItem(fullItem, false)
	})
	flagOutput = origOutput

	var jsonResult map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &jsonResult))
	assert.Equal(t, "full-test-id", jsonResult["id"])
	assert.Equal(t, "Full Test Item", jsonResult["name"])
	fields := jsonResult["fields"].(map[string]interface{})
	assert.Equal(t, "********", fields["password"])
	assert.Equal(t, "testuser", fields["username"])
}

// TestIntegration_EnvWorkflow tests environment tag-based workflow.
func TestIntegration_EnvWorkflow(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Create items in different environments
	envItems := []struct {
		name string
		tags []string
	}{
		{"Prod DB", []string{"env:production", "database"}},
		{"Prod API", []string{"env:production", "api"}},
		{"Staging DB", []string{"env:staging", "database"}},
		{"Staging API", []string{"env:staging", "api"}},
		{"Dev DB", []string{"env:development", "database"}},
	}

	for _, ei := range envItems {
		itm := &types.Item{
			Name: ei.name, Type: types.ItemTypeLogin,
			Fields: map[string]string{types.FieldUsername: "user"},
			Tags:   ei.tags,
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	// List items by env tag
	prodItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:production"}})
	require.NoError(t, err)
	assert.Len(t, prodItems, 2)

	stagingItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:staging"}})
	require.NoError(t, err)
	assert.Len(t, stagingItems, 2)

	devItems, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:development"}})
	require.NoError(t, err)
	assert.Len(t, devItems, 1)

	// Extract environments from items
	allItems, err := mgr.AllItems()
	require.NoError(t, err)
	envs := make(map[string]int)
	for _, itm := range allItems {
		for _, tag := range itm.Tags {
			if strings.HasPrefix(tag, envTagPrefix) {
				envName := strings.TrimPrefix(tag, envTagPrefix)
				envs[envName]++
			}
		}
	}
	assert.Equal(t, 2, envs["production"])
	assert.Equal(t, 2, envs["staging"])
	assert.Equal(t, 1, envs["development"])
}

// TestIntegration_ZPReferenceResolution tests end-to-end zp:// reference resolution.
func TestIntegration_ZPReferenceResolution(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add items that will be referenced
	dbCred := &types.Item{
		Name: "production-db",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "db_admin",
			types.FieldPassword: "super-secret-db-pw!",
			types.FieldURL:      "postgres://db.prod.internal:5432",
		},
		Notes: "Production database credentials",
	}
	require.NoError(t, mgr.AddItem(dbCred))

	apiCred := &types.Item{
		Name: "stripe-production",
		Type: types.ItemTypeAPIKey,
		Fields: map[string]string{
			types.FieldAPIKey:    "sk_live_abc123",
			types.FieldAPISecret: "whsec_xyz789",
		},
	}
	require.NoError(t, mgr.AddItem(apiCred))

	customCred := &types.Item{
		Name:         "deploy-config",
		Type:         types.ItemTypeCustom,
		Fields:       map[string]string{},
		CustomFields: map[string]string{"aws_access_key": "AKIA...", "aws_secret": "secret..."},
	}
	require.NoError(t, mgr.AddItem(customCred))

	// Create .env file with zp:// references
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	envContent := fmt.Sprintf(`# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=zp://production-db/username
DB_PASSWORD=zp://production-db/password
DB_URL=zp://production-db/url
DB_NOTES=zp://production-db/notes

# Stripe
STRIPE_KEY=zp://stripe-production/api_key
STRIPE_SECRET=zp://stripe-production/api_secret

# AWS
AWS_ACCESS_KEY=zp://deploy-config/aws_access_key

# Literals
NODE_ENV=production
PORT=3000
`)
	require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0600))

	// Parse
	vars, hasRefs, err := ParseEnvFile(envPath)
	require.NoError(t, err)
	assert.True(t, hasRefs)

	// Resolve references
	for key, val := range vars {
		resolved, err := resolveZPReference(val, mgr, idx)
		require.NoError(t, err, "resolve %s", key)
		vars[key] = resolved
	}

	// Verify resolutions
	assert.Equal(t, "localhost", vars["DB_HOST"])
	assert.Equal(t, "5432", vars["DB_PORT"])
	assert.Equal(t, "db_admin", vars["DB_USER"])
	assert.Equal(t, "super-secret-db-pw!", vars["DB_PASSWORD"])
	assert.Equal(t, "postgres://db.prod.internal:5432", vars["DB_URL"])
	assert.Equal(t, "Production database credentials", vars["DB_NOTES"])
	assert.Equal(t, "sk_live_abc123", vars["STRIPE_KEY"])
	assert.Equal(t, "whsec_xyz789", vars["STRIPE_SECRET"])
	assert.Equal(t, "AKIA...", vars["AWS_ACCESS_KEY"])
	assert.Equal(t, "production", vars["NODE_ENV"])
	assert.Equal(t, "3000", vars["PORT"])
}

// TestIntegration_VaultLockUnlockCycle tests multiple lock/unlock cycles.
func TestIntegration_VaultLockUnlockCycle(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test-vault")

	v, _, err := store.Create(testMasterPassword, vaultPath, store.DefaultConfig())
	require.NoError(t, err)

	// Add item while unlocked
	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)

	itm := &types.Item{
		Name: "Persistent Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "user",
			types.FieldPassword: "pass",
		},
	}
	require.NoError(t, mgr.AddItem(itm))
	idx.Close()
	v.Lock()

	// Multiple lock/unlock cycles
	for i := 0; i < 3; i++ {
		v2, err := store.Open(vaultPath)
		require.NoError(t, err, "cycle %d open", i)
		assert.True(t, v2.IsLocked(), "cycle %d should be locked", i)

		err = v2.Unlock(testMasterPassword)
		require.NoError(t, err, "cycle %d unlock", i)
		assert.False(t, v2.IsLocked(), "cycle %d should be unlocked", i)

		// Verify item is still there
		idx2, err := index.Open(v2.IndexPath())
		require.NoError(t, err)
		mgr2 := item.NewManager(v2.ItemsPath(), v2.VaultKey, idx2)
		got, err := mgr2.GetItem(itm.ID)
		require.NoError(t, err, "cycle %d get item", i)
		assert.Equal(t, "Persistent Item", got.Name)
		idx2.Close()

		v2.Lock()
	}
}

// TestIntegration_MultipleSearchHits tests search with multiple matching items.
func TestIntegration_MultipleSearchHits(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add items with similar names
	for i := 0; i < 5; i++ {
		itm := &types.Item{
			Name: fmt.Sprintf("Database Server %d", i),
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: fmt.Sprintf("admin%d", i),
				types.FieldPassword: fmt.Sprintf("pass%d", i),
			},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	// Search should find all database items
	ids, err := idx.Search("database")
	require.NoError(t, err)
	assert.Len(t, ids, 5)
}

// TestIntegration_ExportCSVFormat tests CSV export format validity.
func TestIntegration_ExportCSVFormat(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add some items
	items := []*types.Item{
		{
			Name: "CSV Test Login",
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: "csvuser",
				types.FieldPassword: "csvpass",
				types.FieldURL:      "https://csv.com",
			},
		},
		{
			Name: "CSV Test Note",
			Type: types.ItemTypeSecureNote,
			Notes: "A note with, commas and \"quotes\"",
		},
	}
	for _, itm := range items {
		require.NoError(t, mgr.AddItem(itm))
	}

	allItems, err := mgr.AllItems()
	require.NoError(t, err)

	plainItems := make([]types.Item, len(allItems))
	for i, itm := range allItems {
		plainItems[i] = *itm
	}

	var buf bytes.Buffer
	err = importexport.ExportCSV(plainItems, &buf)
	require.NoError(t, err)

	csvOutput := buf.String()
	assert.Contains(t, csvOutput, "CSV Test Login")
	assert.Contains(t, csvOutput, "csvuser")
}

// TestIntegration_ItemVersioning tests that versions increment correctly.
func TestIntegration_ItemVersioning(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	itm := &types.Item{
		Name: "Versioned Item",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "v1user",
			types.FieldPassword: "v1pass",
		},
	}
	require.NoError(t, mgr.AddItem(itm))
	assert.Equal(t, 1, itm.Version)

	// Update multiple times
	for i := 2; i <= 5; i++ {
		itm.Fields[types.FieldPassword] = fmt.Sprintf("v%dpass", i)
		err := mgr.UpdateItem(itm.ID, itm)
		require.NoError(t, err)
		assert.Equal(t, i, itm.Version, "version should be %d", i)
	}

	// Verify final state
	got, err := mgr.GetItem(itm.ID)
	require.NoError(t, err)
	assert.Equal(t, "v5pass", got.Fields[types.FieldPassword])
	assert.Equal(t, 5, got.Version)
}

// TestIntegration_FavoriteToggle tests toggling favorite status.
func TestIntegration_FavoriteToggle(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	itm := &types.Item{
		Name:     "Fav Item",
		Type:     types.ItemTypeLogin,
		Fields:   map[string]string{types.FieldUsername: "u"},
		Favorite: false,
	}
	require.NoError(t, mgr.AddItem(itm))

	// Toggle favorite on
	itm.Favorite = true
	err := mgr.UpdateItem(itm.ID, itm)
	require.NoError(t, err)

	got, err := mgr.GetItem(itm.ID)
	require.NoError(t, err)
	assert.True(t, got.Favorite)

	// Toggle favorite off
	got.Favorite = false
	err = mgr.UpdateItem(got.ID, got)
	require.NoError(t, err)

	got2, err := mgr.GetItem(got.ID)
	require.NoError(t, err)
	assert.False(t, got2.Favorite)
}

// TestIntegration_EncryptedExportImport tests the encrypted export/import cycle.
func TestIntegration_EncryptedExportImport(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, _ := createTestManager(t, v)

	// Add items
	for i := 0; i < 3; i++ {
		itm := &types.Item{
			Name: fmt.Sprintf("Encrypted Test %d", i),
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: fmt.Sprintf("user%d", i),
				types.FieldPassword: fmt.Sprintf("pass%d", i),
			},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	items, err := mgr.AllItems()
	require.NoError(t, err)

	plainItems := make([]types.Item, len(items))
	for i, itm := range items {
		plainItems[i] = *itm
	}

	vaultKey, err := v.VaultKey()
	require.NoError(t, err)

	// Export encrypted
	var buf bytes.Buffer
	err = importexport.ExportEncrypted(plainItems, vaultKey, &buf)
	require.NoError(t, err)

	// Import encrypted
	imported, err := importexport.ImportEncrypted(&buf, vaultKey)
	require.NoError(t, err)
	assert.Len(t, imported, 3)

	for i, imp := range imported {
		assert.Equal(t, fmt.Sprintf("Encrypted Test %d", i), imp.Name)
		assert.Equal(t, fmt.Sprintf("user%d", i), imp.Fields[types.FieldUsername])
		assert.Equal(t, fmt.Sprintf("pass%d", i), imp.Fields[types.FieldPassword])
	}
}

// TestIntegration_CobraCommandSubcommands verifies all subcommands are registered.
func TestIntegration_CobraCommandSubcommands(t *testing.T) {
	cmdNames := make(map[string]bool)
	for _, sub := range rootCmd.Commands() {
		cmdNames[sub.Name()] = true
	}

	expectedCmds := []string{
		"init", "unlock", "lock",
		"add", "get", "search", "list",
		"edit", "delete",
		"generate", "recovery", "health",
		"export", "import",
		"run", "env",
	}

	for _, name := range expectedCmds {
		assert.True(t, cmdNames[name], "expected subcommand %q to be registered", name)
	}
}

// TestIntegration_CobraCommandFlags tests that all expected flags are registered.
func TestIntegration_CobraCommandFlags(t *testing.T) {
	// Root persistent flags
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("vault-path"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("output"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("no-color"))

	// Add command flags
	assert.NotNil(t, addCmd.Flags().Lookup("type"))
	assert.NotNil(t, addCmd.Flags().Lookup("name"))
	assert.NotNil(t, addCmd.Flags().Lookup("username"))
	assert.NotNil(t, addCmd.Flags().Lookup("password"))
	assert.NotNil(t, addCmd.Flags().Lookup("url"))
	assert.NotNil(t, addCmd.Flags().Lookup("tags"))
	assert.NotNil(t, addCmd.Flags().Lookup("favorite"))
	assert.NotNil(t, addCmd.Flags().Lookup("generate-password"))
	assert.NotNil(t, addCmd.Flags().Lookup("notes"))
	assert.NotNil(t, addCmd.Flags().Lookup("key"))
	assert.NotNil(t, addCmd.Flags().Lookup("secret"))
	assert.NotNil(t, addCmd.Flags().Lookup("private-key-file"))
	assert.NotNil(t, addCmd.Flags().Lookup("public-key-file"))
	assert.NotNil(t, addCmd.Flags().Lookup("card-number"))
	assert.NotNil(t, addCmd.Flags().Lookup("card-holder"))
	assert.NotNil(t, addCmd.Flags().Lookup("card-expiry"))
	assert.NotNil(t, addCmd.Flags().Lookup("card-cvv"))
	assert.NotNil(t, addCmd.Flags().Lookup("first-name"))
	assert.NotNil(t, addCmd.Flags().Lookup("last-name"))
	assert.NotNil(t, addCmd.Flags().Lookup("email"))
	assert.NotNil(t, addCmd.Flags().Lookup("phone"))
	assert.NotNil(t, addCmd.Flags().Lookup("address"))

	// Generate command flags
	assert.NotNil(t, generateCmd.Flags().Lookup("length"))
	assert.NotNil(t, generateCmd.Flags().Lookup("no-symbols"))
	assert.NotNil(t, generateCmd.Flags().Lookup("no-digits"))
	assert.NotNil(t, generateCmd.Flags().Lookup("no-uppercase"))
	assert.NotNil(t, generateCmd.Flags().Lookup("passphrase"))
	assert.NotNil(t, generateCmd.Flags().Lookup("words"))
	assert.NotNil(t, generateCmd.Flags().Lookup("separator"))
	assert.NotNil(t, generateCmd.Flags().Lookup("copy"))

	// Get command flags
	assert.NotNil(t, getCmd.Flags().Lookup("copy"))
	assert.NotNil(t, getCmd.Flags().Lookup("field"))
	assert.NotNil(t, getCmd.Flags().Lookup("json"))

	// List command flags
	assert.NotNil(t, listCmd.Flags().Lookup("type"))
	assert.NotNil(t, listCmd.Flags().Lookup("tag"))
	assert.NotNil(t, listCmd.Flags().Lookup("favorite"))
	assert.NotNil(t, listCmd.Flags().Lookup("sort"))

	// Edit command flags
	assert.NotNil(t, editCmd.Flags().Lookup("name"))
	assert.NotNil(t, editCmd.Flags().Lookup("username"))
	assert.NotNil(t, editCmd.Flags().Lookup("password"))

	// Delete command flags
	assert.NotNil(t, deleteCmd.Flags().Lookup("force"))

	// Export command flags
	assert.NotNil(t, exportCmd.Flags().Lookup("format"))
	assert.NotNil(t, exportCmd.Flags().Lookup("file"))

	// Import command flags
	assert.NotNil(t, importCmd.Flags().Lookup("from"))
	assert.NotNil(t, importCmd.Flags().Lookup("file"))

	// Run command flags
	assert.NotNil(t, runCmd.Flags().Lookup("env-file"))

	// Recovery command flags
	assert.NotNil(t, recoveryCmd.Flags().Lookup("validate"))
}

// TestIntegration_CobraHelp tests that help works for all subcommands.
func TestIntegration_CobraHelp(t *testing.T) {
	subcommands := []string{
		"init", "unlock", "lock",
		"add", "get", "search", "list",
		"edit", "delete",
		"generate", "recovery", "health",
		"export", "import", "run", "env",
	}

	for _, subcmd := range subcommands {
		t.Run(subcmd, func(t *testing.T) {
			var found bool
			for _, c := range rootCmd.Commands() {
				if c.Name() == subcmd {
					found = true
					assert.NotEmpty(t, c.Short, "subcommand %q should have Short description", subcmd)
					break
				}
			}
			assert.True(t, found, "subcommand %q should be registered", subcmd)
		})
	}
}

// TestIntegration_EnvSubcommands tests env subcommand registration.
func TestIntegration_EnvSubcommands(t *testing.T) {
	subcmdNames := make(map[string]bool)
	for _, c := range envCmd.Commands() {
		subcmdNames[c.Name()] = true
	}
	assert.True(t, subcmdNames["list"])
	assert.True(t, subcmdNames["create"])
	assert.True(t, subcmdNames["use"])
}

// TestIntegration_RecoveryValidate tests mnemonic validation via runRecovery.
func TestIntegration_RecoveryValidate(t *testing.T) {
	origOutput := flagOutput
	flagOutput = "json"
	origValidate := recoveryValidate
	defer func() {
		flagOutput = origOutput
		recoveryValidate = origValidate
	}()

	recoveryValidate = "invalid mnemonic phrase that should not validate"

	output := captureStdout(t, func() {
		err := runRecovery(recoveryCmd, nil)
		require.NoError(t, err)
	})

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, false, result["valid"])
}

// TestIntegration_LargeVault tests operations on a vault with many items.
func TestIntegration_LargeVault(t *testing.T) {
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	mgr, idx := createTestManager(t, v)

	// Add 50 items
	for i := 0; i < 50; i++ {
		itm := &types.Item{
			Name: fmt.Sprintf("Large Vault Item %03d", i),
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldUsername: fmt.Sprintf("user%d@example.com", i),
				types.FieldPassword: fmt.Sprintf("P@ss%d!Strong", i),
				types.FieldURL:      fmt.Sprintf("https://site%d.example.com", i),
			},
			Tags: []string{fmt.Sprintf("group-%d", i%5)},
		}
		require.NoError(t, mgr.AddItem(itm))
	}

	// Verify count
	all, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, all, 50)

	// Search
	ids, err := idx.Search("Large Vault Item 025")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(ids), 1)

	// Filter by tag
	group0, err := mgr.ListItems(types.ItemFilter{Tags: []string{"group-0"}})
	require.NoError(t, err)
	assert.Len(t, group0, 10)

	// Sort
	sorted, err := mgr.ListItems(types.ItemFilter{SortBy: types.SortByName})
	require.NoError(t, err)
	assert.Len(t, sorted, 50)
	assert.Equal(t, "Large Vault Item 000", sorted[0].Name)

	// Delete half
	for i := 0; i < 25; i++ {
		require.NoError(t, mgr.DeleteItem(all[i].ID))
	}

	remaining, err := mgr.AllItems()
	require.NoError(t, err)
	assert.Len(t, remaining, 25)
}
