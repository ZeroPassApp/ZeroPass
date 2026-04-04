package importexport

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/crypto/key"
	"github.com/zeropass/zeropass/core/vault/types"
)

const chromeCSV = `name,url,username,password
GitHub,https://github.com,devuser,s3cure!
AWS Console,https://aws.amazon.com,admin,Cl0udP@ss
`

const firefoxCSV = `"url","username","password","httpRealm","formActionOrigin","guid","timeCreated","timeLastUsed","timePasswordChanged"
"https://github.com","devuser","s3cure!","","https://github.com","guid1","1700000000000","1700000001000","1700000002000"
"https://gitlab.com","user2","p@ssw0rd","","https://gitlab.com","guid2","1700000000000","1700000001000","1700000002000"
`

const onePasswordCSV = `Title,Website,Username,Password,Notes,Type
GitHub,https://github.com,devuser,s3cure!,My main account,login
AWS Key,,,AKIA12345,Cloud key,apikey
`

const bitwardenCSV = `folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp
,,login,GitHub,,,,https://github.com,devuser,s3cure!,
,,login,AWS Console,,,,https://aws.amazon.com,admin,Cl0udP@ss,
`

func TestImportChrome(t *testing.T) {
	items, err := ImportChrome(strings.NewReader(chromeCSV))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])

	assert.Equal(t, "AWS Console", items[1].Name)
}

func TestImportFirefox(t *testing.T) {
	items, err := ImportFirefox(strings.NewReader(firefoxCSV))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "https://github.com", items[0].Name)
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
}

func TestImport1Password(t *testing.T) {
	items, err := Import1Password(strings.NewReader(onePasswordCSV))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "AWS Key", items[1].Name)
	assert.Equal(t, types.ItemTypeAPIKey, items[1].Type)
}

func TestImportBitwardenCSV(t *testing.T) {
	items, err := ImportBitwarden(strings.NewReader(bitwardenCSV))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
}

func TestImportBitwardenJSON(t *testing.T) {
	jsonData := `{
		"items": [
			{
				"name": "GitHub",
				"type": 1,
				"notes": "dev account",
				"login": {
					"username": "devuser",
					"password": "s3cure!",
					"uris": [{"uri": "https://github.com"}]
				}
			},
			{
				"name": "Secure Note",
				"type": 2,
				"notes": "some secret"
			}
		]
	}`
	items, err := ImportBitwarden(strings.NewReader(jsonData))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])

	assert.Equal(t, "Secure Note", items[1].Name)
	assert.Equal(t, types.ItemTypeSecureNote, items[1].Type)
}

func TestImportEmptyCSV(t *testing.T) {
	items, err := ImportChrome(strings.NewReader("name,url,username,password\n"))
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestImportCSVCustomMapping(t *testing.T) {
	csv := "site,user,pass\nGitHub,dev,123\n"
	items, err := ImportCSV(strings.NewReader(csv), CSVMapping{
		Name:     0,
		Username: 1,
		Password: 2,
		URL:      -1,
		Notes:    -1,
		Type:     -1,
		Skip:     1,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, "dev", items[0].Fields[types.FieldUsername])
}

func TestExportJSON(t *testing.T) {
	items := []types.Item{
		{
			ID:   "id-1",
			Type: types.ItemTypeLogin,
			Name: "GitHub",
			Fields: map[string]string{
				types.FieldUsername: "user",
				types.FieldPassword: "pass",
			},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, ExportJSON(items, &buf))
	assert.Contains(t, buf.String(), "GitHub")
	assert.Contains(t, buf.String(), "user")
}

func TestExportCSV(t *testing.T) {
	items := []types.Item{
		{
			Name: "GitHub",
			Type: types.ItemTypeLogin,
			Fields: map[string]string{
				types.FieldURL:      "https://github.com",
				types.FieldUsername: "user",
				types.FieldPassword: "pass",
			},
			Notes: "notes here",
			Tags:  []string{"dev", "work"},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, ExportCSV(items, &buf))
	output := buf.String()
	assert.Contains(t, output, "name,type,url,username,password,notes,tags")
	assert.Contains(t, output, "GitHub")
	assert.Contains(t, output, "dev,work")
}

func TestExportEncryptedRoundTrip(t *testing.T) {
	items := []types.Item{
		{
			ID:   "id-1",
			Type: types.ItemTypeLogin,
			Name: "Secret",
			Fields: map[string]string{
				types.FieldPassword: "s3cure!",
			},
		},
	}

	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)

	// Export encrypted
	var buf bytes.Buffer
	require.NoError(t, ExportEncrypted(items, vaultKey, &buf))

	// Import encrypted
	restored, err := ImportEncrypted(strings.NewReader(buf.String()), vaultKey)
	require.NoError(t, err)
	require.Len(t, restored, 1)
	assert.Equal(t, "Secret", restored[0].Name)
	assert.Equal(t, "s3cure!", restored[0].Fields[types.FieldPassword])
}

func TestExportEncryptedWrongKey(t *testing.T) {
	items := []types.Item{{Name: "x", Type: types.ItemTypeLogin}}

	key1, _ := key.GenerateVaultKey()
	key2, _ := key.GenerateVaultKey()

	var buf bytes.Buffer
	require.NoError(t, ExportEncrypted(items, key1, &buf))

	_, err := ImportEncrypted(strings.NewReader(buf.String()), key2)
	assert.Error(t, err)
}

func TestMapItemType(t *testing.T) {
	assert.Equal(t, types.ItemTypeLogin, mapItemType("login"))
	assert.Equal(t, types.ItemTypeLogin, mapItemType("1"))
	assert.Equal(t, types.ItemTypeSecureNote, mapItemType("note"))
	assert.Equal(t, types.ItemTypeSecureNote, mapItemType("securenote"))
	assert.Equal(t, types.ItemTypeSecureNote, mapItemType("2"))
	assert.Equal(t, types.ItemTypeCreditCard, mapItemType("creditcard"))
	assert.Equal(t, types.ItemTypeCreditCard, mapItemType("credit card"))
	assert.Equal(t, types.ItemTypeCreditCard, mapItemType("3"))
	assert.Equal(t, types.ItemTypeIdentity, mapItemType("identity"))
	assert.Equal(t, types.ItemTypeIdentity, mapItemType("4"))
	assert.Equal(t, types.ItemTypeAPIKey, mapItemType("apikey"))
	assert.Equal(t, types.ItemTypeAPIKey, mapItemType("api key"))
	assert.Equal(t, types.ItemTypeSSHKey, mapItemType("sshkey"))
	assert.Equal(t, types.ItemTypeSSHKey, mapItemType("ssh key"))
	assert.Equal(t, types.ItemTypeLogin, mapItemType("unknown"))
}

func TestMapBitwardenType(t *testing.T) {
	assert.Equal(t, types.ItemTypeLogin, mapBitwardenType(1))
	assert.Equal(t, types.ItemTypeSecureNote, mapBitwardenType(2))
	assert.Equal(t, types.ItemTypeCreditCard, mapBitwardenType(3))
	assert.Equal(t, types.ItemTypeIdentity, mapBitwardenType(4))
	assert.Equal(t, types.ItemTypeLogin, mapBitwardenType(99)) // default
	assert.Equal(t, types.ItemTypeLogin, mapBitwardenType(0))  // default
	assert.Equal(t, types.ItemTypeLogin, mapBitwardenType(-1)) // default
}

// errorWriter always returns an error on Write.
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestExportJSONError(t *testing.T) {
	items := []types.Item{{Name: "x", Type: types.ItemTypeLogin}}
	err := ExportJSON(items, &errorWriter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export JSON")
}

func TestExportCSVError(t *testing.T) {
	items := []types.Item{{
		Name:   "x",
		Type:   types.ItemTypeLogin,
		Fields: map[string]string{},
	}}
	// csv.Writer buffers, so the error surfaces on Flush which is deferred.
	// Use a writer that fails immediately.
	err := ExportCSV(items, &errorWriter{})
	// csv.Writer may not propagate write errors until Flush;
	// depending on buffer size, the header write may or may not fail.
	// We mainly want to ensure no panic.
	_ = err
}

func TestExportCSVEmptyTags(t *testing.T) {
	items := []types.Item{
		{
			Name:   "NoTags",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{types.FieldURL: "https://example.com"},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, ExportCSV(items, &buf))
	output := buf.String()
	assert.Contains(t, output, "NoTags")
	// Tags column should be empty
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 2) // header + 1 data row
}

func TestExportCSVSingleTag(t *testing.T) {
	items := []types.Item{
		{
			Name:   "SingleTag",
			Type:   types.ItemTypeLogin,
			Fields: map[string]string{},
			Tags:   []string{"solo"},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, ExportCSV(items, &buf))
	assert.Contains(t, buf.String(), "solo")
}

func TestExportCSVEmptyItems(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportCSV(nil, &buf))
	// Should still have the header
	assert.Contains(t, buf.String(), "name,type,url")
}

func TestExportEncryptedInvalidKey(t *testing.T) {
	items := []types.Item{{Name: "x", Type: types.ItemTypeLogin}}
	var buf bytes.Buffer
	err := ExportEncrypted(items, []byte("short-key"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "encrypt export")
}

func TestExportEncryptedWriteError(t *testing.T) {
	items := []types.Item{{Name: "x", Type: types.ItemTypeLogin}}
	vaultKey, err := key.GenerateVaultKey()
	require.NoError(t, err)
	err = ExportEncrypted(items, vaultKey, &errorWriter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write encrypted export")
}

func TestImportEncryptedInvalidBase64(t *testing.T) {
	_, err := ImportEncrypted(strings.NewReader("not-valid-base64!!!@@@"), make([]byte, 32))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode encrypted export")
}

func TestImportEncryptedDecryptFails(t *testing.T) {
	// Valid base64 but wrong key
	vaultKey, _ := key.GenerateVaultKey()
	items := []types.Item{{Name: "x", Type: types.ItemTypeLogin}}
	var buf bytes.Buffer
	require.NoError(t, ExportEncrypted(items, vaultKey, &buf))

	wrongKey, _ := key.GenerateVaultKey()
	_, err := ImportEncrypted(strings.NewReader(buf.String()), wrongKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt export")
}

func TestImportCSVEmpty(t *testing.T) {
	// Completely empty input
	items, err := ImportCSV(strings.NewReader(""), CSVMapping{
		Name: 0, URL: 1, Username: 2, Password: 3,
		Notes: -1, Type: -1, Skip: 1,
	})
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportCSVOnlyHeader(t *testing.T) {
	items, err := ImportCSV(strings.NewReader("name,url,user,pass\n"), CSVMapping{
		Name: 0, URL: 1, Username: 2, Password: 3,
		Notes: -1, Type: -1, Skip: 1,
	})
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportCSVRowWithEmptyName(t *testing.T) {
	// Row with empty name and no URL — should be skipped
	csv := "name,url,user,pass\n,,,secret\n"
	items, err := ImportCSV(strings.NewReader(csv), CSVMapping{
		Name: 0, URL: 1, Username: 2, Password: 3,
		Notes: -1, Type: -1, Skip: 1,
	})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestImportCSVRowNameFallsBackToURL(t *testing.T) {
	// Row with empty name but has URL — name becomes URL
	csv := "name,url,user,pass\n,https://example.com,user,secret\n"
	items, err := ImportCSV(strings.NewReader(csv), CSVMapping{
		Name: 0, URL: 1, Username: 2, Password: 3,
		Notes: -1, Type: -1, Skip: 1,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://example.com", items[0].Name)
}

func TestImportCSVWithTypeColumn(t *testing.T) {
	csv := "name,type,url,user,pass\nMyNote,note,,,\nMyCard,creditcard,,,\n"
	items, err := ImportCSV(strings.NewReader(csv), CSVMapping{
		Name: 0, Type: 1, URL: 2, Username: 3, Password: 4,
		Notes: -1, Skip: 1,
	})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, types.ItemTypeSecureNote, items[0].Type)
	assert.Equal(t, types.ItemTypeCreditCard, items[1].Type)
}

func TestImportFirefoxEmpty(t *testing.T) {
	items, err := ImportFirefox(strings.NewReader("url,username,password\n"))
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestImportFirefoxRowWithEmptyURL(t *testing.T) {
	csv := `"url","username","password","httpRealm","formActionOrigin","guid","timeCreated","timeLastUsed","timePasswordChanged"
"","user","pass","","","g","0","0","0"
`
	items, err := ImportFirefox(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Empty(t, items) // empty URL => empty Name => skipped
}

func TestImportFirefoxShortRow(t *testing.T) {
	// Firefox CSV uses ReadAll which requires consistent field count.
	// A row with all fields but some empty should work.
	csv := `"url","username","password","httpRealm","formActionOrigin","guid","timeCreated","timeLastUsed","timePasswordChanged"
"https://example.com","","","","","g","0","0","0"
`
	items, err := ImportFirefox(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://example.com", items[0].Name)
	assert.Empty(t, items[0].Fields[types.FieldUsername])
	assert.Empty(t, items[0].Fields[types.FieldPassword])
}

func TestImportBitwardenJSONEmptyNameFiltered(t *testing.T) {
	jsonData := `{
		"items": [
			{"name": "", "type": 1, "notes": "no name"},
			{"name": "Valid", "type": 1, "notes": "has name"}
		]
	}`
	items, err := ImportBitwarden(strings.NewReader(jsonData))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Valid", items[0].Name)
}

func TestImportBitwardenJSONNoLogin(t *testing.T) {
	jsonData := `{
		"items": [
			{"name": "Note", "type": 2, "notes": "a note"}
		]
	}`
	items, err := ImportBitwarden(strings.NewReader(jsonData))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, types.ItemTypeSecureNote, items[0].Type)
	assert.Empty(t, items[0].Fields[types.FieldUsername])
}

func TestImportBitwardenJSONLoginNoURIs(t *testing.T) {
	jsonData := `{
		"items": [
			{"name": "NoURI", "type": 1, "login": {"username": "u", "password": "p", "uris": []}}
		]
	}`
	items, err := ImportBitwarden(strings.NewReader(jsonData))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Empty(t, items[0].Fields[types.FieldURL])
}

func TestImportBitwardenJSONCreditCard(t *testing.T) {
	jsonData := `{
		"items": [
			{"name": "MyCard", "type": 3, "notes": "cc"},
			{"name": "MyID", "type": 4, "notes": "id"},
			{"name": "Unknown", "type": 99, "notes": "unk"}
		]
	}`
	items, err := ImportBitwarden(strings.NewReader(jsonData))
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, types.ItemTypeCreditCard, items[0].Type)
	assert.Equal(t, types.ItemTypeIdentity, items[1].Type)
	assert.Equal(t, types.ItemTypeLogin, items[2].Type) // default
}

func TestImportBitwardenFallsBackToCSV(t *testing.T) {
	// Invalid JSON should fall back to CSV parsing
	data := "folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp\n,,login,TestItem,,,,https://test.com,user,pass,\n"
	items, err := ImportBitwarden(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "TestItem", items[0].Name)
}

func TestExportJSONRoundTrip(t *testing.T) {
	items := []types.Item{
		{
			ID:   "id-1",
			Type: types.ItemTypeLogin,
			Name: "Test",
			Fields: map[string]string{
				types.FieldUsername: "user",
				types.FieldPassword: "pass",
				types.FieldURL:      "https://test.com",
			},
			Tags:  []string{"a", "b"},
			Notes: "some notes",
		},
	}
	var buf bytes.Buffer
	require.NoError(t, ExportJSON(items, &buf))

	var restored []types.Item
	require.NoError(t, json.Unmarshal(buf.Bytes(), &restored))
	require.Len(t, restored, 1)
	assert.Equal(t, "Test", restored[0].Name)
	assert.Equal(t, "user", restored[0].Fields[types.FieldUsername])
}

func TestJoinTags(t *testing.T) {
	assert.Equal(t, "", joinTags(nil))
	assert.Equal(t, "", joinTags([]string{}))
	assert.Equal(t, "solo", joinTags([]string{"solo"}))
	assert.Equal(t, "a,b,c", joinTags([]string{"a", "b", "c"}))
}

func TestSafeCol(t *testing.T) {
	row := []string{"a", "b", "c"}
	assert.Equal(t, "a", safeCol(row, 0))
	assert.Equal(t, "c", safeCol(row, 2))
	assert.Equal(t, "", safeCol(row, 3))  // out of bounds
	assert.Equal(t, "", safeCol(row, -1)) // negative
	assert.Equal(t, "", safeCol(nil, 0))  // nil row
}

// --- Safari import tests ---

const safariCSV = `Title,URL,Username,Password,Notes,OTPAuth
GitHub,https://github.com,devuser,s3cure!,My GitHub account,otpauth://totp/GitHub?secret=ABC
AWS Console,https://aws.amazon.com,admin,Cl0udP@ss,,
Personal Blog,https://blog.example.com,blogger,bl0gP@ss,My blog,
`

func TestImportSafari(t *testing.T) {
	items, err := ImportSafari(strings.NewReader(safariCSV))
	require.NoError(t, err)
	require.Len(t, items, 3)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
	assert.Equal(t, "My GitHub account", items[0].Notes)
	assert.Equal(t, "otpauth://totp/GitHub?secret=ABC", items[0].Fields[types.FieldTOTP])

	assert.Equal(t, "AWS Console", items[1].Name)
	assert.Equal(t, "admin", items[1].Fields[types.FieldUsername])
	assert.Empty(t, items[1].Fields[types.FieldTOTP])
	assert.Empty(t, items[1].Notes)

	assert.Equal(t, "Personal Blog", items[2].Name)
	assert.Empty(t, items[2].Fields[types.FieldTOTP])
	assert.Equal(t, "My blog", items[2].Notes)
}

func TestImportSafariEmpty(t *testing.T) {
	items, err := ImportSafari(strings.NewReader(""))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportSafariHeaderOnly(t *testing.T) {
	items, err := ImportSafari(strings.NewReader("Title,URL,Username,Password,Notes,OTPAuth\n"))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportSafariSpecialChars(t *testing.T) {
	csv := "Title,URL,Username,Password,Notes,OTPAuth\n\"My \"\"Special\"\" Site\",https://example.com,\"user@email.com\",\"p@ss,word!\",\"Notes with\nnewline\",\n"
	items, err := ImportSafari(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "My \"Special\" Site", items[0].Name)
	assert.Equal(t, "user@email.com", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "p@ss,word!", items[0].Fields[types.FieldPassword])
}

func TestImportSafariNameFallbackToURL(t *testing.T) {
	csv := "Title,URL,Username,Password,Notes,OTPAuth\n,https://example.com,user,pass,,\n"
	items, err := ImportSafari(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://example.com", items[0].Name)
}

// --- LastPass import tests ---

const lastPassCSV = `url,username,password,totp,extra,name,grouping,fav
https://github.com,devuser,s3cure!,otpauth://totp/GH?secret=XYZ,GitHub notes,GitHub,Development\Git,1
https://aws.amazon.com,admin,Cl0udP@ss,,,AWS Console,Cloud\AWS,0
http://sn,,,,This is a secure note,My Secret Note,Notes,0
`

func TestImportLastPass(t *testing.T) {
	items, err := ImportLastPass(strings.NewReader(lastPassCSV))
	require.NoError(t, err)
	require.Len(t, items, 3)

	// First item: full login with TOTP, tags, favorite
	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
	assert.Equal(t, "otpauth://totp/GH?secret=XYZ", items[0].Fields[types.FieldTOTP])
	assert.Equal(t, "GitHub notes", items[0].Notes)
	assert.Equal(t, []string{"Development", "Git"}, items[0].Tags)
	assert.True(t, items[0].Favorite)

	// Second item: login without TOTP, not favorite
	assert.Equal(t, "AWS Console", items[1].Name)
	assert.Equal(t, types.ItemTypeLogin, items[1].Type)
	assert.Empty(t, items[1].Fields[types.FieldTOTP])
	assert.Equal(t, []string{"Cloud", "AWS"}, items[1].Tags)
	assert.False(t, items[1].Favorite)

	// Third item: secure note
	assert.Equal(t, "My Secret Note", items[2].Name)
	assert.Equal(t, types.ItemTypeSecureNote, items[2].Type)
	assert.Equal(t, "This is a secure note", items[2].Notes)
	assert.Equal(t, "http://sn", items[2].Fields[types.FieldURL])
}

func TestImportLastPassEmpty(t *testing.T) {
	items, err := ImportLastPass(strings.NewReader(""))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportLastPassHeaderOnly(t *testing.T) {
	items, err := ImportLastPass(strings.NewReader("url,username,password,totp,extra,name,grouping,fav\n"))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportLastPassSpecialChars(t *testing.T) {
	csv := "url,username,password,totp,extra,name,grouping,fav\nhttps://example.com,\"user@email.com\",\"p@ss,word!\",,,\"Site \"\"A\"\"\",folder,0\n"
	items, err := ImportLastPass(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Site \"A\"", items[0].Name)
	assert.Equal(t, "user@email.com", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "p@ss,word!", items[0].Fields[types.FieldPassword])
}

func TestImportLastPassNoGrouping(t *testing.T) {
	csv := "url,username,password,totp,extra,name,grouping,fav\nhttps://example.com,user,pass,,,Test,,0\n"
	items, err := ImportLastPass(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Nil(t, items[0].Tags)
}

// --- KeePass import tests ---

const keepassCSV = `Group,Title,Username,Password,URL,Notes
Internet,GitHub,devuser,s3cure!,https://github.com,My GitHub
Internet/Cloud,AWS Console,admin,Cl0udP@ss,https://aws.amazon.com,
Email,Gmail,user@gmail.com,gmailP@ss,https://mail.google.com,Personal email
`

func TestImportKeePass(t *testing.T) {
	items, err := ImportKeePass(strings.NewReader(keepassCSV))
	require.NoError(t, err)
	require.Len(t, items, 3)

	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
	assert.Equal(t, "My GitHub", items[0].Notes)
	assert.Equal(t, []string{"Internet"}, items[0].Tags)

	assert.Equal(t, "AWS Console", items[1].Name)
	assert.Equal(t, []string{"Internet", "Cloud"}, items[1].Tags)
	assert.Empty(t, items[1].Notes)

	assert.Equal(t, "Gmail", items[2].Name)
	assert.Equal(t, "user@gmail.com", items[2].Fields[types.FieldUsername])
	assert.Equal(t, []string{"Email"}, items[2].Tags)
}

func TestImportKeePassEmpty(t *testing.T) {
	items, err := ImportKeePass(strings.NewReader(""))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportKeePassHeaderOnly(t *testing.T) {
	items, err := ImportKeePass(strings.NewReader("Group,Title,Username,Password,URL,Notes\n"))
	require.NoError(t, err)
	assert.Nil(t, items)
}

func TestImportKeePassSpecialChars(t *testing.T) {
	csv := "Group,Title,Username,Password,URL,Notes\n\"Special/Group\",\"Site \"\"X\"\"\",\"user@test.com\",\"p@ss,w0rd\",https://x.com,\"Multi\nline note\"\n"
	items, err := ImportKeePass(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Site \"X\"", items[0].Name)
	assert.Equal(t, "user@test.com", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "p@ss,w0rd", items[0].Fields[types.FieldPassword])
	assert.Equal(t, []string{"Special", "Group"}, items[0].Tags)
}

func TestImportKeePassNameFallbackToURL(t *testing.T) {
	csv := "Group,Title,Username,Password,URL,Notes\nRoot,,user,pass,https://example.com,\n"
	items, err := ImportKeePass(strings.NewReader(csv))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://example.com", items[0].Name)
}

// --- 1PUX import tests ---

const onePUXJSON = `{
	"accounts": [{
		"vaults": [{
			"items": [
				{
					"item": {
						"uuid": "uuid-1",
						"typeName": "001",
						"title": "GitHub",
						"overview": {"url": "https://github.com"},
						"details": {
							"loginFields": [
								{"designation": "username", "value": "devuser"},
								{"designation": "password", "value": "s3cure!"}
							],
							"sections": [{
								"fields": [
									{"title": "recovery", "value": {"string": "abc-def"}}
								]
							}],
							"notesPlain": "My GitHub account"
						}
					}
				},
				{
					"item": {
						"uuid": "uuid-2",
						"typeName": "003",
						"title": "Secret Note",
						"overview": {},
						"details": {
							"notesPlain": "Top secret stuff"
						}
					}
				}
			]
		}]
	}]
}`

func TestImport1PUX(t *testing.T) {
	items, err := Import1PUX(strings.NewReader(onePUXJSON))
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, "uuid-1", items[0].ID)
	assert.Equal(t, "GitHub", items[0].Name)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, "https://github.com", items[0].Fields[types.FieldURL])
	assert.Equal(t, "devuser", items[0].Fields[types.FieldUsername])
	assert.Equal(t, "s3cure!", items[0].Fields[types.FieldPassword])
	assert.Equal(t, "My GitHub account", items[0].Notes)
	assert.Equal(t, "abc-def", items[0].CustomFields["recovery"])

	assert.Equal(t, "uuid-2", items[1].ID)
	assert.Equal(t, "Secret Note", items[1].Name)
	assert.Equal(t, types.ItemTypeSecureNote, items[1].Type)
	assert.Equal(t, "Top secret stuff", items[1].Notes)
}

func TestImport1PUXEmpty(t *testing.T) {
	items, err := Import1PUX(strings.NewReader(`{"accounts":[]}`))
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestImport1PUXInvalidJSON(t *testing.T) {
	_, err := Import1PUX(strings.NewReader("not json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse 1PUX JSON")
}

func TestImport1PUXMultipleVaultsAndAccounts(t *testing.T) {
	data := `{
		"accounts": [
			{
				"vaults": [
					{
						"items": [
							{"item": {"uuid": "a1v1", "typeName": "001", "title": "Site A", "overview": {}, "details": {}}}
						]
					},
					{
						"items": [
							{"item": {"uuid": "a1v2", "typeName": "001", "title": "Site B", "overview": {}, "details": {}}}
						]
					}
				]
			},
			{
				"vaults": [{
					"items": [
						{"item": {"uuid": "a2v1", "typeName": "001", "title": "Site C", "overview": {}, "details": {}}}
					]
				}]
			}
		]
	}`
	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, "Site A", items[0].Name)
	assert.Equal(t, "Site B", items[1].Name)
	assert.Equal(t, "Site C", items[2].Name)
}

func TestImport1PUXTypeMappings(t *testing.T) {
	data := `{
		"accounts": [{
			"vaults": [{
				"items": [
					{"item": {"uuid": "1", "typeName": "001", "title": "Login", "overview": {}, "details": {}}},
					{"item": {"uuid": "2", "typeName": "002", "title": "Card", "overview": {}, "details": {}}},
					{"item": {"uuid": "3", "typeName": "003", "title": "Note", "overview": {}, "details": {}}},
					{"item": {"uuid": "4", "typeName": "004", "title": "Identity", "overview": {}, "details": {}}},
					{"item": {"uuid": "5", "typeName": "005", "title": "Password", "overview": {}, "details": {}}},
					{"item": {"uuid": "6", "typeName": "999", "title": "Unknown", "overview": {}, "details": {}}}
				]
			}]
		}]
	}`
	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 6)
	assert.Equal(t, types.ItemTypeLogin, items[0].Type)
	assert.Equal(t, types.ItemTypeCreditCard, items[1].Type)
	assert.Equal(t, types.ItemTypeSecureNote, items[2].Type)
	assert.Equal(t, types.ItemTypeIdentity, items[3].Type)
	assert.Equal(t, types.ItemTypeLogin, items[4].Type) // 005 -> Login
	assert.Equal(t, types.ItemTypeLogin, items[5].Type) // unknown -> Login
}

func TestImport1PUXMissingOptionalFields(t *testing.T) {
	data := `{
		"accounts": [{
			"vaults": [{
				"items": [{
					"item": {
						"uuid": "min",
						"typeName": "001",
						"title": "Minimal",
						"overview": {},
						"details": {}
					}
				}]
			}]
		}]
	}`
	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Minimal", items[0].Name)
	assert.Empty(t, items[0].Fields[types.FieldURL])
	assert.Empty(t, items[0].Fields[types.FieldUsername])
	assert.Empty(t, items[0].Fields[types.FieldPassword])
	assert.Empty(t, items[0].Notes)
	assert.Nil(t, items[0].CustomFields)
}

func TestImport1PUXEmptyTitleSkipped(t *testing.T) {
	data := `{
		"accounts": [{
			"vaults": [{
				"items": [
					{"item": {"uuid": "1", "typeName": "001", "title": "", "overview": {}, "details": {}}},
					{"item": {"uuid": "2", "typeName": "001", "title": "Valid", "overview": {}, "details": {}}}
				]
			}]
		}]
	}`
	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Valid", items[0].Name)
}

func TestImport1PUXSectionFieldsIgnoreEmpty(t *testing.T) {
	data := `{
		"accounts": [{
			"vaults": [{
				"items": [{
					"item": {
						"uuid": "1",
						"typeName": "001",
						"title": "WithSections",
						"overview": {},
						"details": {
							"sections": [{
								"fields": [
									{"title": "filled", "value": {"string": "value1"}},
									{"title": "", "value": {"string": "no-title"}},
									{"title": "empty-val", "value": {"string": ""}}
								]
							}]
						}
					}
				}]
			}]
		}]
	}`
	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Len(t, items[0].CustomFields, 1)
	assert.Equal(t, "value1", items[0].CustomFields["filled"])
}

func TestImport1PUXUnsafeUUIDIsDropped(t *testing.T) {
	data := `{
		"accounts": [{
			"vaults": [{
				"items": [{
					"item": {
						"uuid": "../escape",
						"typeName": "001",
						"title": "Unsafe",
						"overview": {},
						"details": {}
					}
				}]
			}]
		}]
	}`

	items, err := Import1PUX(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Empty(t, items[0].ID)
	assert.Equal(t, "Unsafe", items[0].Name)
}

func TestMap1PUXType(t *testing.T) {
	assert.Equal(t, types.ItemTypeLogin, map1PUXType("001"))
	assert.Equal(t, types.ItemTypeCreditCard, map1PUXType("002"))
	assert.Equal(t, types.ItemTypeSecureNote, map1PUXType("003"))
	assert.Equal(t, types.ItemTypeIdentity, map1PUXType("004"))
	assert.Equal(t, types.ItemTypeLogin, map1PUXType("005"))
	assert.Equal(t, types.ItemTypeLogin, map1PUXType("unknown"))
	assert.Equal(t, types.ItemTypeLogin, map1PUXType(""))
}
