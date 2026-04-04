package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/term"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/vault/index"
	"github.com/zeropass/zeropass/core/vault/item"
	"github.com/zeropass/zeropass/core/vault/store"
	"github.com/zeropass/zeropass/core/vault/types"
)

// getVaultPath resolves the vault path from flags or default location.
func getVaultPath() string {
	if flagVaultPath != "" {
		return flagVaultPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", store.DefaultVaultDir, "vaults", store.DefaultVaultName)
	}
	return filepath.Join(home, store.DefaultVaultDir, "vaults", store.DefaultVaultName)
}

// promptPassword reads a password from the terminal with no echo.
func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		pw, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		result := string(pw)
		crypto.ZeroBytes(pw)
		return result, nil
	}
	// Non-terminal fallback (piped input)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("no input available")
}

// promptConfirmPassword prompts for password twice and confirms they match.
func promptConfirmPassword() (string, error) {
	pw, err := promptPassword("Enter master password: ")
	if err != nil {
		return "", err
	}
	confirm, err := promptPassword("Confirm master password: ")
	if err != nil {
		return "", err
	}
	if pw != confirm {
		return "", fmt.Errorf("passwords do not match")
	}
	return pw, nil
}

// promptLine reads a single line from stdin.
func promptLine(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// promptYesNo asks a yes/no question and returns true for yes.
func promptYesNo(prompt string) bool {
	answer := promptLine(prompt + " [y/N]: ")
	return strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes"
}

// openAndUnlockVault opens and unlocks the vault, prompting for password.
func openAndUnlockVault() (*store.Vault, error) {
	vaultPath := getVaultPath()

	v, err := store.Open(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("open vault: %w (run '%s' to create a vault)", err, cliUsage("init"))
	}

	pw, err := promptPassword("Enter master password: ")
	if err != nil {
		return nil, err
	}
	defer zeroString(pw)

	if err := v.Unlock(pw); err != nil {
		return nil, fmt.Errorf("unlock vault: %w", err)
	}
	return v, nil
}

// newItemManager creates an item manager with a search index for the given vault.
func newItemManager(v *store.Vault) (*item.Manager, *index.Index, error) {
	idx, err := index.Open(v.IndexPath())
	if err != nil {
		return nil, nil, fmt.Errorf("open index: %w", err)
	}
	mgr := item.NewManager(v.ItemsPath(), v.VaultKey, idx)
	if err := idx.EnsureCurrentPolicy(mgr.AllItems); err != nil {
		_ = idx.Close()
		return nil, nil, fmt.Errorf("initialize search index: %w", err)
	}
	return mgr, idx, nil
}

// copyToClipboard copies text to clipboard with optional auto-clear.
func copyToClipboard(text string, clearAfterSec int) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}
	if clearAfterSec > 0 {
		go func() {
			time.Sleep(time.Duration(clearAfterSec) * time.Second)
			_ = clipboard.WriteAll("")
		}()
		fmt.Fprintf(os.Stderr, "Copied to clipboard. Will clear in %d seconds.\n", clearAfterSec)
	} else {
		fmt.Fprintln(os.Stderr, "Copied to clipboard.")
	}
	return nil
}

// formatOutput prints data as JSON or delegates to a custom formatter.
func formatOutput(data interface{}) {
	if flagOutput == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
		return
	}
	// For non-JSON, caller should handle their own formatting.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

// printItem displays an item in human-readable or JSON format.
func printItem(itm *types.Item, showSecret bool) {
	if flagOutput == "json" {
		out := itemToOutputMap(itm, showSecret)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}
	printItemText(itm, showSecret)
}

// printItemText renders an item as human-readable text.
func printItemText(itm *types.Item, showSecret bool) {
	fmt.Printf("ID:       %s\n", itm.ID)
	fmt.Printf("Name:     %s\n", itm.Name)
	fmt.Printf("Type:     %s\n", itm.Type)

	secretFields := map[string]bool{
		types.FieldPassword:   true,
		types.FieldAPISecret:  true,
		types.FieldPrivateKey: true,
		types.FieldCVV:        true,
		types.FieldPassphrase: true,
	}

	for k, v := range itm.Fields {
		if secretFields[k] && !showSecret {
			fmt.Printf("%-10s%s\n", k+":", "********")
		} else {
			fmt.Printf("%-10s%s\n", k+":", v)
		}
	}
	for k, v := range itm.CustomFields {
		fmt.Printf("%-10s%s\n", k+":", v)
	}
	if itm.Notes != "" {
		fmt.Printf("Notes:    %s\n", itm.Notes)
	}
	if len(itm.Tags) > 0 {
		fmt.Printf("Tags:     %s\n", strings.Join(itm.Tags, ", "))
	}
	if itm.Favorite {
		fmt.Println("Favorite: ★")
	}
	fmt.Printf("Created:  %s\n", itm.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated:  %s\n", itm.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Version:  %d\n", itm.Version)
}

// itemToOutputMap creates a map suitable for JSON output.
func itemToOutputMap(itm *types.Item, showSecret bool) map[string]interface{} {
	secretFields := map[string]bool{
		types.FieldPassword:   true,
		types.FieldAPISecret:  true,
		types.FieldPrivateKey: true,
		types.FieldCVV:        true,
		types.FieldPassphrase: true,
	}

	fields := make(map[string]string, len(itm.Fields))
	for k, v := range itm.Fields {
		if secretFields[k] && !showSecret {
			fields[k] = "********"
		} else {
			fields[k] = v
		}
	}

	return map[string]interface{}{
		"id":            itm.ID,
		"name":          itm.Name,
		"type":          itm.Type,
		"fields":        fields,
		"custom_fields": itm.CustomFields,
		"notes":         itm.Notes,
		"tags":          itm.Tags,
		"favorite":      itm.Favorite,
		"created_at":    itm.CreatedAt,
		"updated_at":    itm.UpdatedAt,
		"version":       itm.Version,
	}
}

// printItemsTable prints items as a formatted table.
func printItemsTable(items []*types.Item) {
	if flagOutput == "json" {
		out := make([]map[string]interface{}, len(items))
		for i, itm := range items {
			out[i] = map[string]interface{}{
				"id":       itm.ID,
				"name":     itm.Name,
				"type":     itm.Type,
				"favorite": itm.Favorite,
				"tags":     itm.Tags,
			}
			if url := itm.Fields[types.FieldURL]; url != "" {
				out[i]["url"] = url
			}
			if user := itm.Fields[types.FieldUsername]; user != "" {
				out[i]["username"] = user
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}

	if len(items) == 0 {
		fmt.Println("No items found.")
		return
	}

	// Calculate column widths
	nameW, typeW := 4, 4 // minimums: "NAME", "TYPE"
	for _, itm := range items {
		if l := len(itm.Name); l > nameW {
			nameW = l
		}
		if l := len(string(itm.Type)); l > typeW {
			typeW = l
		}
	}
	if nameW > 40 {
		nameW = 40
	}
	if typeW > 15 {
		typeW = 15
	}

	hdr := fmt.Sprintf("%-36s  %-*s  %-*s  %s", "ID", nameW, "NAME", typeW, "TYPE", "PREVIEW")
	fmt.Println(hdr)
	fmt.Println(strings.Repeat("─", len(hdr)+20))

	for _, itm := range items {
		name := itm.Name
		if len(name) > nameW {
			name = name[:nameW-1] + "…"
		}
		preview := itemPreview(itm)
		fav := ""
		if itm.Favorite {
			fav = " ★"
		}
		fmt.Printf("%-36s  %-*s  %-*s  %s%s\n", itm.ID, nameW, name, typeW, string(itm.Type), preview, fav)
	}
	fmt.Printf("\nTotal: %d items\n", len(items))
}

// itemPreview returns a short preview string for an item.
func itemPreview(itm *types.Item) string {
	switch itm.Type {
	case types.ItemTypeLogin:
		if u := itm.Fields[types.FieldURL]; u != "" {
			return u
		}
		if u := itm.Fields[types.FieldUsername]; u != "" {
			return u
		}
	case types.ItemTypeAPIKey:
		if k := itm.Fields[types.FieldAPIKey]; k != "" {
			if len(k) > 8 {
				return k[:8] + "…"
			}
			return k
		}
	case types.ItemTypeSSHKey:
		if fp := itm.Fields[types.FieldFingerprint]; fp != "" {
			return fp
		}
	case types.ItemTypeCreditCard:
		if n := itm.Fields[types.FieldCardNumber]; n != "" {
			if len(n) >= 4 {
				return "****" + n[len(n)-4:]
			}
		}
	case types.ItemTypeIdentity:
		parts := []string{}
		if fn := itm.Fields[types.FieldFirstName]; fn != "" {
			parts = append(parts, fn)
		}
		if ln := itm.Fields[types.FieldLastName]; ln != "" {
			parts = append(parts, ln)
		}
		return strings.Join(parts, " ")
	case types.ItemTypeSecureNote:
		if itm.Notes != "" {
			n := itm.Notes
			if len(n) > 40 {
				n = n[:38] + "..."
			}
			return n
		}
	}
	return ""
}

// findItemByQuery searches for an item by name or ID.
// If multiple matches, prompts the user to pick one.
func findItemByQuery(mgr *item.Manager, idx *index.Index, query string) (*types.Item, error) {
	// Try exact ID match first
	itm, err := mgr.GetItem(query)
	if err == nil && itm != nil {
		return itm, nil
	}

	// Search by name/content
	ids, err := idx.Search(query)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	if len(ids) == 0 {
		// Fallback: brute-force name search
		all, err := mgr.AllItems()
		if err != nil {
			return nil, err
		}
		for _, it := range all {
			if strings.EqualFold(it.Name, query) {
				ids = append(ids, it.ID)
			}
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("no items matching %q", query)
		}
	}

	if len(ids) == 1 {
		return mgr.GetItem(ids[0])
	}

	// Multiple matches — let user pick
	fmt.Fprintf(os.Stderr, "Multiple items found for %q:\n", query)
	var items []*types.Item
	for _, id := range ids {
		it, e := mgr.GetItem(id)
		if e == nil {
			items = append(items, it)
		}
	}

	for i, it := range items {
		fmt.Fprintf(os.Stderr, "  [%d] %s (%s) — %s\n", i+1, it.Name, it.Type, it.ID)
	}

	answer := promptLine("Select item number: ")
	var idx2 int
	if _, err := fmt.Sscanf(answer, "%d", &idx2); err != nil || idx2 < 1 || idx2 > len(items) {
		return nil, fmt.Errorf("invalid selection")
	}
	return items[idx2-1], nil
}

// zeroString overwrites a string's memory. Note: Go strings are immutable
// so this is best-effort for security hygiene.
func zeroString(s string) {
	b := []byte(s)
	crypto.ZeroBytes(b)
}

// parseTags splits a comma-separated tags string.
func parseTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
