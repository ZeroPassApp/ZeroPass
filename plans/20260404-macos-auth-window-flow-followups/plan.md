---
title: "Clean up macOS auth/window follow-ups"
description: "Scope auth sheets per window, cover visibility-recovery automation, remove dead no-vault flow, and harden fresh-launch UI tests."
status: pending
priority: P2
effort: 4h
branch: main
tags: [macos, swiftui, auth, ui-tests]
created: 2026-04-04
---

# macOS auth/window follow-ups

## Phases
- P1 pending — move auth sheet presentation off app-global modal state
- P2 pending — add UI automation for no-visible-window recovery
- P3 pending — remove dead no-vault folder-open branch / duplicate flow
- P4 pending — fail fresh-launch helper loudly on stale app instances

## Root-cause summary
- `VaultClient.activeAuthModal` lives on the app-global environment object, but auth UI is rendered inside a `WindowGroup`; any welcome window can therefore open or dismiss another window’s sheet.
- The no-vault `Open Vault…` command depends on `revealMainWindowIfNeeded()` plus a heuristic `⌘N` menu fallback in `AppDelegate`; that recovery path is used in production but not directly protected by UI tests.
- `ZeroPassApp.chooseVaultFolder(replacingCurrent:)` still carries an open-existing/no-vault path even though `OpenVaultSheet` + `VaultFolderPicker` now own that workflow, leaving duplicate picker/error logic.
- `ZeroPassUITests.terminateRunningAppIfNeeded()` waits up to 5s but never asserts success, so stale processes can silently poison “fresh launch” coverage.

## Minimal safe implementation
1. Add window-local auth presentation state at `ContentView`/`WelcomeView` level (small coordinator object or local `@State`) and bind `.sheet` to that local state instead of `VaultClient.activeAuthModal`.
2. Route menu-triggered auth requests to the focused/visible welcome window only. Prefer a focused-scene binding/coordinator; acceptable fallback is a one-shot window request consumed only by the active window after `revealMainWindowIfNeeded()`.
3. Delete `VaultClient.presentCreateVaultSheet()`, `presentOpenVaultSheet()`, `dismissAuthModal()`, and the `activeAuthModal` property after callers are moved to window-local state.
4. Collapse `ZeroPassApp.chooseVaultFolder(replacingCurrent:)` to the replace-current case only, or remove it entirely if `OpenVaultSheet(mode: .replaceCurrent)` can serve both command paths through one picker/open implementation.
5. Harden `terminateRunningAppIfNeeded()` to assert that no matching `NSRunningApplication` remains after the wait loop and fail with diagnostics if termination stalls.
6. Add one UI smoke that closes the last visible window, triggers `⌘O`, and asserts a visible window plus `openVault.title`; keep existing fresh-launch/create/open-cancel tests green.

## Likely files to edit
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift` (if flow consolidation happens)
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift` (optional helper reuse only)

## Test strategy
- Targeted UI tests: fresh launch, create-sheet cancel, open-sheet cancel, new no-visible-window `⌘O` recovery smoke.
- Full macOS test scheme after targeted tests pass.
- Manual multi-window spot check: with two welcome windows open, `Open Vault…` affects only the focused window and dismissing one sheet does not change the other window.
- Manual reopen spot check: last window closed → app reopen restores a visible window before auth presentation.

## Unresolved questions
- If `Commands` cannot cleanly target a focused scene binding in this app structure, use the smallest one-shot request bus that hands presentation off to window-local state without reintroducing persistent global modal state.
- If `⌘W` is flaky in CI for hiding the last window, close the window via standard close controls in the new UI smoke instead of changing the implementation scope.
