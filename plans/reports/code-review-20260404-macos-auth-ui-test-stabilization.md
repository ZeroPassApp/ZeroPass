## Code Review Summary

### Scope
- Files reviewed:
  - `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
  - Nearby dependencies read for context: `AuthSceneScaffold.swift`, `AuthWindowLayoutModifier.swift`, `ZeroPassApp.swift`, `ContentView.swift`, `VaultClient.swift`, `ZeroPassUITestsLaunchTests.swift`
- LOC: 581 LOC in the 4 focus files; 2,206 LOC reviewed including nearby dependencies
- Focus: recent macOS auth UI test stabilization changes
- Scout findings:
  - UI-test launch recovery is exercised in the provided verification logs before welcome controls become visible
  - `UITEST_MODE` and `UITEST_RESET_STATE` are wired through `ZeroPassApp`/`VaultClient`, reducing state leakage and side effects
  - The generic launch screenshot test still does not assert auth-window visibility

### Overall Assessment
Overall severity: **Medium**.

The changes are directionally good: stable accessibility identifiers improve selector quality, auth-sheet folder picking is cleaner, and the UI tests are much more deterministic than before. Fresh verification also passed (`15` unit tests and `6` UI tests). The main issue is semantic rather than mechanical: the new “fresh launch” test now recovers by opening `File > New Window`, so it no longer proves that a fresh launch actually shows a visible auth window.

### Critical Issues
- None.

### High Priority
- None.

### Medium Priority
1. **The fresh-launch test now masks the original launch-window regression**
   - `testWelcomeScreenShowsPrimaryActionsOnFreshLaunch()` calls `ensureMainWindowVisible(in:)` before asserting the welcome buttons.
   - `ensureMainWindowVisible(in:)` silently recovers by selecting `File > New Window` when the welcome buttons are absent.
   - Both `/tmp/zeropass-macos-uitests-rerun-2.log` and `/tmp/zeropass-macos-full-tests.log` show this recovery path being taken before the buttons become available.
   - Impact: the test will pass even if the app still launches without a visible auth window, so the original regression is stabilized around rather than actually covered.
   - Recommendation: keep the recovery helper for sheet interaction tests, but split out a dedicated launch-visibility assertion (or rename the current test so its intent matches its behavior).

### Low Priority
1. **Menu-driven recovery is localization brittle**
   - `ensureMainWindowVisible(in:)` hard-codes `"File"` and `"New Window"`.
   - Impact: brittle on non-English environments or future menu-copy changes.
   - Recommendation: prefer a nonlocalized trigger if possible (for example a keyboard shortcut path after activation, or another helper route not tied to menu titles).

2. **Minor compiler warning in UI tests**
   - The UI-test build emits `warning: result of call to 'launchFreshApp()' is unused` from `testLaunchPerformance()`.
   - Impact: harmless, but adds avoidable noise to local/CI verification.
   - Recommendation: assign to `_ = launchFreshApp()` or inline the launch call.

3. **`ZeroPassUITestsLaunchTests` still does not validate window visibility**
   - `ZeroPassUITestsLaunchTests.testLaunch()` launches and captures a screenshot, but never checks for the auth controls or uses the new visibility helper.
   - Impact: it will not catch a missing main window, and its attachment may continue to capture a menu-bar-only launch state.
   - Recommendation: either assert auth-window visibility there or keep it explicitly screenshot-only and rely on a separate smoke test for the window guarantee.

### Edge Cases Found by Scout
- Windowless launch remains reproducible under UI tests; the fallback path is exercised in the provided logs.
- The menu bar extra/status-item UI remains present in the accessibility tree and can dominate queries when no main window exists.
- `UITEST_RESET_STATE` is correctly wired in `VaultClient.resetStateForUITestsIfNeeded()`, which helps clear bookmarks/defaults-driven state leakage between runs.

### Positive Observations
- Accessibility identifiers were added with clear, stable namespaces: `welcome.*`, `createVault.*`, and `openVault.*`.
- `launchFreshApp()`, `ensureMainWindowVisible(in:)`, and `waitForNonExistence(...)` reduce duplicated test setup and make the intent readable.
- `ZeroPassApp` skips notification and hotkey setup in `UITEST_MODE`, removing unrelated test flake vectors.
- `VaultFolderPicker` centralizes auth-sheet folder selection and prefers sheet presentation on a real window when one exists.

### Recommended Actions
1. Add or restore a dedicated assertion that a fresh UI-test launch presents a visible auth window without manual recovery.
2. Keep `ensureMainWindowVisible(in:)` for flow tests that only need a usable auth window.
3. Silence the unused-result warning in `testLaunchPerformance()`.
4. Replace menu-title-based recovery with a nonlocalized trigger if CI may ever run under other locales.
5. Decide whether `ZeroPassUITestsLaunchTests.testLaunch()` should remain screenshot-only or become a true launch smoke test.

### Metrics
- Reviewed scope: 4 focus files (581 LOC); 10 files total with nearby dependencies (2,206 LOC)
- Verification observed: 15 unit tests passed; 6 UI tests passed
- Test Coverage: not collected in this session
- Type Coverage: N/A (Swift)
- Linting/Build issues observed: 1 compiler warning in UI tests

### Unresolved Questions
- Is a windowless launch under UI tests considered acceptable product behavior for ZeroPass, or should the app itself guarantee a visible auth window on launch?
