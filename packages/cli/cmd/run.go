package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/types"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Run a command with secrets injected as environment variables",
	Long: `Parse a .env file, resolve zp:// references to actual secret values from
the vault, and execute the given command with those environment variables set.
Secrets exist only in process memory, never written to disk.

Format for references: zp://item-name/field-name`,
	RunE:               runRun,
	DisableFlagParsing: false,
}

var (
	runEnvFile string
	runEnvName string
)

func init() {
	runCmd.Flags().StringVar(&runEnvFile, "env-file", ".env", "path to .env file with zp:// references")
	runCmd.Flags().StringVar(&runEnvName, "env", "", "environment name to inject secrets from (uses env:<name> tags)")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified. Usage: %s", cliUsage("run -- <command> [args...]"))
	}

	// Parse .env file
	envVars, hasRefs, err := ParseEnvFile(runEnvFile)
	if err != nil {
		return fmt.Errorf("parse env file: %w", err)
	}

	needVault := hasRefs || runEnvName != ""

	if needVault {
		v, err := openAndUnlockVault()
		if err != nil {
			return err
		}

		mgr, idx, err := newItemManager(v)
		if err != nil {
			return err
		}

		// Resolve all zp:// references
		if hasRefs {
			for key, val := range envVars {
				resolved, err := resolveZPReference(val, mgr, idx)
				if err != nil {
					return fmt.Errorf("resolve %s: %w", key, err)
				}
				envVars[key] = resolved
			}
		}

		// If --env flag is set, inject secrets from tagged items
		if runEnvName != "" {
			envTag := "env:" + runEnvName
			items, err := mgr.ListItems(types.ItemFilter{Tags: []string{envTag}})
			if err != nil {
				return fmt.Errorf("list env items: %w", err)
			}

			for _, itm := range items {
				prefix := strings.ToUpper(strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(itm.Name))
				for fieldName, fieldValue := range itm.Fields {
					envKey := prefix + "_" + strings.ToUpper(fieldName)
					envVars[envKey] = fieldValue
				}
			}
		}

		// Cleanup before syscall.Exec replaces the process (defers won't run).
		idx.Close()
		v.Lock()
	}

	// Build environment: inherit current env + add resolved vars
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}

	// Find the command binary
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("command not found: %s", args[0])
	}

	// Exec replaces the current process
	return syscall.Exec(binary, args, env)
}

// ParseEnvFile reads a .env file and returns key-value pairs.
// Returns true if any value contains a zp:// reference.
func ParseEnvFile(path string) (map[string]string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	vars := make(map[string]string)
	hasRefs := false
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		eqIdx := strings.IndexByte(line, '=')
		if eqIdx < 0 {
			return nil, false, fmt.Errorf("line %d: invalid format (missing '=')", lineNum)
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		// Remove surrounding quotes
		value = unquoteEnvValue(value)

		if strings.HasPrefix(value, "zp://") {
			hasRefs = true
		}

		vars[key] = value
	}

	return vars, hasRefs, scanner.Err()
}

// unquoteEnvValue removes surrounding single or double quotes from a value.
func unquoteEnvValue(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// resolveZPReference resolves a zp:// reference to the actual secret value.
// Format: zp://item-name/field-name
func resolveZPReference(value string, mgr *item.Manager, idx *index.Index) (string, error) {
	if !strings.HasPrefix(value, "zp://") {
		return value, nil
	}

	ref := strings.TrimPrefix(value, "zp://")
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid zp:// reference %q: expected zp://item-name/field-name", value)
	}

	itemQuery := parts[0]
	fieldName := parts[1]

	itm, err := findItemByQuery(mgr, idx, itemQuery)
	if err != nil {
		return "", fmt.Errorf("resolve item %q: %w", itemQuery, err)
	}

	// Check standard fields
	if val, ok := itm.Fields[fieldName]; ok {
		return val, nil
	}
	// Check custom fields
	if val, ok := itm.CustomFields[fieldName]; ok {
		return val, nil
	}
	// Special: "notes" field
	if fieldName == "notes" {
		return itm.Notes, nil
	}

	return "", fmt.Errorf("field %q not found in item %q", fieldName, itm.Name)
}
