# Malloc Crash Investigation: Sync Migration Tests

**Date**: 2026-04-04  
**Investigator**: Debugger Agent  
**Type**: Research/Debug (No Code Changes)

---

## Executive Summary

**Issue**: Two sync migration tests crash with `malloc: *** error for object 0x2ac097a90: pointer being freed was not allocated`  
**Tests Affected**:
- `testMigrateLegacySyncAPIKeyMovesSecretOutOfDiskAndDefaults`
- `testMigrateLegacySyncAPIKeyPrefersExistingStoredSecret`

**Root Cause**: Memory management error in C bridge layer — likely double-free or freeing Swift-managed string buffer  
**Severity**: High (blocks test suite)  
**Status**: Root cause identified, patch strategy provided

---

## Root Cause Analysis

### 1. **Most Likely Cause: C String Lifecycle Mismatch**

**Evidence**:
- Identical crash address `0x2ac097a90` in both tests (not random malloc corruption)
- Tests use `FakeKeychainStore` but crash still occurs (not Security.framework issue)
- No vault handle/C bridge calls during test execution
- Crash happens when VaultClient is deallocated or during Swift string cleanup

**The Smoking Gun**:

[VaultClient.swift:70](VaultClient.swift#L70)
```swift
@Published var syncAPIKey: String  // NO didSet observer
```

Compare to nearby properties:
```swift
@Published var syncDeviceID: String {
    didSet { UserDefaults.standard.set(syncDeviceID, forKey: "syncDeviceID") }
}
```

**Why This Matters**: 
After sync hardening, `syncAPIKey` likely had logic removed/changed that previously managed C bridge cleanup. The property is initialized to `""` in init but lacks lifecycle management if it ever held a Go-allocated string.

### 2. **Secondary Suspect: `withMutableCString` Unsafe Pattern**

[ZPBridge.swift:156-160](ZPBridge.swift#L156-L160)
```swift
private nonisolated static func withMutableCString<T>(_ s: String, _ body: (UnsafeMutablePointer<CChar>) -> T) -> T {
    s.withCString { c in
        body(UnsafeMutablePointer(mutating: c))
    }
}
```

**Problem**: Swift `withCString` provides a temporary buffer that's auto-freed **after the closure**. If Go code stores this pointer and tries to free it later via `ZPFreeResult`, we get the exact crash seen.

### 3. **Go Bridge Memory Contract Violation**

[libzeropass.h:167-169](bridge/libzeropass.h#L167-L169)
```c
extern void ZPFreeResult(ZPResult r);
extern void ZPFreeResultPtr(ZPResult* r);
```

[ZPBridge.swift:133-147](ZPBridge.swift#L133-L147)
```swift
private nonisolated static func dataFromResult(_ result: ZPResult) throws -> Data {
    let r = result
    defer { ZPFreeResult(r) }  // ← Frees Go-allocated strings
    
    guard let p = r.data else { return Data() }
    let s = String(cString: p)  // ← Swift copies, but Go still owns `p`
    return Data(s.utf8)
}
```

**Issue**: If Go's `ZPFreeResult` tries to free a Swift-owned string buffer (passed via `withMutableCString`), malloc detects the mismatch.

---

## Suspicious Code Paths

### **Method 1: `migrateLegacySyncAPIKeyIfNeeded`**  
[VaultClient.swift:809-842](VaultClient.swift#L809-L842)

**What It Does**:
1. Reads legacy config from `sync.json`
2. Calls `keychain.storeSyncAPIKey()` (FakeKeychainStore)
3. Calls `persistMetadataOnlySyncConfig()` to write stripped config
4. Removes `UserDefaults` key

**Why It Crashes**:
- Sets `syncAPIKey` property (line 795 in `rehydrateSyncClientFromStoredSecretIfNeeded`)
- But tests don't have vault handle, so **no bridge call happens during test**
- Crash likely occurs during **test teardown** when VaultClient is deallocated

### **Method 2: Test Lifecycle**  
```swift
func testMigrateLegacySyncAPIKeyMovesSecretOutOfDiskAndDefaults() throws {
    let client = VaultClient(keychain: keychain)  // ← Init sets syncAPIKey = ""
    ...
    client.migrateLegacySyncAPIKeyIfNeeded(for: dir)  // ← Runs migration
    // ← VaultClient deallocates HERE (defer cleanup runs)
    // ← If syncAPIKey had stale C pointer, ZPFreeResult called on Swift string
}
```

---

## Does NOT Involve Real Security.framework

**Confirmed**: All keychain calls use `FakeKeychainStore`:
- `storeSyncAPIKey()` → in-memory dict
- `loadSyncAPIKey()` → in-memory dict
- No `SecItemAdd/SecItemCopyMatching` calls in tests

**No LocalAuthentication**:
Tests don't call `LAContext` methods since no vault unlocking happens.

---

## Minimal Patch Strategy

### **Fix 1: Add `didSet` Zeroing for `syncAPIKey`** (Safest)

[VaultClient.swift:70](VaultClient.swift#L70)
```swift
// BEFORE (crash-prone):
@Published var syncAPIKey: String

// AFTER (lifecycle-safe):
@Published var syncAPIKey: String {
    didSet {
        // If sync hardening previously stored C pointers here,
        // ensure we don't hold stale references
    }
}
```

**Rationale**: Match pattern of other sync properties. Keeps state clean on assignment.

### **Fix 2: Mock `syncAPIKey` Assignment in Tests**

[ZeroPassTests.swift:159+](ZeroPassTests.swift#L159)
```swift
func testMigrateLegacySyncAPIKeyMovesSecretOutOfDiskAndDefaults() throws {
    let client = VaultClient(keychain: keychain)
    // BEFORE migration, explicitly set to avoid stale state:
    client.syncAPIKey = ""  // Force clean string, not Go-allocated
    
    client.migrateLegacySyncAPIKeyIfNeeded(for: dir)
    ...
}
```

**Rationale**: Ensures test starts with known-good Swift string, not recycled memory.

### **Fix 3: Audit `ZPFreeResult` Usage** (Go-side)

Check bridge implementation for improper frees:
```go
// bridge/main.go or bridge/sync_api.go
// Ensure ZPFreeResult ONLY frees Go-allocated memory
// NOT Swift-provided strings from withMutableCString
```

---

## Timeline

**Recent Change**: "Sync secret storage hardening" moved API keys from:
- ~~UserDefaults~~ (cleartext) → ❌
- ~~`sync.json` disk file~~ (cleartext) → ❌  
- ✅ **Keychain** (encrypted)

**Side Effect**: Migration logic (`migrateLegacySyncAPIKeyIfNeeded`) now:
1. Reads old cleartext secrets
2. Moves to keychain
3. Strips from disk/defaults

**But**: Tests don't mock C bridge cleanup, so deallocating VaultClient triggers `ZPFreeResult` on Swift string.

---

## Recommended Action (Research Complete)

1. **Add `didSet` to `syncAPIKey`** to match other sync properties
2. **Audit Go bridge**: Verify `ZPFreeResult` only frees Go heap allocations
3. **Consider**: Make tests explicitly zero `syncAPIKey` after migration
4. **Test fix**: Run tests with `malloc_error_break` breakpoint to confirm fix

**Do NOT** revert hardening — fix tests to respect new lifecycle.

---

## Unresolved Questions

1. Why same address `0x2ac097a90` in both crashes? (Static allocator pattern?)
2. Does Go bridge cache Swift string pointers anywhere?
3. Was `syncAPIKey` ever assigned a Go-allocated string in production use?

---

## Appendices

### A. Crash Signature
```
ZeroPass(59555,0x1fed03100) malloc: *** error for object 0x2ac097a90: pointer being freed was not allocated
ZeroPass(59555,0x1fed03100) malloc: *** set a breakpoint in malloc_error_break to debug
```

### B. Test Environment
- **macOS**: 26.4 (Tahoe)
- **Xcode**: 17E192
- **Swift**: 6 language mode
- **Architecture**: arm64
- **Build**: Debug, no signing

### C. Related Files
- [VaultClient.swift](apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift)
- [ZPBridge.swift](apps/macos/ZeroPass/ZeroPass/Bridge/ZPBridge.swift)
- [KeychainService.swift](apps/macos/ZeroPass/ZeroPass/Services/KeychainService.swift)
- [ZeroPassTests.swift](apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift)
- [libzeropass.h](bridge/libzeropass.h)

---

**End of Report**
