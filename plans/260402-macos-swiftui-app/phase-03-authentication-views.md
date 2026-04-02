# Phase 3: Authentication Views

## Context
- Depends on: [Phase 2 — Xcode Project Setup](./phase-02-xcode-project-setup.md)
- **Design Reference**: [UI/UX Design Guideline §8 Authentication Flows](./reports/ui-ux-design-guideline.md#8-authentication-flows)

## Overview
- **Priority:** P0
- **Status:** Pending
- **Description:** Build the vault setup wizard (first-time), unlock screen (returning user), and recovery flow views. These are the gateway to the entire app.

## Key Insights
- First-time setup must feel fast (< 2 min per PRD)
- Recovery key display is ONE TIME ONLY — must be prominent and unmissable
- Argon2id KDF takes 200-500ms — show loading indicator, use async/await
- TouchID integration: store vault key in Keychain with biometric ACL after first password unlock

## Design Specs (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

### Setup Wizard Window
- **Size**: Fixed 480×520pt, centered, non-resizable
- **Step indicator**: Horizontal dots at top (5 dots)
- **Navigation**: Back arrow in top-left, Continue in bottom-right
- **Transitions**: Cross-fade between steps (200ms `easeInOut`)
- **Recovery key grid**: 4×3 layout, SF Mono 14pt, each word in bordered cell
- **CRITICAL**: `NSWindow.sharingType = .none` during recovery key step

### Unlock Screen
- **Layout**: Full-window overlay with `.ultraThinMaterial` background
- **App icon**: 64pt, centered
- **Title**: "ZeroPass" — 26pt semibold
- **Password field**: `SecureField`, 300pt width, centered
- **Enter key** triggers unlock (no need to click button explicitly)
- **TouchID**: Attempt automatically on appear if enabled
- **Error**: Field shakes (3x, 4pt amplitude, 300ms), red error text fades in
- **Loading**: Replace "Unlock" button with `ProgressView` during Argon2id

### Password Strength Bar
- **Height**: 4pt, rounded corners
- **Colors**: `.red` (0) → `.orange` (1) → `.yellow` (2) → `.green` (3) → dark green (4)
- **Font**: 11pt `.secondary` for feedback text
- **Debounce**: 300ms keystroke debounce

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Views/Auth/SetupView.swift` | First-time vault creation wizard |
| `macos/ZeroPass/Views/Auth/RecoveryKeyView.swift` | Recovery mnemonic display + confirmation |
| `macos/ZeroPass/Views/Auth/UnlockView.swift` | Master password entry + TouchID button |
| `macos/ZeroPass/Views/Auth/RecoveryUnlockView.swift` | Recovery flow: enter mnemonic → set new password |
| `macos/ZeroPass/Views/Auth/PasswordStrengthView.swift` | Real-time zxcvbn strength indicator |

## Implementation Steps

### SetupView (first-time wizard)
1. Step 1: Welcome screen with brief ZeroPass intro
2. Step 2: Master password creation
   - Password field + confirm field
   - Real-time strength meter (calls `ZPScorePassword` via bridge)
   - Enforce: length ≥ 12, zxcvbn score ≥ 3
   - Show requirements checklist
3. Step 3: Recovery key display
   - Show 12-word BIP-39 mnemonic in large, clear grid
   - **⚠️ Mnemonic comes from `ZPCreateVault` result** — must `ZPFreeResult` immediately after extracting
   - **Handle mnemonic as `Data` + `UnsafeMutableBufferPointer`, NOT plain Swift `String`** — zero after display
   - "I've written this down" checkbox
   - Optional: copy to clipboard button (with warning + auto-clear)
4. Step 4: Recovery key verification
   - Ask user to enter words 3, 7, 11 (random selection)
   - Confirm match before proceeding
5. Step 5: Vault created → transition to main app

### UnlockView (returning user)
1. App icon + "ZeroPass" branding
2. Master password SecureField
3. "Unlock" button (⏎ Enter shortcut)
4. TouchID button (if available and enabled)
   - Calls `BiometricService.authenticate()` → retrieves vault key from Keychain
   - Calls **`ZPUnlockWithKey`** (NOT `ZPUnlock`) — passes raw vault key, bypasses KDF entirely
   - Prereq: `ZPUnlockWithKey` in Phase 1 + `store.UnlockWithKey()` in Go core
5. "Use Recovery Key" link → RecoveryUnlockView
6. Loading state during Argon2id derivation (200-500ms)
7. Error state: "Incorrect password" with shake animation

### RecoveryUnlockView
1. 12 text fields for mnemonic words (or paste-all area)
2. Validate mnemonic format
3. On success: prompt for new master password
4. Reset vault with new password → back to UnlockView

### PasswordStrengthView (reusable component)
1. Score bar (0-4 colored segments: red → orange → yellow → green → dark green)
2. Feedback text from zxcvbn ("Add a number", "Avoid common words")
3. Estimated crack time display
4. Real-time updates on keystroke (debounced 300ms)

## Success Criteria
- [ ] SetupView creates vault and displays recovery mnemonic
- [ ] Recovery key verification step works (random word check)
- [ ] UnlockView unlocks vault with correct password
- [ ] UnlockView shows error for wrong password with shake animation
- [ ] TouchID flow: first unlock with password → subsequent unlocks with TouchID
- [ ] RecoveryUnlockView restores vault access and sets new password
- [ ] Password strength meter updates in real-time
- [ ] KDF loading state displayed (no frozen UI)
- [ ] Setup wizard window is fixed 480×520pt, non-resizable
- [ ] All text uses system font APIs (13pt body, 26pt title) — NOT iOS scale
- [ ] All colors are semantic (`.primary`, `.red` for errors)
- [ ] Reduced motion respected (`@Environment(\.accessibilityReduceMotion)`)
- [ ] VoiceOver: password field labeled "Password, hidden", strength announced

## Security Considerations
- Password never stored in plaintext (only held in memory during KDF)
- **Recovery mnemonic**: handle as `Data`, zero after user confirms written down. Disable screenshots via `NSWindow.sharingType = .none`
- Clear password from memory after unlock completes
- TouchID vault key stored with `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly`
- TouchID key invalidated when fingerprints change (`.biometryCurrentSet`)
- **TouchID stores raw vault key** (32 bytes), NOT master password — Keychain retrieval → `ZPUnlockWithKey` → instant unlock (no KDF)
