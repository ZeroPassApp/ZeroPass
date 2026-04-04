---
plan: 20260404-macos-auth-window-flow-followups
date: 2026-04-04
status: COMPLETED
validation: CODE_REVIEW + FULL_MACOS_TESTS ✓
---

# macOS Auth/Window Follow-up Cleanup — Plan Sync Report

## Executive Summary
Completed. All 4 phases implemented, code reviewed, and validated via full macOS test scheme.

**Test Result:** `** TEST SUCCEEDED **` on 2026-04-04

## Deliverables

### Phase 1: Window-Local Auth State ✓
- Moved auth sheet presentation from `VaultClient.activeAuthModal` (app-global) to window-local state
- Eliminated cross-window sheet bleed
- Files: `ZeroPassApp.swift`, `WelcomeView.swift`, `VaultClient.swift`

### Phase 2: No-Visible-Window Recovery ✓
- Added UI automation for `⌘O` recovery when all windows closed
- Covered by new UI smoke test
- UI tests now reset persisted state on launch + assert shutdown

### Phase 3: Dead Flow Removal ✓
- Removed dead no-vault folder-open branch in `ZeroPassApp.chooseVaultFolder()`
- Deleted `presentCreateVaultSheet()`, `presentOpenVaultSheet()`, `dismissAuthModal()`, `activeAuthModal` from VaultClient
- Cleaned up duplicate picker/error logic

### Phase 4: Launch Stability Hardening ✓
- `terminateRunningAppIfNeeded()` now asserts success + fails loudly on stale processes
- Fresh-launch lifecycle checks tightened
- Prevents CI/slow-startup contamination

## Validation Evidence
- **Code Review:** 6 files reviewed | 490 insertions / 61 deletions | 0 critical/high/medium issues
- **Test Coverage:** Full macOS scheme passed (ZeroPassTests + ZeroPassUITests)
- **Edge Cases:** Multi-window targeting, no-visible-window reopen, stale launch-state — all covered

## Known Observations (Low Priority)
1. `WelcomeView` button handlers should clear `vault.authFlowError` for parity with menu-command path
2. One-shot notification handoff after `waitUntilMainWindowIsVisible()` — adequate for current timing but can queue if flakiness appears
3. Replace-current command path still duplicates `OpenVaultSheet`/`VaultFolderPicker` flow (maintainability, not regression)

## Impact
- ✓ Eliminates process-global modal state — safer multi-window behavior
- ✓ Improves fresh-launch robustness
- ✓ Removes dead code paths
- ✓ All existing tests green

## Status
**PLAN CLOSED** — Ready for merge. No blocking issues. Recommended actions documented for future polish iteration.
