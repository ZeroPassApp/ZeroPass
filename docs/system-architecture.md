# ZeroPass: System Architecture

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        ZeroPass (Developer)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  CLI Application (Cobra)                                │   │
│  │  Commands: init, add, get, search, export, recovery...   │   │
│  └────────────────────┬────────────────────────────────────┘   │
│                       │                                          │
│  ┌────────────────────▼────────────────────────────────────┐   │
│  │  Core Layer (Interface-Driven DI)                       │   │
│  │  ╔════════════════╦═════════════╦═══════════════╗      │   │
│  │  ║    Crypto      ║   Vault     ║     Sync      ║      │   │
│  │  ╠════════════════╬═════════════╬═══════════════╣      │   │
│  │  ║ • KDF (Argon)  ║ • Item CRUD ║ • ClntPull    ║      │   │
│  │  ║ • Cipher (AES) ║ • Search    ║ • SrvPush     ║      │   │
│  │  ║ • Key Manage   ║ • Version   ║ • Conflict    ║      │   │
│  │  ║ • RNG/Recovery ║ • Health    ║ • DeviceReg   ║      │   │
│  │  ║ • PassGen      ║ • Import    ║               ║      │   │
│  │  ╚════════════════╩═════════════╩═══════════════╝      │   │
│  └────────────────────┬────────────────────────────────────┘   │
│                       │                                          │
│  ┌────────────────────▼────────────────────────────────────┐   │
│  │  Storage Layer                                          │   │
│  │  • Encrypted JSON files ({id}.json)                     │   │
│  │  • SQLite FTS5 index (index.db; locked → index.db.enc)  │   │
│  │  • Version history ({id}.versions.json)                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              Self-Hosted Sync Server (Optional)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  REST API Handlers                                    │     │
│  │  /devices/register  /sync/pull  /sync/push  /health   │     │
│  └───────────────────┬─────────────────────────────────┘     │
│                      │                                         │
│  ┌───────────────────▼─────────────────────────────────┐     │
│  │  Sync Handler Logic                                  │     │
│  │  • Delta sync (timestamps)                           │     │
│  │  • Conflict resolution (last-write-wins)             │     │
│  │  • Device registration                               │     │
│  └───────────────────┬─────────────────────────────────┘     │
│                      │                                         │
│  ┌───────────────────▼─────────────────────────────────┐     │
│  │  SQLite Storage (WAL mode)                           │     │
│  │  • sync_items table (encrypted items + metadata)     │     │
│  │  • devices table (device registration)               │     │
│  └──────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Module Dependencies

```
┌──────────────────────┐
│   CLI (packages/)    │
└─────────────┬────────┘
              │ imports
              ▼
┌───────────────────────────────────────────────────┐
│          Core Layer (core/)                       │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ crypto/      │  │ vault/       │ ◄─────┐      │
│  │              │  │              │       │      │
│  │• interfaces  │  │• types       │ imports crypto
│  │• kdf/        │  │• item/       │       │      │
│  │• cipher/     │  │• store/      │       │      │
│  │• encoding/   │  │• index/      │  ┌─────────┐ │
│  │• key/        │  │• version/    │  │sync/    │ │
│  │• password/   │  │• health/     │  │         │ │
│  └──────────────┘  │• importexport│  │• client │ │
│                    └──────────────┘  │• server │ │
│                                      └─────────┘ │
└───────────────────────────────────────────────────┘
              │ imports
              ▼
┌──────────────────────────────────┐
│  Services (services/syncserver)  │
│  • main.go (HTTP server)         │
│  • storage.go (SQLite wrapper)   │
└──────────────────────────────────┘
```

---

## Key Hierarchy & Encryption Flow

### Key Derivation Chain

```
User enters password: "my-vault-password"
  │
  ├─ Argon2id(password, salt=random_128bit)
  │  Memory: 64MB, Iterations: 3, Parallelism: 4
  │  Produces: Master Key (256-bit)
  │
  ├─ Master Key stored in vault (never transmitted)
  │  and used to lock/unlock Vault Key
  │
  └─ Vault Key
     │  (Random 256-bit key, encrypted under Master Key)
     │  AES-256-GCM(Vault Key, Master Key)
     │
     ├─ Stored in vault metadata (encrypted)
     │  Decrypted on unlock, wiped on lock
     │
     └─ For each item: HKDF-SHA256(Vault Key, Item ID)
        Produces: Item Key (256-bit, per-item unique)
        
        └─ AES-256-GCM(Item Data, Item Key, random nonce per operation)
           Produces: Encrypted Item
           
           + SHA-256(Encrypted Item)  ← Integrity check
```

### Recovery Path

```
User loses master password
  │
  ├─ Enters 12-word BIP-39 mnemonic (stored externally/printed)
  │
  ├─ PBKDF2-HMAC-SHA512(mnemonic, "recovery-salt")
  │  Produces: Recovery Key (256-bit)
  │
  ├─ AES-256-GCM.decrypt(Vault Key, Recovery Key)
  │  Produces: Plaintext Vault Key
  │
  ├─ (Now can proceed with item decryption using recovered Vault Key)
  │
  └─ Auto-rotation (PRD §8.5: one-time use):
     ├─ Generate new BIP-39 mnemonic + new recovery key
     ├─ Re-encrypt vault key under new recovery key
     ├─ Return new mnemonic — user MUST save it
     └─ Previous mnemonic is invalidated
```

### Recovery Validation (Non-Destructive)

```
User wants to verify they still have the correct mnemonic
  │
  ├─ ValidateRecovery(mnemonic)
  │  ├─ Derive recovery key from mnemonic
  │  ├─ Attempt to decrypt vault key
  │  └─ Return success/error WITHOUT unlocking vault
  │
  └─ No side effects: vault remains locked, recovery key is NOT rotated
```

---

## Vault Lifecycle State Machine

```
┌──────────┐
│ Nonexist │  (cli: zp init)
└─────┬────┘
      │ Create(password)
      ▼
  ┌────────────────────────┐
  │ Created (Locked)       │ ◄──────────┐
  │ • Master key derived   │            │
  │ • Vault key created    │            │
  │ • Metadata encrypted   │            │ (manual lock)
  │ • Stored on disk       │            │
  │ • SQLite index init    │            │
  └────────┬───────────────┘            │
           │                            │
           │ Unlock(password|recovery|vaultKey)  │
           ▼                            │
  ┌────────────────────────┐            │
  │ Unlocked (Open)        │     ┌──────┴─────┐
  │ • Master key loaded    │────►│  Lock()    │
  │ • Vault key decrypted  │     └────────────┘
  │ • Items decryptable    │
  │ • Search index live    │ (manual lock or timeout)
  │ • Auto-lock per vault config (default 15 min) │
  └────────────────────────┘

  Note: UnlockWithRecovery(mnemonic) auto-rotates the recovery key
  and returns a NEW mnemonic. The old mnemonic is invalidated (one-time use).
```

> Bridge note: the CGO bridge disables the Go auto-lock timer at runtime (`DisableAutoLock()`); the host app is responsible for inactivity locking.


---

## Add Item Flow (Encryption Path)

```
User: zp add --type=login --name="GitHub" --username="octocat" --password="generated"

1. CLI prompts for fields (interactive)
   └─ Password validated by zxcvbn (min score 3 or reject)

2. Create Item struct
   ├─ ID: UUID
   ├─ Type: login
   ├─ Name: "GitHub"
   ├─ Fields: {username: "octocat", password: "..."}
   └─ Metadata: created_at, updated_at, tags

3. Item Encryption
   ├─ Lookup Vault Key (unlocked in memory)
   ├─ Derive Item Key = HKDF-SHA256(Vault Key, Item ID)
   ├─ Serialize Item as JSON
   ├─ Encrypt JSON payload
   │  └─ AES-256-GCM(JSON bytes, Item Key, random_nonce)
   │     Produces: (ciphertext, auth_tag, nonce)
   ├─ Compute SHA-256(ciphertext + auth_tag)
   │  └─ Stored separately for integrity verification
   └─ Defer ZeroBytes(&ItemKey)

4. Storage
   ├─ Write {item_id}.json (EncryptedItem JSON wrapper: base64 ciphertext + checksum)
   ├─ Index derived metadata in SQLite FTS5 (index is encrypted at rest when the vault is locked)
   └─ Version history ({item_id}.versions.json) is written on update/restore snapshots (encrypted at rest)

5. Return success to CLI
   └─ Display: "Added GitHub (ID: abc-123)"
```

---

## Retrieval & Decryption Path

```
User: zp get "GitHub" --copy

1. Search index
   ├─ Query: "GitHub" in SQLite FTS5
   └─ Match item ID: abc-123

2. Load encrypted item
   ├─ Read {item_id}.json (EncryptedItem JSON)
   ├─ Decode `data` (base64) → ciphertext bytes
   └─ Compute SHA-256(ciphertext) and compare to `checksum`

3. Item Decryption
   ├─ Lookup Vault Key (require unlocked vault)
   ├─ Derive Item Key = HKDF-SHA256(Vault Key, Item ID)
   ├─ Decrypt ciphertext
   │  └─ AES-256-GCM.decrypt(ciphertext, Item Key, nonce)
   │     Produces: plaintext JSON bytes
   ├─ Parse JSON → Item struct
   └─ Defer ZeroBytes(&ItemKey)

4. Display/Copy
   ├─ CLI renders item fields (username, password, etc.)
   ├─ User chooses: --copy (clipboard) or --show (terminal)
   ├─ If clipboard:
   │  └─ Copy to clipboard
   │  └─ Start 30-second timer (auto-clear after 30s)
   └─ Return to CLI
```

---

## Sync Protocol Flow (Current Preview)

> Important: the current sync stack provides transport, conflict reporting, and bridge/macOS preview flows. It is not yet a full generic “pull and automatically merge into the vault” pipeline across every product surface.

### Registration

```
Client (Device 1)
  │
  ├─ Generate device_id (UUID)
  │
  └─► POST /devices/register
      {
        device_id: "uuid-...",
        device_name: "MacBook Pro"
      }
      
Server
  │
  ├─ Store device registration in devices table
  └─ Response: 201 Created
```

### Pull (Download Changes)

```
Client (Device 1, last sync: 2026-04-01 10:00:00)
  │
  ├─ Read last_sync_timestamp from local state
  │
  └─► GET /sync/pull?device_id=uuid-...&since=1712054400

Server
  │
  ├─ Query: SELECT * FROM sync_items WHERE timestamp > 1712054400
  ├─ Construct response with changed items
  │  (Each item: encrypted payload + version + timestamp)
  │
  └─ Response: 200 OK
     {
       items: [
         {
           item_id: "item-123",
           version: 5,
           device_id: "device-from-another-phone",
           payload: "base64-...",
           timestamp: 1712054500,
           checksum: "sha256...",
           deleted: false
         },
         ...
       ],
       server_time: 1712054600
     }

Client
  │
  ├─ For each remote item:
  │  ├─ Build a remote view of changed encrypted items
  │  ├─ Compare against local pending changes when performing `Sync()`
  │  ├─ Record conflicts for local resolution/reporting
  │  └─ Product-specific layers decide how/when to apply pulled items into local vault state
  │
  └─ Continue sync flow or surface the pulled payloads to the caller
```

### Push (Upload Changes)

```
Client (Device 1)
  │
  ├─ Collect locally changed items since last sync
  │  (Items modified or deleted locally)
  │
  └─► POST /sync/push
      {
        device_id: "uuid-...",
        items: [
          {
            item_id: "item-456",
            version: 6,
            payload: "base64-...",
            timestamp: 1712054550,
            checksum: "sha256...",
            deleted: false
          },
          ...
        ]
      }

Server
  │
  ├─ For each item:
  │  ├─ Check if server has this item
  │  ├─ If not: store new (no conflict)
  │  ├─ If exists:
  │  │  ├─ If server timestamp is newer-or-equal and came from another device: report conflict
  │  │  ├─ Else: accept client item and persist it
  │  └─ Update device last_sync_at
  │
  ├─ Detect any conflicts
  │
  └─ Response: 200 OK
     {
       accepted: ["item-456"],
       conflicts: [
         {
           item_id: "item-456",
           server_item: { ... },
           message: "server has newer or equal version"
         }
       ],
       server_time: 1712054600
     }

Client
  │
  ├─ If conflicts returned:
  │  ├─ Re-pull / inspect server copy
  │  ├─ Resolve in client/product layer
  │  └─ Retry as needed
  │
  └─ Sync complete
```

---

## Conflict Resolution Algorithm

```
For each item where both sides modified the same record:
  - Client sends `(timestamp, version, device_id)`
  - Server primarily compares timestamp + originating device

Conflict Check:
  if server.Timestamp > client.Timestamp and server.DeviceID != client.DeviceID
    → Server is newer, reject client version
    → Return server item in conflict payload

  if server.Timestamp == client.Timestamp
    → Current server behavior is conservative
    → Cross-device equal timestamps conflict
    → Client re-pulls and resolves locally

  otherwise
    → Accept client version and persist it
```

**Example:**

```
Device A: Item 'GitHub' edited at 10:00:00, version: 3
Device B: Item 'GitHub' edited at 10:00:05, version: 2

Server receives push from B at 10:00:10:
  client.Timestamp > server.Timestamp
  → Client is newer by timestamp
  → Accept B's version

If both timestamps had tied across devices:
  → Current server path reports a conflict
  → Client re-pulls and resolves locally

If server timestamp had been newer from another device:
  → Reject B's version, respond with "use_remote"
  → B re-pulls and gets A's version
```

---

## Data Formats

### EncryptedItem (On Disk)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "data": "base64([nonce(12)][ciphertext+gcm_tag])",
  "version": 1,
  "checksum": "sha256_hex_of_decoded_data"
}
```

Notes:
- The plaintext item fields (type/name/fields/notes/tags/...) are inside the encrypted `data` payload.
- AES-GCM output is stored as `[nonce][ciphertext+tag]` (nonce prepended).


### Plaintext Item (In Memory)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "login",
  "name": "GitHub",
  "username": "octocat",
  "password": "super-secret-token",
  "url": "https://github.com",
  "notes": "Personal GitHub account",
  "created_at": "2026-04-01T10:00:00Z",
  "updated_at": "2026-04-01T10:05:00Z",
  "tags": ["prod", "critical"]
}
```

### SyncItem Protocol

```json
{
  "item_id": "550e8400-e29b-41d4-a716-446655440000",
  "version": 5,
  "device_id": "device-uuid",
  "payload": "base64(...)",
  "timestamp": 1712054400,
  "checksum": "sha256(...)",
  "deleted": false
}
```

### VaultMetadata (vault.json)

```json
{
  "salt": "hex(...)",
  "encrypted_vault_key": "base64(...)",
  "encrypted_recovery_key": "base64(...)",
  "vault_key_check": "base64(...)",
  "created_at": "2026-04-01T10:00:00Z",
  "config": {
    "auto_lock_timeout": 900000000000,
    "clipboard_clear_sec": 30,
    "max_versions": 10
  }
}
```

Notes:
- `vault_key_check` enables validating a raw vault key for `UnlockWithKey` (e.g. TouchID/keychain flows).
- Vault metadata reads/writes preserve unknown JSON fields (forward/backward compatibility).
- Vault metadata writes use atomic replace semantics (tmp + sync + rename + best-effort directory sync).

---

## Storage Architecture

### Local Storage

```
~/.zeropass/vaults/default/
├── vault.json                 # Vault metadata (salt, encrypted keys, vault_key_check, config)
├── index.db                   # SQLite FTS5 (present while unlocked)
├── index.db.enc               # Encrypted index at rest (present while locked)
├── index.db-wal               # SQLite write-ahead log (when unlocked)
├── index.db-shm               # SQLite shared-memory file (when unlocked)
├── items/
│   ├── 550e8400-...json            # Encrypted item 1
│   ├── 550e8400-...versions.json   # Version history 1 (encrypted at rest)
│   └── [more items...]
└── vault.lock                 # Advisory lock used by the CGO bridge (flock)
```

### Server Storage (sync_items)

```sql
CREATE TABLE sync_items (
  item_id TEXT PRIMARY KEY,
  version INTEGER NOT NULL DEFAULT 1,
  device_id TEXT NOT NULL,
  payload TEXT NOT NULL,
  timestamp INTEGER NOT NULL,
  checksum TEXT NOT NULL DEFAULT '',
  deleted BOOLEAN NOT NULL DEFAULT FALSE
);
```

### Server Storage (devices)

```sql
CREATE TABLE devices (
  device_id TEXT PRIMARY KEY,
  device_name TEXT NOT NULL,
  last_sync_at INTEGER NOT NULL DEFAULT 0
);
```

---

## Performance & Scalability

### Local Vault
- **Encrypt/Decrypt:** ~10-50ms (Argon2id KDF, depends on password length)
- **Add Item:** ~5ms (depends on item size)
- **Search (FTS5):** <1ms for typical queries (<1000 items)
- **Vault with 1000 items:** Loads in <500ms

### Multi-Device Sync
- **Pull (100 items):** ~2-3 seconds (network latency dominant)
- **Push (10 items):** ~1-2 seconds
- **Conflict resolution:** <1 second (in-memory)
- **Full sync:** ~5-10 seconds depending on item count

### Server Scalability
- **Single instance handles:** 1000+ devices concurrently
- **Database:** SQLite with WAL mode suitable for small-to-medium deployments
- **Upgrade path:** Replace with PostgreSQL for enterprise scale

---

**Document Version:** 1.1  
**Last Updated:** June 2025  
**Owner:** Architecture Team
