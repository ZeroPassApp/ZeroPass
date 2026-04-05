---
title: "Add stable macOS UI tests for auth and vault lifecycle"
description: "Focused plan to make create/open/unlock/recovery UI flows deterministic under macOS UI automation."
status: completed
priority: P1
effort: 5h
branch: main
tags: [macos, swiftui, ui-tests, auth, vault]
created: 2026-04-05
completed: 2026-04-05
validated: FRESH_XCODEBUILD_UI_TESTS + REVIEW_APPROVED_WITH_NITS + TESTER_READY_TO_KEEP
validation_timestamp: 2026-04-05
---

# Stable macOS UI tests for auth/vault lifecycle

## Focused scope
- Replace folder-selection flakiness in UI tests with deterministic launch-driven vault path injection.
- Cover four stable end-to-end UI flows: create vault, recovery phrase continue, locked-state unlock, and open existing vault.
- Keep production folder picking on `NSOpenPanel` unchanged outside `UITEST_MODE`.
- Limit touch points to the auth/vault entry flow plus the UI test target.

## Progress Snapshot
- Completed: deterministic folder selection now uses `UITEST_PICK_DIRECTORY_PATH` via the UI-test-only `VaultFolderPicker` override path.
- Completed: stable accessibility identifiers now cover auth and unlocked-shell assertions, including create/open sheet actions, recovery continue, unlock password/submit, and `mainShell.root` / `mainShell.newItemButton`.
- Completed: `ZeroPassUITests.swift` now covers the critical full-flow lifecycle paths:
   - create vault → recovery phrase → unlocked shell
   - relaunch existing vault → locked state → unlock
   - open existing vault via deterministic picker → unlock
- Validation completed successfully with fresh `xcodebuild` UI test runs.
- Reviewer outcome: approve with nits.
- Tester outcome: ready to keep.

## Likely files
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- Optional small helper near app startup if launch-option parsing becomes noisy.

## Implementation steps
1. **Add a UI-test launch configuration seam**
   - Centralize parsing of `UITEST_MODE` plus explicit vault-path overrides from launch arguments/environment.
   - Keep the seam tiny and readable so auth views do not each parse `ProcessInfo` ad hoc.

2. **Make folder selection deterministic in UI-test mode**
   - Update the create/open flows to resolve a pre-supplied vault folder when a UI-test override is present.
   - Fall back to `VaultFolderPicker` + `NSOpenPanel` for every non-test launch and for UI-test runs without an override.

3. **Add stable UI assertions for each auth state**
   - Add explicit accessibility identifiers for the create inputs/actions, recovery continue action, unlock inputs/actions, and a clear unlocked-state sentinel in the main shell.
   - Prefer identifiers over text-only assertions so tests survive copy/styling changes.

4. **Expand the UI-test harness into full-flow coverage**
   - Add helpers that create unique temporary vault directories, launch with or without `UITEST_RESET_STATE`, and pass vault-path overrides through `launchEnvironment`/`launchArguments`.
   - Implement stable tests for:
     - create vault → recovery phrase shown
     - recovery phrase continue → unlocked shell shown
     - relaunch existing vault without reset → locked state → unlock succeeds
     - open existing vault via deterministic path → locked state → unlock succeeds

5. **Tighten validation around state markers, not timing guesses**
   - Reuse the existing stale-process cleanup, but wait on explicit screen markers instead of arbitrary sleeps.
   - Keep each test self-contained and filesystem-backed so failures point to one broken lifecycle step.

## Guardrails
- **Determinism first:** do not automate `NSOpenPanel`; use launch-driven path injection for all full-flow UI tests.
- **Security intact:** never log master passwords or recovery phrases; keep test vaults in per-test temporary folders and clean them up after use.
- **No production behavior change:** any new hook must be gated behind `UITEST_MODE`, with current production behavior preserved when hooks are absent.
- **Minimal surface area:** prefer one small launch-options helper over scattering test conditionals across multiple views.
- **State isolation:** only use `UITEST_RESET_STATE` for tests that need a fresh app state; relaunch/unlock coverage should deliberately preserve the created vault between launches inside the same test.

## Validation command
`xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' -only-testing:ZeroPassUITests CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`

## Unresolved questions
- None.

---

## Completion Summary

**Completed:** 2026-04-05  
**Validation:** Fresh macOS UI-test `xcodebuild` runs passed  
**Outcome:** Focused auth/vault lifecycle stabilization work is implemented, validated, and ready to keep.
