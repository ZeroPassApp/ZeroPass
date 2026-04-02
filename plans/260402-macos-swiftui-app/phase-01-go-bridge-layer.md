# Phase 1: Go Bridge Layer

## Context
- [Go↔Swift Interop Research](./research/go-swift-interop-research.md)
- [Codebase API Surface](./reports/codebase-api-surface.md)

## Overview
- **Priority:** P0 — blocks all subsequent phases
- **Status:** Done
- **Description:** Create a Go `bridge/` package that exposes the ZeroPass core API as C-callable functions via `//export`. Uses `c-archive` build mode to produce `libzeropass.a` + `libzeropass.h`.

## Key Insights
- WireGuard uses same pattern: `//export` → `c-archive` → Swift bridging header
- JSON-over-FFI eliminates complex C struct mapping (adds ~300ns, negligible vs 200-500ms KDF)
- Every exported function MUST have `defer recover()` — Go panic = process crash
- Handle pattern: store `*Vault` in Go-side `map[int64]`, pass `int64` to C/Swift
- Memory rule: `C.CString()` returns malloc'd memory, caller MUST free via `ZPFree()`
- **`ZPCreateVault` returns `{handle, mnemonic}`** — Swift MUST free result immediately & zero mnemonic copy
- **File locking (`flock`)** required for CLI/GUI concurrent vault access — advisory lock on `vault.lock`
- **Disable Go-side auto-lock** when vault opened via bridge (`vault.DisableAutoLock()` runtime-only); Swift handles auto-lock
- **Import functions** take `io.Reader` in Go core — bridge must open file, read, import, add items to vault
- **Export functions** take `io.Writer` in Go core — bridge must create file, get all items, export, close file

## Requirements

### Functional
- Expose all vault lifecycle operations (create, open, unlock, lock, close)
- Expose `UnlockWithKey` for TouchID flow (bypasses KDF, uses raw vault key)
- Expose `ChangeMasterPassword` for settings (prereq: implement in Go core first)
- Expose all item CRUD operations (add, get, update, delete, list, search)
- Expose item version history (get history, restore version)
- Expose password generator and HIBP breach check
- Expose import/export operations (bridge orchestrates file I/O + item persistence)
- Expose health analysis
- Expose sync client operations (including device registration)
- Expose recovery key regeneration (NOT "get" — mnemonic is never stored)
- Implement advisory file locking (`flock`) for concurrent CLI/GUI vault access

### Non-Functional
- Thread-safe (Go-side mutex per vault handle)
- Panic-safe (defer recover in every export)
- Memory-safe (documented ownership for every allocation)
- Universal binary support (arm64 + x86_64 via lipo)

## Architecture

### ZPResult Pattern (all functions return this)

```c
typedef struct {
    char* data;   // JSON string (NULL if no data), caller frees via ZPFree
    char* error;  // Error message (NULL on success), caller frees via ZPFree
    int   code;   // 0=ok, 1=locked, 2=not_found, 3=auth_failed, 4=internal, 5=busy
} ZPResult;
```

### Handle Registry

```go
var (
    mu       sync.RWMutex
    handles  = make(map[int64]*vaultSession)
    nextID   int64
)

type vaultSession struct {
    vault    *store.Vault
    mgr      *item.Manager
    idx      *index.Index
    verMgr   *version.Manager  // item version history
    lockFile *os.File          // flock advisory lock (released on close)
    syncCli  *client.SyncClient // nil until sync configured
}
```

## Related Code Files

### Files to CREATE

| File | Action | Description |
|------|--------|-------------|
| `bridge/main.go` | Create | Package main, import "C", ZPResult typedef |
| `bridge/vault_api.go` | Create | ZPCreateVault, ZPOpenVault, ZPUnlock, ZPLock, ZPCloseVault, ZPIsLocked |
| `bridge/item_api.go` | Create | ZPListItems, ZPGetItem, ZPCreateItem, ZPUpdateItem, ZPDeleteItem |
| `bridge/search_api.go` | Create | ZPSearch, ZPRebuildIndex |
| `bridge/crypto_api.go` | Create | ZPGeneratePassword, ZPGeneratePassphrase, ZPCheckBreach, ZPScorePassword |
| `bridge/recovery_api.go` | Create | ZPRegenerateRecovery, ZPUnlockWithRecovery (NO ZPGetRecoveryKey — mnemonic is never stored) |
| `bridge/importexport_api.go` | Create | ZPImportCSV, ZPImportChrome, ZPExportJSON, etc. |
| `bridge/health_api.go` | Create | ZPAnalyzeHealth |
| `bridge/sync_api.go` | Create | ZPSyncRegister, ZPSyncSetup, ZPSyncPull, ZPSyncPush, ZPSyncFull |
| `bridge/version_api.go` | Create | ZPGetVersionHistory, ZPRestoreVersion |
| `bridge/handle_registry.go` | Create | Handle map, register/get/remove vault sessions |
| `bridge/filelock.go` | Create | Advisory flock() on vault.lock file — acquire on open, release on close |
| `bridge/helpers.go` | Create | jsonResponse(), ZPFree(), ZPFreeResult(), error codes |
| `bridge/Makefile` | Create | Build targets for arm64, amd64, universal |

## Implementation Steps

1. Create `bridge/` directory at project root
2. Create `bridge/main.go` with package main, C header typedef, `ZPFree`/`ZPFreeResult`
3. Create `bridge/handle_registry.go` — thread-safe handle map with register/get/remove
4. Create `bridge/filelock.go` — advisory `flock()` on `vault.lock` file:
   - `acquireLock(vaultPath) (*os.File, error)` — exclusive lock on open
   - `releaseLock(f *os.File)` — release on close
   - Stored in `vaultSession.lockFile`
5. Create `bridge/helpers.go` — `jsonResponse()`, `errorResult()`, error codes
6. Create `bridge/vault_api.go` — vault lifecycle exports (~9 functions):
   - `ZPCreateVault` → returns `{"handle": N, "mnemonic": "..."}` — vault starts unlocked after create
   - `ZPOpenVault` → acquires flock, sets `AutoLockTimeout=0` (Swift handles auto-lock)
   - `ZPUnlock`, `ZPUnlockWithRecovery`
   - **`ZPUnlockWithKey`** — raw vault key unlock for TouchID (prereq: `store.UnlockWithKey()`)
   - **`ZPChangeMasterPassword`** — (prereq: `store.ChangeMasterPassword()`)
   - `ZPLock` (returns success always — Go `Lock()` has no error return, log index encryption failures)
   - `ZPCloseVault` → releases flock
   - `ZPIsLocked` → returns `C.int` (0/1, not ZPResult — hot path)
7. Create `bridge/item_api.go` — item CRUD exports (~5 functions)
8. Create `bridge/version_api.go` — version history exports (~2 functions):
   - `ZPGetVersionHistory(handle, itemID)` → `[]{version, saved_at}`
   - `ZPRestoreVersion(handle, itemID, version)` → restores item to specific version
9. Create `bridge/search_api.go` — search exports (~2 functions)
10. Create `bridge/crypto_api.go` — password/breach exports (~4 functions)
11. Create `bridge/recovery_api.go` — recovery exports (~2 functions, NOT ZPGetRecoveryKey)
12. Create `bridge/importexport_api.go` — import/export exports (~12 functions):
    - **Import orchestration:** `os.Open(path)` → reader → `ImportX(reader)` → loop `mgr.AddItem(&item)` → return `{"imported": count}`
    - **Export orchestration:** `mgr.AllItems()` → `os.Create(path)` → `ExportX(items, writer)` → close
    - **ExportEncrypted:** uses vault key from `vault.VaultKey()`
13. Create `bridge/health_api.go` — health analysis export (~1 function)
14. Create `bridge/sync_api.go` — sync exports (~5 functions):
    - **`ZPSyncRegister`** — registers device with server (must happen first!)
    - `ZPSyncSetup` — persist config (server URL, device ID, API key) to `sync.json` next to `vault.json`
    - `ZPSyncPull`, `ZPSyncPush`, `ZPSyncFull`
    - Note: `SyncClient` is stateful — create on `ZPSyncSetup`, store in `vaultSession.syncCli`
15. Create `bridge/Makefile` with targets: `build-arm64`, `build-amd64`, `build-universal`, `clean`
16. Verify build: `cd bridge && CGO_ENABLED=1 go build -buildmode=c-archive -o libzeropass.a .`
17. Verify generated header contains all ~42 function declarations
18. Write integration tests: `bridge/bridge_test.go` (test Go API calls directly before crossing C boundary)

## Exported Functions Inventory (~42 functions)

```
// Lifecycle
ZPCreateVault(path, masterPassword *C.char) C.ZPResult  → {"handle": int, "mnemonic": "word1 word2 ..."}
                                                          ⚠️ Vault starts UNLOCKED after create
                                                          ⚠️ Swift MUST ZPFreeResult immediately + zero mnemonic
ZPOpenVault(path *C.char) C.ZPResult                    → {"handle": int} (acquires flock, AutoLockTimeout=0)
ZPUnlock(handle C.long, masterPassword *C.char) C.ZPResult
ZPUnlockWithKey(handle C.long, vaultKeyBase64 *C.char) C.ZPResult  // TouchID bypass — raw vault key
ZPUnlockWithRecovery(handle C.long, mnemonic *C.char) C.ZPResult
ZPLock(handle C.long) C.ZPResult                        // Always succeeds (Go Lock() has no error return)
ZPCloseVault(handle C.long) C.ZPResult                  // Releases flock
ZPIsLocked(handle C.long) C.int                         // 0=unlocked, 1=locked (hot path, not ZPResult)
ZPChangeMasterPassword(handle C.long, oldPassword, newPassword *C.char) C.ZPResult

// Item CRUD
ZPListItems(handle C.long, filterJSON *C.char) C.ZPResult
ZPGetItem(handle C.long, itemID *C.char) C.ZPResult
ZPCreateItem(handle C.long, itemJSON *C.char) C.ZPResult
ZPUpdateItem(handle C.long, itemID, itemJSON *C.char) C.ZPResult
ZPDeleteItem(handle C.long, itemID *C.char) C.ZPResult

// Version History
ZPGetVersionHistory(handle C.long, itemID *C.char) C.ZPResult  → [{version, saved_at, item}]
ZPRestoreVersion(handle C.long, itemID *C.char, version C.int) C.ZPResult

// Search
ZPSearch(handle C.long, query *C.char) C.ZPResult
ZPRebuildIndex(handle C.long) C.ZPResult

// Password & Security
ZPGeneratePassword(length C.int, optionsJSON *C.char) C.ZPResult
ZPGeneratePassphrase(words C.int, separator *C.char) C.ZPResult
ZPScorePassword(password *C.char) C.ZPResult
ZPCheckBreach(password *C.char) C.ZPResult

// Recovery
ZPRegenerateRecovery(handle C.long) C.ZPResult          → {"mnemonic": "new words..."}

// Health
ZPAnalyzeHealth(handle C.long) C.ZPResult

// Import — bridge orchestrates: open file → reader → ImportX → loop mgr.AddItem → return count
ZPImportCSV(handle C.long, path *C.char) C.ZPResult     → {"imported": count}
ZPImportChrome(handle C.long, path *C.char) C.ZPResult
ZPImportFirefox(handle C.long, path *C.char) C.ZPResult
ZPImport1Password(handle C.long, path *C.char) C.ZPResult
ZPImportBitwarden(handle C.long, path *C.char) C.ZPResult
ZPImportSafari(handle C.long, path *C.char) C.ZPResult
ZPImportLastPass(handle C.long, path *C.char) C.ZPResult
ZPImportKeePass(handle C.long, path *C.char) C.ZPResult
ZPImport1PUX(handle C.long, path *C.char) C.ZPResult

// Export — bridge orchestrates: AllItems → create file → ExportX → close
ZPExportJSON(handle C.long, path *C.char) C.ZPResult
ZPExportCSV(handle C.long, path *C.char) C.ZPResult
ZPExportEncrypted(handle C.long, path *C.char) C.ZPResult  // Uses vault key

// Sync — stateful: ZPSyncSetup creates/persists config, ZPSyncRegister registers device
ZPSyncSetup(handle C.long, configJSON *C.char) C.ZPResult   // Persists to sync.json
ZPSyncRegister(handle C.long, deviceName *C.char) C.ZPResult // Must call before pull/push
ZPSyncPull(handle C.long) C.ZPResult
ZPSyncPush(handle C.long) C.ZPResult
ZPSyncFull(handle C.long) C.ZPResult

// Memory
ZPFree(ptr *C.char)
ZPFreeResult(r C.ZPResult)
```

## Success Criteria
- [ ] `go build -buildmode=c-archive` succeeds for arm64
- [ ] `lipo` universal binary builds successfully
- [ ] Generated `libzeropass.h` contains all ~42 function declarations
- [ ] All exports have `defer recover()` panic guard
- [ ] All C.CString allocations have documented ownership
- [ ] File locking: `flock()` acquired on `ZPOpenVault`, released on `ZPCloseVault`
- [ ] `ZPCreateVault` returns `{handle, mnemonic}` correctly
- [ ] `ZPUnlockWithKey` works (prereq: `store.UnlockWithKey` implemented)
- [ ] `ZPChangeMasterPassword` works (prereq: `store.ChangeMasterPassword` implemented)
- [ ] Import orchestration: file → reader → import → AddItem loop → count
- [ ] Export orchestration: AllItems → file → Export → close
- [ ] Version history: `ZPGetVersionHistory` + `ZPRestoreVersion` functional
- [ ] Integration test covers vault create → unlock → CRUD → lock → close cycle

## Risk Assessment
- **Go panic crash:** Mitigated by `defer recover()` in every export
- **Memory leak:** Mitigated by strict `ZPFreeResult` pattern + documentation
- **Thread safety:** Mitigated by per-handle mutex in registry
- **CGo overhead:** Benchmarked at 50-400ns per call — negligible
- **File locking:** Advisory only — won't prevent non-ZeroPass processes from writing. Acceptable for CLI/GUI coexistence.
- **Mnemonic exposure:** `ZPCreateVault` returns mnemonic in C-heap JSON. Swift must free immediately and zero its own copy.
- **Go auto-lock conflict:** Bridge sets `AutoLockTimeout=0` on open. Swift-side `AutoLockService` handles all lock timing.
