# Tester Review: macOS Auth HIG Refactor

**Date:** April 4, 2026  
**Work Context:** /Volumes/DATA/Developments/ZeroPass/  
**Test Environment:** macOS arm64, Xcode UI test target  

---

## Executive Summary

**Current Status:** ✅ **BUILD PASSES** | ❌ **INSUFFICIENT UI/AUTH COVERAGE**

- 15 unit tests passed (0 failed)
- All tests are backend logic (vault operations, crypto, import/export)
- **ZERO automated tests cover the refactored auth UI components**
- High-risk auth flows untested by automation: welcome → create/open, unlock (password + recovery), recovery phrase display, keyboard navigation, sheet transitions, accessibility

**Risk Level:** 🔴 **HIGH** for UI regressions; ✅ **LOW** for backend  

**Recommendation:** Manual smoke test all 6 auth state transitions before shipping. Do NOT rely solely on unit tests. If time allows, add 2–3 focused UI smoke tests to catch future regressions.

---

## Test Coverage Analysis

### What's Tested (Backend-Only)

- ✅ Vault creation & unlock with password
- ✅ Password change round-trip
- ✅ Version history & restoration
- ✅ JSON export & CSV import
- ✅ Biometric key storage (via Keychain mock)

### What's NOT Tested (Auth UI)

- ❌ Welcome view state machine
- ❌ Create vault sheet (folder selection, password validation, strength indicator)
- ❌ Open vault sheet error handling
- ❌ Unlock view (password field, recovery phrase field, method toggle)
- ❌ Password section (caps lock warning, biometric button, show password toggle)
- ❌ Recovery section (phrase parsing, validation feedback, focus states)
- ❌ Recovery phrase display (copy button, continue flow)
- ❌ Window layout transitions (size, appearance, focus restoration)
- ❌ HIG compliance: keyboard shortcuts, accessibility labels/hints, dark mode, reduce motion, reduce transparency
- ❌ Error message rendering & tone (info/success/warning/error)
- ❌ Form validation UI feedback (highlight states, border colors)

### Test Count Breakdown

| Category | Count | Status |
|----------|-------|--------|
| Unit tests (backend) | 15 | ✅ All pass |
| UI tests (empty template) | 1 | ⚠️ Placeholder only |
| Auth component tests | 0 | ❌ **MISSING** |
| Integration tests | 0 | ❌ **MISSING** |

---

## Highest-Risk Manual Flows

**Priority order for smoke testing:**

### 1. **Welcome → Create Vault** (CRITICAL)
   - **Flow:** App launch → "Create Vault…" button → sheet opens → pick folder → set password → confirm → vault created
   - **What can break:** Sheet presentation, folder picker, password strength display, page focus after creation, error message if folder invalid
   - **Keyboard test:** Tab through buttons, use Tab + Space to create vault, verify focus returns to welcome on cancel
   - **Accessibility test:** Labels on buttons, error message announced

### 2. **Welcome → Open Vault** (CRITICAL)
   - **Flow:** "Open Vault…" button → folder picker → existing vault loads → unlock screen or main app
   - **What can break:** File picker crashes, wrong vault fails to detect, error message doesn't display, focus doesn't move to password field
   - **Keyboard test:** Tab/Shift+Tab, Cmd+O shortcut, default action on valid selection
   - **Accessibility test:** Sheet title, file path display readability, error tone

### 3. **Locked → Unlock with Password** (CRITICAL)
   - **Flow:** Wrong password → error highlight + message → retype → success → main app
   - **What can break:** Highlight doesn't appear, error clears too fast, focus returns to password field, caps lock warning missing, input validation fails
   - **Keyboard test:** Return key submits, Tab cycle (password → show toggle → biometric button → unlock button), Cmd+L lock + tab order
   - **Accessibility test:** "Master Password" label, caps lock warning visibility, error message tone

### 4. **Locked → Unlock with Recovery Phrase** (CRITICAL)
   - **Flow:** Toggle method → recovery phrase field → paste phrase → error if invalid → success
   - **What can break:** Toggle disabled during input, boundary validation (12/15/18/21/24 words), editor doesn't scroll, phrase display (icon/color changes) misses user, success doesn't clear message
   - **Keyboard test:** Tab enters/exits text editor, Return doesn't submit (recovery case), focus behavior in multiline editor
   - **Accessibility test:** Field label, placeholder visibility, validation feedback color + icon + text

### 5. **Create → Recovery Phrase Display** (HIGH)
   - **Flow:** Vault created → recovery phrase shown → "Copy" button → "Continue" → success
   - **What can break:** Copy button fails silently, copied state doesn't reset, clipboard auto-clear timing breaks, continue button doesn't move to main app, phrase not visually highlightable for manual copy
   - **Keyboard test:** Tab to copy/continue buttons, Return on continue, Cmd+C native copy if phrase selected
   - **Accessibility test:** Warning message tone, phrase readability, button purpose (copy vs continue)

### 6. **Window Size & Focus Transitions** (HIGH)
   - **Flow:** Welcome (small) → create sheet (overlay) → sheet closes → welcome (restore size) → unlock screen (possibly different size) → main app (large)
   - **What can break:** Sheet doesn't center, window resizes unexpectedly, focus lost after sheet, first field in new view not focused, appearance doesn't match HIG (light/dark switch during transition)
   - **Test method:** Manual observation of window frame, frame debug + screenshot comparison
   - **Accessibility test:** Reduce motion doesn't hide transitions; reduce transparency applied consistently

---

## Missing Automated Test Coverage Worth Adding

**Feasibility: Add 2–3 focused smoke tests in this session**

### Smoke Test 1: Welcome State & Button Accessibility (HIGH VALUE)
```
Goal: Verify WelcomeView renders, buttons are accessible, error dismisses.
Coverage: Welcome presence, button labels, keyboard shortcuts (Cmd+O), error message UI.
Time: 5 min to add.
```

### Smoke Test 2: Unlock Password Field Focus & Keyboard (MEDIUM VALUE)
```
Goal: Verify UnlockVaultView password field captures focus, keyboard shortcuts work, caps lock warning appears.
Coverage: Focus on app launch, Return key submits, Cmd+L lock shortcut, show password toggle.
Time: 8 min to add.
```

### Smoke Test 3: Recovery Sheet Cancellation & Focus Restoration (MEDIUM VALUE)
```
Goal: Verify CreateVaultView sheet opens/closes, focus returns to welcome, no lingering error.
Coverage: Sheet presentation, cancel button (Escape), focus restoration, cleanup state.
Time: 5 min to add.
```

**Total effort:** ~18 min. **Payoff:** Catches regressions in focus, keyboard navigation, and sheet behavior across future refactors.

---

## Suspected Regressions from Changed Code

### 🟡 **Medium Risk: Window Layout Transitions**
**File:** `AuthWindowLayoutModifier.swift`  
**Risk:** `applyAppearance()` and dynamic window sizing during auth state changes may cause:
- Focus loss if window updates before first responder is restored
- Frame jitter if transitions fire during sheet dismiss
- Dark mode not applied if `updateNSView()` timing is off

**Mitigation:** Manually verify window size change from welcome → unlock → create; check focused field after sheet close.

### 🟡 **Medium Risk: Recovery Phrase Accessibility**
**Files:** `UnlockRecoverySection.swift`, `RecoveryPhraseView.swift`  
**Risk:** Recovery phrase parsing (`RecoveryPhraseSupport.summary()`) may not announce validation status correctly:
- Icon + color + text should combine for accessibility, but only text is accessible
- `AccessibilityHint` missing on phrase editor
- Error highlight color-only may fail with reduced transparency

**Mitigation:** Use VoiceOver to read phrase validation feedback; verify error state has text + icon + color.

### 🟡 **Medium Risk: Biometric Button Visibility**
**File:** `UnlockPasswordSection.swift`  
**Risk:** Touch ID button hidden if `canUseBiometrics == false`, but helper text displays:
- May confuse users if helper text says "unavailable" but button is gone entirely
- Focus order change if button removed dynamically

**Mitigation:** Test unlock flow on machine with + without biometric vault keys; verify focus tab order.

### 🟠 **Low Risk: Error Message Tones**
**File:** `AuthSceneScaffold.swift`  
**Risk:** `AuthMessageTone` enum applies colors + backgrounds, but:
- Color contrast must pass WCAG AA on both light/dark modes
- Reduced transparency may reduce contrast further

**Mitigation:** Use Xcode accessibility calculator; verify error messages readable with reduce transparency enabled.

---

## Build & Compile Status

✅ **Build:** Success  
✅ **Tests:** 15/15 passed (no failures)  
✅ **Compile errors:** None detected  
✅ **Warnings:** None logged

**Command run:**
```bash
$ xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj \
  -scheme ZeroPass \
  -destination 'platform=macOS,arch=arm64' \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=''
```

---

## Accessibility Audit (Code-Level)

| Aspect | Status | Notes |
|--------|--------|-------|
| **Labels** | 🟡 Partial | Buttons labeled (✅), but field hints missing on some editors (⚠️) |
| **Keyboard** | 🟡 Partial | Shortcuts present (✅), but Tab order in complex fields untested (⚠️) |
| **Color Contrast** | 🟠 Unknown | Design tokens use system colors; manual check needed (⚠️) |
| **Reduce Motion** | 🟠 Unknown | `@Environment(\.accessibilityReduceMotion)` checked in UnlockVaultView; others untested (⚠️) |
| **Dark Mode** | 🟢 Good | All colors use `ZPTheme` tokens; should adapt automatically (✅) |
| **VoiceOver** | ❌ Untested | No UI test coverage; manual spot check required |

---

## Recommendations by Priority

### Phase A: Do Before Ship (This Session)

1. **Manual smoke test all 6 auth flows** (~20 min)
   - Welcome create/open, unlock password/recovery, recovery display, + cancel flows
   - Use keyboard-only (no mouse) for each transition
   - Verify focus is in the right field after each view change

2. **VoiceOver spot check** (~10 min)
   - Enable VoiceOver (Cmd+F5), read welcome screen, primary buttons, error messages
   - Verify error messages announced (tone should be clear in text, not color alone)
   - Check recovery phrase field is readable (text-heavy, make sure pause behavior is OK)

3. **Light/dark/reduce motion verification** (~5 min)
   - Toggle System Preferences > Appearance > Light/Dark during unlock flow
   - Enable Accessibility > Display > Reduce Motion; verify no jarring transitions
   - Check error message colors readable in both modes

### Phase B: Do If Time Allows

4. **Add 2–3 UI smoke tests** (~20 min)
   - Focus restoration after sheet close (WelcomeView → CreateVaultView → dismiss → focus check)
   - Keyboard shortcut regression (Cmd+O, Cmd+L, return key unlock)
   - Optional: Recovery phrase field validation feedback

5. **Build and commit** (~5 min)
   - Ensure all lint + tests pass
   - Document any deferred follow-ups

### Phase C: Track for Future Work

- [ ] Full auth UI test suite (requires Xcode UI test infrastructure)
- [ ] Snapshot tests for auth component visual consistency
- [ ] Biometric unlock flow on actual Touch ID hardware
- [ ] NSOpenPanel sheet replacement if needed for better UX

---

## Summary Table

| Aspect | Coverage | Risk | Action |
|--------|----------|------|--------|
| **Backend (vault ops)** | ✅ 100% | ✅ Low | No action needed |
| **Auth UI flow** | ❌ 0% | 🔴 High | **MANUAL TEST** required |
| **Keyboard navigation** | ❌ 0% | 🔴 High | **MANUAL TEST** + optional 1 focused UI test |
| **Accessibility (HIG)** | ⚠️ Code-level only | 🟡 Medium | **MANUAL TEST** + VoiceOver spot check |
| **Window layout** | ⚠️ Code review only | 🟡 Medium | **MANUAL TEST** window transitions |
| **Dark mode/appearance** | ✅ Token-based | ✅ Low | Toggle during manual test |

---

## Unresolved Questions

1. **Is the biometric unlock flow tested on actual hardware or only via mock?**  
   Current: Mock via `FakeKeychainStore`. Recommendation: Manual Touch ID test if sensitive.

2. **Does the NSOpenPanel sheet respect dark mode consistently?**  
   Current: Unknown; macOS system control. Recommendation: Observe during manual smoke test.

3. **Are there any hidden focus order bugs in the recovery phrase editor?**  
   Current: No UI test to detect. Recommendation: Keyboard-only smoke test; consider narrow UI test if bugs found.

4. **Does clipboard auto-clear timing interact with the copy button state?**  
   Current: Code review suggests it should work; not tested. Recommendation: Manual test copy → wait → observe cleared state.

---

## Conclusion

**Current build compiles and passes all backend tests.** However, the refactored auth UI has **zero automated coverage**. The refactor introduces several new components (`AuthSceneScaffold`, `UnlockPasswordSection`, `UnlockRecoverySection`, `AuthWindowLayoutModifier`) with complex state, keyboard interaction, and accessibility requirements.

**Before shipping:**
- Manually smoke test all 6 auth entry/exit flows (20 min)
- Spot-check with VoiceOver and accessibility settings (15 min)
- Add 2–3 focused UI regression tests if feasible (20 min)

**Post-ship (separate track):**
- Build comprehensive auth UI test suite
- Add snapshot tests for visual regression
- Continue keyboard + accessibility automation

**Recommendation:** ✅ **UNBLOCK SHIP** after manual smoke test completes. Track deferred automated coverage separately.
