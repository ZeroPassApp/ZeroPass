# Phase 05: Settings + Platform Features

## Context Links
- [plan.md](plan.md) | [phase-04](phase-04-search-quick-search.md)
- [macOS SettingsView](../../apps/macos/ZeroPass/ZeroPass/Views/Settings/SettingsView.swift)

## Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Settings tabs, system tray, auto-lock, clipboard auto-clear, biometric, import/export

## Implementation Steps

- [ ] 1. Settings store (Zustand): autoLockTimeout, clipboardClearTimeout, biometricEnabled, syncConfig
- [ ] 2. Tauri commands: `commands/settings.rs`, `commands/biometric.rs`, `commands/clipboard.rs`
- [ ] 3. Tauri commands: `commands/autolock.rs`, `commands/importexport.rs`, `commands/system.rs`
- [ ] 4. SettingsDialog — tab container (General, Security, Sync, About)
- [ ] 5. GeneralSettings — appearance, launch at login, menu bar
- [ ] 6. SecuritySettings — auto-lock, clipboard clear, biometric toggle, change password, import/export
- [ ] 7. SyncSettings — server URL, device name, sync now, device list
- [ ] 8. AboutSettings — version, links
- [ ] 9. ChangePasswordDialog — old + new + confirm
- [ ] 10. ImportDialog — source selector (Chrome, Firefox, 1Password, Bitwarden, Safari, LastPass, KeePass, CSV, 1PUX) + file picker
- [ ] 11. ExportDialog — format selector (JSON, CSV, Encrypted) + file picker
- [ ] 12. System tray: `tray.rs` — icon + menu (lock, quick search, settings, quit)
- [ ] 13. Auto-lock: `autolock.rs` — Rust timer, reset on user activity events
- [ ] 14. Biometric: `biometric.rs` — keyring crate store/load vault key
- [ ] 15. Clipboard: `clipboard.rs` — copy + setTimeout clear
- [ ] 16. Tests: settings-store, component tests, Rust command tests

## Success Criteria
- All 4 settings tabs functional
- System tray icon with working menu
- Auto-lock triggers after configured timeout
- Clipboard auto-clears sensitive data
- Biometric unlock stores/retrieves vault key
- Import from 9 sources works
- Export to 3 formats works
