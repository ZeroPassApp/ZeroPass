## Code Review Summary

### Scope
- Files: `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`, `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`, `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- Adjacent files checked: `ContentView.swift`, `UnlockVaultView.swift`, `AuthSceneScaffold.swift`, `AuthWindowLayoutModifier.swift`
- LOC reviewed: 2,556
- Focus: current uncommitted final state for the macOS auth UI smoke-test stabilization work
- Scout findings: manual scout of the command-routing and window lifecycle found one remaining no-window command gap, one English-only UITest fallback, and one duplicated folder-picker path

### Overall Assessment
The stabilization pass is meaningfully better. The app-side launch/reopen recovery is now much more robust, the no-vault `Open Vault…` command no longer fights a separate panel path when a welcome window is present, and the exact fresh verification log the user cited (`/tmp/zeropass-macos-full-tests-4.log`) shows `6/6` UI tests passing with `** TEST SUCCEEDED **`.

I do not see any critical or high-severity problems in the reviewed final state. The remaining risk is mostly around windowless edge cases and small maintainability/testing brittleness rather than the primary smoke-test path.

### Critical Issues
- None.

### High Priority
- None.

### Medium Priority
1. **`Open Vault…` still has a windowless edge case**
   - `ZeroPassApp.swift:61-65` now routes the no-vault command to `vault.presentOpenVaultSheet()`, but the actual presenter lives in `WelcomeView.swift:52` via `.sheet(item: activeAuthModalBinding)`.
   - If the user closes every main window and then invokes `Open Vault…` / `⌘O` from the app menu, the command mutates shared state but does not create or surface a window. `applicationShouldHandleReopen` in `ZeroPassApp.swift:156-161` only helps when the app is explicitly reopened, not when this command is invoked while the app is already active and windowless.
   - Impact: the primary command can still appear to do nothing in a real macOS edge case that the current smoke tests do not cover.
   - Recommendation: centralize the no-vault command behind a helper that ensures a main window exists before setting `activeAuthModal`, or route the command through the same reopen/materialize logic used by the app delegate.


### Low Priority
1. **UITest recovery helper is still English-locale brittle**
   - `ZeroPassUITests.swift:110-114` falls back to `File` → `New Window` using literal menu titles.
   - Impact: if the fallback path is ever needed on a localized runner, or if menu copy changes, the safety net becomes flaky.
   - Recommendation: prefer a nonlocalized activation strategy for the fallback (for example a test hook, a keyboard-shortcut-based path, or a more selector-oriented query if available).

2. **Folder-picking behavior is still split across two implementations**
   - Auth sheets use the shared async picker in `OpenVaultSheet.swift:91-133` and `CreateVaultView.swift:172`, but the app command path still carries its own `NSOpenPanel.runModal()` implementation in `ZeroPassApp.swift:114-145`.
   - Impact: the smoke-test conflict is fixed for the no-vault path, but the repo still has two folder-picking/opening flows to keep in sync.
   - Recommendation: consolidate the command path onto the same picker/open helper used by the auth surfaces.

### Edge Cases Found by Scout
- No-window `⌘O` / app-menu `Open Vault…` does not itself materialize a presenter window.
- UITest fallback recovery depends on English menu labels.
- Folder-picker behavior can still drift between auth-sheet and app-command entry points.

### Positive Observations
- `AppDelegate` window recovery is stronger now: it activates the app, retries for the main window, deminiaturizes existing windows, and only then falls back to opening a new one.
- App-side new-window recovery is no longer tied to literal `File` / `New Window` titles; it searches by `⌘N`, which is less brittle than the earlier title-based approach.
- The welcome screen now owns the no-vault create/open modal state through `VaultClient.AuthModal`, which removes the earlier command-path divergence.
- `VaultClient.resetStateForUITestsIfNeeded()` clears persisted defaults and bookmarks early, which makes the smoke tests more deterministic.
- The exact fresh log provided by the user shows the four auth UITests plus the two launch UITests all passing.

### Recommended Actions
1. Fix the remaining windowless `Open Vault…` command path by ensuring the command can surface or create a presenter window before setting `activeAuthModal`.
2. Replace the English-only UITest fallback with a nonlocalized recovery path.
3. Consolidate the remaining command-path folder picker onto the shared auth picker/helper to reduce drift.

### Metrics
- Type Coverage: N/A (Swift review; no type-coverage metric available)
- Test Coverage: Not measured in this review
- Test Evidence: `/tmp/zeropass-macos-full-tests-4.log` shows `Executed 6 tests, with 0 failures (0 unexpected)` for `ZeroPassUITests.xctest` and `** TEST SUCCEEDED **`
- Linting Issues: Not checked in this review

### Unresolved Questions
- None that block the review; the remaining issues are follow-up hardening, not evidence gaps.
