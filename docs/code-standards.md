# ZeroPass: Code Standards & Conventions

This document defines the coding standards, patterns, and best practices observed throughout the ZeroPass codebase.

---

## Module & Naming

### Go Module
```
github.com/zeropass/zeropass
```

All imports reference this root module.

### Package Naming

1. **Lowercase, no underscores** — `crypto`, `vault`, `sync`, not `crypto_utils` or `CryptoUtils`
2. **Short and specific** — `cipher`, `kdf`, `key`, not `encryption` or `crypto_operations`
3. **Avoid generic names** — Use `health`, `importexport`, not `utils` or `helpers`
4. **Group related concerns** — `crypto/cipher/`, `vault/item/`, `core/sync/`

**Examples:**
```
core/crypto/kdf      ✅ (specific, hierarchical)
core/crypto/encoding ✅ (groups related encoding types)
core/helpers         ❌ (too generic)
core/crypto_utils    ❌ (underscores in package name)
```

### File Naming

1. **Snake_case for multi-word files** — `aes_gcm.go`, `master_key.go`, not `aesGCM.go`
2. **Descriptive names** — `password_strength.go` (not `strength.go`), `sync_handler.go`
3. **Test files** — `{filename}_test.go` in same package (not separate `tests/` directory)
4. **Avoid generic prefixes** — `kdf/argon2id.go` (not `kdf/kdf_impl.go`)

**Examples:**
```
crypto/cipher/aes_gcm.go        ✅ (clear, snake_case)
crypto/kdf/argon2id.go          ✅ (specific algorithm named)
crypto/encoding/base64.go       ✅ (descriptive)
vault/item_manager.go           ❌ (too generic, use context in package)
vault/item/item.go              ✅ (package context makes it clear)
```

---

## Type Definitions & Interfaces

### Type Naming

```go
// Domain types: struct name = domain + specific concept
type Item struct { }              // ✅ (context: vault items)
type EncryptedItem struct { }     // ✅ (clear what's encrypted)
type MasterKey struct { }         // ✅ (clear purpose)

type Config struct { }            // ❌ (too generic without context)
type Data struct { }              // ❌ (meaningless)
```

### Interface Definitions

```go
// Placed in interfaces.go for each major module
// Exported, descriptive names
package crypto

type KDF interface {
  Derive(password string, salt []byte) ([32]byte, error)
}

type Cipher interface {
  Encrypt(data, key []byte) ([]byte, error)
  Decrypt(data, key []byte) ([]byte, error)
}

// Package implementations are internal/concrete
type argon2idKDF struct { }

func (a *argon2idKDF) Derive(...) { }
```

**Rationale:** Interfaces define contracts; concrete types are unexported to enforce dependency injection.

---

## Function & Method Naming

### Exported Functions

```go
// Verbs for action functions
func Encrypt(data, key []byte) ([]byte, error)  // ✅ verb first
func Decrypt(...)                               // ✅
func NewItem(...) *Item                         // ✅ constructor

// Descriptive names for queries/getters
func FindItemByName(name string) (*Item, error) // ✅
func CalculatePasswordStrength(pwd string) int  // ✅

// Generic getters: Get{PropertyName}
func GetItemID() string                         // ✅
func GetCreatedAt() time.Time                   // ✅

// Boolean predicates: Is{Adjective}, Has{Property}
func IsExpired() bool                           // ✅
func HasPassword() bool                         // ✅
```

### Receiver Names

```go
// Single-character, standard receivers
func (v *Vault) AddItem(...) error      // ✅ (v for Vault)
func (i *Item) Encrypt(...) error       // ✅ (i for Item)
func (c *Cipher) Lock(...) error        // ✅ (c for Cipher)

// Never 'receiver' or 'this'
func (receiver *Vault) AddItem(...) error  // ❌
func (this *Item) Encrypt(...) error       // ❌
```

---

## Comments & Documentation

### Exported Types

```go
// Item represents a credential stored in the vault.
// All sensitive fields (Password, PrivateKey, etc.) are encrypted at rest.
type Item struct {
  ID        string    // Unique identifier (UUID)
  Type      ItemType  // Item type (login, apikey, etc.)
  Name      string    // Display name (encrypted)
  CreatedAt time.Time // Creation timestamp
}
```

### Exported Functions

```go
// Unlock decrypts the vault using the provided password.
// Returns an error if the password is incorrect or vault is corrupted.
func (v *Vault) Unlock(password string) error {
  // ...
}

// GeneratePassword creates a random password with specified options.
// Length must be 8-128 characters.
func GeneratePassword(length int, opts Options) (string, error) {
  // ...
}
```

### Complex Logic

```go
// Derive derives a key from the password using Argon2id with hardened parameters:
// - Memory: 64MB
// - Iterations: 3
// - Parallelism: 4
// These parameters balance security and performance for local use.
func (k *Argon2id) Derive(password string, salt []byte) ([32]byte, error) {
  // ... implementation
}
```

### Security-Critical Sections

```go
// Clean memory before returning. This ensures the key material is not
// recoverable from process memory after the function exits.
defer ZeroBytes(&key)

// Constant-time comparison prevents timing attacks when checking passwords.
if subtle.ConstantTimeCompare(hash[:], argon2Hash[:]) == 0 {
  return ErrInvalidPassword
}
```

---

## Error Handling

### Error Naming

```go
// Sentinel errors: Err{Concept} or Error{Verb}
var (
  ErrInvalidPassword     error  // ✅
  ErrVaultNotFound       error  // ✅
  ErrVaultLocked         error  // ✅
  ErrChecksumMismatch    error  // ✅
  
  ErrBad                 error  // ❌ (too vague)
  ErrFailed              error  // ❌ (too vague)
)
```

### Error Wrapping

```go
// Wrap errors with context. Do NOT lose stack trace.
hash, err := deriveKey(password, salt)
if err != nil {
  return [32]byte{}, fmt.Errorf("derive master key: %w", err)
}

// Later: errors.Unwrap() or errors.Is() to inspect
if errors.Is(err, ErrInvalidPassword) {
  // Handle specific error
}
```

### Error Files

Store sentinel errors in dedicated `errors.go` files:

```go
// core/vault/errors.go
package vault

var (
  ErrNotFound    = errors.New("item not found")
  ErrNotUnlocked = errors.New("vault is locked")
  ErrExists      = errors.New("item already exists")
)
```

---

## Testing Conventions

### Test File Organization

```go
// core/crypto/aes_gcm_test.go
package crypto

// Table-driven tests for variants
func TestEncrypt(t *testing.T) {
  tests := []struct {
    name      string
    data      []byte
    key       []byte
    wantErr   bool
    wantEqual bool
  }{
    {
      name:    "encrypts 32-byte key",
      data:    []byte("hello world"),
      key:     [32]byte{1, 2, 3, ...},
      wantErr: false,
    },
    {
      name:    "fails on invalid key size",
      data:    []byte("hello"),
      key:     [16]byte{1, 2, 3, ...},
      wantErr: true,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      ciphertext, err := AESGCMEncrypt(tt.data, tt.key)
      if (err != nil) != tt.wantErr {
        t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
      }
    })
  }
}
```

### Test Assertions

```go
import "github.com/stretchr/testify/assert"
import "github.com/stretchr/testify/require"

func TestVaultUnlock(t *testing.T) {
  v, _ := createTestVault()
  
  // Use require for setup (failures stop test)
  err := v.Create("password123")
  require.NoError(t, err)
  
  // Use assert for conditions (failures log but continue)
  assert.True(t, v.IsUnlocked(), "vault should be unlocked after create")
  assert.Equal(t, 0, len(v.Items()), "new vault should have 0 items")
}
```

### Test Helpers

```go
// Helpers in *_test.go files to avoid polluting main code
func createTestVault(t *testing.T) *Vault {
  tmpDir := t.TempDir()
  v := New(tmpDir)
  require.NoError(t, v.Create("test123"))
  return v
}

func generateTestPassword(t *testing.T, length int) string {
  pwd, err := GeneratePassword(length, DefaultOptions())
  require.NoError(t, err)
  return pwd
}
```

### Coverage

- Target: **90%+ coverage** for core modules (crypto, vault)
- Target: **80%+ coverage** for CLI, sync
- Required: No uncovered branches in security-critical code (KDF, cipher, key management)

---

## Security Patterns

### Memory Wiping

```go
// Always zero sensitive data before returning
func deriveKey(password string, salt []byte) ([32]byte, error) {
  key := argon2id.Key(...)
  defer ZeroBytes(&key)  // ✅ Immediate cleanup
  
  return key, nil
}

// ZeroBytes uses unsafe.Pointer + memory barrier
func ZeroBytes(b *[32]byte) {
  for i := range b {
    b[i] = 0
  }
  runtime.KeepAlive(b)
}
```

### Randomness

```go
// Always use crypto/rand, never math/rand
import "crypto/rand"

key := make([]byte, 32)
_, err := rand.Read(key)  // ✅ Secure random

// Never
rand.Read(key)  // ❌ math.rand, not crypto.rand
```

### Constant-Time Operations

```go
import "crypto/subtle"

// Always for password/checksum comparison
if subtle.ConstantTimeCompare(computed, stored) == 0 {
  return ErrInvalidChecksum
}

// Never
if computed == stored {  // ❌ Timing attack vector
  // ...
}
```

### Per-Item Keys

```go
// Derive unique key per item, never reuse master key directly
itemKey := hkdf.Expand(masterKey, itemID)

// Encrypt item with item key (prevents bulk decryption)
encrypted := aesGCM.Encrypt(itemData, itemKey)
defer ZeroBytes(&itemKey)
```

---

## Build & Compilation

### Build Commands

```bash
# Build CLI
go build -o zp ./packages/cli/

# Build sync server
go build -o syncserver ./services/syncserver/

# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark (for performance-critical paths)
go test -bench=. -benchmem ./core/crypto/
```

### Lint & Vet

```bash
# Run go vet (static analysis)
go vet ./...

# Format code (gofmt)
go fmt ./...

# Optional: use golangci-lint for comprehensive linting
golangci-lint run ./...
```

### Version Management

**Go Version:** 1.26.1 minimum

```go
// go.mod
go 1.26
```

---

## Concurrency & Goroutines

### Safe Resource Management

```go
// Use context.Context for cancellation
func (s *SyncClient) Pull(ctx context.Context) error {
  select {
  case <-ctx.Done():
    return ctx.Err()
  default:
  }
  // ... pull logic
}

// Channels for communication, not shared memory
results := make(chan *Item, 10)
errors := make(chan error)

go func() {
  item, err := vault.AddItem(...)
  if err != nil {
    errors <- err
  } else {
    results <- item
  }
}()
```

### Avoid Race Conditions

```bash
# Check for races
go test -race ./...

# Never directly access shared state without sync
type UnsafeVault struct {
  items map[string]*Item  // ❌ Race condition if accessed from multiple goroutines
}

// Instead, use sync.RWMutex or channels
type SafeVault struct {
  mu    sync.RWMutex
  items map[string]*Item  // ✅ Protected by mutex
}
```

---

## Configuration & constants

### Hardened Parameters

All security-relevant parameters are constants, not configurable at runtime:

```go
// core/crypto/kdf/argon2id.go
const (
  ArgonMemoryMiB = 64          // 64 MB (fixed)
  ArgonIterations = 3          // (fixed)
  ArgonParallelism = 4         // (fixed)
  EncryptionKeySize = 32       // 256-bit (fixed)
  SaltSize = 16                // 128-bit (fixed)
)
```

**Rationale:** Users should NOT be able to weaken security via configuration.

### Configurable Convenience Options

```go
// core/crypto/password/generator.go
type Options struct {
  IncludeSymbols    bool         // Configurable
  ExcludeAmbiguous  bool         // Configurable
  MinStrengthScore  int          // Configurable (1-4)
}

// But minimum requirements are enforced
if opts.MinStrengthScore < 1 {
  opts.MinStrengthScore = 1  // ✅ Can't disable security
}
```

---

## Dependency Management

### Direct Dependencies

```
github.com/spf13/cobra           # CLI framework
github.com/golang/x/crypto       # Argon2id, HKDF, SHA-256
github.com/golang/x/term         # Secure terminal input
github.com/modernc.org/sqlite    # Pure Go SQLite
github.com/go-bip39/go-bip39     # BIP-39 mnemonics
github.com/nbutton23/zxcvbn-go   # Password strength
github.com/atotto/clipboard      # Clipboard access
github.com/stretchr/testify      # Testing assertions
```

### No External API Dependencies

- All crypto implemented locally (no external services at runtime)
- HIBP integration is optional and uses k-anonymity (minimal data sent)
- No cloud provider lock-in

---

## Documentation Requirements

**All exported functions/types MUST have godoc comments.**

```go
// ❌ Missing comment
func (v *Vault) AddItem(item *Item) error {
}

// ✅ Proper godoc
// AddItem adds a new credential to the vault and returns an error if the item
// already exists or vault is locked.
func (v *Vault) AddItem(item *Item) error {
}
```

---

**Document Version:** 1.1  
**Last Updated:** June 2025  
**Owner:** Engineering
