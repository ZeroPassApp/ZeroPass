# ZeroPass: System Architecture

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        ZeroPass (Developer)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  CLI Application (Cobra)                                │   │
│  │  Commands: init, add, get, search, export, sync, etc.   │   │
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
│  │  • SQLite FTS5 index (searchable metadata)              │   │
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
│  │  /sync/register  /sync/pull  /sync/push  /health      │     │
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
  └─ (Now can proceed with item decryption using recovered Vault Key)
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

## Sync Protocol Flow

### Registration

```
Client (Device 1)
  │
  ├─ Generate device_id (UUID)
  ├─ Generate public_key for this device
  │
  └─► POST /sync/register
      {
        device_id: "uuid-...",
        name: "MacBook Pro",
        public_key: "base64-encoded-key"
      }
      
Server
  │
  ├─ Store device registration in devices table
  ├─ Return device_token for future auth
  │
  └─ Response: 200 OK (device registered)

Client
  └─ Store device_token locally for future syncs
```

### Pull (Download Changes)

```
Client (Device 1, last sync: 2026-04-01 10:00:00)
  │
  ├─ Read last_sync_timestamp from local state
  │
  └─► POST /sync/pull
      {
        device_id: "uuid-...",
        last_sync_at: 1712054400,
        limit: 100
      }

Server
  │
  ├─ Query: SELECT * FROM sync_items WHERE updated_at > 1712054400
  ├─ Construct response with changed items
  │  (Each item: encrypted data + version vector + timestamp)
  │
  └─ Response: 200 OK
     {
       items: [
         {
           id: "item-123",
           device_id: "device-from-another-phone",
           encrypted_data: "base64-...",
           version_vector: {device1: 5, device2: 3, ...},
           updated_at: 1712054500,
           conflict: false
         },
         ...
       ],
       server_timestamp: 1712054600
     }

Client
  │
  ├─ For each remote item:
  │  ├─ Check if local item exists
  │  ├─ If no local: apply remote (merge)
  │  ├─ If local exists:
  │  │  ├─ Compare version vectors
  │  │  ├─ If remote newer: apply remote
  │  │  ├─ If tied: prefer remote (authority)
  │  │  ├─ If local newer: skip (will push local)
  │  └─ Decrypt item (still encrypted, just merged metadata)
  │
  └─ Update last_sync_timestamp = server_timestamp
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
            id: "item-456",
            encrypted_data: "base64-...",
            version_vector: {device1: 6, device2: 3},
            operation: "upsert"  # or "delete"
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
  │  │  ├─ Compare version vectors
  │  │  ├─ If client vector > server vector: accept (client newer)
  │  │  ├─ If server vector > client vector: reject (server newer)
  │  │  ├─ If tied: reject (conflict, ask client to re-pull)
  │  └─ Update server version vector for device
  │
  ├─ Detect any conflicts (versions tied)
  │
  └─ Response: 200 OK or 409 Conflict
     {
       conflicts: [
         {
           id: "item-456",
           server_version_vector: {...},
           server_updated_at: 1712054550,
           resolution: "use_remote"  # or "use_local"
         }
       ]
     }

Client
  │
  ├─ If conflicts returned:
  │  ├─ Merge per server resolution
  │  ├─ Re-pull to fetch resolved state
  │  └─ Clear conflict markers
  │
  └─ Sync complete
```

---

## Conflict Resolution Algorithm

```
For each item where:
  - Client has version vector V_client
  - Server has version vector V_server

Conflict Check:
  if V_client > V_server (client-side clocks advanced more)
    → Client is newer, accept client version
    → Server stores client version
    
  if V_server > V_client (server-side clocks advanced more)
    → Server is newer, reject client version
    → Server returns "use_remote"
    
  if V_client == V_server (simultaneous edits on different devices)
    → Conflict detected!
    → Apply tiebreaker:
       1. Last-write-wins (compare timestamps, account for clock skew)
       2. If timestamps equal: prefer remote (server authority)
       3. Mark conflict for client attention
```

**Example:**

```
Device A: Item 'GitHub' edited at 10:00:00, version_vector: {A: 3, B: 1}
Device B: Item 'GitHub' edited at 10:00:05, version_vector: {A: 2, B: 2}

Server receives push from B at 10:00:10:
  V_B (2,2) vs V_server(3,1)
  → (2,2) < (3,1)  [B's 2 < A's 3]
  → Server-side is newer
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
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "device_id": "device-uuid",
  "encrypted_data": "base64(...)",
  "version_vector": {
    "device-uuid-1": 5,
    "device-uuid-2": 3
  },
  "updated_at": 1712054400,
  "created_at": 1712054300,
  "operation": "upsert"
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
  id TEXT PRIMARY KEY,
  device_id TEXT NOT NULL,
  encrypted_data BLOB NOT NULL,
  version_vector JSON NOT NULL,
  updated_at INTEGER NOT NULL,
  created_at INTEGER NOT NULL,
  INDEX idx_device_id (device_id),
  INDEX idx_updated_at (updated_at)
);
```

### Server Storage (devices)

```sql
CREATE TABLE devices (
  device_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  public_key TEXT,
  last_sync_at INTEGER,
  created_at INTEGER
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

**Document Version:** 1.0  
**Last Updated:** April 1, 2026  
**Owner:** Architecture Team
