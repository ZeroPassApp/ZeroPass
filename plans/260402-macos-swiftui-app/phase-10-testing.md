# Phase 10: Testing & QA

## Context
- Depends on: All implementation phases

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Comprehensive testing: Go bridge unit tests, Swift unit tests, UI tests, integration tests, performance benchmarks.

## Test Strategy

### Layer 1: Go Bridge Tests (`bridge/bridge_test.go`)
- Test each exported function at the Go level (before C boundary)
- Vault lifecycle: create → open → unlock → CRUD → lock → close
- Error paths: locked vault access, bad password, invalid JSON
- Memory: verify no leaks in handle registry
- Panic recovery: verify no crashes on edge inputs
- **`ZPUnlockWithKey`**: unlock with raw vault key bytes (bypasses Argon2)
- **`ZPChangeMasterPassword`**: verify old password → new password flow, wrong old password rejection
- **`ZPAcquireLock` / `ZPReleaseLock`**: flock advisory locking — verify exclusive lock prevents second open
- **`ZPGetVersionHistory` / `ZPRestoreVersion`**: create item, update 3x, verify 3 versions, restore version 1, verify content
- **Flock concurrency**: two goroutines racing to open same vault — only one succeeds

### Layer 2: Swift Bridge Tests (`ZeroPassTests/BridgeTests.swift`)
- Round-trip: Swift → C → Go → JSON → Swift decoding
- Memory: BridgeResult properly frees C allocations
- Async: heavy operations don't block main thread
- Error mapping: Go error codes → Swift ZeroPassError
- **UnlockWithKey round-trip**: create vault → get vault key → lock → unlock with key → verify items accessible
- **ChangeMasterPassword round-trip**: change password → lock → unlock with new password → verify old password fails
- **Mnemonic zeroing**: after ZPCreateVault, verify C-heap result is freed and Swift copy is zeroed after use
- **JSON date decoding**: verify ISO 8601 dates decode correctly with `.iso8601` strategy

### Layer 3: Swift Unit Tests (`ZeroPassTests/`)
- VaultManager state transitions (locked/unlocked)
- ItemFilter logic
- AppSettings persistence
- ClipboardService timer behavior
- KeychainService store/retrieve (mock Keychain or Keychain test target)
- **AutoLockService**: verify Swift-side timer fires correctly, verify no conflict with Go-side timer
- **BiometricService**: mock LAContext, verify unlock-with-key flow (Keychain → ZPUnlockWithKey)
- **QuickSearchPanel (NSPanel)**: verify panel shows/hides on hotkey, non-activating behavior

### Layer 4: UI Tests (`ZeroPassUITests/`)
- Setup wizard flow: password → recovery → verify → complete
- Unlock: correct password → unlocked, wrong password → error
- Item CRUD: create → verify in list → edit → delete
- Search: type query → results appear → select → detail shown
- Settings: change preference → verify applied
- Keyboard shortcuts: ⌘N, ⌘K, ⌘⇧L
- **Change Master Password**: open settings → security → change password → verify new password works
- **Version History**: edit item 3x → open version history → verify 3 entries → restore oldest → verify content
- **Auto-lock transition**: set 1-second timeout → wait → verify UI transitions to UnlockView
- **Passkey view-only**: open passkey item → verify no "Use" button (view/copy only)

### Layer 5: Performance Tests
- Vault unlock latency: < 300ms (excluding KDF)
- **Vault unlock with key (TouchID path)**: < 50ms (no KDF, raw key)
- Search latency: < 50ms for 1000 items
- App launch to ready: < 1 second
- FFI round-trip overhead: < 1ms per call
- **Version history retrieval**: < 100ms for item with 50 versions

## Files to CREATE

| File | Description |
|------|-------------|
| `bridge/bridge_test.go` | Go bridge integration tests (including flock, UnlockWithKey, ChangeMasterPassword, version history) |
| `bridge/bridge_bench_test.go` | Go bridge benchmarks (including UnlockWithKey latency) |
| `macos/ZeroPassTests/BridgeTests.swift` | Swift-Go bridge tests (including key unlock, password change) |
| `macos/ZeroPassTests/VaultManagerTests.swift` | State management tests |
| `macos/ZeroPassTests/BridgeTypesTests.swift` | JSON decoding tests |
| `macos/ZeroPassUITests/SetupFlowTests.swift` | UI: setup wizard |
| `macos/ZeroPassUITests/UnlockFlowTests.swift` | UI: unlock flow |
| `macos/ZeroPassUITests/ItemCRUDTests.swift` | UI: create/edit/delete |
| `macos/ZeroPassUITests/SearchTests.swift` | UI: search + quick search |

## Success Criteria
- [ ] Go bridge tests pass (100% of exported functions covered, including new ZPUnlockWithKey, ZPChangeMasterPassword, ZPAcquireLock/ReleaseLock, ZPGetVersionHistory/RestoreVersion)
- [ ] Flock concurrency test: two processes cannot open same vault simultaneously
- [ ] Swift bridge tests pass (round-trip JSON decoding, mnemonic zeroing)
- [ ] Unit tests cover VaultManager state machine + AutoLockService + BiometricService
- [ ] UI tests cover critical flows (setup, unlock, CRUD, search, change password, version history)
- [ ] Performance within PRD targets (including < 50ms for UnlockWithKey)
- [ ] No memory leaks in Instruments (Leaks + Allocations)
- [ ] No crashes on edge cases (empty vault, locked operations, bad input, auto-lock during operation)
