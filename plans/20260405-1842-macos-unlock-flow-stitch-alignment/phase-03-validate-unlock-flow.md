# Phase 03 — validate unlock flow

## Context Links
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- `apps/macos/ZeroPass/ZeroPassTests/RecoveryPhraseSupportTests.swift`
- `docs/project-changelog.md`
- `plans/20260405-1230-unlock-vault-stitch-fidelity-pass/`
- `plans/reports/tester-unlock-screen-validation-2026-04-05.md`

## Overview
- Priority: P1
- Status: pending
- Brief description: Close the remaining confidence gap with deterministic UI coverage for auth-state transitions plus a manual accessibility and Touch ID pass.

## Key Insights
- Full macOS scheme validation already exists, but recovery unlock UI, keyboard navigation, Touch ID hardware, and resize/VoiceOver behavior remain under-validated.
- The latest unlock-focused plan history already calls manual unlock-specific regression the remaining open gate.

## Requirements
- Cover create → recovery → unlocked, open existing → locked → unlock, recovery phrase unlock, choose different vault, and window-state transitions.
- Run manual checks for Touch ID hardware, keyboard-only flow, VoiceOver labels/hints, dark/light appearance, and lock/unlock resize behavior.
- Update docs/changelog only if user-visible behavior changes, not for pure internal cleanup.

## Architecture
- Extend the current deterministic UI test harness (`UITEST_MODE`, `UITEST_RESET_STATE`, `UITEST_PICK_DIRECTORY_PATH`) instead of adding new test infrastructure.
- Keep auth coverage close to the existing `ZeroPassUITests.swift` flow helpers.

## Related Code Files
- Modify: `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- Optional: `apps/macos/ZeroPass/ZeroPassTests/RecoveryPhraseSupportTests.swift`
- Optional if behavior changes: `docs/project-changelog.md`
- Create: none
- Delete: none

## Implementation Steps
1. Extend UI coverage for recovery phrase unlock, validation feedback, choose different vault, and post-recovery continue/unlock transitions.
2. Re-run the macOS scheme tests with `xcodebuild test`.
3. Execute a manual matrix for password, recovery, Touch ID, keyboard-only flow, VoiceOver, resize cycle, and dark/light appearance.
4. Update changelog/docs only if the shipped flow behavior changed.

## Todo List
- [ ] Add deterministic UI coverage for the remaining auth-state transitions
- [ ] Re-run the macOS scheme tests
- [ ] Perform the manual accessibility and Touch ID regression pass
- [ ] Sync changelog/docs only if behavior changed

## Success Criteria
- The full unlock journey is covered by automated smoke tests where practical and manual QA where hardware/accessibility require it.
- The prior “manual unlock-specific regression pending” note can be closed.
- No secrets are captured in screenshots, logs, or test fixtures.

## Risk Assessment
- Risk: UI tests become brittle if they assert layout details instead of state transitions.  
  Mitigation: assert stable identifiers and state outcomes, not pixel-perfect geometry.
- Risk: manual QA misses hardware-specific Touch ID issues.  
  Mitigation: run the final pass on a Mac with actual biometric support before calling the flow done.

## Security Considerations
- Keep test vaults in temporary directories and clean them up.
- Do not persist or document live recovery phrases, passwords, or vault keys.

## Next Steps
- Once validation is green, close the older unlock-specific pending notes and leave the locked-screen plan as historical context.
