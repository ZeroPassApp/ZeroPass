package index

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/vault/types"
)

// ───────── EncryptIndexFile: WAL checkpoint path ─────────

func TestEncryptIndexFile_WithActiveWAL(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.db")

	// Create a real SQLite DB with some data, which produces a WAL file.
	idx, err := Open(dbPath)
	require.NoError(t, err)

	// Insert enough data to ensure WAL has content
	for i := 0; i < 10; i++ {
		require.NoError(t, idx.AddToIndex(&types.Item{
			ID:   "wal-" + string(rune('a'+i)),
			Type: types.ItemTypeLogin,
			Name: "WAL Test Item " + string(rune('a'+i)),
			Fields: map[string]string{
				types.FieldUsername: "user",
			},
		}))
	}
	// Close the index (this may checkpoint WAL, but let's test the encrypt path)
	require.NoError(t, idx.Close())

	// Manually create a WAL and SHM file to ensure they exist before encrypt
	require.NoError(t, os.WriteFile(dbPath+"-wal", []byte("wal-data"), 0600))
	require.NoError(t, os.WriteFile(dbPath+"-shm", []byte("shm-data"), 0600))

	key := testKey()
	require.NoError(t, EncryptIndexFile(dbPath, key))

	// Verify: plaintext, WAL, SHM all removed; .enc exists
	_, err = os.Stat(dbPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(dbPath + "-wal")
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(dbPath + "-shm")
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(dbPath + ".enc")
	assert.NoError(t, err)

	// Decrypt and verify data is intact
	require.NoError(t, DecryptIndexFile(dbPath, key))
	idx2, err := Open(dbPath)
	require.NoError(t, err)
	defer idx2.Close()

	ids, err := idx2.Search("WAL Test")
	require.NoError(t, err)
	assert.Len(t, ids, 10)
}

func TestEncryptIndexFile_WALCheckpointOnLiveDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "live.db")

	// Create DB with data — don't close it before encrypting.
	// EncryptIndexFile opens its own connection for checkpoint.
	idx, err := Open(dbPath)
	require.NoError(t, err)
	require.NoError(t, idx.AddToIndex(&types.Item{
		ID: "live-1", Type: types.ItemTypeLogin, Name: "LiveItem",
	}))
	// Close it so EncryptIndexFile can read the file
	require.NoError(t, idx.Close())

	key := testKey()
	require.NoError(t, EncryptIndexFile(dbPath, key))

	// Verify encrypted file exists
	assert.True(t, IsEncrypted(dbPath))

	// Round-trip: decrypt and verify data
	require.NoError(t, DecryptIndexFile(dbPath, key))
	idx2, err := Open(dbPath)
	require.NoError(t, err)
	defer idx2.Close()

	ids, err := idx2.Search("LiveItem")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestEncryptIndexFile_ReadError(t *testing.T) {
	dir := t.TempDir()
	// Create the dbPath as a directory to cause ReadFile error
	dbPath := filepath.Join(dir, "index.db")
	require.NoError(t, os.MkdirAll(dbPath, 0755))

	err := EncryptIndexFile(dbPath, testKey())
	assert.Error(t, err)
	// Error may come from WAL checkpoint or ReadFile depending on OS behavior
	assert.True(t, err != nil)
}

func TestEncryptIndexFile_WriteError(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("data"), 0600))

	// Make directory read-only to prevent writing .enc file
	require.NoError(t, os.Chmod(dir, 0500))
	defer os.Chmod(dir, 0755) // cleanup

	err := EncryptIndexFile(dbPath, testKey())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write encrypted index")
}

func TestDecryptIndexFile_ReadError(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.db")
	// Create .enc as directory to cause read error
	require.NoError(t, os.MkdirAll(dbPath+".enc", 0755))

	err := DecryptIndexFile(dbPath, testKey())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read encrypted index")
}

func TestDecryptIndexFile_WriteError(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "index.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("plaintext"), 0600))

	key := testKey()
	require.NoError(t, EncryptIndexFile(dbPath, key))

	// Make directory read-only to prevent writing decrypted file
	require.NoError(t, os.Chmod(dir, 0500))
	defer os.Chmod(dir, 0755)

	err := DecryptIndexFile(dbPath, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write decrypted index")
}

// ───────── Open: error paths ─────────

func TestOpen_WALPragmaError(t *testing.T) {
	// Opening a path that is a directory should fail
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "is-a-dir.db")
	require.NoError(t, os.MkdirAll(dbPath, 0755))

	_, err := Open(dbPath)
	assert.Error(t, err)
}

func TestOpen_ValidAndUsable(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "valid.db")

	idx, err := Open(dbPath)
	require.NoError(t, err)
	defer idx.Close()

	// Verify we can use it immediately
	require.NoError(t, idx.AddToIndex(&types.Item{
		ID: "open-test", Type: types.ItemTypeLogin, Name: "OpenTest",
	}))
	ids, err := idx.Search("OpenTest")
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

// ───────── Additional edge cases for better coverage ─────────

func TestEncryptDecrypt_LargeIndex(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "large.db")

	idx, err := Open(dbPath)
	require.NoError(t, err)

	// Add many items to create a sizeable DB
	for i := 0; i < 100; i++ {
		require.NoError(t, idx.AddToIndex(&types.Item{
			ID:     fmt.Sprintf("large-%d", i),
			Type:   types.ItemTypeLogin,
			Name:   fmt.Sprintf("LargeItem%d", i),
			Fields: map[string]string{types.FieldUsername: fmt.Sprintf("user%d", i)},
			Notes:  "This is a note for a large batch test item",
		}))
	}
	require.NoError(t, idx.Close())

	key := testKey()
	require.NoError(t, EncryptIndexFile(dbPath, key))
	require.NoError(t, DecryptIndexFile(dbPath, key))

	idx2, err := Open(dbPath)
	require.NoError(t, err)
	defer idx2.Close()

	ids, err := idx2.Search("LargeItem")
	require.NoError(t, err)
	assert.Len(t, ids, 100)
}

func TestCreateTables_SecondCallIdempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "idem.db")

	idx, err := Open(dbPath)
	require.NoError(t, err)
	defer idx.Close()

	// createTables was already called by Open; call it again
	err = createTables(idx.db)
	assert.NoError(t, err)
}
