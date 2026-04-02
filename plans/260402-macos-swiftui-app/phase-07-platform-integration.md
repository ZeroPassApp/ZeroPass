# Phase 7: Platform Integration (Keychain, TouchID, Clipboard, MenuBar)

## Context
- Depends on: [Phase 3](./phase-03-authentication-views.md) (TouchID in unlock), [Phase 4](./phase-04-main-ui-layout.md) (MenuBar)
- **Design Reference**: [UI/UX Guideline §12 Menu Bar Widget](./reports/ui-ux-design-guideline.md#12-menu-bar-widget) + [§19 Security UX Patterns](./reports/ui-ux-design-guideline.md#19-security-ux-patterns)

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Integrate native macOS platform services: Keychain for secure key storage, TouchID/FaceID for biometric unlock, clipboard management with auto-clear, menu bar widget for quick access, and system notifications.

## Key Insights
- Keychain policy: `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` — most secure, deleted if passcode removed
- TouchID ACL: `.biometryCurrentSet` — invalidated when fingerprints change
- Clipboard: `NSPasteboard.general` + scheduled clear timer
- MenuBarExtra: `.menuBarExtraStyle(.window)` for popover panel
- Lock-on-sleep: `NSWorkspace.willSleepNotification`

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Services/KeychainService.swift` | Store/retrieve vault key in Keychain |
| `macos/ZeroPass/Services/BiometricService.swift` | TouchID/FaceID authentication |
| `macos/ZeroPass/Services/ClipboardService.swift` | Copy with auto-clear timer |
| `macos/ZeroPass/Services/AutoLockService.swift` | Auto-lock on idle/sleep |
| `macos/ZeroPass/Services/NotificationService.swift` | System notifications |
| `macos/ZeroPass/Views/MenuBar/MenuBarView.swift` | Menu bar popover widget |

## Implementation Steps

### KeychainService
1. `storeVaultKey(_ key: Data, forVault id: String)` — non-biometric
2. `storeVaultKeyBiometric(_ key: Data, forVault id: String)` — with `.biometryCurrentSet` ACL
3. `retrieveVaultKey(forVault id: String, biometric: Bool) -> Data?`
4. `removeVaultKey(forVault id: String)` — on logout/reset
5. Error handling: keychain full, user denied, biometric failed

### BiometricService
1. `canUseBiometrics() -> Bool` — check LAContext availability
2. `authenticate() async throws -> Bool` — trigger TouchID prompt
3. Flow: authenticate → retrieve **raw vault key** from Keychain → call **`ZPUnlockWithKey`** (NOT `ZPUnlock`) → instant unlock (no KDF)
4. Fallback: "Use Master Password" button in TouchID prompt
5. **First unlock with password:** after successful `ZPUnlock`, get vault key via `vault.VaultKey()` and store in Keychain for future biometric unlocks

### ClipboardService
1. `copy(_ text: String, clearAfter seconds: Int = 30)`
2. Check if clipboard still contains our string before clearing
3. Cancel previous timer when new copy happens
4. Optional: notification "Copied to clipboard (auto-clear in 30s)"

### AutoLockService
1. **⚠️ Go-side auto-lock is DISABLED** when vault opened via bridge (`AutoLockTimeout=0` set in Phase 1). Swift handles ALL auto-lock logic.
2. Observe `NSWorkspace.willSleepNotification` → lock vault via `ZPLock`
3. Observe `NSWorkspace.screensDidSleepNotification` → lock vault (if setting enabled)
4. Idle timer: reset on any vault operation, lock after `autoLockTimeout`
5. **Guard against mid-operation lock:** track in-flight bridge operations and defer lock until completion
4. Configurable: 1/5/15/30 min or never

### MenuBarView
1. Lock shield icon in menu bar — `lock.fill` (locked) / `lock.open.fill` (unlocked), 18×18pt
2. Popover panel (`.menuBarExtraStyle(.window)`):
   - If locked: "Unlock Vault" button + "Enter Password" option
   - If unlocked: search field (13pt compact) + recent items (top 5)
   - Quick copy password from results — copy button appears on hover per row
   - "Open ZeroPass" → activate main window
   - "Lock Vault" button
   - Divider → "Quit"
3. **Width**: 280pt (per [Guideline §12](./reports/ui-ux-design-guideline.md#12-menu-bar-widget))
4. Menu bar icon adapts to light/dark mode automatically
5. Sections separated by `Divider()`

### Toast Notifications (Copy Feedback)
- **Position**: Bottom-center of active window, 16pt from bottom
- **Format**: "✓ Copied to clipboard (30s)" with subtle progress bar
- **Background**: `.thickMaterial`, rounded corners (8pt)
- **Animation**: Fade in 150ms, hold 2s, fade out 300ms
- **Auto-clear countdown**: Subtle progress bar under text

### NotificationService
1. Clipboard cleared notification (subtle)
2. Vault auto-locked notification
3. Sync completed/failed notification
4. Use `UNUserNotificationCenter`

## Success Criteria
- [ ] Vault key stored in Keychain after first password unlock
- [ ] TouchID unlocks vault without password prompt
- [ ] TouchID key invalidated when fingerprints change
- [ ] Clipboard auto-clears after configured timeout
- [ ] Copy action shows toast notification (fade in 150ms, hold 2s, fade out 300ms)
- [ ] Toast shows auto-clear countdown progress bar
- [ ] Vault locks on system sleep
- [ ] Vault locks after idle timeout
- [ ] Menu bar: lock.fill/lock.open.fill icon (18×18pt), adapts to dark mode
- [ ] Menu bar popover: 280pt width, search + 5 recent items when unlocked
- [ ] Menu bar: quick copy with hover-visible copy button per row
- [ ] System notifications for auto-lock and clipboard clear
- [ ] All UI follows [Guideline §19 Security UX Patterns](./reports/ui-ux-design-guideline.md#19-security-ux-patterns)
