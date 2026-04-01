package importexport

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/zeropass/zeropass/core/crypto/cipher"
	"github.com/zeropass/zeropass/core/crypto/encoding"
	"github.com/zeropass/zeropass/core/vault/types"
)

// ExportJSON writes items as unencrypted JSON.
func ExportJSON(items []types.Item, writer io.Writer) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("export JSON: %w", err)
	}
	return nil
}

// ExportCSV writes items as unencrypted CSV.
// Columns: name, type, url, username, password, notes, tags
func ExportCSV(items []types.Item, writer io.Writer) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"name", "type", "url", "username", "password", "notes", "tags"}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, item := range items {
		row := []string{
			item.Name,
			string(item.Type),
			item.Fields[types.FieldURL],
			item.Fields[types.FieldUsername],
			item.Fields[types.FieldPassword],
			item.Notes,
			joinTags(item.Tags),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return nil
}

// ExportEncrypted writes items as an encrypted backup.
// Format: base64-encoded AES-256-GCM ciphertext of the JSON payload.
func ExportEncrypted(items []types.Item, key []byte, writer io.Writer) error {
	plaintext, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}

	c := cipher.New()
	ciphertext, err := c.Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("encrypt export: %w", err)
	}

	encoded := encoding.Base64StdEncode(ciphertext)
	if _, err := io.WriteString(writer, encoded); err != nil {
		return fmt.Errorf("write encrypted export: %w", err)
	}
	return nil
}

// ImportEncrypted reads an encrypted backup and returns the items.
func ImportEncrypted(reader io.Reader, key []byte) ([]types.Item, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read encrypted export: %w", err)
	}

	ciphertext, err := encoding.Base64StdDecode(string(data))
	if err != nil {
		return nil, fmt.Errorf("decode encrypted export: %w", err)
	}

	c := cipher.New()
	plaintext, err := c.Decrypt(ciphertext, key)
	if err != nil {
		return nil, fmt.Errorf("decrypt export: %w", err)
	}

	var items []types.Item
	if err := json.Unmarshal(plaintext, &items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}
	return items, nil
}

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for _, t := range tags[1:] {
		result += "," + t
	}
	return result
}
