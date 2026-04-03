package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ───────── runRun additional error paths ─────────

func TestRunRun_EnvFlagWithNoVault(t *testing.T) {
	// Create .env with no refs, but --env flag set — should fail opening vault
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("PLAIN=value\n"), 0600))

	origEnvFile := runEnvFile
	origEnvName := runEnvName
	origVaultPath := flagVaultPath
	defer func() {
		runEnvFile = origEnvFile
		runEnvName = origEnvName
		flagVaultPath = origVaultPath
	}()

	runEnvFile = envFile
	runEnvName = "production"
	flagVaultPath = "/nonexistent/vault"

	stderrCleanup := suppressStderr(t)
	defer stderrCleanup()

	err := runRun(runCmd, []string{"echo"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open vault")
}

// ───────── ParseEnvFile additional edge cases ─────────

func TestParseEnvFile_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "DOUBLE=\"hello world\"\nSINGLE='foo bar'\nNONE=plain\n"
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0600))

	vars, _, err := ParseEnvFile(envFile)
	require.NoError(t, err)
	assert.Equal(t, "hello world", vars["DOUBLE"])
	assert.Equal(t, "foo bar", vars["SINGLE"])
	assert.Equal(t, "plain", vars["NONE"])
}

func TestParseEnvFile_WithMultipleZPRefs(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "DB_PASS=zp://mydb/password\nAPI_KEY=zp://api/key\nPLAIN=regular\n"
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0600))

	vars, hasRefs, err := ParseEnvFile(envFile)
	require.NoError(t, err)
	assert.True(t, hasRefs)
	assert.Equal(t, "zp://mydb/password", vars["DB_PASS"])
	assert.Equal(t, "zp://api/key", vars["API_KEY"])
	assert.Equal(t, "regular", vars["PLAIN"])
}

func TestParseEnvFile_InvalidLineFormat(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := "VALID=ok\nINVALID_LINE_NO_EQUALS\n"
	require.NoError(t, os.WriteFile(envFile, []byte(content), 0600))

	_, _, err := ParseEnvFile(envFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

// ───────── Env var naming convention ─────────

func TestEnvVarNamingConvention(t *testing.T) {
	replacer := strings.NewReplacer("-", "_", " ", "_", ".", "_")

	tests := []struct {
		input    string
		expected string
	}{
		{"my-database", "MY_DATABASE"},
		{"api.service", "API_SERVICE"},
		{"web server", "WEB_SERVER"},
		{"my-api.service-v2", "MY_API_SERVICE_V2"},
		{"simple", "SIMPLE"},
		{"a-b.c d", "A_B_C_D"},
	}

	for _, tt := range tests {
		got := strings.ToUpper(replacer.Replace(tt.input))
		assert.Equal(t, tt.expected, got, "naming for %q", tt.input)
	}
}

// ───────── unquoteEnvValue edge cases ─────────

func TestUnquoteEnvValue_MismatchedQuotes(t *testing.T) {
	assert.Equal(t, `"hello'`, unquoteEnvValue(`"hello'`))
	assert.Equal(t, `'hello"`, unquoteEnvValue(`'hello"`))
}

func TestUnquoteEnvValue_SingleChar(t *testing.T) {
	assert.Equal(t, "x", unquoteEnvValue("x"))
	assert.Equal(t, `"`, unquoteEnvValue(`"`))
}

func TestUnquoteEnvValue_EmptyString(t *testing.T) {
	assert.Equal(t, "", unquoteEnvValue(""))
}

func TestUnquoteEnvValue_EmptyQuoted(t *testing.T) {
	assert.Equal(t, "", unquoteEnvValue(`""`))
	assert.Equal(t, "", unquoteEnvValue(`''`))
}

