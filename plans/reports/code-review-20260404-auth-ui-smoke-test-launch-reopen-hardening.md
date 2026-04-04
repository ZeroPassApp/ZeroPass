# Code Review Summary

## Scope
- Files: `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`, `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- Focus: final macOS auth UI smoke-test stabilization after the latest launch/reopen hardening
- Evidence: `/tmp/zeropass-macos-full-tests-3.log` (`Executed 6 tests, with 0 failures`; `** TEST SUCCEEDED **`)
- Scout findings: shortcut-path ambiguity, reopen logic coverage gap, localized-menu brittleness in test fallback, stale plan tracking

## Overall Assessment
The latest state is materially better. App-side launch/reopen recovery is now locale-agnostic, minimized main windows are restored instead of being treated as missing, and the fresh-launch test now directly proves the welcome actions render on first launch. I found no critical or high-severity issues. The main remaining risk is that the `⌘O` path is still wired in two places with different behavior, which leaves both product behavior and the sheet smoke test slightly brittle.

## Critical Issues
- None.

## High Priority
- None.

## Medium Priority
1. **Conflicting `⌘O` handlers still drive different UX paths**
   - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:61-64` binds `⌘O` in the app command menu to `chooseVaultFolder(replacingCurrent:)`, which opens an `NSOpenPanel` directly.
   - `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift:45-52` also binds `⌘O`, but uses it to present `OpenVaultSheet`.
   - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:56` assumes `⌘O` resolves to the sheet path.
   - Impact: the shortcut behavior depends on responder precedence and current focus. Today the test passes, but the same user shortcut is still mapped to two different flows.

## Low Priority
1. **Launch/reopen hardening is only partially covered by the smoke suite**
   - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:152-238` contains the actual Dock reopen, deminiaturize, and menu-action fallback logic.
   - The latest suite directly verifies fresh launch, but it does not exercise `applicationShouldHandleReopen`, miniaturized-window restoration, or duplicate-window avoidance.
   - Impact: regressions in the reopen path could slip past the current tests.

2. **The UITest recovery helper is still English-menu brittle**
   - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:110-115` recovers by clicking `File` → `New Window`.
   - Impact: likely fine on English CI, but weaker on localized runners and less robust than the app-side key-equivalent lookup.

3. **Plan tracking is stale**
   - `plans/260404-macos-auth-ui-overhaul/plan.md` and `plans/260404-macos-auth-ui-overhaul/phase-04-qa-and-doc-sync.md` still show pending status and unchecked TODOs.
   - Impact: low code risk, but future contributors lose an accurate source of truth.

## Edge Cases Found by Scout
- The no-vault `⌘O` shortcut currently has two possible destinations.
- The launch test proves first-launch visibility, but not Dock reopen/minimize recovery.
- Test-side recovery still depends on English menu titles.

## Positive Observations
- `ZeroPassApp` now restores a minimized main window instead of opening a duplicate first.
- App-side window recovery no longer depends on localized menu titles.
- `testWelcomeScreenShowsPrimaryActionsOnFreshLaunch()` is now an honest first-launch assertion instead of a self-healing test.
- `CreateVaultView` and `OpenVaultSheet` now share the folder picker path, which reduces auth-sheet drift.

## Recommended Actions
1. Unify the no-vault `⌘O` path so both the menu/shortcut and welcome button reach the same open-vault flow.
2. Add one dedicated reopen/minimize regression test if feasible.
3. If the UITest recovery helper stays, make it locale-agnostic.
4. Sync the related plan status once this work is accepted.

## Metrics
- UI smoke tests: 6 passed, 0 failed
- Critical issues: 0
- High issues: 0
- Medium issues: 1
- Low issues: 3

## Unresolved Questions
- None blocking.
