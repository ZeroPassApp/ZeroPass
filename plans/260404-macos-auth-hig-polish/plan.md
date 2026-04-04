---
title: "ZeroPass macOS Auth HIG Polish"
description: "Tighten the macOS auth/onboarding flow around standard window chrome, true sheets, and keyboard/a11y polish without changing vault behavior."
status: completed
priority: P1
effort: 1 session (~5h)
completion: phases 1-2 completed; phase 3 automated tests passed (15/15), manual validation deferred to follow-up
branch: main
tags: [macos, auth, hig, swiftui, appkit]
created: 2026-04-04
---

# ZeroPass macOS Auth HIG Polish

## Outcome
- Remove the biggest remaining non-native auth behaviors: custom floating chrome, faux modal overlays, and brittle focus/accessibility edges.
- Preserve the existing create/open/lock/unlock/recovery behavior and state machine.
- Finish with a focused validation sweep for keyboard, VoiceOver, appearance, and window sizing.

## Reuse vs supersede
- **Reuse:** `plans/260404-macos-auth-ui-overhaul/` as historical context and a source of already-landed structural work.
- **Supersede for active work:** that plan assumed the scaffold, split sections, and recovery card still had to be created. They now exist, so the remaining task is a smaller HIG-polish pass, not another full rewrite.

## Phases

| # | Phase | Effort | Purpose | Link |
|---|-------|--------|---------|------|
| 1 | Window Chrome + Native Sheets | 2h | Restore standard macOS window behavior and replace custom auth overlays with true sheets | [phase-01](./phase-01-window-chrome-and-native-sheets.md) |
| 2 | Auth Surface Hierarchy + Focus | 2h | Tighten welcome/unlock/create/recovery hierarchy, action placement, focus, and control emphasis | [phase-02](./phase-02-auth-surface-hierarchy-and-focus.md) |
| 3 | Accessibility + Validation Sweep | 1h | Verify keyboard, VoiceOver, reduce motion/transparency, and no-regression build/tests | [phase-03](./phase-03-accessibility-and-validation.md) |

## Primary Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`

## Secondary Files Likely Needed
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` *(only if focus or one-shot biometric coordination must move out of the views)*

## Key Risks
- Restoring standard chrome changes the current “floating auth” look; keep the visual shell subtle so the app still feels intentional.
- Replacing custom overlays with true sheets can expose focus/order bugs; do the window/sheet work before touch-up styling.
- Fixed auth sizing currently hides some issues; once the window becomes more standard and more resizable, spacing and truncation regressions may appear.

## Docs Impact
- No broad docs rewrite required.
- If implementation materially changes the auth shell behavior, add a short changelog/plan sync-back note referencing this plan and the superseded `260404-macos-auth-ui-overhaul` plan.
