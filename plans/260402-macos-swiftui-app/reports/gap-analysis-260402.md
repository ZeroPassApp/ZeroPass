# Gap Analysis Report: ZeroPass macOS SwiftUI App Plan

**Date:** 2026-04-02  
**Scope:** Review all 10 phases + research reports vs actual codebase + PRD  
**Methodology:** Cross-referenced plan against Go core source, PRD requirements, Apple platform docs, Swift conventions

---

## Executive Summary

Plan is **solid architecturally** — c-archive + JSON-over-FFI + handle pattern is proven (WireGuard precedent). Research reports are thorough and benchmarked. However, there are **12 critical gaps, 8 medium gaps, and 6 minor issues** that will cause implementation problems if not addressed before coding.

**Top 3 showstoppers:**
1. No `ChangeMasterPassword()` API exists in Go core — Phase 8 Settings references it but it's unimplemented
2. No file-locking (`flock`) mechanism exists — plan claims Phase 1 adds it but no implementation steps
3. Bridge `ZPCreateVault` returns handle + mnemonic, but vault starts UNLOCKED after Create — handle registry must track this inconsistent state

---

## Critical Gaps (Must Fix Before Implementation)

### GAP-01: Missing `ChangeMasterPassword` in Go Core

**Severity:** CRITICAL  
**Phases affected:** 1 (bridge), 8 (settings)

Phase 8 SecuritySettingsView shows "Change Master Password..." button. Phase 1 exported functions inventory has NO `ZPChangeMasterPassword` function.

**Root cause:** `store.go` has no `ChangeMasterPassword()` method. Only `RotateVaultKey()` exists in `key/vault_key.go` — but that's vault key rotation, not master password change. The flow would be:
1. Verify old password (derive master key with old salt)
2. Generate new salt
3. Derive new master key from new password + new salt
4. Re-encrypt vault key with new master key
5. Update vault.json metadata (salt, encrypted_vault_key)

**Fix:** Implement `func (v *Vault) ChangeMasterPassword(oldPassword, newPassword string) error` in `store.go` BEFORE starting Phase 1. Then add `ZPChangeMasterPassword` to bridge API.

---

### GAP-02: No File Locking (flock) Implementation

**Severity:** CRITICAL  
**Phases affected:** 1, all runtime phases

Plan says "Add `flock()` mechanism in Phase 1" in Risk table, but Phase 1 implementation steps have ZERO mention of file locking. No lock file, no advisory lock, nothing.

CLI and GUI sharing `~/.zeropass/` will corrupt:
- `vault.json` (concurrent metadata writes)
- `index.db` (SQLite — has its own locking but Go's `Unlock/Lock` does `EncryptIndexFile/DecryptIndexFile` which is file-level, not SQLite WAL)
- Item files (concurrent encrypt/write)

**Fix:** Add step in Phase 1: create `bridge/filelock.go` implementing `flock()` advisory locking on `~/.zeropass/default/vault.lock`. Acquire exclusive lock on vault open, release on close. Use shared lock for read-only operations.

---

### GAP-03: `ZPCreateVault` State Inconsistency

**Severity:** CRITICAL  
**Phases affected:** 1, 2, 3

Phase 1 shows `ZPCreateVault(path, masterPassword) → ZPResult` returning `{"handle": int}`. But looking at `store.Create()`:
- Returns `(*Vault, *CreateResult, error)` where `CreateResult` contains `Mnemonic`
- Vault is **already unlocked** after creation (vaultKey is set, locked=false)

The plan's exported function must return BOTH the handle AND the mnemonic. Current ZPResult pattern only has a single `data` JSON field, so:
```json
{"handle": 42, "mnemonic": "word1 word2 ... word12"}
```

**BUT** — returning mnemonic as JSON means it stays in Go-allocated C memory until `ZPFreeResult`. This is a security concern — mnemonic is the "show once" recovery secret.

**Fix:** 
1. Document that `ZPCreateVault` returns `{handle, mnemonic}` in the data field
2. Add explicit warning in Phase 1 that Swift side MUST call `ZPFreeResult` immediately after extracting mnemonic, and mnemonic handling in Swift must use `UnsafeMutableBufferPointer` + `memset_s` zeroing (as the research report suggests for passwords)

---

### GAP-04: Missing Version History in Bridge API

**Severity:** CRITICAL  
**Phases affected:** 1, 5

`core/vault/version/version.go` has a full Version Manager (`SaveVersion`, `GetVersionHistory`, `RestoreVersion`). Phase 5 EditItemView mentions "History section showing version count."

**But Phase 1 has ZERO exported functions for version history.** No `ZPGetVersionHistory`, no `ZPRestoreVersion`.

Also: `item.Manager` doesn't appear to use `version.Manager` — versioning seems disjoint from CRUD. Need to verify if `AddItem`/`UpdateItem` automatically call `SaveVersion`.

**Fix:** 
1. Add `ZPGetVersionHistory(handle, itemID)` and `ZPRestoreVersion(handle, itemID, version)` to Phase 1 exports
2. Verify/wire version.Manager into item.Manager in the bridge layer if not already connected

---

### GAP-05: `Lock()` Method has No Return Value

**Severity:** CRITICAL  
**Phases affected:** 1

Plan shows `ZPLock(handle C.long) C.ZPResult` — expects a `ZPResult` return. But `store.Vault.Lock()` signature is:
```go
func (v *Vault) Lock()   // no error return!
```

Lock does `EncryptIndexFile` + `ZeroBytes` but silently ignores errors. The bridge export returns `ZPResult` which expects an error field. Inconsistency — bridge must handle this.

Also: Lock triggers index encryption which does file I/O. If file I/O fails, vault key is still zeroed (data loss of index until next rebuild). This should at minimum log an error.

**Fix:** Bridge `ZPLock` should call `Lock()` and return success (no error possible from signature). But consider adding error return to `Lock()` in core for index encryption failure awareness.

---

### GAP-06: Deployment Target Mismatch: PRD says macOS 13+, Plan says macOS 14+

**Severity:** CRITICAL  
**Phases affected:** 2, all UI phases

PRD Section 9 says:
> **Compatibility:** macOS 13+ (Ventura and later)

Plan says:
> **macOS 14 (Sonoma)** deployment target for @Observable + NavigationSplitView maturity

These are incompatible. `@Observable` requires macOS 14+ (Swift 5.9, Xcode 15). If you target macOS 13, you'd need `ObservableObject` + `@Published` fallback.

**Fix:** Update PRD to macOS 14+ or accept macOS 14 minimum in plan. Recommendation: **macOS 14+** is correct — @Observable is a huge DX win, and Sonoma adoption is >85% among Mac users by now.

---

## Medium Gaps (Should Fix Before Implementation)

### GAP-07: No `ZPGetRecoveryKey` Export — but Listed in File

Phase 1 file table lists `bridge/recovery_api.go` with functions "ZPGetRecoveryKey, ZPRegenerateRecovery, ZPUnlockWithRecovery". However:
- `ZPGetRecoveryKey` does NOT exist in the exported functions inventory (~35 functions list)  
- Go core has no "get recovery key" — the mnemonic is returned only during creation or regeneration
- This function makes no sense — the mnemonic is never stored in plaintext

**Fix:** Remove `ZPGetRecoveryKey` from the file description. Recovery key is returned by `ZPCreateVault` (mnemonic) and `ZPRegenerateRecovery` (new mnemonic). Phase 8 "View Recovery Key" button should be removed or replaced with "Regenerate Recovery Key" only.

---

### GAP-08: TouchID Architecture Gap — Vault Key vs Password

Phase 3 says:
> TouchID: store vault key in Keychain with biometric ACL after first password unlock
> Passes vault key directly to Go bridge (bypasses password)

But there's no `ZPUnlockWithVaultKey(handle, vaultKeyData)` in the exports! Only:
- `ZPUnlock(handle, masterPassword)` — derives key via Argon2
- `ZPUnlockWithRecovery(handle, mnemonic)` — derives key via BIP-39

For TouchID to work without re-entering password, you need a way to unlock with the raw vault key. The Go core's `Vault.Unlock()` takes a password and does KDF.

**Fix:** Add `ZPUnlockWithKey(handle C.long, vaultKeyBase64 *C.char) C.ZPResult` that bypasses KDF and directly sets the vault key. This needs a new method in `store.go`:
```go
func (v *Vault) UnlockWithKey(vaultKey []byte) error
```
Security: validate the key actually decrypts a test item before accepting it.

---

### GAP-09: Import Functions Take `io.Reader`, Not File Paths

Phase 1 bridge shows:
```
ZPImportCSV(handle C.long, path *C.char) C.ZPResult
```

But actual Go import functions take `io.Reader`:
```go
func ImportChrome(reader io.Reader) ([]types.Item, error)
```

The bridge layer must:
1. Open the file from path
2. Create reader
3. Call import function
4. Get returned items
5. Add each item via `Manager.AddItem()`
6. Return count

This is a non-trivial orchestration step NOT documented in Phase 1. The import functions return `[]types.Item` but don't save to vault — the bridge must loop and add each item.

**Fix:** Document in Phase 1 that bridge import functions must:
1. `os.Open(path)` → reader
2. Call specific importer → `[]types.Item`
3. Loop: `manager.AddItem(&item)` for each
4. Return `{"imported": count}`

---

### GAP-10: No Export File Path Handling in Bridge

Same issue as GAP-09 but for export. `ExportJSON/ExportCSV/ExportEncrypted` all take `io.Writer`. Bridge must:
1. Create output file
2. Get all items via `Manager.AllItems()`
3. Pass to export function
4. Close file
5. For encrypted export: need vault key

**Fix:** Document export bridge orchestration. Also: `ExportEncrypted` needs a key parameter — plan doesn't specify whether to use vault key or a separate user-provided password for export encryption.

---

### GAP-11: `SyncClient` Requires Device Registration First

Phase 1 exports list `ZPSyncSetup`, `ZPSyncPull`, `ZPSyncPush`, `ZPSyncFull`. But `sync_client.go` shows:
- `NewSyncClient(serverURL, deviceID, opts...)` — creates client
- `Register(deviceName string)` — registers device with server (must happen first!)
- `Pull()`, `Push()`, `FullSync()` — actual sync ops

The bridge needs to:
1. Persist sync config (server URL, device ID, API key) — WHERE? Not in vault.json currently
2. Handle device registration as separate step
3. Store `lastSyncTime` persistently
4. Create `SyncClient` on demand with stored config

Currently no sync config persistence exists in the Go core.

**Fix:** Add sync config storage (perhaps in `vault.json` Config or a separate `sync.json`). Add `ZPSyncRegister` to exports. Document stateful lifecycle of sync client in bridge.

---

### GAP-12: Bridge Handle Registry — Single vs Multi Vault

Phase 1 handle registry allows multiple vaults (`map[int64]*vaultSession`). But the app design throughout Phases 3-8 assumes **single vault** (one unlock, one item list, etc.).

The PRD also seems single-vault for MVP. CLI currently uses `~/.zeropass/vaults/default/`.

**Fix:** Decision needed: support multi-vault in bridge (future-proof) or simplify to single vault? Recommendation: keep multi-vault in bridge API (cheap), but Swift VaultManager can just track one handle.

---

### GAP-13: PRD Lists Passkey Type but No Passkey Auth Flow

Item type `passkey` exists in types.go with fields (credential_id, rp_id, etc.). Phase 5 has `PasskeyDetailView.swift`.

But passkeys are credentials FOR other services, not an auth mechanism for ZeroPass itself. The plan correctly treats them as viewable items. However — there's no provision for actually USING passkeys (acting as a Credential Provider to autofill into browsers/apps).

PRD says "Passkey provider (Phase 3)" is out of scope, but displaying passkey items without the ability to use them is confusing UX.

**Fix:** Add a note in Phase 5 that passkey items are "view/copy only" for MVP, with a placeholder UI indicating "Credential Provider coming in Phase 3."

---

### GAP-14: Quick Search Window — SwiftUI `Window` Scene Limitations

Phase 6 uses `Window("Quick Search", id: "quick-search")` with `.windowStyle(.hiddenTitleBar)`.

Issues:
1. SwiftUI `Window` scenes can't be opened from global hotkey outside the app easily — need `NSApp.activate()` first
2. `.hiddenTitleBar` doesn't give the true "floating panel" behavior (no entry in Dock, no window management)
3. Better approach: `NSPanel` with `.nonactivatingPanel` behavior (used by spotlight-style UIs)

**Fix:** Phase 6 should use `NSPanel` wrapped in `NSViewRepresentable` or AppKit coordinator, not pure SwiftUI Window scene. HotkeyService needs to `NSApp.activate(ignoringOtherApps: true)` before showing panel.

---

## Minor Issues

### MINOR-01: Research Report Shows c-shared, Plan Uses c-archive

Research SwiftUI report (section 2) initially recommends **c-shared (.dylib)** as "Option B", while the interop research and plan correctly chose **c-archive (.a)**. The SwiftUI report's code sample shows `libzeropass.dylib` in the project structure. Confusing but non-blocking — c-archive is the correct choice.

### MINOR-02: `ZPIsLocked` Return Type

Plan shows `ZPIsLocked(handle C.long) C.int` — returns int, not ZPResult. This breaks the uniform ZPResult pattern. Consider returning ZPResult with `{"locked": true/false}` for consistency, or keep the C.int for minor performance gain (this is a hot method).

### MINOR-03: TOTP Timer Not Planned

Phase 5 mentions "TOTP: show countdown timer if TOTP seed present" but there's no TOTP code generation in Go core. The `totp` field stores a seed/URI but no function generates the time-based code. Either implement TOTP generation in Go bridge or Swift-side library (e.g., SwiftOTP).

### MINOR-04: Effort Estimates Undercount

Total planned effort: 24 days. With the gaps identified (especially GAP-01, GAP-02, GAP-08 requiring Go core changes), add ~3-4 days. Real estimate: **28-30 days**.

### MINOR-05: No Error Recovery in Auto-Lock

`store.go` `startAutoLock` fires `v.Lock()` via `time.AfterFunc`. If this runs while a bridge operation is mid-flight (e.g., decrypting an item), the vault key gets zeroed and the operation fails mid-way. Need a read-write lock counter or "operation in progress" guard.

### MINOR-06: Swift Conventions Mismatch

`swift-conventions.instructions.md` specifies:
- Modular Swift Packages under `Modules/` — but plan uses `macos/ZeroPass/` flat structure
- Protocol suffix `Protocol` — but plan uses no protocol-based DI for bridge
- File naming matches primary type — plan follows this

Conventions are from a different project context. No action needed unless strict compliance required.

---

## Architectural Observations (Not Bugs, Just Notes)

### OBS-1: Index Encryption/Decryption on Lock/Unlock Cycle

Every `Vault.Unlock()` calls `DecryptIndexFile()` and every `Lock()` calls `EncryptIndexFile()`. This is file-level AES encryption of the entire SQLite database. With large vaults (10K+ items), this adds latency to lock/unlock beyond KDF time.

### OBS-2: No Batch Operations in Bridge

All item operations are one-at-a-time (`ZPCreateItem`, `ZPGetItem`, etc.). For import of 1000+ items, this means 1000+ FFI calls + 1000+ index updates. Consider adding `ZPImportBatch` that takes JSON array.

### OBS-3: Memory Pressure from JSON-over-FFI

For vaults with many items, `ZPListItems` returns ALL items as one JSON blob. With 1000 items × ~500 bytes each = ~500KB of JSON allocated in C heap. Not a problem for typical use but consider pagination for large vaults.

### OBS-4: Go AutoLock Timer vs Swift AutoLock

Both the Go core (`store.go` auto-lock timer) and Swift (`AutoLockService` in Phase 7) implement auto-lock. These will conflict. The Go-side timer calls `Lock()` internally, while Swift's service also tries to lock via bridge. 

**Recommendation:** Disable Go-side auto-lock when used via bridge (set `AutoLockTimeout=0` in VaultConfig). Let Swift handle all auto-lock logic since it has UI context (idle detection, sleep notifications).

---

## Prioritized Action Items

| Priority | Gap | Effort | When |
|----------|-----|--------|------|
| P0 | GAP-01: Implement ChangeMasterPassword | 0.5d | Before Phase 1 |
| P0 | GAP-02: Implement flock file locking | 0.5d | Phase 1 |
| P0 | GAP-03: ZPCreateVault return + mnemonic security | 0.25d | Phase 1 |
| P0 | GAP-06: Confirm deployment target macOS 14+ | Decision | Before Phase 2 |
| P0 | GAP-08: Add ZPUnlockWithKey for TouchID | 0.5d | Phase 1 |
| P1 | GAP-04: Add version history exports | 0.25d | Phase 1 |
| P1 | GAP-05: Bridge Lock error handling | 0.1d | Phase 1 |
| P1 | GAP-09: Document import bridge orchestration | 0.1d | Phase 1 |
| P1 | GAP-10: Document export bridge orchestration | 0.1d | Phase 1 |
| P1 | GAP-11: Sync config persistence + registration | 0.5d | Phase 1 |
| P1 | GAP-14: Use NSPanel for quick search | 0.25d | Phase 6 |
| P2 | GAP-07: Remove phantom ZPGetRecoveryKey | 0.05d | Phase 1 |
| P2 | GAP-12: Single vs multi-vault decision | Decision | Before Phase 1 |
| P2 | GAP-13: Passkey "view only" note | 0.05d | Phase 5 |
| P2 | OBS-4: Disable Go auto-lock in bridge mode | 0.1d | Phase 1 |

**Total additional effort: ~3-4 days**

---

## Unresolved Questions

1. **Export encryption key:** Should encrypted export use vault key or a user-provided separate password? User-provided is more portable but requires UI for password entry.
2. **Clipboard service scope:** Phase 7 ClipboardService clears clipboard after N seconds. If user copies something else before timer fires, should we still clear? (Plan says check if clipboard still contains our string — good.)
3. **Hardened Runtime + Go:** WireGuard ships with `com.apple.security.cs.allow-unsigned-executable-memory` — confirmed via code search. Go runtime definitely needs this. But Apple may reject or warn. Need to test early (Phase 2, not Phase 9).
4. **Auto-update signing:** Sparkle 2.7+ uses EdDSA. Who manages the key pair? Where stored? Not addressed in Phase 9.
5. **Index rebuild on corruption:** If index.db gets corrupted (crash during encrypt/decrypt), what's the recovery path? `RebuildIndex` exists but requires all items decrypted first (vault must be unlocked).
