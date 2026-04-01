---
title: ZeroPass PRD
version: 2.0
author: Justin
last_updated: 2026-04-01
status: Draft
changelog:
  - v1.0 (2026-03): Initial draft
  - v2.0 (2026-04-01): Deep research integration, competitive analysis, security hardening, developer DX spec
---

# ZeroPass – Product Requirements Document (PRD)

## 1. Problem Statement (WHY)

### Pain Points (Research-Backed)

**For Developers:**

1. **Secrets Sprawl** — API keys, tokens, SSH credentials phân tán qua `.env` files, CI/CD vars, Slack, email. Không thể track, rotate, revoke hiệu quả. Trung bình mỗi developer quản lý 50+ credentials across 10+ services.
2. **Hardcoded Secrets** — Pressure to ship fast → credentials in source code → exposed within minutes by automated GitHub scanners (gitleaks, truffleHog bots).
3. **No Developer-Native PM** — 1Password có CLI mạnh nhưng proprietary + đắt ($36/yr). Bitwarden CLI basic. KeePass powerful nhưng UX legacy. Không có PM nào coi CLI là first-class citizen.
4. **"Secret Zero" Problem** — Dùng vault để lưu secrets, nhưng authenticate tới vault bằng gì? Vòng lặp đệ quy credentials.
5. **Multi-Environment Chaos** — Dev/staging/production secrets cần tách biệt nhưng quản lý manual dẫn đến config sai environment.

**For Privacy-Focused Users:**

1. **Cloud Trust Erosion** — LastPass 2022 breach: encrypted vaults bị đánh cắp, crypto thefts xảy ra nhiều năm sau. UK ICO phạt tiền + class-action. Users mất niềm tin vào cloud PMs.
2. **Vendor Lock-in** — Commercial PMs giữ data hostage. Export/migration is painful.
3. **Metadata Leakage** — Even "zero-knowledge" PMs leak unencrypted URLs, item names, timestamps — cho phép attacker profile user behavior.
4. **Password Fatigue** — 250+ accounts per user average. Password reuse vẫn phổ biến dù users biết rủi ro.

### Why Now

* Explosion of API keys sau AI/LLM boom (mỗi AI service = 1+ API key)
* ETH Zürich 2026 research: "malicious server" bypass zero-knowledge encryption during sync → offline-first design giải quyết triệt để
* NIST SP 800-63B Rev. 4 tăng yêu cầu credential security
* Passkey/FIDO2 transition tạo cơ hội cho "credential manager" thay vì chỉ "password manager"
* Growing demand for self-sovereign, auditable security tools

### Market Opportunity

* Password management market: $2.05B (2024) → projected $7.13B (2030)
* Developer-specific segment underserved — current solutions are either too enterprise (HashiCorp Vault) or too consumer (1Password)
* Open-source trust advantage post-LastPass breach

---

## 2. Product Vision & Concept

### Vision

ZeroPass is a **developer-first, zero-knowledge credential manager** designed for speed, security, and extensibility. CLI-native, offline-first, open-source.

### Product Concept

* **Offline-first vault** — works fully without internet
* **Zero-knowledge encryption** — server (if used) cannot decrypt data
* **CLI-native** — `zeropass` CLI is first-class, GUI is complementary
* **Modular architecture** — pluggable crypto, storage, sync layers
* **Developer secrets management** — native support for API keys, SSH keys, .env files, tokens

### Unique Value Proposition

> "The developer's command-line vault: offline-first, zero-knowledge, secrets management that fits your terminal workflow."

| Differentiator | ZeroPass | Others |
|---|---|---|
| CLI-first workflow | `zeropass run --env-file` native | CLI is afterthought |
| Truly offline | Full functionality without internet | Require cloud for core features |
| Dev secrets native | API keys, SSH keys, .env injection | Designed for passwords only |
| Self-sovereign | User owns 100% data, no vendor dependency | Data on vendor servers |
| Pluggable crypto | Community can audit/replace crypto layer | Black box |

---

## 3. Target Users

### Persona 1: "DevOps Dan" (Primary)

* **Role:** Full-stack developer, runs multiple side projects and freelance clients
* **Tools:** VS Code, Terminal (zsh/fish), Git, Docker, AWS/GCP
* **Pain:** Has 50+ API keys scattered across `.env` files, sticky notes, and a random Google Doc. Lost an AWS key last month → $200 surprise bill. Shares credentials with contractors via Slack DM.
* **JTBD:** Store, retrieve, and inject secrets into dev workflow seamlessly without leaving the terminal

### Persona 2: "Security-First Sarah" (Primary)

* **Role:** Privacy advocate, runs own server infrastructure, contributes to open-source
* **Tools:** Linux/macOS, self-hosted services, GPG, YubiKey
* **Pain:** Tried KeePass but UI is unbearable. Doesn't trust 1Password's cloud. Wants to audit the crypto implementation herself.
* **JTBD:** Self-sovereign credential management with auditable, open-source crypto

### Persona 3: "Indie Ian" (Secondary)

* **Role:** Solo founder, manages 15 SaaS subscriptions for his startup
* **Tools:** macOS, Chrome, basic terminal usage
* **Pain:** Reuses same 3 passwords everywhere. Knows it's wrong but PM setup feels overwhelming. Scared of forgetting master password.
* **JTBD:** Simple, fast migration from browser-saved passwords to a secure vault with safety net (recovery key)

### Jobs To Be Done

* Store and retrieve credentials securely (< 2 seconds end-to-end)
* Inject secrets into development workflow without plaintext `.env` files
* Manage secrets across dev/staging/production environments
* Import existing credentials from browsers and other PMs effortlessly
* Recover vault access if master password is forgotten

---

## 4. Goals & Success Metrics

### Product Metrics (Phase 1)

| Metric | Target | How to Measure |
|---|---|---|
| Vault unlock latency | < 300ms | Client instrumentation |
| Search latency | < 50ms | Client instrumentation |
| Onboarding completion rate | > 80% | Analytics: setup → first credential stored |
| CLI daily invocations | Track per user | `zeropass` command telemetry (opt-in) |
| Import success rate | > 95% | Import flow analytics |

### Security Metrics

| Metric | Target | How to Measure |
|---|---|---|
| Security incidents | 0 | Incident tracking |
| Time to patch critical vuln | < 48 hours | Release process |
| Code coverage (crypto module) | > 95% | CI/CD |
| Third-party audit findings | 0 critical | Annual audit |

### Community Metrics (Open Source — Year 1)

| Metric | Target |
|---|---|
| GitHub stars | 1,000+ |
| Contributors | 20+ |
| Community bug reports resolved | Track velocity |
| Documentation coverage | 100% of public APIs |

---

## 5. Scope Definition

### In Scope (MVP — Phase 1)

* Local encrypted vault (AES-256-GCM, per-item encryption)
* Crypto engine (Argon2id KDF + AES-256-GCM)
* CLI tool (`zeropass`) — core commands
* Basic macOS app (SwiftUI vault browser)
* Import from Chrome, Firefox, 1Password, Bitwarden (CSV)
* Export encrypted vault backup
* Password/passphrase generator
* Recovery key generation (BIP-39 mnemonic)
* Full-text search across all fields
* Tagging & categories
* Item types: Login, API Key, SSH Key, Secure Note, Credit Card, Identity
* Master password strength enforcement (zxcvbn-based)
* Clipboard auto-clear (30s)
* Auto-lock on idle (configurable)

### Out of Scope (Deferred)

* Enterprise admin dashboard (Phase 4)
* Team collaboration / shared vaults (Phase 3)
* Mobile apps — iOS/Android (Phase 4)
* Browser autofill extension (Phase 3)
* Cloud sync service (Phase 2 — self-hosted only)
* Passkey provider (Phase 3)

---

## 6. Competitive Analysis

### Competitive Matrix

| Feature | **ZeroPass** | **Bitwarden** | **1Password** | **KeePass** |
|---|---|---|---|---|
| Zero-knowledge | ✅ Core | ✅ | ✅ + Secret Key | ✅ |
| Offline-first | ✅ Core | ❌ Cloud-first | ❌ Cloud-first | ✅ |
| Open Source | ✅ MIT/Apache | ✅ AGPLv3 | ❌ Proprietary | ✅ GPL |
| Self-hostable | ✅ Planned | ✅ | ❌ | ✅ Local |
| Developer CLI | ✅ First-class | ✅ Basic | ✅ `op` (strong) | ❌ |
| .env injection | ✅ Native | ❌ | ✅ `op run` | ❌ |
| Autofill | Phase 3 | ✅ | ✅ Best-in-class | ✅ Plugins |
| Passkey support | Phase 3 | ✅ | ✅ | ❌ |
| Team sharing | Phase 3 | ✅ | ✅ Best-in-class | ❌ |
| Recovery mechanism | ✅ BIP-39 | ✅ | ✅ | ❌ |
| Price | Free | Free + $10/yr | $36/yr | Free |
| UI/UX quality | Modern (planned) | Good | Excellent | Poor/Legacy |

### Positioning Strategy

Avoid head-to-head competition with Bitwarden/1Password on consumer features. Focus on **sharp angle differentiation**:

1. **CLI-native** — `zeropass` CLI is the PRIMARY interface, not an add-on
2. **Offline-first** — full functionality without internet (unlike Bitwarden/1Password)
3. **Developer secrets** — native `.env` injection, SSH key agent, environment management
4. **Self-sovereign** — user owns 100% of data, zero vendor dependency
5. **Auditable crypto** — pluggable, open-source crypto layer

---

## 7. User Experience (High-Level)

### Core Flows (MVP)

1. **First-Time Setup**
   * Create master password (with strength meter — zxcvbn)
   * Generate & display recovery key (BIP-39 12-word mnemonic)
   * Confirm recovery key written down (verification step)
   * Create first vault → guided "add first credential" flow

2. **Unlock Vault**
   * Enter master password OR biometric (TouchID on macOS)
   * Auto-lock after configurable idle time (default: 5 min)

3. **Add Credential (GUI)**
   * Click "+" → select type (Login / API Key / SSH Key / Secure Note / Credit Card / Identity)
   * Fill fields → auto-generate password if login type
   * Tag, categorize, add to favorites

4. **Add Credential (CLI)**
   * `zeropass add --type=apikey --name="AWS Prod" --value="AKIA..."`
   * `zeropass add --type=login --name="GitHub" --generate-password`

5. **Search & Retrieve**
   * GUI: Cmd+K spotlight-style search across all fields
   * CLI: `zeropass get "github" --copy` → copies to clipboard (auto-clear 30s)
   * CLI: `zeropass search "aws"` → lists matching items

6. **Inject Secrets to Dev Workflow**
   * `zeropass run --env-file=.env -- npm start`
   * Replace `zp://vault/item/field` references with real values at runtime
   * Secrets exist only in memory, never written to disk

7. **Recovery Flow**
   * Enter BIP-39 mnemonic → derive master key → access vault
   * Set new master password → re-encrypt vault key

### UX Principles

* **Fast** — < 300ms for all interactions
* **Keyboard-first** — vim-style keybindings, Cmd+K search
* **CLI-first** — terminal is primary interface for developers
* **Zero-config start** — works out of the box, no account creation needed
* **Progressive disclosure** — simple by default, powerful when needed
* **Secure by default** — auto-lock, clipboard clear, strength enforcement

---

## 8. Functional Requirements

### 8.1 Vault Management

* CRUD items with multiple types (Login, API Key, SSH Key, Secure Note, Credit Card, Identity)
* Custom fields support (key-value pairs)
* Tagging & smart categories (auto-categorize by type)
* Favorites & recently used
* Full-text search across title, username, URL, tags, notes
* Sort & filter (by type, date modified, tag, favorite)
* Vault lock/unlock with configurable auto-lock timer (1/5/15/30 min)
* Item version history (track changes)

### 8.2 CLI Tool (`zeropass`)

```
zeropass init                              # Setup new vault
zeropass unlock                            # Unlock vault (biometric or password)
zeropass lock                              # Lock vault
zeropass add --type=<type> --name=<name>   # Add credential
zeropass get <query> [--copy|--json]       # Retrieve credential
zeropass search <query>                    # Search credentials
zeropass list [--type=<type>] [--tag=<tag>]# List items
zeropass edit <query>                      # Edit credential
zeropass delete <query>                    # Delete credential
zeropass run --env-file=<file> -- <cmd>    # Inject secrets & run command
zeropass run --env=<environment> -- <cmd>  # Inject environment secrets
zeropass export --format=[json|csv|encrypted]
zeropass import --from=[1password|bitwarden|chrome|firefox|csv]
zeropass generate [--length=32] [--no-symbols] [--passphrase]
zeropass health                            # Password health report
zeropass recovery                          # Show/regenerate recovery key
```

* Shell completions: zsh, bash, fish
* Secret reference format: `zp://[vault]/[item]/[field]`
* JSON output mode for scripting: `--output=json`
* Interactive TUI mode: `zeropass` (no args)

### 8.3 Encryption

* **KDF:** Argon2id (memory-hard, resistant to GPU/ASIC attacks)
  * Configurable parameters: memory=64MB, iterations=3, parallelism=4
* **Cipher:** AES-256-GCM (authenticated encryption)
  * Unique nonce/IV per encryption operation (prevent pattern analysis)
* **Client-side only** — all crypto operations on user device
* **Metadata encryption** — encrypt item names, URLs, timestamps (lesson from LastPass)

### 8.4 Password Generation

* Random password: configurable length (8-128), character sets
* Passphrase mode: diceware word list (4-8 words, configurable separator)
* Copy to clipboard with auto-clear (configurable: 15/30/60s)
* Password strength indicator (zxcvbn-based)

### 8.5 Recovery System

* BIP-39 mnemonic phrase (12 words) generated at vault creation
* Recovery key file option (encrypted PDF export)
* Recovery flow: mnemonic → derive recovery key → decrypt vault key → set new master password
* One-time use: recovery key is rotated after successful recovery

### 8.6 Security Features

* Master password strength enforcement (minimum: length ≥ 12, entropy check via zxcvbn score ≥ 3)
* Breach detection integration (Have I Been Pwned API — opt-in, k-anonymity model)
* Password health dashboard (weak, reused, old passwords)
* Clipboard auto-clear after 30 seconds (configurable)
* Auto-lock on idle (configurable: 1/5/15/30 min, default 5 min)
* No plaintext in memory longer than needed — zeroise after use
* No sensitive data in logs

### 8.7 Import/Export

* **Import from:** Chrome, Firefox, Safari, 1Password (1PUX/CSV), Bitwarden (JSON/CSV), KeePass (KDBX/CSV), LastPass (CSV), generic CSV
* **Export:** Encrypted vault backup (ZeroPass format), unencrypted CSV (with warning), JSON

### 8.8 Storage

* Encrypted local JSON blobs (one file per vault)
* Each item encrypted individually with unique key derived from vault key
* SQLite index for search (encrypted metadata index)

---

## 9. Non-Functional Requirements

### Security

* Zero-knowledge architecture — no plaintext storage, no sensitive logs
* All data encrypted at rest (AES-256-GCM)
* Metadata encrypted (URLs, item names, timestamps)
* Memory security — clear sensitive data from memory immediately after use
* Constant-time comparison operations (prevent timing attacks)
* No telemetry of vault contents (ever)

### Performance

* Vault unlock: < 300ms (after KDF, which is intentionally slow: ~500ms-1s)
* Search: < 50ms across 10,000 items
* CLI command response: < 200ms (excluding unlock)
* App launch to ready: < 1 second

### Reliability

* Zero data loss — atomic writes, WAL journaling
* Graceful degradation — works fully offline
* Backup on every vault modification (configurable retention)
* Conflict-safe sync with versioning (Phase 2)

### Compatibility

* macOS 13+ (Ventura and later)
* CLI: macOS, Linux (x86_64, arm64)
* Future: Windows, iOS, Android

---

## 10. System Architecture (High-Level)

### Modules

```
zeropass/
├── apps/
│   ├── macos/              # SwiftUI macOS app
│   ├── web/                # Next.js web app (Phase 3)
│   └── extension/          # Browser extension (Phase 3)
├── core/
│   ├── crypto-engine/      # Argon2id KDF + AES-256-GCM
│   ├── vault-engine/       # CRUD, search, versioning
│   ├── sync-engine/        # Delta sync, conflict resolution (Phase 2)
│   └── autofill-engine/    # Browser autofill logic (Phase 3)
├── services/
│   └── sync-server/        # Self-hosted sync backend (Phase 2)
├── packages/
│   ├── sdk/                # ZeroPass SDK (Go library)
│   └── cli/                # `zeropass` CLI tool
└── docs/
    ├── security-whitepaper.md
    └── architecture.md
```

### Platform Strategy

| Platform | Technology | Phase |
|---|---|---|
| CLI | Go (cross-platform binary) | Phase 1 |
| macOS App | SwiftUI + Go core (via FFI/cgo) | Phase 1 |
| Sync Server | Go + Postgres + S3/MinIO | Phase 2 |
| Browser Extension | TypeScript + WebExtensions API | Phase 3 |
| Web App | Next.js + WASM (Go crypto compiled to WASM) | Phase 3 |
| iOS/Android | Swift/Kotlin + Go core | Phase 4 |

---

## 11. Security Model & Threat Analysis

### Key Hierarchy

```
Master Password
  → Argon2id KDF (salt=random, memory=64MB, iter=3, par=4)
  → Master Key (256-bit)
    → Encrypts Vault Key (random 256-bit, one per vault)
      → Derives Item Keys (HKDF from vault key + item ID)
        → AES-256-GCM per-item encryption (unique nonce per operation)

Recovery Key (BIP-39 mnemonic)
  → Derives Recovery Master Key
    → Can decrypt Vault Key (backup path)
```

### Vault Encryption Model

Each item encrypted individually:

```json
{
  "id": "uuid-v4",
  "type_encrypted": "aes-gcm-ciphertext",
  "data_encrypted": "aes-gcm-ciphertext",
  "meta_encrypted": "aes-gcm-ciphertext",
  "nonce": "unique-per-encryption",
  "version": 3,
  "created_at_encrypted": "aes-gcm-ciphertext",
  "updated_at_encrypted": "aes-gcm-ciphertext"
}
```

> **Note:** Unlike LastPass, ALL metadata (type, timestamps, URLs) is encrypted. Only the item `id` and `version` are plaintext for sync purposes.

### Threat Model

| Threat | Protection | Status |
|---|---|---|
| Server compromise | Zero-knowledge: server has no decryption keys | ✅ By design |
| Network interception | All sync data encrypted before transmission | ✅ Phase 2 |
| Database theft | All stored data AES-256-GCM encrypted | ✅ Phase 1 |
| Malicious server (ETH Zürich 2026) | Offline-first: no server dependency | ✅ By design |
| Weak master password | zxcvbn enforcement ≥ score 3, min 12 chars | ✅ Phase 1 |
| Memory scraping malware | Zeroise sensitive data after use | ✅ Phase 1 |
| Timing attacks | Constant-time comparison operations | ✅ Phase 1 |
| Metadata profiling | Full metadata encryption | ✅ Phase 1 |

### What Requires User Responsibility

* Master password strength (policy enforced but user chooses)
* Recovery key storage (user must store safely offline)
* Device security (ZeroPass cannot protect against keyloggers/malware on device)
* Phishing (ZeroPass does not prevent entering credentials on fake sites — until autofill Phase 3)

### Security Audit Plan

| Phase | Activity |
|---|---|
| Phase 1 | Internal code review + automated SAST scanning (gosec, semgrep) |
| Phase 2 | Third-party penetration testing (crypto module focus) |
| Phase 3 | Public bug bounty program (HackerOne or self-hosted) |
| Annual | Independent security audit — results published publicly |

### Security Rules

* No plaintext in memory longer than needed — zeroise immediately
* Zero logging of sensitive data (passwords, keys, mnemonics)
* Key rotation supported (re-encrypt vault key with new master key)
* All randomness from crypto/rand (CSPRNG), never math/rand

---

## 12. Technical Specification (Crypto + Sync)

### 12.1 Crypto Engine Design

#### Module Structure

```
crypto-engine/
├── kdf/
│   └── argon2id.go         # Argon2id key derivation
├── cipher/
│   └── aes_gcm.go          # AES-256-GCM encrypt/decrypt
├── key/
│   ├── master_key.go       # Master key derivation
│   ├── vault_key.go        # Vault key generation/encryption
│   ├── item_key.go         # Per-item key derivation (HKDF)
│   └── recovery_key.go     # BIP-39 mnemonic generation/recovery
├── encoding/
│   └── base64.go           # Safe encoding utilities
├── password/
│   ├── generator.go        # Password/passphrase generation
│   └── strength.go         # zxcvbn strength scoring
├── interfaces.go
└── crypto_test.go           # >95% coverage required
```

#### Interfaces

```go
type KDF interface {
    DeriveKey(password []byte, salt []byte) ([]byte, error)
}

type Cipher interface {
    Encrypt(plaintext []byte, key []byte) ([]byte, error)
    Decrypt(ciphertext []byte, key []byte) ([]byte, error)
}

type RecoveryManager interface {
    GenerateMnemonic() (string, error)
    DeriveKeyFromMnemonic(mnemonic string) ([]byte, error)
}

type PasswordGenerator interface {
    GenerateRandom(length int, opts GeneratorOptions) (string, error)
    GeneratePassphrase(words int, separator string) (string, error)
    ScoreStrength(password string) (int, string, error)  // score 0-4, feedback
}
```

#### Unlock Flow

1. User inputs master password
2. Retrieve salt from vault metadata
3. Derive master key via Argon2id (salt, memory=64MB, iter=3, par=4)
4. Decrypt vault key using master key
5. Vault key used to derive per-item keys (HKDF)
6. Per-item keys decrypt individual items on-demand (lazy decryption)

### 12.2 Vault Engine

#### Module Structure

```
vault-engine/
├── store.go          # Vault storage operations
├── item.go           # Item CRUD
├── index.go          # Full-text search index (encrypted)
├── version.go        # Item versioning
├── types.go          # Item type definitions
├── import.go         # Import from other PMs
├── export.go         # Export vault
└── health.go         # Password health analysis
```

#### Item Types

```go
type ItemType string

const (
    ItemTypeLogin      ItemType = "login"
    ItemTypeAPIKey     ItemType = "apikey"
    ItemTypeSSHKey     ItemType = "sshkey"
    ItemTypeSecureNote ItemType = "note"
    ItemTypeCreditCard ItemType = "creditcard"
    ItemTypeIdentity   ItemType = "identity"
    ItemTypeCustom     ItemType = "custom"
)
```

#### Versioning

* Each item has monotonic version counter
* Increment on every update
* Keep last N versions (configurable, default: 10)
* Version history enables undo and sync conflict resolution

### 12.3 Sync Engine Design (Phase 2)

#### Goals

* Eventually consistent
* Conflict-safe
* Offline-first (sync is optional, not required)

#### Module Structure

```
sync-engine/
├── client/
│   └── sync_client.go
├── server/
│   └── sync_handler.go
├── protocol/
│   └── sync.proto
├── conflict/
│   └── resolver.go
```

#### Data Model

```json
{
  "item_id": "uuid",
  "version": 3,
  "device_id": "device-uuid",
  "payload": "encrypted_blob",
  "timestamp": 1711929600,
  "checksum": "sha256-of-encrypted-payload"
}
```

#### Sync Strategy

* Delta sync (only changed items since last sync timestamp)
* Pull + Push model
* All payloads encrypted BEFORE leaving device

#### Sync Flow

1. Client sends last sync timestamp + device ID
2. Server returns items updated since timestamp
3. Client merges (conflict resolution)
4. Client pushes local changes
5. Server acknowledges with new server timestamp

### 12.4 Conflict Resolution

* **Phase 2 (MVP sync):** Last-write-wins (compare timestamps)
* **Phase 4:** CRDT-based (operation-transform for concurrent edits)

Conflict case: same item edited on 2 devices → compare version + timestamp → keep latest → store conflicted version in history

### 12.5 Sync API

```
GET  /sync/pull?since=<timestamp>&device_id=<id>
POST /sync/push
     Body: { items: [...encrypted_items], device_id: "..." }

Response: { items: [...], server_time: 1711929600, conflicts: [...] }
```

### 12.6 Storage Layer

| Layer | Technology |
|---|---|
| Client vault | Encrypted JSON blobs (local filesystem) |
| Client search index | SQLite with encrypted columns |
| Sync server metadata | PostgreSQL |
| Sync server blobs | S3/MinIO (encrypted payloads) |

---

## 13. Developer Experience (DX) Specification

### Design Principles

1. **"Secure way = easiest way"** — security must not add friction to dev workflow
2. **CLI is first-class citizen** — not an afterthought or plugin
3. **Works WITH existing tools** — integrates into shell, editor, CI/CD
4. **Zero-configuration start** — `zeropass init` and you're running
5. **Progressive customization** — simple defaults, power when needed

### Secret Reference Format

```
zp://[vault-name]/[item-name]/[field-name]
```

Examples:
```
zp://default/aws-prod/access-key
zp://work/github-token/token
zp://default/postgres-staging/password
```

### .env Integration Workflow

**Before ZeroPass:**
```env
# .env (PLAINTEXT - DANGER!)
AWS_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
DATABASE_URL=postgres://user:password@localhost/db
```

**After ZeroPass:**
```env
# .env (SAFE - references only)
AWS_ACCESS_KEY=zp://prod/aws-api/access-key
AWS_SECRET_KEY=zp://prod/aws-api/secret-key
DATABASE_URL=zp://staging/postgres/connection-string
```

```bash
# Inject secrets at runtime (never written to disk)
zeropass run --env-file=.env -- npm start
```

### Environment Management

```bash
# Switch between environments
zeropass run --env=development -- npm start
zeropass run --env=staging -- npm run test:e2e
zeropass run --env=production -- npm run deploy
```

### Integration Points (Roadmap)

| Integration | Phase | Description |
|---|---|---|
| Shell completions | Phase 1 | zsh, bash, fish auto-complete |
| SSH Agent | Phase 3 | ZeroPass as SSH key agent |
| VS Code Extension | Phase 3 | Inline secret reference, auto-complete |
| GitHub Actions | Phase 3 | Service account for CI/CD secrets |
| GitLab CI | Phase 3 | Service account for CI/CD secrets |
| Docker | Phase 3 | Secrets injection via env vars |
| Kubernetes | Phase 4 | Secrets provider (CSI driver) |

---

## 14. User Stories

### Developer Workflow

* As a developer, I want to inject vault secrets into my dev server so that I never have plaintext `.env` files
* As a developer, I want to store SSH keys in my vault so that I have a single secure location for all credentials
* As a developer, I want a CLI that works in my terminal so that I don't have to context-switch to a GUI app
* As a developer, I want to import my existing `.env` files so that migration is effortless
* As a developer, I want to manage secrets per environment (dev/staging/prod) so that I never mix up credentials
* As a developer, I want shell completions so that I can tab-complete vault item names

### Security

* As a user, I want my data encrypted locally so that no one else can read it
* As a user, I want a recovery key generated at setup so that I can recover my vault if I forget my master password
* As a user, I want my vault auto-locked after inactivity so that my credentials are safe if I walk away
* As a user, I want to know if my passwords have been in data breaches so that I can change them proactively
* As a user, I want clipboard auto-cleared after copying a password so that it's not exposed indefinitely
* As a user, I want all metadata encrypted so that even leaked vault files reveal nothing

### Onboarding

* As a new user, I want to import from Chrome/Firefox/Safari so that I can switch without re-entering passwords
* As a new user, I want a setup wizard that takes < 2 minutes so that I'm not overwhelmed by configuration
* As a new user, I want to understand WHY offline-first is safer so that I trust the product

### Vault Management

* As a user, I want to create a vault so that I can store credentials securely
* As a user, I want to search credentials instantly across all fields
* As a user, I want custom item types so that I can store credit cards, identities, API keys, and secure notes
* As a user, I want to tag and categorize items so that my vault stays organized as it grows
* As a user, I want to see password health metrics so that I know which passwords need updating

---

## 15. Risks & Mitigations

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|---|---|---|---|
| Crypto implementation bugs | Critical | Medium | Third-party audit, >95% test coverage, well-known Go crypto libraries |
| Memory-scraping malware | High | Low | Zeroise sensitive data after use, document as user responsibility |
| Side-channel attacks | High | Low | Argon2id, constant-time operations, no branching on secrets |
| Data loss during sync | High | Medium | Atomic writes, version history, conflict-safe merge |

### Product Risks

| Risk | Impact | Probability | Mitigation |
|---|---|---|---|
| Low adoption (saturated market) | High | High | Sharp niche focus: developer CLI-first |
| "Another KeePass?" perception | Medium | Medium | Modern UX + CLI integration = clear differentiation |
| Passkey disruption reduces PM relevance | Medium | Medium | Evolve into "credential manager" (passwords + passkeys + secrets) |
| UX too complex for non-technical users | Medium | Medium | Progressive disclosure, guided onboarding |

### Business Risks

| Risk | Impact | Probability | Mitigation |
|---|---|---|---|
| No revenue → unmaintained OSS | High | High | Open-core model, premium features for teams |
| Single vuln kills trust | Critical | Low | Bug bounty, annual audits, responsible disclosure policy |
| Key contributor leaves | Medium | Medium | Modular architecture, documentation, bus factor > 1 |

### Assumptions

* Users increasingly distrust cloud-hosted credential storage post-breach era
* Developers prefer keyboard-first, terminal-native tools
* Open-source transparency builds trust faster than marketing
* Self-sovereign data ownership is a growing market demand

---

## 16. Roadmap

### Phase 1: Foundation (Month 1-3)

* [ ] crypto-engine (Argon2id + AES-256-GCM + per-item encryption)
* [ ] vault-engine (CRUD + full-text search + versioning + item types)
* [ ] CLI tool (`zeropass`) — init, add, get, search, list, generate, import/export
* [ ] Local encrypted storage (JSON blobs + SQLite index)
* [ ] Recovery key system (BIP-39 mnemonic)
* [ ] Password/passphrase generator
* [ ] macOS app — basic SwiftUI vault browser
* [ ] Import from CSV / Chrome / Firefox / 1Password / Bitwarden
* [ ] Master password strength enforcement
* [ ] Clipboard auto-clear + auto-lock
* [ ] Security: memory zeroing, no sensitive logs

### Phase 2: Polish & Sync (Month 4-6)

* [ ] macOS app — full-featured UI (search, categories, tags, favorites)
* [ ] Biometric unlock (TouchID)
* [ ] Sync engine — self-hosted server (Go + Postgres + MinIO)
* [ ] Multi-device support (delta sync)
* [ ] Passkey storage & management in vault
* [ ] Password health dashboard (weak, reused, breached)
* [ ] HIBP breach detection integration (opt-in)
* [ ] `.env` runtime injection (`zeropass run`)
* [ ] Environment management (dev/staging/prod)

### Phase 3: Ecosystem (Month 7-12)

* [ ] Browser extension (Chrome, Firefox, Safari)
* [ ] Autofill engine
* [ ] Passkey provider (WebAuthn authenticator)
* [ ] Team/shared vaults (basic RBAC)
* [ ] SSH Agent integration
* [ ] VS Code extension
* [ ] CI/CD service accounts (GitHub Actions, GitLab CI)
* [ ] Web app (Next.js + Go WASM)
* [ ] Emergency access (trusted contacts with waiting period)

### Phase 4: Growth (Year 2)

* [ ] Mobile apps (iOS, Android)
* [ ] Enterprise SSO integration (SAML, OIDC)
* [ ] Audit logs & compliance reporting
* [ ] CRDT-based sync (replaces last-write-wins)
* [ ] Hardware key support (YubiKey, FIDO2)
* [ ] FIDO Credential Exchange (CXP) — import/export passkeys
* [ ] Kubernetes secrets provider
* [ ] Admin dashboard for teams

---

## 17. Monetization & Sustainability

### Open-Core Model

#### Free (Community Edition)

* Full vault (unlimited items, all item types)
* CLI tool (all commands)
* macOS app
* Local storage + encrypted backup
* Import/export (all formats)
* Password generator
* Recovery key
* Self-hosted sync server
* Browser extension
* Community support

#### Premium ($3-5/month — Individual)

* Managed cloud sync (hosted by ZeroPass)
* Password health dashboard + breach monitoring (HIBP)
* Priority email support
* Advanced passkey management
* Emergency access feature

#### Team ($7-10/user/month)

* Shared vaults with RBAC
* Audit logs
* SSO integration (SAML, OIDC)
* Admin dashboard
* Service accounts for CI/CD
* Priority support + SLA
* Centralized policy enforcement

---

## 18. Open Source Strategy

* **License:** MIT (core libraries) / Apache 2.0 (applications)
* **Repository:** Public GitHub monorepo
* **Contributions:** Modular architecture enables focused contributions
* **Security:** Responsible disclosure policy, bug bounty (Phase 3)
* **Documentation:** Security whitepaper, architecture docs, contributor guide
* **Community:** GitHub Discussions, Discord server, monthly dev updates

---

## 19. Open Questions

* ~~Should we support mobile early?~~ → **Decided: Phase 4.** Focus on CLI + macOS first.
* ~~Should sync be self-hosted only?~~ → **Decided: Phase 2 self-hosted, Phase 3+ managed cloud (premium).**
* ~~How to handle key recovery?~~ → **Decided: BIP-39 mnemonic phrase at vault creation.**
* Should we support hardware key (YubiKey) for vault unlock in Phase 2 or defer to Phase 4?
* Should the sync server be compatible with existing protocols (WebDAV, S3) or custom-only?
* What is the minimum viable browser extension scope for Phase 3?

---

## 20. Accessibility & Internationalization

### Accessibility (Phase 1)

* Full keyboard navigation for all UI elements
* Screen reader support (VoiceOver on macOS)
* High contrast mode option
* Minimum font size: 14px
* Focus indicators visible on all interactive elements

### Internationalization (Phase 2+)

* English (primary)
* Vietnamese, Japanese, Chinese (community-contributed)
* RTL support (future consideration)
* Locale-aware date/time formatting

---

## 21. PRD Nature (AI-era)

This PRD is a **living document**:

* Continuously updated with discovery insights and user feedback
* Driven by real-world research (user pain points, competitive analysis, security landscape)
* Versioned with changelog for traceability
* Refined through iterative development cycles

> Modern PRDs are not static specs but evolving alignment tools that bridge product vision with engineering execution.

---

## 22. Implementation Order

1. `crypto-engine` — Argon2id KDF + AES-256-GCM + recovery key
2. `vault-engine` — CRUD + search + versioning + item types
3. `cli` — `zeropass` CLI tool (core commands)
4. Local storage — encrypted JSON + SQLite index
5. macOS app — SwiftUI vault browser
6. Import/export — Chrome, Firefox, 1Password, Bitwarden
7. `sync-engine` — client first, then server (Phase 2)
8. Browser extension — autofill (Phase 3)

> **Principle:** Always build crypto first, vault second, sync last. Never compromise security for speed of delivery.
