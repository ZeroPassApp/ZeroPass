# Code Review — macOS auth/window follow-up final focused pass

Date: 2026-04-04

## Scope
- Reviewed files:
  - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
  - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
  - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- Related dependents spot-checked for edge cases:
  - `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- Fresh validation evidence: `/tmp/zeropass-macos-full-tests-followups.log` ends with `** TEST SUCCEEDED **` on 2026-04-04
- Scope size: 6 files / 1845 LOC reviewed

## Severity Counts
- Critical: 0
- High: 0
- Medium: 0
- Low: 2

## Overall Assessment
The auth/window cleanup looks solid. The welcome-sheet state is now window-local instead of app-global, the no-visible-window `⌘O` recovery path is covered by a direct UI smoke, and the stale-process preflight now fails loudly instead of silently contaminating fresh-launch coverage. I did not find any blocking correctness, security, or cross-window state regressions in the reviewed surface.

## Findings
### Low
1. `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
   The fresh full-scheme log still shows one project warning at `testLaunchPerformance`: `result of call to 'launchFreshApp()' is unused`. This is harmless at runtime, but it adds warning noise and can hide future signal in CI logs.

2. `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`, `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift`
   The stale-process preflight logic is duplicated and checks survivors immediately after `pkill -9`. Current validation is green, but this remains a small long-tail flake risk on slower/busier runners if process-table teardown lags for a moment. A shared helper with a short retry window would reduce future drift and startup flakiness.

## Edge Cases Checked
- Menu-triggered `Open Vault…` routing to the focused welcome scene instead of app-global modal state
- No-visible-window recovery after closing the last welcome window, followed by `⌘O`
- Welcome-sheet dismissal when vault state leaves `.noVault`
- Recovery phrase focus restoration after method changes / sheet dismissal
- Sensitive recovery input treatment via `.privacySensitive()`

## Positive Observations
- `WelcomeView` correctly localizes auth modal presentation with `@State` + focused-scene handoff.
- `ZeroPassApp` only falls back to window recovery when no focused welcome binding is available.
- `UnlockRecoverySection` improves privacy and focus ergonomics without widening app state.
- The follow-up plan file already shows all four phases complete and the fresh full-scheme validation backs that up.

## Remaining Risks
- Test-only startup flake risk from the immediate post-kill survivor assertion.
- Minor warning debt from the unused `launchFreshApp()` result in the performance test.

## Recommendation
Safe to merge as-is for the current scope. Optional follow-up cleanup: remove the one warning and centralize the stale-process preflight helper with a brief retry loop.
