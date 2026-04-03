---
title: "ZeroPass macOS Auth UI Overhaul"
description: "Make the locked/unlock flow feel native on macOS, adaptive in appearance, and consistent across welcome and recovery screens."
status: pending
priority: P1
effort: 4d
branch: main
tags: [macos, swiftui, auth, ui]
created: 2026-04-04
---

# ZeroPass macOS Auth UI Overhaul

## Outcome
- Replace the sparse hero-style auth experience with a compact native macOS auth shell.
- Remove the unlock screen’s forced dark mode and use adaptive system colors/materials.
- Replace the recovery checkbox with a clearer auth-mode switch and better recovery phrase entry.
- Make Touch ID obvious, and move vault-closing/switching out of the bottom footer.
- Resize auth states to fit their content instead of opening the app at full main-window size.

## Phases

| # | Phase | Effort | Purpose | Link |
|---|-------|--------|---------|------|
| 1 | Auth Shell + Window Behavior | 1d | Shared native auth scaffold, adaptive tokens, state-aware window sizing | [phase-01](./phase-01-auth-shell-and-window-behavior.md) |
| 2 | Unlock Screen + Recovery Mode | 1.5d | Full unlock flow rewrite, Touch ID discoverability, safer vault actions | [phase-02](./phase-02-unlock-screen-and-recovery-mode.md) |
| 3 | Welcome + Recovery Consistency | 1d | Align welcome, recovery display, create/regenerate flows with the new auth shell | [phase-03](./phase-03-related-auth-screen-alignment.md) |
| 4 | QA + Doc Sync | 0.5d | Validate accessibility/window behavior and update outdated auth design docs | [phase-04](./phase-04-qa-and-doc-sync.md) |

## Primary Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

## Supporting Files Likely Needed for Full Consistency
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` (only if one-shot Touch ID auto-attempt or auth-state helpers are added)

## Files to Create
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseCardView.swift`

## Key Risks
- State-driven window resizing can feel jumpy if not coordinated with auth ↔ main transitions.
- Touch ID becomes more visible, but auto-prompting can get annoying if it fires repeatedly.
- Recovery phrase normalization must accept pasted whitespace/newlines/numbered lists without false negatives.
- If `SecuritySettingsView.swift` is left untouched, recovery-key regeneration will remain visually inconsistent.

## Docs Impact
- **Yes, minor but important:** update `plans/260402-macos-swiftui-app/phase-03-authentication-views.md` and `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md` because they currently codify the old hero/full-window unlock pattern.
- `docs/code-standards.md` likely does **not** need changes.
