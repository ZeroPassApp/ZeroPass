// Package importexport provides import from various password managers and export functionality.
package importexport

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/zeropass/zeropass/core/vault/types"
)

// CSVMapping maps CSV column indices to item fields.
type CSVMapping struct {
	Name     int
	URL      int
	Username int
	Password int
	Notes    int
	Type     int // -1 if not present
	Skip     int // header rows to skip (usually 1)
}

// ImportCSV parses a generic CSV using the provided column mapping.
func ImportCSV(reader io.Reader, mapping CSVMapping) ([]types.Item, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	if len(records) <= mapping.Skip {
		return nil, nil
	}

	var items []types.Item
	for _, row := range records[mapping.Skip:] {
		item := types.Item{
			Type:   types.ItemTypeLogin,
			Fields: make(map[string]string),
		}

		item.Name = safeCol(row, mapping.Name)
		if url := safeCol(row, mapping.URL); url != "" {
			item.Fields[types.FieldURL] = url
		}
		if user := safeCol(row, mapping.Username); user != "" {
			item.Fields[types.FieldUsername] = user
		}
		if pw := safeCol(row, mapping.Password); pw != "" {
			item.Fields[types.FieldPassword] = pw
		}
		item.Notes = safeCol(row, mapping.Notes)

		if mapping.Type >= 0 {
			if t := safeCol(row, mapping.Type); t != "" {
				item.Type = mapItemType(t)
			}
		}

		if item.Name == "" && item.Fields[types.FieldURL] != "" {
			item.Name = item.Fields[types.FieldURL]
		}

		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// ImportChrome parses Chrome CSV export: name, url, username, password
func ImportChrome(reader io.Reader) ([]types.Item, error) {
	return ImportCSV(reader, CSVMapping{
		Name:     0,
		URL:      1,
		Username: 2,
		Password: 3,
		Notes:    -1,
		Type:     -1,
		Skip:     1,
	})
}

// ImportFirefox parses Firefox CSV export: url, username, password, httpRealm, formActionOrigin, guid, timeCreated, timeLastUsed, timePasswordChanged
func ImportFirefox(reader io.Reader) ([]types.Item, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse Firefox CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var items []types.Item
	for _, row := range records[1:] {
		item := types.Item{
			Type:   types.ItemTypeLogin,
			Fields: make(map[string]string),
		}

		url := safeCol(row, 0)
		item.Name = url
		if url != "" {
			item.Fields[types.FieldURL] = url
		}
		if user := safeCol(row, 1); user != "" {
			item.Fields[types.FieldUsername] = user
		}
		if pw := safeCol(row, 2); pw != "" {
			item.Fields[types.FieldPassword] = pw
		}

		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// Import1Password parses 1Password CSV export: Title, Website, Username, Password, Notes, Type
func Import1Password(reader io.Reader) ([]types.Item, error) {
	return ImportCSV(reader, CSVMapping{
		Name:     0,
		URL:      1,
		Username: 2,
		Password: 3,
		Notes:    4,
		Type:     5,
		Skip:     1,
	})
}

// bitwardenJSON is the structure for Bitwarden JSON export.
type bitwardenJSON struct {
	Items []bitwardenItem `json:"items"`
}

type bitwardenItem struct {
	Name  string          `json:"name"`
	Type  int             `json:"type"`
	Notes string          `json:"notes"`
	Login *bitwardenLogin `json:"login,omitempty"`
}

type bitwardenLogin struct {
	Username string         `json:"username"`
	Password string         `json:"password"`
	URIs     []bitwardenURI `json:"uris"`
}

type bitwardenURI struct {
	URI string `json:"uri"`
}

// ImportBitwarden imports from Bitwarden CSV or JSON format.
// It auto-detects the format by trying JSON first.
func ImportBitwarden(reader io.Reader) ([]types.Item, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	// Try JSON first.
	var bwJSON bitwardenJSON
	if err := json.Unmarshal(data, &bwJSON); err == nil && len(bwJSON.Items) > 0 {
		return importBitwardenJSON(bwJSON), nil
	}

	// Fall back to CSV: folder, favorite, type, name, notes, fields, reprompt, login_uri, login_username, login_password, login_totp
	return ImportCSV(strings.NewReader(string(data)), CSVMapping{
		Name:     3,
		URL:      7,
		Username: 8,
		Password: 9,
		Notes:    4,
		Type:     2,
		Skip:     1,
	})
}

func importBitwardenJSON(bw bitwardenJSON) []types.Item {
	var items []types.Item
	for _, bi := range bw.Items {
		item := types.Item{
			Type:   mapBitwardenType(bi.Type),
			Name:   bi.Name,
			Notes:  bi.Notes,
			Fields: make(map[string]string),
		}
		if bi.Login != nil {
			item.Fields[types.FieldUsername] = bi.Login.Username
			item.Fields[types.FieldPassword] = bi.Login.Password
			if len(bi.Login.URIs) > 0 {
				item.Fields[types.FieldURL] = bi.Login.URIs[0].URI
			}
		}
		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items
}

// --- helpers ---

func safeCol(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func mapItemType(s string) types.ItemType {
	switch strings.ToLower(s) {
	case "login", "1":
		return types.ItemTypeLogin
	case "note", "securenote", "2":
		return types.ItemTypeSecureNote
	case "creditcard", "credit card", "3":
		return types.ItemTypeCreditCard
	case "identity", "4":
		return types.ItemTypeIdentity
	case "apikey", "api key":
		return types.ItemTypeAPIKey
	case "sshkey", "ssh key":
		return types.ItemTypeSSHKey
	default:
		return types.ItemTypeLogin
	}
}

func mapBitwardenType(t int) types.ItemType {
	switch t {
	case 1:
		return types.ItemTypeLogin
	case 2:
		return types.ItemTypeSecureNote
	case 3:
		return types.ItemTypeCreditCard
	case 4:
		return types.ItemTypeIdentity
	default:
		return types.ItemTypeLogin
	}
}

// ErrEmptyInput is returned when the input is empty.
var ErrEmptyInput = errors.New("empty input")

// ImportSafari parses Safari CSV export: Title, URL, Username, Password, Notes, OTPAuth
func ImportSafari(reader io.Reader) ([]types.Item, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse Safari CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var items []types.Item
	for _, row := range records[1:] {
		item := types.Item{
			Type:   types.ItemTypeLogin,
			Fields: make(map[string]string),
		}

		item.Name = safeCol(row, 0)
		if url := safeCol(row, 1); url != "" {
			item.Fields[types.FieldURL] = url
		}
		if user := safeCol(row, 2); user != "" {
			item.Fields[types.FieldUsername] = user
		}
		if pw := safeCol(row, 3); pw != "" {
			item.Fields[types.FieldPassword] = pw
		}
		item.Notes = safeCol(row, 4)
		if otp := safeCol(row, 5); otp != "" {
			item.Fields[types.FieldTOTP] = otp
		}

		if item.Name == "" && item.Fields[types.FieldURL] != "" {
			item.Name = item.Fields[types.FieldURL]
		}
		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// ImportLastPass parses LastPass CSV export: url, username, password, totp, extra, name, grouping, fav
func ImportLastPass(reader io.Reader) ([]types.Item, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse LastPass CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var items []types.Item
	for _, row := range records[1:] {
		item := types.Item{
			Type:   types.ItemTypeLogin,
			Fields: make(map[string]string),
		}

		url := safeCol(row, 0)
		item.Name = safeCol(row, 5)
		if url != "" {
			item.Fields[types.FieldURL] = url
		}
		if user := safeCol(row, 1); user != "" {
			item.Fields[types.FieldUsername] = user
		}
		if pw := safeCol(row, 2); pw != "" {
			item.Fields[types.FieldPassword] = pw
		}
		if totp := safeCol(row, 3); totp != "" {
			item.Fields[types.FieldTOTP] = totp
		}
		item.Notes = safeCol(row, 4)

		// Tags from grouping (backslash-separated).
		if grouping := safeCol(row, 6); grouping != "" {
			item.Tags = strings.Split(grouping, "\\")
		}
		// Favorite.
		if safeCol(row, 7) == "1" {
			item.Favorite = true
		}
		// Secure note detection.
		if url == "http://sn" {
			item.Type = types.ItemTypeSecureNote
		}

		if item.Name == "" && item.Fields[types.FieldURL] != "" {
			item.Name = item.Fields[types.FieldURL]
		}
		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// ImportKeePass parses KeePass CSV export: Group, Title, Username, Password, URL, Notes
func ImportKeePass(reader io.Reader) ([]types.Item, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse KeePass CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var items []types.Item
	for _, row := range records[1:] {
		item := types.Item{
			Type:   types.ItemTypeLogin,
			Fields: make(map[string]string),
		}

		item.Name = safeCol(row, 1)
		if url := safeCol(row, 4); url != "" {
			item.Fields[types.FieldURL] = url
		}
		if user := safeCol(row, 2); user != "" {
			item.Fields[types.FieldUsername] = user
		}
		if pw := safeCol(row, 3); pw != "" {
			item.Fields[types.FieldPassword] = pw
		}
		item.Notes = safeCol(row, 5)

		// Tags from Group (slash-separated).
		if group := safeCol(row, 0); group != "" {
			item.Tags = strings.Split(group, "/")
		}

		if item.Name == "" && item.Fields[types.FieldURL] != "" {
			item.Name = item.Fields[types.FieldURL]
		}
		if item.Name != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// --- 1PUX JSON structures ---

type onePUXExport struct {
	Accounts []onePUXAccount `json:"accounts"`
}

type onePUXAccount struct {
	Vaults []onePUXVault `json:"vaults"`
}

type onePUXVault struct {
	Items []onePUXItemWrapper `json:"items"`
}

type onePUXItemWrapper struct {
	Item onePUXItem `json:"item"`
}

type onePUXItem struct {
	UUID     string         `json:"uuid"`
	TypeName string         `json:"typeName"`
	Title    string         `json:"title"`
	Overview onePUXOverview `json:"overview"`
	Details  onePUXDetails  `json:"details"`
}

type onePUXOverview struct {
	URL string `json:"url"`
}

type onePUXDetails struct {
	LoginFields []onePUXLoginField `json:"loginFields"`
	Sections    []onePUXSection    `json:"sections"`
	NotesPlain  string             `json:"notesPlain"`
}

type onePUXLoginField struct {
	Designation string `json:"designation"`
	Value       string `json:"value"`
}

type onePUXSection struct {
	Fields []onePUXSectionField `json:"fields"`
}

type onePUXSectionField struct {
	Title string             `json:"title"`
	Value onePUXSectionValue `json:"value"`
}

type onePUXSectionValue struct {
	String string `json:"string"`
}

// Import1PUX parses 1Password 1PUX export (export.data JSON).
func Import1PUX(reader io.Reader) ([]types.Item, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read 1PUX input: %w", err)
	}

	var export onePUXExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("parse 1PUX JSON: %w", err)
	}

	var items []types.Item
	for _, account := range export.Accounts {
		for _, vault := range account.Vaults {
			for _, wrapper := range vault.Items {
				src := wrapper.Item
				item := types.Item{
					ID:     src.UUID,
					Type:   map1PUXType(src.TypeName),
					Name:   src.Title,
					Fields: make(map[string]string),
					Notes:  src.Details.NotesPlain,
				}

				// URL from overview.
				if src.Overview.URL != "" {
					item.Fields[types.FieldURL] = src.Overview.URL
				}

				// Username/password from loginFields.
				for _, lf := range src.Details.LoginFields {
					switch lf.Designation {
					case "username":
						if lf.Value != "" {
							item.Fields[types.FieldUsername] = lf.Value
						}
					case "password":
						if lf.Value != "" {
							item.Fields[types.FieldPassword] = lf.Value
						}
					}
				}

				// Custom fields from sections.
				for _, section := range src.Details.Sections {
					for _, f := range section.Fields {
						if f.Title != "" && f.Value.String != "" {
							if item.CustomFields == nil {
								item.CustomFields = make(map[string]string)
							}
							item.CustomFields[f.Title] = f.Value.String
						}
					}
				}

				if item.Name != "" {
					items = append(items, item)
				}
			}
		}
	}
	return items, nil
}

func map1PUXType(typeName string) types.ItemType {
	switch typeName {
	case "001":
		return types.ItemTypeLogin
	case "002":
		return types.ItemTypeCreditCard
	case "003":
		return types.ItemTypeSecureNote
	case "004":
		return types.ItemTypeIdentity
	case "005":
		return types.ItemTypeLogin
	default:
		return types.ItemTypeLogin
	}
}
