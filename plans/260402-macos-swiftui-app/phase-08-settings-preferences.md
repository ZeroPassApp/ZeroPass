# Phase 8: Settings & Preferences

## Context
- Depends on: [Phase 7 — Platform Integration](./phase-07-platform-integration.md) (settings control services)
- **Design Reference**: [UI/UX Guideline §13 Settings Window](./reports/ui-ux-design-guideline.md#13-settings-window)

## Prerequisites

> **CRITICAL:** This phase requires `ZPChangeMasterPassword(handle, oldPassword, newPassword)` in the bridge API. This function does NOT exist in the Go core yet — `store.go` has no `ChangeMasterPassword()` method. It must be implemented in Go core BEFORE this phase can be fully completed. See [Gap Analysis GAP-01](./reports/gap-analysis-260402.md#gap-01-missing-changemasterpassword-in-go-core) for the required implementation flow (verify old → new salt → new master key → re-encrypt vault key → update metadata).

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Build the Settings window (⌘,) with all configurable preferences.

## Design Specs (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

### Settings Window
- **Size**: 550×450pt (fixed, centered)
- **Tab style**: System `TabView` with icons + text at top
- **Tabs**: General (⚙️), Security (🔒), Sync (🔄), About (ℹ️)
- **Section headers**: 13pt semibold
- **Form fields**: Standard SwiftUI `Form` with `.formStyle(.grouped)`
- **Spacing**: 12pt between sections, 8pt between fields
- **Buttons**: `.bordered` style, destructive actions in `.red`

### Destructive Action Confirmation (from [Guideline §19](./reports/ui-ux-design-guideline.md#19-security-ux-patterns))
- **Change Password**: Require current password entry in sheet
- **Regenerate Recovery Key**: Confirmation alert + mnemonic display
- **Export Vault**: Warning about unencrypted file + type-to-confirm pattern

### General Tab Layout
```
Vault Location:     ~/.zeropass/vaults/default  [Change...]
Launch at login:    [✓]
Show in menu bar:   [✓]
Quick Search hotkey: [⌘⇧P] [Record...]
Appearance:         ○ System  ○ Light  ○ Dark
```

### Security Tab Layout
```
Auto-lock after:        [5 minutes ▾]
Clipboard clear:        [30 seconds ▾]
Lock on sleep:          [✓]
Lock on screen saver:   [✓]
Unlock with TouchID:    [✓]

Master Password
[Change Master Password...]

Recovery Key
[Regenerate Recovery Key...]
⚠ Recovery key shown ONLY during creation/regeneration (never stored).

Data
[Export Vault...]  [Import...]
```

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Views/Settings/SettingsView.swift` | Main settings TabView |
| `macos/ZeroPass/Views/Settings/GeneralSettingsView.swift` | General preferences |
| `macos/ZeroPass/Views/Settings/SecuritySettingsView.swift` | Security preferences |
| `macos/ZeroPass/Views/Settings/SyncSettingsView.swift` | Sync server config |
| `macos/ZeroPass/Views/Settings/AboutView.swift` | Version, credits, links |

## Implementation Steps

### SettingsView (TabView)
1. General tab: vault path, launch at login, default vault
2. Security tab: auto-lock timeout, clipboard clear time, TouchID toggle, lock on sleep
3. Sync tab: server URL, API key, device name, sync interval
4. About tab: version, open-source credits, GitHub link

### SecuritySettingsView
```
┌──────────────────────────────────────────┐
│  Security                                │
│                                          │
│  Auto-lock after:    [5 minutes ▾]       │
│  Clipboard clear:    [30 seconds ▾]      │
│  Lock on sleep:      [✓]                 │
│  Lock on screen saver: [✓]              │
│  Unlock with TouchID: [✓]               │
│                                          │
│  Master Password                         │
│  [Change Master Password...]             │
│                                          │
│  Recovery Key                            │
│  [Regenerate Recovery Key...]            │
│  ⚠ Recovery key is shown ONLY during     │
│  creation or regeneration (never stored). │
│  "View Recovery Key" is NOT possible.     │
│                                          │
│  Vault                                   │
│  [Export Vault...]  [Import...]          │
└──────────────────────────────────────────┘
```

### SyncSettingsView
1. Enable/disable sync toggle
2. Server URL text field
3. API key secure field
4. Device name text field
5. "Test Connection" button
6. Last sync timestamp display
7. "Sync Now" button

## Implementation Details

### Change Master Password Flow
1. User clicks "Change Master Password..." → opens sheet
2. Sheet has 3 fields: Current Password, New Password, Confirm New Password
3. Validate new password !== current, confirm matches new
4. Show password strength indicator for new password (reuse from setup wizard)
5. Call `ZPChangeMasterPassword(handle, oldPassword, newPassword)` via bridge
6. On success: show confirmation alert
7. On error (wrong current password): show inline error, do NOT close sheet
8. **Security:** If TouchID is enabled, update Keychain with new vault key (salt changed → vault key changed)

### Regenerate Recovery Key Flow
1. User clicks "Regenerate Recovery Key..." → confirm dialog ("This will invalidate your current recovery key")
2. Call `ZPRegenerateRecovery(handle)` → returns new 12-word mnemonic
3. Show mnemonic in RecoveryKeyView (same component as setup wizard)
4. User must confirm they wrote it down before dismissing
5. **IMPORTANT:** C-heap mnemonic must be copied to Swift String, then immediately `ZPFreeResult`. Zero Swift copy after user dismisses.

### Export/Import Notes
- Export/Import buttons call `ZPExportJSON`/`ZPImportCSV` etc. via bridge
- Bridge functions handle file I/O internally (Go core uses `io.Reader`/`io.Writer`, bridge opens files from paths)
- For encrypted export: uses vault key — no separate password prompt needed

## Success Criteria
- [ ] Settings window opens with ⌘, (550×450pt, fixed, centered)
- [ ] Settings uses system `TabView` with 4 tabs (General, Security, Sync, About)
- [ ] All preferences persist across app launches (UserDefaults)
- [ ] Section headers: 13pt semibold, form fields: `.formStyle(.grouped)`
- [ ] Security settings control actual behavior (auto-lock, clipboard, TouchID)
- [ ] Sync settings save server configuration
- [ ] Change master password flow: sheet with current + new + confirm fields
- [ ] Recovery key regenerate: confirmation alert + mnemonic display + zeroing
- [ ] "View Recovery Key" is NOT offered (impossible — mnemonic is never stored)
- [ ] Export: warning about unencrypted file with type-to-confirm
- [ ] Destructive buttons styled in `.red`
- [ ] All text uses macOS scale (13pt body), semantic colors only
- [ ] Hotkey recording field for Quick Search shortcut customization
