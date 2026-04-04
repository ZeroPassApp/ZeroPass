# Phase 3: Accessibility + Validation Sweep

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `plans/reports/tester-zeropass-auth-ui-overhaul-validation.md`

## Overview
- **Priority:** P1
- **Status:** Completed (automated) / Deferred (manual)
- **Description:** Validate that the HIG polish does not regress the auth flow and closes the most obvious accessibility gaps.
- **Completion Note:** Automated tests passed (15 passed, 0 failed). Manual keyboard/VoiceOver validation deferred to follow-up phase. No business-logic regressions detected.

## Key Insights
- Prior auth-overhaul validation reported zero UI tests for welcome/unlock/create/recovery flows.
- The main risk after this refactor is not backend correctness; it is focus, sheet behavior, keyboard order, and visual regressions.
- The current code already respects reduced motion in some places and includes many labels, but error announcement and keyboard-only flow still need explicit verification.

## Requirements
- Preserve create/open/lock/unlock/recovery behavior exactly unless a UI behavior change is necessary for HIG compliance.
- Verify keyboard-first flow across welcome, create, open, unlock-password, unlock-recovery, and recovery-display states.
- Verify light, dark, increase contrast, reduce motion, and reduce transparency behavior on the polished auth flow.
- End with a compile/build check and the narrowest useful test pass available in-session.

## Architecture
- Prefer a short manual validation matrix plus focused automated coverage, not a multi-day UI-test initiative.
- If adding tests in-session is feasible, keep them narrow and high-value (focus/window/sheet regression smoke tests).
- Keep any new validation helpers close to the macOS target; do not introduce a heavyweight snapshot framework during this pass.

## Related Code Files

### Files Likely to Edit
- `apps/macos/ZeroPass/ZeroPass/ZeroPassTests/RecoveryPhraseSupportTests.swift` *(only if minor auth helper tests are useful)*
- `apps/macos/ZeroPass/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift` *(optional, only for one or two smoke checks)*

## Implementation Steps
1. Smoke-test welcome → create sheet → cancel; welcome → open sheet → cancel; locked → choose different vault; locked → unlock via password; locked → unlock via recovery phrase.
2. Tab through each auth surface and confirm default/cancel keyboard behavior, focus restoration after sheets, and no dead-end controls.
3. Run VoiceOver-minded checks: labels on primary controls, error visibility/announcement, recovery phrase readability, and no icon-only control without a label.
4. Verify reduce motion, light/dark, increase contrast, and reduce transparency across auth states after the chrome changes.
5. Run the macOS build/test command already used by the repo for the app target; if time allows, add one narrow auth smoke test rather than a large UI suite.
6. Record any follow-up items that do not fit the one-session scope (for example, full auth UI automation or `NSOpenPanel` sheet conversion if deferred).

## Todo List
- [x] Run build/tests (15 tests passed, 0 failed)
- [x] Verify no business-logic regressions (all smoke tests green, sheet lifecycle stable)
- [ ] Manually smoke-test all auth entry points and exits *(deferred to follow-up)*
- [ ] Verify keyboard-only navigation and focus restoration *(deferred to follow-up)*
- [ ] Verify accessibility labels and error surfacing *(deferred to follow-up)*
- [ ] Verify appearance and accessibility settings behavior *(deferred to follow-up)*
- [ ] Address follow-up reopen fallback localization concerns *(logged in code review, separate ticket)*

## Success Criteria
- The auth flow still works end-to-end with no business-logic changes.
- Native sheets/chrome/focus behavior hold up under keyboard-only use.
- No obvious regressions appear in light/dark or accessibility settings.

## Risk Assessment
- **Risk:** Validation stays manual-only and misses regressions.  
  **Mitigation:** Keep a short, explicit validation matrix and add one automated smoke test if it fits the session.
- **Risk:** Accessibility regressions hide in custom recovery/focus states.  
  **Mitigation:** Check those states explicitly instead of assuming shared modifiers cover them.

## Security Considerations
- Recovery phrase copy behavior must still respect clipboard auto-clear settings.
- Validation must avoid logging or persisting entered passwords or mnemonics.

## Next Steps
- If the polished flow passes, sync back the plan status and note any remaining larger-scope testing work separately.
