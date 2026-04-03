# Phase 3: Recovery Key Auto-Rotation After Recovery Unlock

## Context

- **Priority:** HIGH | **Effort:** 2h | **Risk:** Medium
- PRD §8.5: "One-time use: recovery key is rotated after successful recovery"
- Current `UnlockWithRecovery()` returns `error` only — same mnemonic works forever
- `RegenerateRecovery()` exists in store.go but is NOT called after recovery unlock
- Bridge has `ZPUnlockWithRecovery` but NO `ZPRegenerateRecovery` export
- **Decision:** Auto-rotate + force user to record new mnemonic

## Related Files

| Action | File |
|--------|------|
| MODIFY | `core/vault/store/store.go` |
| MODIFY | `packages/cli/cmd/unlock.go` |
| MODIFY | `bridge/vault_api.go` |
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` |
| NEW | Test: `core/vault/store/store_test.go` (add TestUnlockWithRecoveryAutoRotation) |

## Implementation Steps

### 1. Modify `UnlockWithRecovery()` signature in store.go

Change return type to include new mnemonic:

```go
func (v *Vault) UnlockWithRecovery(mnemonic string) (newMnemonic string, err error) {
```

After successful decrypt + unlock, auto-rotate:

```go
// Auto-rotate recovery key (PRD: one-time use)
newMnemonic, err = v.regenerateRecoveryLocked(vaultKey)
if err != nil {
    // Unlock succeeded but rotation failed — still unlocked
    return "", fmt.Errorf("vault unlocked but recovery rotation failed: %w", err)
}
return newMnemonic, nil
```

### 2. Add private `regenerateRecoveryLocked()` helper

Same logic as `RegenerateRecovery()` but assumes mutex is already held (avoids deadlock):

```go
func (v *Vault) regenerateRecoveryLocked(vaultKey []byte) (string, error) {
    mnemonic, encRecovery, err := key.RegenerateRecoveryKey(vaultKey)
    if err != nil {
        return "", fmt.Errorf("regenerate recovery key: %w", err)
    }
    v.meta.EncryptedRecoveryKey = encRecovery
    if err := v.persistMetaLocked(); err != nil {
        return "", fmt.Errorf("persist recovery key: %w", err)
    }
    return mnemonic, nil
}
```

### 3. Update CLI unlock.go

After recovery unlock, display new mnemonic with VIP box:

```go
if unlockRecovery {
    mnemonic := promptLine("Enter recovery mnemonic: ")
    newMnemonic, err := v.UnlockWithRecovery(mnemonic)
    if err != nil {
        return fmt.Errorf("unlock with recovery: %w", err)
    }
    fmt.Fprintln(cmd.ErrOrStderr(), "✓ Vault unlocked with recovery phrase.")
    fmt.Fprintln(cmd.ErrOrStderr())
    fmt.Fprintln(cmd.ErrOrStderr(), "╔══════════════════════════════════════════════════════════════╗")
    fmt.Fprintln(cmd.ErrOrStderr(), "║  ⚠ RECOVERY KEY HAS BEEN ROTATED                           ║")
    fmt.Fprintln(cmd.ErrOrStderr(), "║                                                             ║")
    fmt.Fprintln(cmd.ErrOrStderr(), "║  Your previous recovery phrase is now INVALID.              ║")
    fmt.Fprintln(cmd.ErrOrStderr(), "║  Write down your NEW recovery phrase:                       ║")
    fmt.Fprintf(cmd.ErrOrStderr(),  "║  %s\n", newMnemonic)
    fmt.Fprintln(cmd.ErrOrStderr(), "║                                                             ║")
    fmt.Fprintln(cmd.ErrOrStderr(), "╚══════════════════════════════════════════════════════════════╝")
    if !promptYesNo("I have saved my new recovery phrase") {
        fmt.Fprintln(cmd.ErrOrStderr(), "⚠ Please save it NOW. This phrase will NOT be shown again.")
        _ = promptYesNo("I confirm I have saved my new recovery phrase")
    }
}
```

### 4. Update bridge vault_api.go

Change `ZPUnlockWithRecovery` to return new mnemonic string:

```go
//export ZPUnlockWithRecovery
func ZPUnlockWithRecovery(handle C.long, mnemonic *C.char) (res C.ZPResult) {
    // ... existing validation ...
    newMnemonic, err := s.vault.UnlockWithRecovery(m)
    if err != nil {
        return errorResult(err)
    }
    // ... ensure managers ...
    return okJSON(map[string]string{"newMnemonic": newMnemonic})
}
```

### 5. Update VaultClient.swift

Handle new mnemonic from bridge response:

```swift
func unlockWithRecovery(mnemonic: String) async throws -> String {
    guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
    let result = try await Task.detached(priority: .userInitiated) {
        try ZPBridge.unlockWithRecovery(handle: h, mnemonic: mnemonic)
    }.value
    // Parse newMnemonic from result JSON
    state = .unlocked
    autoLock.setVaultUnlocked(true)
    autoLock.recordActivity()
    try await refreshItems()
    await persistVaultKeyIfNeeded()
    return result.newMnemonic
}
```

Transition to `.showingRecovery(newMnemonic)` state after recovery unlock in the view layer.

### 6. Add test: TestUnlockWithRecoveryAutoRotation

```go
func TestUnlockWithRecoveryAutoRotation(t *testing.T) {
    // Setup: create vault, get initial mnemonic
    // Act: UnlockWithRecovery(oldMnemonic) → returns newMnemonic
    // Assert: oldMnemonic no longer works
    // Assert: newMnemonic works for subsequent recovery
    // Assert: each recovery produces a different new mnemonic
}
```

## Risk Mitigation

- **Deadlock risk:** Using `regenerateRecoveryLocked()` (no mutex re-acquisition) instead of public `RegenerateRecovery()`
- **Partial failure:** If rotation fails after unlock, vault is still usable — error message indicates the state clearly
- **Bridge compatibility:** Changing return type from `okNoData()` to `okJSON(...)` — Swift must parse the new response format

## Todo List

- [x] Change `UnlockWithRecovery()` signature to return `(string, error)`
- [x] Add `regenerateRecoveryLocked()` helper
- [x] Update all callers of `UnlockWithRecovery()` in store
- [x] Update CLI unlock.go with VIP box display
- [x] Update bridge `ZPUnlockWithRecovery` to return newMnemonic
- [x] Update VaultClient.swift to handle new mnemonic
- [x] Update Swift view layer for recovery rotation UI
- [x] Add TestUnlockWithRecoveryAutoRotation
- [x] Run full test suite

## Success Criteria

- Old mnemonic no longer works after recovery unlock
- New mnemonic returned works for subsequent recovery
- CLI displays VIP box with new mnemonic
- macOS app shows recovery rotation screen
- All existing tests pass
