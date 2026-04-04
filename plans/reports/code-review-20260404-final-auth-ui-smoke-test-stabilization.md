# Code Review — Final macOS Auth UI Smoke-Test Stabilization
**Date:** 2026-04-04
**Overall severity:** Medium

## Verdict

The earlier review concern is resolved: `testWelcomeScreenShowsPrimaryActionsOnFreshLaunch()` in `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:28` now performs a direct fresh-launch assertion instead of recovering the window first. The provided verification logs are green (`6/6` UI tests, `15/15` unit tests).

## Findings

### Medium

1. **Reopen fallback is localization brittle**
   - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:152` and `:161-201` reopen/surface the main window by searching the main menu for the literal English titles `File` and `New Window`.
   - On localized macOS systems, or after menu-copy changes, that fallback can fail and the app may not surface a main window when recovery is needed.

2. **Reopen can create a duplicate window instead of restoring a minimized one**
   - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:195` filters `mainWindow` to `!window.isMiniaturized`.
   - If the user reopens the app from the Dock with a minimized ZeroPass window, `applicationShouldHandleReopen` can skip that existing window and fall back to opening a brand-new one.

### Low

1. **Folder picking behavior still diverges between auth sheets and the app command path**
   - `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift:118-139` now centralizes sheet-aware folder picking via `VaultFolderPicker`.
   - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:110-117` still uses a separate blocking `NSOpenPanel.runModal()` path, which leaves duplicated behavior and different UX depending on entry point.

2. **UI test build still emits a small warning**
   - `/tmp/zeropass-macos-uitests-rerun-6.log` and `/tmp/zeropass-macos-full-tests-2.log` both show `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:69:13: warning: result of call to 'launchFreshApp()' is unused`.
   - Not a blocker, just noisy.

## Positive observations

- The fresh-launch smoke test is now honest again: it directly asserts the welcome actions without helper recovery.
- The sheet smoke tests are simpler and more stable by driving the default/cancel keyboard paths.
- `VaultFolderPicker` is a solid refactor for the auth-sheet flows and reduces repeated panel setup.

## Recommended follow-up

1. Replace title-based “New Window” recovery with a nonlocalized window-opening path.
2. Restore/de-miniaturize an existing ZeroPass window before opening a second one.
3. Reuse the shared folder picker from `ZeroPassApp` to avoid behavior drift.
4. Silence the unused-result warning in `testLaunchPerformance()`.

## Evidence

- `/tmp/zeropass-macos-uitests-rerun-6.log` — `6/6` UI tests passed
- `/tmp/zeropass-macos-full-tests-2.log` — `15/15` unit tests and `6/6` UI tests passed
