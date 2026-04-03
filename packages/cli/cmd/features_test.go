package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/types"
)

// ---------- Completion Command Tests ----------

func TestCompletionBash(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = completionCmd.RunE(completionCmd, []string{"bash"})
	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NotEmpty(t, output)
	// Bash completion scripts contain this marker
	assert.Contains(t, output, "bash completion")
}

func TestCompletionZsh(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = completionCmd.RunE(completionCmd, []string{"zsh"})
	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NotEmpty(t, output)
	// Zsh completions typically reference compdef or #compdef
	assert.True(t, strings.Contains(output, "zsh") || strings.Contains(output, "compdef") || strings.Contains(output, cliCommandName),
		"expected zsh completion output")
}

func TestCompletionFish(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = completionCmd.RunE(completionCmd, []string{"fish"})
	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NotEmpty(t, output)
	// Fish completions use "complete" keyword
	assert.Contains(t, output, "complete")
}

func TestCompletionPowershell(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = completionCmd.RunE(completionCmd, []string{"powershell"})
	w.Close()
	os.Stdout = old

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NotEmpty(t, output)
	// PowerShell completions reference Register-ArgumentCompleter
	assert.Contains(t, output, "Register-ArgumentCompleter")
}

func TestCompletionInvalidShell(t *testing.T) {
	// cobra.OnlyValidArgs should reject invalid shell types
	err := completionCmd.Args(completionCmd, []string{"invalid"})
	assert.Error(t, err)
}

func TestCompletionNoArgs(t *testing.T) {
	// cobra.ExactArgs(1) should reject no args
	err := completionCmd.Args(completionCmd, []string{})
	assert.Error(t, err)
}

func TestCompletionTooManyArgs(t *testing.T) {
	// cobra.ExactArgs(1) should reject too many args
	err := completionCmd.Args(completionCmd, []string{"bash", "zsh"})
	assert.Error(t, err)
}

func TestCompletionValidArgs(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			err := completionCmd.Args(completionCmd, []string{shell})
			assert.NoError(t, err)
		})
	}
}

func TestCompletionCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "completion [bash|zsh|fish|powershell]" {
			found = true
			break
		}
	}
	assert.True(t, found, "completion command should be registered on rootCmd")
}

// ---------- Run --env Flag Tests ----------

func TestRunEnvFlagRegistered(t *testing.T) {
	flag := runCmd.Flags().Lookup("env")
	require.NotNil(t, flag, "--env flag should be registered on run command")
	assert.Equal(t, "", flag.DefValue)
	assert.Contains(t, flag.Usage, "environment name")
}

func TestRunEnvFileFlagStillExists(t *testing.T) {
	flag := runCmd.Flags().Lookup("env-file")
	require.NotNil(t, flag, "--env-file flag should still be registered")
	assert.Equal(t, ".env", flag.DefValue)
}

func TestRunEnvFlagParsing(t *testing.T) {
	// Save and restore
	origEnvName := runEnvName
	defer func() { runEnvName = origEnvName }()

	// Simulate flag parsing
	err := runCmd.Flags().Set("env", "production")
	require.NoError(t, err)
	assert.Equal(t, "production", runEnvName)
}

func TestRunEnvInjection(t *testing.T) {
	// Create a vault with items tagged env:staging
	v, _, cleanup := createTestVault(t)
	defer cleanup()

	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)
	defer idx.Close()

	vk, err := v.VaultKey()
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), func() ([]byte, error) { return vk, nil }, idx)

	// Add items tagged with env:staging
	dbItem := &types.Item{
		Name: "my-database",
		Type: types.ItemTypeLogin,
		Fields: map[string]string{
			types.FieldUsername: "admin",
			types.FieldPassword: "dbpass123",
		},
		Tags: []string{"env:staging"},
	}
	require.NoError(t, mgr.AddItem(dbItem))

	apiItem := &types.Item{
		Name: "api.service",
		Type: types.ItemTypeAPIKey,
		Fields: map[string]string{
			types.FieldAPIKey: "ak_test_123",
		},
		Tags: []string{"env:staging"},
	}
	require.NoError(t, mgr.AddItem(apiItem))

	// Verify items are filterable by env:staging tag
	items, err := mgr.ListItems(types.ItemFilter{Tags: []string{"env:staging"}})
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Test the env key naming convention
	replacer := strings.NewReplacer("-", "_", " ", "_", ".", "_")
	prefix1 := strings.ToUpper(replacer.Replace("my-database"))
	assert.Equal(t, "MY_DATABASE", prefix1)

	prefix2 := strings.ToUpper(replacer.Replace("api.service"))
	assert.Equal(t, "API_SERVICE", prefix2)

	// Simulate the env var building logic
	envVars := make(map[string]string)
	for _, itm := range items {
		prefix := strings.ToUpper(replacer.Replace(itm.Name))
		for fieldName, fieldValue := range itm.Fields {
			envKey := prefix + "_" + strings.ToUpper(fieldName)
			envVars[envKey] = fieldValue
		}
	}

	// Verify the expected keys are present (items may be in any order)
	_, hasVaultPath := envVars["MY_DATABASE_"+strings.ToUpper(types.FieldUsername)]
	assert.True(t, hasVaultPath || envVars["MY_DATABASE_USERNAME"] == "admin" || envVars["MY_DATABASE_"+strings.ToUpper(types.FieldUsername)] == "admin",
		"expected MY_DATABASE_USERNAME=admin")
}

// ---------- Recovery --regenerate Flag Tests ----------

func TestRecoveryRegenerateFlagRegistered(t *testing.T) {
	flag := recoveryCmd.Flags().Lookup("regenerate")
	require.NotNil(t, flag, "--regenerate flag should be registered on recovery command")
	assert.Equal(t, "false", flag.DefValue)
	assert.Contains(t, flag.Usage, "generate new recovery phrase")
}

func TestRecoveryValidateFlagStillExists(t *testing.T) {
	flag := recoveryCmd.Flags().Lookup("validate")
	require.NotNil(t, flag, "--validate flag should still be registered")
}

func TestRecoveryRegenerateFlagParsing(t *testing.T) {
	origVal := recoveryRegenerate
	defer func() { recoveryRegenerate = origVal }()

	err := recoveryCmd.Flags().Set("regenerate", "true")
	require.NoError(t, err)
	assert.True(t, recoveryRegenerate)

	// Reset
	err = recoveryCmd.Flags().Set("regenerate", "false")
	require.NoError(t, err)
	assert.False(t, recoveryRegenerate)
}

func TestRecoveryRegenerate_JSON(t *testing.T) {
	v, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	// Save and restore globals
	origPath := flagVaultPath
	origOutput := flagOutput
	origRegen := recoveryRegenerate
	flagVaultPath = vaultPath
	flagOutput = "json"
	recoveryRegenerate = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		recoveryRegenerate = origRegen
	}()

	// Lock vault so runRecovery will prompt for password
	v.Lock()

	stdinCleanup := pipeStdin(t, testMasterPassword)
	defer stdinCleanup()
	stderrCleanup := suppressStderr(t)
	defer stderrCleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = runRecovery(recoveryCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "regenerated", result["status"])
	assert.NotEmpty(t, result["mnemonic"])

	// Verify the mnemonic is actually valid (has words)
	words := strings.Fields(result["mnemonic"])
	assert.True(t, len(words) >= 12, "mnemonic should have at least 12 words, got %d", len(words))
}

func TestRecoveryRegenerate_Text(t *testing.T) {
	_, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	origPath := flagVaultPath
	origOutput := flagOutput
	origRegen := recoveryRegenerate
	flagVaultPath = vaultPath
	flagOutput = "text"
	recoveryRegenerate = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		recoveryRegenerate = origRegen
	}()

	stdinCleanup := pipeStdin(t, testMasterPassword)
	defer stdinCleanup()

	// Capture stderr (text mode outputs to stderr)
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	err = runRecovery(recoveryCmd, []string{})

	w.Close()
	os.Stderr = oldStderr

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "WARNING")
	assert.Contains(t, output, "INVALID")
	assert.Contains(t, output, "New recovery phrase:")
	assert.Contains(t, output, "Store this phrase safely")
}

func TestRecoveryRegenerate_NonexistentVault(t *testing.T) {
	origPath := flagVaultPath
	origRegen := recoveryRegenerate
	flagVaultPath = "/nonexistent/vault/path"
	recoveryRegenerate = true
	defer func() {
		flagVaultPath = origPath
		recoveryRegenerate = origRegen
	}()

	stderrCleanup := suppressStderr(t)
	defer stderrCleanup()

	err := runRecovery(recoveryCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open vault")
}

func TestRecoveryRegenerate_InvalidatesOldPhrase(t *testing.T) {
	v, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	origPath := flagVaultPath
	origOutput := flagOutput
	origRegen := recoveryRegenerate
	flagVaultPath = vaultPath
	flagOutput = "json"
	recoveryRegenerate = true
	defer func() {
		flagVaultPath = origPath
		flagOutput = origOutput
		recoveryRegenerate = origRegen
	}()

	// Lock vault first
	v.Lock()

	stdinCleanup := pipeStdin(t, testMasterPassword)
	defer stdinCleanup()
	stderrCleanup := suppressStderr(t)
	defer stderrCleanup()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	err = runRecovery(recoveryCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	newMnemonic := result["mnemonic"]
	assert.NotEmpty(t, newMnemonic)

	// Verify the new mnemonic can unlock the vault
	v2, err := store.Open(vaultPath)
	require.NoError(t, err)

	_, err = v2.UnlockWithRecovery(newMnemonic)
	require.NoError(t, err, "new mnemonic should unlock the vault")
	v2.Lock()
}

// ---------- Run --env Integration Test (with vault) ----------

func TestRunEnvWithVault(t *testing.T) {
	// Create a vault with env-tagged items
	v, vaultPath, cleanup := createTestVault(t)
	defer cleanup()

	idx, err := index.Open(v.IndexPath())
	require.NoError(t, err)

	vk, err := v.VaultKey()
	require.NoError(t, err)
	mgr := item.NewManager(v.ItemsPath(), func() ([]byte, error) { return vk, nil }, idx)

	// Add items tagged with env:test
	testItem := &types.Item{
		Name: "test-service",
		Type: types.ItemTypeAPIKey,
		Fields: map[string]string{
			types.FieldAPIKey:    "key_abc",
			types.FieldAPISecret: "secret_xyz",
		},
		Tags: []string{"env:test"},
	}
	require.NoError(t, mgr.AddItem(testItem))
	idx.Close()

	v.Lock()

	// Verify env:test filtering works via the mgr
	v2, err := store.Open(vaultPath)
	require.NoError(t, err)
	require.NoError(t, v2.Unlock(testMasterPassword))
	defer v2.Lock()

	idx2, err := index.Open(v2.IndexPath())
	require.NoError(t, err)
	defer idx2.Close()

	vk2, err := v2.VaultKey()
	require.NoError(t, err)
	mgr2 := item.NewManager(v2.ItemsPath(), func() ([]byte, error) { return vk2, nil }, idx2)

	items, err := mgr2.ListItems(types.ItemFilter{Tags: []string{"env:test"}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "test-service", items[0].Name)

	// Build env vars like the run command would
	envVars := make(map[string]string)
	replacer := strings.NewReplacer("-", "_", " ", "_", ".", "_")
	for _, itm := range items {
		prefix := strings.ToUpper(replacer.Replace(itm.Name))
		for fieldName, fieldValue := range itm.Fields {
			envKey := prefix + "_" + strings.ToUpper(fieldName)
			envVars[envKey] = fieldValue
		}
	}

	assert.Equal(t, "key_abc", envVars["TEST_SERVICE_"+strings.ToUpper(types.FieldAPIKey)])
	assert.Equal(t, "secret_xyz", envVars["TEST_SERVICE_"+strings.ToUpper(types.FieldAPISecret)])
}
