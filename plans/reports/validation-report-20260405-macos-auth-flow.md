# macOS Auth Flow Validation Report
**Date:** April 5, 2026  
**Focus:** WelcomeView, UnlockVaultView, AuthWindowLayoutModifier  
**Status:** ⚠️ Test Failure Detected

---

## Test Execution Summary

### Overall Results
- **Total Tests:** 18
- **Passed:** 17 ✅
- **Failed:** 1 ❌
- **Success Rate:** 94%

### Test Breakdown

#### Unit Tests (Passed)
- ✅ RecoveryPhraseSupportTests (2 tests)
- ✅ ZeroPassTests (13 tests)

#### UI Tests
- ✅ testCreateVaultSheetCanBeOpenedAndCancelled (8.8s)
- ✅ testLaunchPerformance (29.3s)
- ⚠️ **testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell (FAILED - 35.0s)**

---

## Failed Test Analysis

### Test: `testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell`
**Location:** [ZeroPassUITests.swift](apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift#L192)  
**Failure Point:** Line 338

#### Error Details
```
Element SecureTextField, {{3415.0, 756.0}, {450.0, 18.0}}, 
identifier: 'createVault.masterPasswordField', 
label: 'Master password', 
placeholderValue: 'Enter a master password' is not hittable
```

#### Test Flow
1. Launch fresh app (reset state)
2. Open create vault sheet → ✅ Works
3. Select vault location → ✅ Works
4. **Click master password field → ❌ FAILS** (element not hittable)
5. (Expected) Enter password
6. (Expected) Complete vault creation
7. (Expected) Verify recovery phrase screen
8. (Expected) Click continue → unlocked shell

#### Root Cause Analysis
The master password SecureTextField in CreateVaultView is rendering at position **{3415.0, 756.0}** which is **off-screen**. 
- Position 3415.0 indicates excessive horizontal offset
- Standard macOS window width: ~500-600 logical points max
- **Conclusion:** Layout constraint issue causing field to render outside visible bounds

---

## Build Status

### Compilation: ✅ SUCCESS
- No errors in modified files
- `AuthWindowLayoutModifier.swift` → ✅ Compiles cleanly
- `UnlockVaultView.swift` → ✅ Compiles cleanly  
- `WelcomeView.swift` → ✅ Compiles cleanly

### Build Output
```
BUILD SUCCEEDED
```

---

## Compilation Warnings

### Modified Files (Scope)
No warnings or errors detected in the three target files.

### Other File Warnings (Context)
- AutoLockService (2 warnings) - Non-Sendable type captures in @Sendable closure
- RecoveryPhraseSupport (1 warning) - MainActor isolation mismatch
- ZPBridge (30+ warnings) - MainActor conformance isolation issues in Swift 6 language mode

**Assessment:** Unrelated to recent auth-flow changes.

---

## Risk Assessment

### 🔴 Critical Risk
1. **UI Test Failure in Core Flow** 
   - Create vault is critical onboarding path
   - Security-sensitive password field failing affects user experience
   - Regression: Element becomes unhittable due to layout changes

### 🟡 Medium Risk
2. **Window Layout Modifier Impact**
   - `AuthWindowLayoutModifier` applies layout changes to auth window
   - Changes may have introduced an off-screen rendering state
   - Could affect other auth flows (not yet detected in tests)

### 🟢 Low Risk  
3. **Other Warnings**
   - Pre-existing concurrency warnings unaffected by changes
   - Non-critical to functionality

---

## Files Modified

| File | Status | Issues |
|------|--------|--------|
| [WelcomeView.swift](apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift) | ✅ Compiles | None |
| [UnlockVaultView.swift](apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift) | ✅ Compiles | None |
| [AuthWindowLayoutModifier.swift](apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift) | ✅ Compiles | **Suspect** |

---

## Regressions Detected

### Primary Regression
- **Symptom:** CreateVaultView password field off-screen, not hittable
- **Impact:** Blocks vault creation flow (critical onboarding function)  
- **Introduced By:** `AuthWindowLayoutModifier` layout calculations
- **Evidence:**
  - Field renders at x=3415 (far off-screen)
  - Same field hittable in `testCreateVaultSheetCanBeOpenedAndCancelled` 
  - Difference: layout modifier application before/after credentials entry

### Secondary Risk
- Other auth UI fields may also be affected by layout calculation
- Not yet caught by existing tests

---

## Next Steps

### 1. Immediate Investigation Required
- Review `AuthWindowLayoutModifier.apply(to:rememberedUnlockedSize:)` logic
- Check window size/frame calculations in unlocked vs locked states
- Verify layout calculations don't exceed screen bounds

### 2. Potential Fixes  
- Validate window frame bounds before applying layout
- Add clipping or safe-frame boundaries
- Test CreateVaultView frame output directly

### 3. Test Coverage Gap
- Add frame validation test for auth views
- Test SecureField hittability in all auth sheets
- Add regression test for off-screen rendering

---

## Recommendations

**Priority 1 (Blocker):** Fix AuthWindowLayoutModifier to prevent off-screen rendering  
**Priority 2:** Add frame validation tests  
**Priority 3:** Address Swift 6 concurrency warnings in ZPBridge  

---

## Unresolved Questions

1. Why does `testCreateVaultSheetCanBeOpenedAndCancelled` pass if layout is broken? (Different window state?)
2. Does unlock flow (UnlockVaultView) have the same off-screen rendering issue?
3. What conditions trigger the layout calculation error?

