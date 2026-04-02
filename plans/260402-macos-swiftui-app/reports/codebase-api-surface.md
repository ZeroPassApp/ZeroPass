Agent is idle (waiting for messages). agent_id: codebase-analysis, agent_type: explore, status: idle, description: Analyze Go core API surface, elapsed: 108s, total_turns: 1

[Turn 0]
Perfect! Now I have comprehensive knowledge of the codebase. Let me create a thorough summary:

## ZeroPass Go Public API Surface — Complete Mapping for macOS SwiftUI Integration

Based on a complete analysis of the ZeroPass codebase, here is the comprehensive public API surface that a macOS SwiftUI app needs to call:

---

### **MODULE 1: VAULT STORE API** 
**File:** `core/vault/store/store.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/store`

#### Key Types
```go
type Vault struct {
    // Handle to an open vault instance (stores path, metadata, keys, lock state)
}

type VaultMetadata struct {
    Salt                 string           // hex-encoded KDF salt
    EncryptedVaultKey    string           // base64-encoded encrypted vault key
    EncryptedRecoveryKey string           // base64-encoded encrypted recovery key
    CreatedAt            time.Time
    Config               VaultConfig
}

type VaultConfig struct {
    AutoLockTimeout      time.Duration    // 0 = disabled
    ClipboardClearSec    int              // seconds before clearing clipboard
    MaxVersions          int              // item version history limit
}

type CreateResult struct {
    Mnemonic string  // BIP-39 12-word recovery phrase (show once only)
}
```

#### Exported Functions
```go
// Vault Lifecycle
func Create(masterPassword string, vaultPath string, cfg VaultConfig) 
    → (*Vault, *CreateResult, error)
func Open(vaultPath string) → (*Vault, error)

// Authentication
func (v *Vault) Unlock(masterPassword string) → error
func (v *Vault) UnlockWithRecovery(mnemonic string) → error
func (v *Vault) Lock() → error

// State Management
func (v *Vault) IsLocked() → bool
func (v *Vault) VaultKey() → ([]byte, error)  // Returns nil if locked
func (v *Vault) Touch() → error               // Reset auto-lock timer

// Metadata Access
func (v *Vault) Path() → string
func (v *Vault) ItemsPath() → string
func (v *Vault) IndexPath() → string
func (v *Vault) Metadata() → *VaultMetadata
func (v *Vault) Config() → VaultConfig

// Recovery
func (v *Vault) RegenerateRecovery() → (string, error)  // Must be unlocked

// Configuration
func DefaultConfig() → VaultConfig

// Constants
const (
    DefaultVaultDir  = ".zeropass"
    DefaultVaultName = "default"
    VaultMetaFile    = "vault.json"
    ItemsDir         = "items"
    IndexFile        = "index.db"
)
```

#### State Management Notes
- **Vault is locked after `Open()`** — caller must call `Unlock()` or `UnlockWithRecovery()`
- **Vault key is stored in memory after unlock** — zeroed on `Lock()`
- **Auto-lock timer** — restarted with `Touch()`, automatically calls `Lock()` after `AutoLockTimeout`
- **Thread-safe** — internal `sync.RWMutex` protects concurrent access

---

### **MODULE 2: VAULT ITEM API**
**File:** `core/vault/item/item.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/item`

#### Manager Type
```go
type Manager struct {
    // Internal: manages CRUD, encryption, and search index integration
}

func NewManager(itemsPath string, vaultKeyFn func() ([]byte, error), 
    indexer Indexer) *Manager
    // vaultKeyFn: callback that returns vault key or error if locked
    // indexer: optional search index (pass nil to disable)
```

#### Indexer Interface (for Manager integration)
```go
type Indexer interface {
    AddToIndex(item *types.Item) error
    UpdateIndex(item *types.Item) error
    RemoveFromIndex(id string) error
}
```

#### Item CRUD Operations
```go
func (m *Manager) AddItem(item *types.Item) → error
    // - Auto-generates ID if empty
    // - Auto-categorizes type if empty (calls AutoCategorize)
    // - Sets CreatedAt, UpdatedAt, Version=1
    // - Encrypts and persists to disk

func (m *Manager) GetItem(id string) → (*types.Item, error)
    // - Decrypts from disk
    // - Updates LastAccessedAt (best-effort, doesn't fail on save error)

func (m *Manager) UpdateItem(id string, item *types.Item) → error
    // - Increments version
    // - Updates UpdatedAt to now
    // - Encrypts and persists

func (m *Manager) DeleteItem(id string) → error
    // - Removes item file and version history

func (m *Manager) ListItems(filter types.ItemFilter) → ([]*types.Item, error)
    // - Filters by type, tags, favorite, search query
    // - Sorts by name, created_at, updated_at, type, last_accessed_at

func (m *Manager) AllItems() → ([]*types.Item, error)
    // - Returns all items unfiltered

// Auto-Categorize Helper
func AutoCategorize(fields map[string]string) types.ItemType
    // Infers type from presence of specific fields
```

---

### **MODULE 3: VAULT TYPES**
**File:** `core/vault/types/types.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/types`

#### Item Type Constants
```go
const (
    ItemTypeLogin      ItemType = "login"       // Username/password
    ItemTypeAPIKey     ItemType = "apikey"      // API key/secret
    ItemTypeSSHKey     ItemType = "sshkey"      // SSH keypair
    ItemTypeSecureNote ItemType = "note"        // Text note
    ItemTypeCreditCard ItemType = "creditcard"  // Card details
    ItemTypeIdentity   ItemType = "identity"    // Personal info
    ItemTypePasskey    ItemType = "passkey"     // WebAuthn credential
    ItemTypeCustom     ItemType = "custom"      // User-defined
)
```

#### Field Name Constants (by type)
```go
// Login fields
const (
    FieldUsername = "username"
    FieldPassword = "password"
    FieldURL      = "url"
    FieldTOTP     = "totp"          // TOTP seed or QR code
)

// API Key fields
const (
    FieldAPIKey    = "api_key"
    FieldAPISecret = "api_secret"
    FieldEndpoint  = "endpoint"
)

// SSH Key fields
const (
    FieldPublicKey  = "public_key"
    FieldPrivateKey = "private_key"
    FieldPassphrase = "passphrase"
    FieldFingerprint = "fingerprint"
)

// Credit Card fields
const (
    FieldCardNumber = "card_number"
    FieldExpiry     = "expiry"
    FieldCVV        = "cvv"
    FieldCardHolder = "card_holder"
)

// Identity fields
const (
    FieldFirstName = "first_name"
    FieldLastName  = "last_name"
    FieldEmail     = "email"
    FieldPhone     = "phone"
    FieldAddress   = "address"
)

// Passkey fields
const (
    FieldCredentialID    = "credential_id"
    FieldPasskeyPublicKey = "passkey_public_key"
    FieldRelyingPartyID  = "rp_id"
    FieldUserHandle      = "user_handle"
    FieldSignCount       = "sign_count"
)
```

#### Item Struct
```go
type Item struct {
    ID             string            `json:"id"`             // UUID-like identifier
    Type           ItemType          `json:"type"`
    Name           string            `json:"name"`           // Required
    Fields         map[string]string `json:"fields"`         // Typed fields (username, password, etc.)
    Notes          string            `json:"notes"`
    Tags           []string          `json:"tags"`
    Favorite       bool              `json:"favorite"`
    CustomFields   map[string]string `json:"custom_fields"`  // User-defined fields
    CreatedAt      time.Time         `json:"created_at"`
    UpdatedAt      time.Time         `json:"updated_at"`
    LastAccessedAt time.Time         `json:"last_accessed_at"`
    Version        int               `json:"version"`        // Incremented on each update
}

type EncryptedItem struct {
    ID       string `json:"id"`
    Data     string `json:"data"`     // base64-encoded encrypted JSON
    Version  int    `json:"version"`
    Checksum string `json:"checksum"` // SHA-256 hex of ciphertext (for corruption detection)
}
```

#### Item Filtering & Sorting
```go
type ItemFilter struct {
    Type        ItemType  `json:"type,omitempty"`
    Tags        []string  `json:"tags,omitempty"`
    Favorite    *bool     `json:"favorite,omitempty"`
    SearchQuery string    `json:"search_query,omitempty"`
    SortBy      string    `json:"sort_by,omitempty"`
    SortOrder   string    `json:"sort_order,omitempty"`  // "asc" or "desc"
}

const (
    SortByName         = "name"
    SortByCreatedAt    = "created_at"
    SortByUpdatedAt    = "updated_at"
    SortByType         = "type"
    SortByLastAccessed = "last_accessed_at"
    SortAsc            = "asc"
    SortDesc           = "desc"
)
```

---

### **MODULE 4: SEARCH/INDEX API**
**File:** `core/vault/index/index.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/index`

#### Index Type
```go
type Index struct {
    // SQLite FTS5 full-text search backed index
}

// Lifecycle
func Open(dbPath string) → (*Index, error)       // Creates or opens index
func (idx *Index) Close() → error

// Search Operations
func (idx *Index) Search(query string) → ([]string, error)
    // Returns list of item IDs matching query
    // Query sanitized to prevent FTS5 injection
    // Prefix matching enabled (e.g., "pass" matches "password")

func (idx *Index) AddToIndex(item *types.Item) → error
    // Indexes: name, type, tags, notes, custom_keys, field_values

func (idx *Index) UpdateIndex(item *types.Item) → error
    // Updates existing index entry (calls AddToIndex internally)

func (idx *Index) RemoveFromIndex(id string) → error

func (idx *Index) RebuildIndex(items []*types.Item) → error
    // Clears and rebuilds entire index

// At-Rest Encryption
func IsEncrypted(dbPath string) → bool           // Checks if .enc file exists

func EncryptIndexFile(dbPath string, key []byte) → error
    // AES-256-GCM encryption, atomic write (temp + rename)
    // WAL checkpoint before encryption
    // Removes plaintext and sidecar files

func DecryptIndexFile(dbPath string, key []byte) → error
    // Decrypts .enc file and writes plaintext
    // Removes .enc file after success
```

---

### **MODULE 5: CRYPTO — KEY MANAGEMENT**
**File:** `core/crypto/key/` (multiple files)
**Module Path:** `github.com/zeropass/zeropass/core/crypto/key`

#### Master Key Derivation
```go
type MasterKey struct {
    // Holds derived key and salt (call Zero() to securely wipe)
}

func DeriveNewMasterKey(password []byte, k *kdf.Argon2id) 
    → (*MasterKey, error)
    // Generates new 16-byte salt, derives key
    // Use for new vault creation

func DeriveMasterKeyWithSalt(password []byte, salt []byte, k *kdf.Argon2id)
    → (*MasterKey, error)
    // Derives key from password and existing salt
    // Use for unlocking existing vault

// MasterKey Methods
func (mk *MasterKey) Key() → []byte         // Returns raw key bytes
func (mk *MasterKey) Salt() → []byte        // Returns salt
func (mk *MasterKey) Zero() → error         // Securely wipe from memory
```

#### Vault Key Management
```go
const VaultKeySize = 32  // 256 bits

type EncryptedVaultKey struct {
    Ciphertext []byte
}

// Generate & Encrypt
func GenerateVaultKey() → ([]byte, error)
    // Random 32-byte vault key

func EncryptVaultKey(vaultKey []byte, masterKey []byte) 
    → (*EncryptedVaultKey, error)
    // AES-256-GCM encryption

func DecryptVaultKey(encrypted *EncryptedVaultKey, masterKey []byte) 
    → ([]byte, error)
    // Returns vault key or error on auth failure

func RotateVaultKey(encrypted *EncryptedVaultKey, oldMasterKey, newMasterKey []byte)
    → (*EncryptedVaultKey, error)
    // Re-encrypt vault key with new master key
```

#### Recovery Key (BIP-39 Mnemonic)
```go
type RecoveryKeyManager struct{}

func NewRecoveryKeyManager() *RecoveryKeyManager

// BIP-39 Operations
func (r *RecoveryKeyManager) GenerateMnemonic() → (string, error)
    // Returns 12-word BIP-39 mnemonic (entropy=128 bits)

func (r *RecoveryKeyManager) ValidateMnemonic(mnemonic string) → bool
    // Validates BIP-39 mnemonic format

func (r *RecoveryKeyManager) DeriveKeyFromMnemonic(mnemonic string) 
    → ([]byte, error)
    // Derives 32-byte key using BIP-39 seed (PBKDF2-HMAC-SHA512, 2048 iterations)

// Vault Key Recovery Functions (standalone)
func EncryptVaultKeyWithRecovery(vaultKey []byte, mnemonic string)
    → (*EncryptedVaultKey, error)

func DecryptVaultKeyWithRecovery(encrypted *EncryptedVaultKey, mnemonic string)
    → ([]byte, error)

func RegenerateRecoveryKey(vaultKey []byte)
    → (string, *EncryptedVaultKey, error)
    // Generates new mnemonic and re-encrypts vault key
```

#### Item Key Derivation
```go
const ItemKeySize = 32

func DeriveItemKey(vaultKey []byte, itemID string) → ([]byte, error)
    // HKDF-SHA256 derivation: unique per-item key
    // Uses itemID as salt, "zeropass-item-key" as info
```

#### KDF (Argon2id)
```go
type Argon2id struct{}

type Argon2idParams struct {
    Memory      uint32  // KiB (default 65536 = 64 MB)
    Iterations  uint32  // Passes (default 3)
    Parallelism uint8   // Threads (default 4)
    KeyLen      uint32  // Output length (default 32)
    SaltLen     int     // Salt length (default 16)
}

func NewDefault() *Argon2id
func New(params Argon2idParams) *Argon2id
func (a *Argon2id) DeriveKey(password []byte, salt []byte) → ([]byte, error)
func (a *Argon2id) GenerateSalt() → ([]byte, error)
func (a *Argon2id) Params() → Argon2idParams
```

#### Cipher (AES-256-GCM)
```go
type AESGCM struct{}

const (
    KeySize   = 32  // 256 bits
    NonceSize = 12  // 96 bits (GCM standard)
)

func New() *AESGCM

func (a *AESGCM) Encrypt(plaintext []byte, key []byte) → ([]byte, error)
    // Output: [12-byte nonce][ciphertext + 16-byte GCM tag]

func (a *AESGCM) Decrypt(ciphertext []byte, key []byte) → ([]byte, error)
    // Extracts nonce from first 12 bytes, decrypts remainder
```

#### Encoding Helpers
```go
func Base64StdEncode(data []byte) → string
func Base64StdDecode(s string) → ([]byte, error)
func Base64URLEncode(data []byte) → string
func Base64URLDecode(s string) → ([]byte, error)
func HexEncode(data []byte) → string
func HexDecode(s string) → ([]byte, error)
```

#### Memory Security
```go
func ZeroBytes(b []byte) → void
    // Securely wipes byte slice from memory
    // Uses volatile write pattern to prevent compiler optimization

func SecureCompare(a, b []byte) → bool
    // Constant-time comparison (timing side-channel resistant)

func SecureCompareStrings(a, b string) → bool
    // Constant-time string comparison
```

---

### **MODULE 6: PASSWORD GENERATOR API**
**File:** `core/crypto/password/generator.go`
**Module Path:** `github.com/zeropass/zeropass/core/crypto/password`

#### Generator Type
```go
type Generator struct{}

func NewGenerator() *Generator

type GeneratorOptions struct {
    Uppercase        bool
    Lowercase        bool
    Digits           bool
    Symbols          bool
    ExcludeAmbiguous bool  // Excludes: l, 1, I, O, 0
}

const (
    MinPasswordLength = 8
    MaxPasswordLength = 128
    MinPassphraseWords = 4
    MaxPassphraseWords = 8
)
```

#### Password Generation
```go
func (g *Generator) GenerateRandom(length int, opts GeneratorOptions) 
    → (string, error)
    // Generates cryptographically secure random password
    // Ensures at least one character from each required set
    // Default (no options): lowercase + uppercase + digits + symbols

func (g *Generator) GeneratePassphrase(words int, separator string) 
    → (string, error)
    // Generates diceware passphrase (e.g., "correct-horse-battery-staple")
    // Uses word list from go-bip39
    // words: 4-8 words

func (g *Generator) ScoreStrength(password string) 
    → (StrengthResult, error)
    // Uses zxcvbn library for strength estimation
```

#### Strength Score
```go
type StrengthResult struct {
    Score         int      // 0-4 (0=very weak, 4=very strong)
    Feedback      string   // Human-readable feedback
    CrackTimeSecs float64  // Estimated offline crack time in seconds
}
```

---

### **MODULE 7: HIBP BREACH CHECK API**
**File:** `core/vault/health/hibp.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/health`

#### HIBP Client
```go
type HIBPClient struct{}

type BreachResult struct {
    ItemID          string
    ItemName        string
    Breached        bool
    OccurrenceCount int  // How many times password appeared in breaches
}

// Constructor
func NewHIBPClient(opts ...HIBPOption) *HIBPClient
    // Default base URL: https://api.pwnedpasswords.com
    // Default HTTP timeout: 10 seconds

// Options
func WithBaseURL(url string) HIBPOption       // For testing
func WithHTTPClient(hc *http.Client) HIBPOption
```

#### Breach Checking
```go
func (c *HIBPClient) CheckPassword(password string) → (bool, int, error)
    // Returns (breached, occurrence_count, error)
    // Uses k-anonymity: only first 5 chars of SHA-1 hash sent to API
    // Rate limited: 1500ms minimum between requests (free tier limit)

func (c *HIBPClient) CheckBatch(items []*batchItem) → ([]BreachResult, error)
    // Batch check multiple passwords with automatic rate limiting

func BuildBatchItems(passwords map[string]string, names map[string]string) 
    → []*batchItem
    // Helper: constructs batch input from ID→password map
```

---

### **MODULE 8: PASSWORD HEALTH ANALYSIS**
**File:** `core/vault/health/health.go`
**Module Path:** `github.com/zeropass/zeropass/core/vault/health`

#### Analyzer
```go
type Analyzer struct{}

type AnalyzerConfig struct {
    MinScore       int           // Min zxcvbn score (0-4), default 3
    MaxPasswordAge time.Duration // Max age before warning, default 90 days
}

type Finding struct {
    ItemID   string   // Item ID
    ItemName string
    Severity Severity // "critical", "warning", "info"
    Category string   // "weak", "reused", "old", "breached"
    Message  string   // Human-readable message
}

type BreachFinding struct {
    ItemID          string
    ItemName        string
    OccurrenceCount int
    Recommendation  string
}

type HealthReport struct {
    TotalItems        int
    LoginItems        int
    WeakCount         int
    ReusedCount       int
    OldCount          int
    BreachedCount     int
    BreachedPasswords []BreachFinding
    Findings          []Finding
    AnalyzedAt        time.Time
    OverallScore      int  // 0-100
}

const (
    SeverityCritical Severity = "critical"
    SeverityWarning  Severity = "warning"
    SeverityInfo     Severity = "info"
)
```

#### Analysis Operations
```go
func NewAnalyzer(cfg AnalyzerConfig) *Analyzer

func DefaultAnalyzerConfig() AnalyzerConfig

func (a *Analyzer) Analyze(items []*types.Item) *HealthReport
    // Checks:
    // - Weak passwords (zxcvbn score < config.MinScore)
    // - Reused passwords (same password on multiple accounts)
    // - Old passwords (UpdatedAt > MaxPasswordAge)
    // - Calculates overall score (0-100) based on findings
    // - Does NOT check HIBP (separate call to HIBPClient)
```

---

### **MODULE 9: IMPORT/EXPORT API**
**File:** `core/vault/importexport/` (export.go, import.go)
**Module Path:** `github.com/zeropass/zeropass/core/vault/importexport`

#### Export Functions
```go
func ExportJSON(items []types.Item, writer io.Writer) → error
    // Exports unencrypted JSON array of items

func ExportCSV(items []types.Item, writer io.Writer) → error
    // Exports unencrypted CSV
    // Columns: name, type, url, username, password, notes, tags
    // Useful for: simple backups, migration to other password managers

func ExportEncrypted(items []types.Item, key []byte, writer io.Writer) → error
    // Exports AES-256-GCM encrypted backup
    // Output: base64-encoded ciphertext
    // Use for: encrypted backups at rest

func ImportEncrypted(reader io.Reader, key []byte) → ([]types.Item, error)
    // Decrypts and deserializes encrypted backup
```

#### Import Functions (Auto-Detect Password Manager Format)
```go
// Generic CSV with mapping
type CSVMapping struct {
    Name     int  // Column index for item name
    URL      int
    Username int
    Password int
    Notes    int
    Type     int  // -1 if not present
    Skip     int  // Header rows to skip (usually 1)
}

func ImportCSV(reader io.Reader, mapping CSVMapping) → ([]types.Item, error)

// Specific Password Manager Importers
func ImportChrome(reader io.Reader) → ([]types.Item, error)
func ImportFirefox(reader io.Reader) → ([]types.Item, error)
func Import1Password(reader io.Reader) → ([]types.Item, error)
func Import1PUX(reader io.Reader) → ([]types.Item, error)      // 1Password .1pux export
func ImportBitwarden(reader io.Reader) → ([]types.Item, error)  // Auto-detects CSV/JSON
func ImportSafari(reader io.Reader) → ([]types.Item, error)
func ImportLastPass(reader io.Reader) → ([]types.Item, error)
func ImportKeePass(reader io.Reader) → ([]types.Item, error)

// Format Detection
// Bitwarden importer auto-detects JSON vs CSV
```

---

### **MODULE 10: SYNC ENGINE API**
**File:** `core/sync/client/sync_client.go` + `core/sync/protocol/types.go`
**Module Path:** `github.com/zeropass/zeropass/core/sync/client` + `core/sync/protocol`

#### Sync Protocol Types
```go
type SyncItem struct {
    ItemID    string `json:"item_id"`
    Version   int    `json:"version"`
    DeviceID  string `json:"device_id"`
    Payload   string `json:"payload"`    // Base64 encrypted blob (server never sees plaintext)
    Timestamp int64  `json:"timestamp"` // Unix nanoseconds
    Checksum  string `json:"checksum"`  // SHA-256 hex of payload
    Deleted   bool   `json:"deleted"`   // Tombstone for deletes
}

type PullResponse struct {
    Items      []SyncItem
    ServerTime int64
}

type PushResponse struct {
    Accepted   []string       // Item IDs that were accepted
    Conflicts  []ConflictInfo // Items with conflicts
    ServerTime int64
}

type ConflictInfo struct {
    ItemID     string
    ServerItem SyncItem
    Message    string
}

type SyncResult struct {
    Pulled    int  // Number of items pulled from server
    Pushed    int  // Number of items pushed successfully
    Conflicts int  // Number of conflicts encountered
}

type DeviceInfo struct {
    DeviceID   string
    DeviceName string
    LastSyncAt int64
}

type ErrorResponse struct {
    Error string
}
```

#### Sync Client
```go
type SyncClient struct{}

type Option func(*SyncClient)

// Constructor
func NewSyncClient(serverURL, deviceID string, opts ...Option) *SyncClient
    // serverURL: e.g., "https://sync.example.com"
    // deviceID: unique identifier for this device

// Options
func WithHTTPClient(c *http.Client) Option
func WithAPIKey(key string) Option
func WithLastSyncTime(ts int64) Option
    // Persist lastSyncTime between sessions
```

#### Client Operations
```go
func (sc *SyncClient) Register(deviceName string) → error
    // Register device with sync server

func (sc *SyncClient) Pull() → ([]protocol.SyncItem, error)
    // Fetch items changed since last sync
    // Uses internally stored lastSyncTime

func (sc *SyncClient) Push(items []protocol.SyncItem) 
    → (*protocol.PushResponse, error)
    // Upload local changes to server

func (sc *SyncClient) Sync(localItems []protocol.SyncItem) 
    → (*protocol.SyncResult, error)
    // Full sync cycle:
    // 1. Pull remote changes
    // 2. Resolve conflicts (last-write-wins)
    // 3. Push resolved items
    // 4. Update lastSyncTime
    // Returns aggregate counts (pulled, pushed, conflicts)

func (sc *SyncClient) LastSyncTime() → int64
    // Returns last successful sync timestamp (Unix nanoseconds)

func (sc *SyncClient) ConflictHistory() → []conflict.ConflictRecord
    // Returns conflict history for UI feedback
```

#### Conflict Resolution
```go
type Resolver struct{}  // From core/sync/conflict/resolver.go

type ConflictRecord struct {
    ItemID string
    Winner protocol.SyncItem
    Loser  protocol.SyncItem
    Reason string  // e.g., "local has later timestamp"
}

// Resolution Strategy (deterministic):
// 1. Last-write-wins: higher timestamp wins
// 2. If timestamps equal: higher version wins
// 3. If still tied: server (remote) wins for determinism
```

---

### **MODULE 11: CLI COMMANDS (Exposing Operations)**
**Directory:** `packages/cli/cmd/`
**Maps to:** Vault operations available to native code

Key command operations show required APIs:
```go
// Vault Initialization
"init" → Create vault (master password → recovery mnemonic)

// Authentication
"unlock"    → Unlock with password or recovery phrase
"lock"      → Lock vault (wipe keys from memory)
"status"    → Check lock state

// Item Management
"add"       → Create item (auto-categorization from fields)
"get <query>" → Retrieve by name/ID (supports copy-to-clipboard)
"list"      → List items (filters, sorting)
"edit"      → Update item
"delete"    → Remove item
"search <query>" → Full-text search

// Password/Key Generation
"generate" → Generate password (length, character set)
"generate --passphrase" → Generate passphrase (words, separator)

// Vault Health
"health" → Analyze vault: weak passwords, reused, old, breached

// Recovery
"recovery --regenerate" → Generate new recovery mnemonic

// Import/Export
"import" → Import from password managers (Chrome, Firefox, 1Password, Bitwarden, etc.)
"export" → Export to JSON, CSV, or encrypted backup

// Sync (Multi-device)
"sync" → Synchronize with server (pull + resolve + push)
```

---

### **DEPENDENCY GRAPH — Module Relationships**

```
┌─────────────────────────────────────────────────────┐
│  SwiftUI App (calling native Go modules)           │
└──────────────────┬──────────────────────────────────┘
                   │
       ┌───────────┴───────────┐
       ▼                       ▼
  ┌─────────────┐        ┌──────────────┐
  │ Vault Store │        │Item Manager  │
  │ (vault.go)  │        │(item.go)     │
  └──────┬──────┘        └──────┬───────┘
         │                       │
         └───────────┬───────────┘
                     ▼
         ┌───────────────────────┐
         │   Crypto Module       │
         ├───────────────────────┤
         │ • Key derivation      │
         │   (KDF: Argon2id)     │
         │ • Master key mgmt     │
         │ • Recovery key (BIP39)│
         │ • Item key derivation │
         │ • Cipher (AES-256-GCM)│
         │ • Encoding            │
         │ • Memory security     │
         └───────────┬───────────┘
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
    ┌─────────────┐      ┌──────────────┐
    │ Index (FTS) │      │Import/Export │
    │ (SQLite)    │      │              │
    └─────────────┘      └──────────────┘
    
    ┌──────────────────────────────────────────┐
    │ Health Analysis                          │
    ├──────────────────────────────────────────┤
    │ • Password strength (zxcvbn)             │
    │ • Reuse detection                        │
    │ • Age detection                          │
    │ • HIBP breach checking (k-anonymity)     │
    └──────────────────────────────────────────┘
    
    ┌──────────────────────────────────────────┐
    │ Sync Engine (Multi-device)               │
    ├──────────────────────────────────────────┤
    │ • HTTP client to sync server             │
    │ • Conflict resolution (LWW + version)    │
    │ • Device registration                    │
    └──────────────────────────────────────────┘
```

---

### **GO MODULE INFORMATION**

**go.mod Path:** `github.com/zeropass/zeropass`

**Key Dependencies:**
```
github.com/atotto/clipboard v0.1.4              // Clipboard access
github.com/lib/pq v1.12.1                       // PostgreSQL driver (optional)
github.com/nbutton23/zxcvbn-go v0.0.0          // Password strength estimation
github.com/spf13/cobra v1.10.2                  // CLI framework (not needed for SwiftUI)
github.com/tyler-smith/go-bip39 v1.1.0         // BIP-39 mnemonics
github.com/stretchr/testify v1.11.1            // Testing only
golang.org/x/crypto v0.49.0                    // Argon2id, HKDF, curves
golang.org/x/term v0.41.0                      // Terminal I/O (not needed for SwiftUI)
modernc.org/sqlite v1.48.0                     // Pure-Go SQLite (FTS5)
```

---

### **CRITICAL STATE MANAGEMENT FOR SWIFT BRIDGE**

1. **Vault Lock State**
   - Vault must be explicitly unlocked before operations
   - Lock state should be monitored (auto-lock timer)
   - Key material is zeroed on lock

2. **Vault Key Callback Pattern**
   - Item Manager requires callback to get vault key
   - Callback must error if vault is locked
   - Prevents accidental operations on locked vault

3. **Item ID Generation**
   - Don't supply IDs from Swift — let Go generate them (UUID-like)
   - IDs are used as salt for per-item key derivation (HKDF)

4. **Encryption at Rest**
   - Index file encrypted when vault is locked (AES-256-GCM)
   - Item files always encrypted on disk
   - Symmetric keys derived from vault key using HKDF-SHA256

5. **Error Handling**
   - Locked vault errors: return special error type for UI
   - Auth errors: "incorrect master password" vs. "vault is locked"
   - Corruption errors: "item checksum mismatch" (SHA-256 verified)

6. **Concurrency**
   - Vault uses sync.RWMutex for thread safety
   - Safe to call from multiple Swift threads
   - Lock/unlock operations are atomic

---

### **ENTRY POINTS FOR SWIFTUI APP**

```swift
// 1. Create new vault
let (vault, recoveryInfo, error) = store.Create(masterPassword, vaultPath, config)
// Show recoveryInfo.Mnemonic once, never again

// 2. Open existing vault
let vault = try store.Open(vaultPath)

// 3. Authenticate
try vault.Unlock(masterPassword)  // or UnlockWithRecovery(mnemonic)

// 4. Create item manager
let indexer = index.Open(vault.IndexPath())
let manager = item.NewManager(vault.ItemsPath(), vault.VaultKey, indexer)

// 5. Perform CRUD
let item = types.Item { name: "GitHub", type: .Login, fields: [...] }
try manager.AddItem(item)
let retrieved = try manager.GetItem(id)
try manager.UpdateItem(id, item)

// 6. Search
let results = try indexer.Search("github")

// 7. Health analysis
let analyzer = health.NewAnalyzer(health.DefaultAnalyzerConfig())
let report = analyzer.Analyze(allItems)

// 8. HIBP check (if desired)
let hibpClient = health.NewHIBPClient()
let breached, count, error = hibpClient.CheckPassword(password)

// 9. Export
try importexport.ExportJSON(items, writer)

// 10. Lock vault
vault.Lock()  // Wipes keys, encrypts index
```

---

This comprehensive mapping provides all the Go APIs, types, constants, and operations needed to design a complete Go↔Swift FFI bridge for the macOS SwiftUI application.