# Code Review — macOS auth/window follow-up cleanup (2026-04-04)

## Scope
- Reviewed files:
  - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
  - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
  - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- Related dependency spot-check: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- Plan context: `plans/20260404-macos-auth-window-flow-followups/plan.md`
- Validation evidence: `/tmp/zeropass-macos-full-tests-followups.log` ended with `** TEST SUCCEEDED **` on 2026-04-04; full macOS UI test run executed 7 tests with 0 failures in 50.016s.

## Severity counts
- Critical: 0
- High: 0
- Medium: 1
- Low: 1

## Findings

### Medium — fresh-launch UI tests still do not preflight bundle-level stale app instances
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:18` defines `bundleIdentifier`, but `launchFreshApp()` (`:108-124`) only terminates `launchedApp` from the current test session.
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift:21-36` similarly launches and terminates only the `XCUIApplication` it created.
- There is no bundle-level preflight using `NSRunningApplication` or equivalent before launch.
- Impact: if a manual or stale `com.tuanle.ZeroPass` process is already running, the suite can still attach to the wrong process and weaken the intended “fresh launch” guarantee, even though the current run passed.
- Follow-up: add a bundle-id preflight kill/assert helper before each launch path and fail loudly if any matching process survives.

### Low — reopen/window recovery still depends on a generic first-`⌘N` menu match
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:183-206` falls back to `openNewWindowFromMenuIfAvailable()` when no visible main window is available.
- That helper scans the menu tree for the first command-key `n` item (`ZeroPassApp.swift:248-271`).
- Impact: it works today and is covered by `testOpenVaultCommandRevealsWindowAfterClosingWelcomeWindow`, but a future unrelated `⌘N` command could retarget this recovery path.
- Follow-up: prefer a dedicated window-scene action, or narrow the lookup to the File → New Window command more explicitly.

## Positive observations
- Window-local auth presentation is now correctly owned by `WelcomeView`, with `focusedSceneValue` plus a targeted notification fallback. That removes the prior app-global modal coupling and should prevent one welcome window from dismissing or opening another window’s sheet.
- `VaultClient.resetStateForUITestsIfNeeded()` now clears `UserDefaults` state and bookmarks when UI tests request a reset, which materially improves deterministic fresh-launch behavior.
- Recovery unlock focus restoration is explicit through `focusRequestID` plumbing, and the recovery `TextEditor` is now marked `privacySensitive()`.
- The latest validation evidence is strong: the full macOS UI test pass completed successfully with 0 failures.

## Overall assessment
This looks good for a low-risk cleanup pass. I would consider the product-code changes mergeable as-is; the only follow-up I’d prioritize is tightening the UI-test preflight so “fresh launch” stays trustworthy even when a stale local app instance is hanging around.
