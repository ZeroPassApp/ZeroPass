# Phase 02: Auth Flow

## Context Links
- [plan.md](plan.md) | [phase-01](phase-01-project-scaffold-rust-bridge.md)
- [macOS WelcomeView](../../apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift)
- [macOS UnlockVaultView](../../apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift)
- [macOS VaultClient](../../apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift)

## Overview
- **Priority**: P0 — Core auth flow required for any vault interaction
- **Status**: pending
- **Description**: Welcome → Create Vault → Recovery Phrase → Lock → Unlock (password + recovery + biometric)

## Implementation Steps

- [ ] 1. Design system: `styles/theme.css` (CSS variables from ZeroPassTheme), `reset.css`, `animations.css`
- [ ] 2. Auth store (Zustand): state machine (restoring|noVault|locked|showingRecovery|unlocked)
- [ ] 3. UI store: toasts, modals, loading states
- [ ] 4. TypeScript types: `src/lib/types.ts` — VaultItem, ItemType, PasswordScore, etc.
- [ ] 5. IPC bindings: `src/lib/commands.ts` — type-safe invoke wrappers
- [ ] 6. Tauri commands: `commands/vault.rs` — create_vault, open_vault, unlock, lock, close, change_password
- [ ] 7. WelcomePage — hero + Create Vault / Open Vault CTAs
- [ ] 8. CreateVaultDialog — folder picker + password + strength bar + confirm
- [ ] 9. OpenVaultDialog — folder picker
- [ ] 10. RecoveryPhrasePage — BIP-39 12-word display + copy + confirm checkbox
- [ ] 11. UnlockPage — password form + recovery form + biometric button
- [ ] 12. UnlockPasswordForm — password input + submit + error display
- [ ] 13. UnlockRecoveryForm — 12-word mnemonic input
- [ ] 14. PasswordStrengthBar component — 0-4 score bar
- [ ] 15. RecoveryPhraseCard component — numbered word grid
- [ ] 16. ToastOverlay component — animated notifications
- [ ] 17. App.tsx — state-based routing (noVault→Welcome, locked→Unlock, unlocked→VaultPage)
- [ ] 18. Tests: auth-store, commands mock, component tests, Rust command tests

## Success Criteria
- Complete auth lifecycle: create → recovery → lock → unlock → lock
- Biometric unlock via keyring (stores/loads vault key)
- Password strength shown during creation
- Recovery phrase displayed and confirmable
- All auth store states tested
