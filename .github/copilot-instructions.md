# Copilot Instructions — ZeroPass

ZeroPass is an offline-first, zero-knowledge secrets manager. The primary interface is the `zp` CLI (Go/Cobra). A macOS SwiftUI app consumes Go via a CGO C-archive bridge. Sync is an optional self-hosted preview.

## Build, Test, Lint

```bash
# Build
go build -o zp ./packages/cli/
go build -o syncserver ./services/syncserver/

# Test
go test ./...                                # all tests
go test ./core/crypto/...                    # one package tree
go test ./core/vault/store/ -run TestUnlock  # single test
go test -race ./...                          # race detector
go test -coverprofile=coverage.out ./...     # coverage

# Lint / vet
go vet ./...
```

### macOS app (requires Xcode + Go)

```bash
# Unit tests (no signing)
xcodebuild test \
  -project apps/macos/ZeroPass/ZeroPass.xcodeproj \
  -scheme ZeroPass \
  -destination 'platform=macOS,arch=arm64' \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""

# Unsigned release archive → dist/macos/
CODE_SIGNING_ALLOWED=NO bash apps/macos/scripts/build.sh
```

### Bridge (CGO library for macOS app)

```bash
cd bridge && make build-universal   # arm64 + amd64 → libzeropass.a
```

### Sync server smoke test

```bash
bash services/syncserver/smoke-test.sh both   # binary + Docker
```

## Architecture

```
packages/cli/cmd/        Cobra commands → the `zp` CLI
core/
  crypto/                KDF, cipher, key hierarchy, recovery, password gen
  vault/                 Store lifecycle, item CRUD, FTS5 index, versioning, health, import/export
  sync/                  Delta sync client/server, conflict resolution, protocol types
services/syncserver/     Self-hosted sync HTTP server (SQLite or Postgres backend)
bridge/                  CGO C-archive exposing Go to Swift via C API
apps/macos/              SwiftUI macOS app consuming the bridge
```

### Key hierarchy (encryption flow)

```
Password → Argon2id → Master Key (256-bit)
  → encrypts Vault Key (random 256-bit, stored in vault.json)
    → HKDF-SHA256(Vault Key, Item ID) → Item Key (unique per item)
      → AES-256-GCM(item JSON, Item Key, random nonce) → encrypted item + SHA-256 checksum
```

Each item has its own derived key — compromising one item does not expose others. Recovery uses a BIP-39 12-word mnemonic with auto-rotation (one-time use).

### Vault on disk

```
~/.zeropass/vaults/default/
├── vault.json              Metadata (salt, encrypted vault key, config)
├── items/{id}.json         Encrypted items
├── items/{id}.versions.json  Version history
├── index.db                SQLite FTS5 search index (WAL mode)
└── index.db.enc            Encrypted index (when locked)
```

### Interface-driven DI

Major components define interfaces in `interfaces.go` (e.g., `KDF`, `Cipher`, `RecoveryManager`). Concrete implementations are unexported. Managers accept dependencies (vault key function, indexer) as constructor arguments.

### Bridge pattern

Go compiles to a C-archive (`libzeropass.a`). The bridge uses a thread-safe handle registry (`sync.Map`) to track open vault sessions. `ZPFree`/`ZPFreeResult` wipe memory before deallocation. The bridge disables Go's auto-lock timer — the host app is responsible for lock policy.

### Sync protocol (preview)

Delta sync via Pull/Push with conflict resolution: last-write-wins → version tiebreak → server-wins. The server stores encrypted blobs — it never sees plaintext. End-to-end local merge path is not yet complete.

## Conventions

### Security rules (non-negotiable)

- **Zero sensitive data after use**: `defer ZeroBytes(&key)` immediately after deriving/decrypting. Use `runtime.KeepAlive()` after zeroing to prevent compiler elision.
- **Constant-time comparison**: Always `crypto/subtle.ConstantTimeCompare()` for passwords, checksums, tokens.
- **Randomness**: Exclusively `crypto/rand`. Never `math/rand`.
- **Fixed-size key arrays**: `[32]byte`, `[16]byte` — stack-allocated, not heap slices.
- **No plaintext in logs**: Never log passwords, keys, mnemonics, or decrypted fields.
- **Hardened constants**: Security parameters (Argon2id memory=64MB, iterations=3, parallelism=4) are compile-time constants, not configurable.
- **Per-operation nonces**: AES-GCM requires a unique random nonce per encryption; never reuse.

### Error handling

- Sentinel errors as `Err{Concept}` in dedicated `errors.go` files (e.g., `ErrInvalidPassword`, `ErrVaultNotFound`).
- Wrap with context: `fmt.Errorf("derive master key: %w", err)`.
- Inspect with `errors.Is()`.

### Naming

- Packages: lowercase, no underscores (`crypto`, `vault`, `sync`).
- Files: snake_case for multi-word (`aes_gcm.go`).
- Exported functions: verb-first (`Encrypt`, `NewItem`).
- Receiver names: single character (`v` for Vault, `i` for Item).

### Testing

- **Table-driven tests** with `t.Run()` subtests.
- `testify/require` for setup (fail-fast); `testify/assert` for assertions (continue on failure).
- Tests live in the same package (`{file}_test.go`), not a separate `tests/` directory.
- Coverage targets: 90%+ for `core/crypto` and `core/vault`; 80%+ for CLI and sync; 100% for security-critical paths.
- Database mocks via `go-sqlmock`.

### Item types and fields

Item types: `login`, `apikey`, `sshkey`, `note`, `creditcard`, `identity`, `passkey`, `custom`. Each type has standard field constants (e.g., `FieldUsername`, `FieldAPIKey`, `FieldPrivateKey`) defined in `core/vault/types/types.go`.

### Vault metadata persistence

`vault.json` uses atomic-replace writes (platform-specific: `atomic_replace_unix.go` / `atomic_replace_windows.go`). Unknown fields are preserved on read/write for forward compatibility.

### Dependencies

- CLI: `spf13/cobra`
- Crypto: `golang.org/x/crypto` (Argon2id, HKDF)
- SQLite: `modernc.org/sqlite` (pure Go, no CGO for main builds)
- Recovery: `go-bip39`
- Password strength: `zxcvbn-go`
- Clipboard: `atotto/clipboard`
- Sync server Postgres: `lib/pq`

No external API calls at runtime. HIBP is opt-in and uses k-anonymity (only first 5 chars of SHA-1 hash sent).
