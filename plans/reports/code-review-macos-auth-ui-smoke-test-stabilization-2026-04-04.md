# Code Review Report: macOS auth UI smoke-test stabilization
**Date:** April 4, 2026  
**Scope:** `ZeroPassApp.swift`, `ZeroPassUITests.swift`, `WelcomeView.swift`, `CreateVaultView.swift`, `OpenVaultSheet.swift`, `VaultClient.swift`  
**Supporting reads:** `ContentView.swift`, `AuthSceneScaffold.swift`, `AuthWindowLayoutModifier.swift`  
**Status:** ⚠️ Solid stabilization overall, with 2 medium and 2 low remaining risks

---

## Verdict

The final state is directionally good:
- the welcome flow and the no-vault `Open Vault…` command now converge on the same auth-modal path,
- launch visibility is actively recovered,
- UI smoke tests assert a true fresh launch before exercising the auth sheets,
- fresh verification passed in `/tmp/zeropass-macos-full-tests-5.log` (`TEST SUCCEEDED`).

I did **not** find a blocking correctness issue in the reviewed path, but I did find a couple of remaining brittleness points worth tracking.

---

## Review scope and evidence

- Reviewed current implementations plus the exact diff against `HEAD~1`
- Read the relevant launch/window/auth dependencies (`ContentView`, auth scaffold/layout)
- Checked the fresh macOS test log at `/tmp/zeropass-macos-full-tests-5.log`
- Focus diff size: **385 insertions / 94 deletions** across the 6 requested files
- Reviewed file footprint: **1,801 LOC** across the 6 requested files; **2,293 LOC** including supporting reads

### Fresh verification evidence

From `/tmp/zeropass-macos-full-tests-5.log`:
- `testWelcomeScreenShowsPrimaryActionsOnFreshLaunch` ✅
- `testCreateVaultSheetCanBeOpenedAndCancelled` ✅
- `testOpenVaultSheetCanBeOpenedAndCancelled` ✅
- `testLaunchPerformance` ✅
- `TEST SUCCEEDED` ✅

---

## Medium severity

### 1. Global auth sheet state is now shared across all windows
**Files:** `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift:13-25,205-216`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift:52-66`, `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:170-249`

`WelcomeView` moved from local `@State` sheet flags to the app-wide `VaultClient.activeAuthModal`. That solved the menu-command routing problem, but it also changed the presentation state from **window-local** to **process-global**.

Because the app uses a `WindowGroup`, multiple welcome windows are possible. In that case:
- one window can trigger sheet presentation for another window,
- one window’s state change can dismiss another window’s sheet via `dismissAuthModal()`,
- the new window-rescue path can make this edge case more likely by explicitly opening windows with `⌘N` semantics.

**Impact:** not a current blocker for the single-window smoke path, but it is a real correctness risk for multi-window macOS behavior.

**Recommendation:** track auth modal presentation in window-scoped state (scene model / scene storage / per-window coordinator) and have menu commands target a specific visible window instead of a global `VaultClient` modal flag.

### 2. The launch/reopen rescue branch is still heuristic and not directly covered by automation
**Files:** `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:157-249`, `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:26-61,103-114`

The rescue logic is implemented as:
1. try to find an existing main window,
2. poll for up to ~1s,
3. recursively scan the app menu for any `⌘N` item,
4. send that item’s action,
5. poll again.

That is workable, but it is still a heuristic. It can silently drift if menu structure, shortcut assignments, or window-identification behavior changes.

Current UI tests **do not directly exercise**:
- `applicationShouldHandleReopen`,
- the app-side `openNewWindowFromMenuIfAvailable()` fallback,
- the no-visible-window `Open Vault…` menu path that is supposed to reveal a window and then present the shared auth sheet.

**Impact:** the most brittle part of the stabilization is currently the least explicitly protected by tests.

**Recommendation:** add a dedicated UI smoke covering “no visible main window → `⌘O` or reopen → visible window + open-vault sheet”. That would guard the exact branch this change introduced.

---

## Low severity

### 3. `ZeroPassApp.chooseVaultFolder(replacingCurrent:)` still carries a dead no-vault branch
**Files:** `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift:58-70,115-143`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift:87-141`

After the no-vault `Open Vault…` command was rerouted through `activeAuthModal`, `chooseVaultFolder(replacingCurrent: false)` is no longer called from `ZeroPassApp`.

That leaves:
- an unused code path,
- duplicated folder-open behavior between `ZeroPassApp` and `OpenVaultSheet`/`VaultFolderPicker`,
- a higher chance of subtle drift in error handling and panel behavior.

**Recommendation:** collapse `chooseVaultFolder` to the replace-current case only, or centralize both flows behind one folder-picking/open coordinator.

### 4. The fresh-launch test helper does not fail hard if the previous app instance survives termination
**Files:** `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift:71-101`

`terminateRunningAppIfNeeded()` force-terminates existing app instances and waits up to 5 seconds, but if the app is still alive after that deadline the helper just continues.

That means the tests can silently stop being truly “fresh launch” tests if termination ever stalls.

**Recommendation:** assert that no matching process remains after the wait loop, so state contamination fails loudly instead of becoming background flake material.

---

## Positive observations

- **Correct direction of travel:** routing no-vault `Open Vault…` through the same auth sheet state as the welcome screen removes one whole class of “UI path A != UI path B” bugs.
- **Good UI-test anchors:** accessibility identifiers on the welcome buttons and sheet titles make the smoke tests much less string-fragile.
- **Keyboard flows are consistent:** default/cancel actions line up with the test interactions (`Return`, `Escape`, `⌘O`).
- **Fresh-state handling improved:** `UITEST_RESET_STATE` clears defaults/bookmarks before the rest of `VaultClient` initialization proceeds.
- **Launch verification exists now:** the dedicated fresh-launch assertion prevents the sheet tests from masking a broken initial window state.

---

## Bottom line

**Overall assessment:** good final polish, **not blocked**, but still carrying **one real multi-window correctness risk** and **one meaningful automation gap** around the launch/reopen rescue branch.

### Severity summary
- **Critical:** 0
- **High:** 0
- **Medium:** 2
- **Low:** 2

### Recommended next actions
1. Make auth modal presentation window-scoped instead of global.
2. Add one targeted UI smoke for reopen / no-visible-window recovery.
3. Remove the dead no-vault branch in `chooseVaultFolder(replacingCurrent:)`.
4. Make the fresh-launch helper fail hard if termination does not actually complete.
