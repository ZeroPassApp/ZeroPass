# ZeroPass: Codebase Summary

## Overview

ZeroPass codebase (~20K LOC, ~70 Go files) is organized into four tiers:

1. **Core Layer** (`core/`) — Reusable crypto, vault, and sync engines
2. **Bridge Layer** (`bridge/`) — CGO C ABI (Go `c-archive`) for native app integration
3. **Packages Layer** (`packages/`) — CLI application and planned SDK
4. **Services Layer** (`services/`) — Self-hosted sync server

**Module Structure:**
```
github.com/zeropass/zeropass/
├── core/
│   ├── crypto/      # Encryption engine, key management, password generation
│   ├── vault/       # Vault CRUD, search, versioning, health, import/export
│   └── sync/        # Delta sync client/server, conflict resolution
├── bridge/          # CGO bridge (C ABI) for native apps (SwiftUI, etc.)
├── packages/
│   ├── cli/         # CLI application (Cobra-based)
│   └── sdk/         # SDK (planned, empty)
├── services/
│   └── syncserver/  # Self-hosted HTTP sync server
├── go.mod
└── go.sum
```

---

## Core Layer

### `/core/crypto/` — Encryption Engine

**Purpose:** All cryptographic operations for ZeroPass.

**Submodules:**

| Module | Responsibility |
|--------|-----------------|
| `interfaces.go` | KDF, Cipher, RecoveryManager, PasswordGenerator interfaces (interface-based DI) |
| `memory.go` | Secure memory wiping with `unsafe.Pointer` and memory barriers |
| `cipher/aes_gcm.go` | AES-256-GCM symmetric encryption with unique random nonce per operation |
| `encoding/base64.go` | URL-safe and standard base64 encoding/decoding for secrets |
| `encoding/hex.go` | Hex encoding/decoding for binary data |
| `kdf/argon2id.go` | Argon2id key derivation function (64MB memory, 3 iterations, 4 parallelism) |
| `key/master_key.go` | Master key derivation from password via Argon2id → 256-bit key |
| `key/vault_key.go` | Random 256-bit vault key, encrypted under master key |
| `key/item_key.go` | Per-item 256-bit key derived via HKDF-SHA256 from vault key + item ID |
| `key/recovery_key.go` | BIP-39 12-word mnemonic, PBKDF2-HMAC-SHA512 for recovery path |
| `password/generator.go` | Random password generation (configurable length, charset), diceware passphrase generation |
| `password/strength.go` | Password strength scoring via zxcvbn-go (0–4 scale, minimum 3 required) |
| `password/wordlist.go` | EFF short diceware wordlist (1,296 words for passphrase generation) |

**Key Hierarchy:**
```
User Password
  ↓ (Argon2id, 64MB, 3 iter, 4 par)
Master Key (256-bit)
  ↓ (AES-256-GCM encryption, stored in vault metadata)
Vault Key (256-bit random)
  ↓ (HKDF-SHA256 + item ID)
Item Key (256-bit, per-item, reused per key derivation)
  ↓ (AES-256-GCM encryption, unique nonce per operation)
Encrypted Item
```

**Security Patterns:**
- All keys stored as `[32]byte` (Go fixed arrays for stack allocation)
- Memory wiping via `ZeroBytes()` before key deletion
- `crypto/rand` for all randomness (no WeakRand)
- Constant-time comparison via `subtle.ConstantTimeCompare`
- Per-operation unique nonces (no IV reuse)
- SHA-256 checksums on encrypted items

---

### `/core/vault/` — Vault Engine

**Purpose:** Vault lifecycle management, item CRUD, search, versioning, health checks, import/export.

**Submodules:**

| Module | Responsibility |
|--------|-----------------|
| `types/types.go` | Item, EncryptedItem, ItemFilter, ItemType constants (login, apikey, sshkey, note, creditcard, identity, custom) |
| `item/item.go` | Item CRUD manager: auto encrypt/decrypt, per-item keys, checksums, versioning metadata |
| `store/store.go` | Vault lifecycle: Create, Open, Unlock (password/recovery), Lock, auto-lock on timeout, `ValidateRecovery()` (non-destructive mnemonic check), `UnlockWithRecovery()` (auto-rotates recovery key per PRD §8.5) |
| `index/index.go` | SQLite FTS5 full-text search with query sanitization (prevent injection) |
| `version/version.go` | Item version history tracking ({id}.versions.json files) |
| `health/health.go` | Password health analyzer (weak, reused, old passwords, entropy analysis) |
| `health/hibp.go` | Haveibeenpwned integration (k-anonymity: only first 5 chars of SHA-1 sent) |
| `importexport/import.go` | Import from 9 sources: Chrome, Firefox, Safari, 1Password (CSV), 1PUX, Bitwarden, LastPass, KeePass, generic CSV |
| `importexport/export.go` | Export vault: JSON (structured), CSV (tabular), encrypted archive |

**Vault Lifecycle:**
```
Create(vaultPath, password)
  → Derive master key
  → Generate random vault key
  → Encrypt vault key under master key
  → Write vault metadata + BIP-39 mnemonic
  → Initialize SQLite FTS5 index
  → Lock vault

Open(vaultPath)
  → Read vault metadata (encrypted)
  → Load SQLite index (unindexed until unlock)
  
Unlock(password | recovery_mnemonic)
  → Derive key from password/mnemonic
  → Decrypt vault key
  → (Best-effort) upgrade/repair vault metadata for key-based unlock
  → Load items into memory (encrypted)
  → Build searchable index

UnlockWithKey(vaultKey)
  → Validate vaultKey against vault metadata (`vault_key_check`)
  → Load items into memory (encrypted)
  → Build searchable index
  
Lock()
  → Clear vault key from memory
  → Items remain encrypted on disk
```

**Storage Format:**
```
vault/
├── vault.json                  # Vault metadata (salt, encrypted keys, vault_key_check, config)
├── vault.lock                  # Advisory lock (bridge only; held for session lifetime)
├── index.db                    # SQLite FTS5 index while unlocked (may include plaintext indexed fields)
├── index.db-wal                # SQLite sidecar (unlocked)
├── index.db-shm                # SQLite sidecar (unlocked)
├── index.db.enc                # Encrypted index at rest when locked (plaintext index removed)
└── items/
    ├── {item_id}.json          # Encrypted item (base64 ciphertext + checksum)
    └── {item_id}.versions.json # Encrypted version snapshots (legacy plaintext migrated on read)
```

**Key Patterns:**
- Interface-based DI for storage backends (SwappableStore interface)
- Auto lock on N minutes of inactivity
- Versioning for rollback capability
- FTS5 query sanitization (strip special chars; user wildcards removed; prefix matching appended internally)
- Vault metadata writes are atomic (temp + sync + rename) to reduce corruption risk
- Vault metadata preserves unknown JSON fields for forward/backward compatibility
- `vault_key_check` enables safe raw-key unlock (`UnlockWithKey`) for future TouchID/keychain flows

---

### `/core/sync/` — Sync Engine

**Purpose:** Delta sync between client and server, conflict resolution, multi-device coordination.

**Submodules:**

| Module | Responsibility |
|--------|-----------------|
| `protocol/types.go` | SyncItem, Pull/Push request/response, ConflictInfo, DeviceInfo, timestamps |
| `client/sync_client.go` | HTTP sync client: Pull, Push, FullSync modes with exponential backoff |
| `server/sync_handler.go` | HTTP handlers: `/sync/pull`, `/sync/push`, `/devices/register`, health check |
| `conflict/resolver.go` | Conflict resolution: last-write-wins → version tiebreak → remote preference |

**Sync Protocol:**
```
Client: Register device (device_id, name, public_key)
  ↓
Client: Pull request (last_sync_timestamp)
  ← Server: Changed items since timestamp
  
Client: Apply remote changes, resolve conflicts, merge local changes
  
Client: Push request (local changed items + version vectors)
  → Server: Apply, detect conflicts, respond with resolution hints
  
Client: Re-pull and merge if conflicts detected
```

**Conflict Resolution Rules:**
1. Last-write-wins (compare `updated_at` timestamps)
2. If equal timestamps, compare version vectors (logical clocks)
3. If still tied, prefer remote (authority = server)

**Storage:**
- Server: SQLite (sync_items table + devices table)
- Client: JSON files + in-memory merge state

---

## Packages Layer

### `/packages/cli/` — CLI Application

**Purpose:** User-facing CLI interface built on Cobra framework.

**Root Command:** `zeropass`

**Global Flags:**
- `--vault-path` — Vault directory (default: `~/.zeropass`)
- `--output` — Output format (json, table, csv)
- `--no-color` — Disable colored output

**Commands:**

| Command | Description |
|---------|-------------|
| `init` | Create new vault, generate recovery mnemonic |
| `unlock` | Unlock vault (password or recovery mnemonic) |
| `lock` | Lock vault manually |
| `add` | Add new credential (interactive prompts) |
| `get` | Retrieve credential by name or ID (`--copy` to clipboard) |
| `search` | Full-text search (`zp search "github"`) |
| `list` | List items with filters (`--type=login`, `--tags=prod`) |
| `edit` | Edit credential fields |
| `delete` | Delete credential (with confirmation) |
| `generate` | Generate password or passphrase (`--length=32`, `--type=passphrase`) |
| `health` | Password health report (weak, reused, old) |
| `recovery` | Manage recovery: validate (`ValidateRecovery`), test, regenerate mnemonic; auto-rotates on recovery unlock |
| `import` | Import from other PMs (Chrome, Firefox, Safari, 1Password, 1PUX, Bitwarden, LastPass, KeePass, CSV) |
| `export` | Export vault (json, csv, encrypted); `--force` flag skips plaintext warning |
| `run` | Inject secrets via `.env` file (`zp run -- npm start`) |
| `env` | Environment management (dev/staging/prod tagging) |

**Helper Functions** (`cmd/helpers.go`):
- `promptPassword()` — Securely prompt password (no echo)
- `openAndUnlockVault()` — Shared vault open+unlock logic
- `copyToClipboard()` — Copy secret to clipboard + auto-clear after 30s
- `findItemByQuery()` — Fuzzy search by name or ID

**File Structure:**
```
cmd/
├── root.go         # Root command + global flags
├── helpers.go      # Shared utilities
├── init.go         # Init command
├── add.go          # Add command
├── get.go          # Get command
├── ... (other commands)
└── main.go         # Entry point → Execute()
```

---

### `/packages/sdk/` — SDK (Planned)

**Purpose:** Programmatic library for integrating ZeroPass into other applications.

**Current Status:** Empty, reserved for Phase 2.

**Planned API:**
```go
client := zeropass.NewClient("~/.zeropass")
client.Init(password)
client.Unlock(password)
items, err := client.Search("api") // Returns decrypted items
```

---

## Services Layer

### `/services/syncserver/` — Self-Hosted Sync Server

**Purpose:** REST API server for multi-device sync, stores encrypted items and metadata.

**Main Entry Point:** `main.go`

**Flags:**
- `--port` — HTTP listen port (default: 8443)
- `--db-path` — SQLite database path (default: `./sync.db`)
- `--api-key` — Bearer token for authentication (optional)

**Database Schema:**
```sql
CREATE TABLE sync_items (
  id TEXT PRIMARY KEY,
  device_id TEXT NOT NULL,
  encrypted_data BLOB NOT NULL,
  version_vector JSON NOT NULL,
  updated_at INTEGER NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE devices (
  device_id TEXT PRIMARY KEY,
  name TEXT,
  public_key TEXT,
  last_sync_at INTEGER,
  created_at INTEGER
);
```

**HTTP Endpoints:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/sync/register` | POST | Register device |
| `/sync/pull` | POST | Download changed items |
| `/sync/push` | POST | Upload changed items |
| `/health` | GET | Server health check |

**Security:**
- Optional bearer token auth via `--api-key` flag
- All item data encrypted (server cannot read)
- WAL mode (write-ahead logging) for crash recovery
- CORS headers configurable

---

## Dependencies

**Production Dependencies:**
- `cobra` — CLI framework
- `x/crypto` — Argon2id, SHA-256, HKDF
- `x/term` — Secure password input
- `modernc.org/sqlite` — Pure Go SQLite driver
- `go-bip39` — BIP-39 mnemonic generation
- `zxcvbn-go` — Password strength scoring
- `clipboard` — Clipboard integration

**Test Dependencies:**
- `testify/assert` — Test assertions (table-driven tests)
- `testify/require` — Test requirements

**No external API calls** (except optional HIBP, k-anonymous model).

---

## Code Organization Principles

### Interface-Based DI
All major components expose interfaces:
```go
// crypto/interfaces.go
type KDF interface {
  Derive(password string, salt []byte) ([32]byte, error)
}

type Vault interface {
  Create(password string) error
  Unlock(password string) error
  AddItem(item *Item) error
  // ...
}
```

### Per-Item Encryption
Each credential encrypted with unique key (HKDF + item ID), preventing bulk decryption.

### FTS5 Indexing
SQLite full-text search on metadata (service name, username, tags) with query sanitization.

### Memory Safety
All cryptographic keys wiped immediately after use:
```go
defer ZeroBytes(&key)
```

---

## File Statistics

**Total Files:** ~70 Go files  
**Total LOC:** ~20,000 (excluding tests, vendor, .git)

**Distribution:**
- `core/crypto/` — ~4,500 LOC
- `core/vault/` — ~6,500 LOC
- `core/sync/` — ~2,500 LOC
- `packages/cli/` — ~5,000 LOC
- `services/syncserver/` — ~1,500 LOC
- Tests — ~5,000 LOC

---

## Entry Points

### CLI
```bash
go build -o zeropass ./packages/cli/
./zeropass init
./zeropass add --type=login
```

### Sync Server
```bash
go build -o syncserver ./services/syncserver/
./syncserver --port=8443 --api-key=secret
```

### Library (Planned SDK)
```go
import "github.com/zeropass/zeropass/core/vault"
v := vault.New(vaultPath)
v.Create(password)
```

---

## Key Design Patterns

1. **Interface-driven architecture** — Swappable implementations for crypto, storage, sync
2. **Defer cleanup** — Immediate memory wiping via deferred `ZeroBytes()` calls
3. **Type safety** — Fixed-size arrays for keys (`[32]byte`, `[16]byte`)
4. **Error wrapping** — `fmt.Errorf("%w", err)` for stack traces
5. **Table-driven testing** — `[]struct { name, input, expected }` patterns
6. **Single responsibility** — Each module handles one concern (KDF, Cipher, Item CRUD, etc.)

---

**Document Version:** 1.1  
**Last Updated:** June 2025  
**Lines of Code:** ~20,000 (excluding tests)
