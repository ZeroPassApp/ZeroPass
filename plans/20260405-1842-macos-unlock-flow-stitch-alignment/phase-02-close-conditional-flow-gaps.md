# Phase 02 — close conditional flow gaps

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

## Overview
- Priority: P1
- Status: pending
- Brief description: Only touch stateful unlock-flow behavior if the approved Stitch flow still differs from the current sequence after Phase 01.

## Key Insights
- The current state machine already covers `noVault`, `locked`, `showingRecovery`, and `unlocked` cleanly.
- Most remaining drift appears visual and journey-level, not core behavior-level.
- The riskiest files are `VaultClient.swift` and window/layout routing; they should be touched last and only with a concrete design reason.

## Requirements
- Keep password unlock, recovery unlock, Touch ID, generic error handling, and vault actions intact unless the design explicitly changes them.
- Any flow change must preserve accessibility, security copy, focus restore, and deterministic UI automation.
- Prefer view-local changes over state-model changes whenever both satisfy the design.

## Architecture
- Keep the existing state machine and modal routing.
- If a flow gap is real, add the smallest possible helper/flag inside the current structure rather than introducing a new auth coordinator.

## Related Code Files
- Modify if needed: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- Modify if needed: `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- Modify if needed: `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- Modify if needed: `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- Optional: `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- Create: none
- Delete: none

## Implementation Steps
1. Compare the approved Stitch flow to the current journey: `welcome/create/open → locked → showingRecovery? → unlocked`.
2. If the design difference is only presentation or progressive disclosure, keep the change view-local in `UnlockVaultView.swift`.
3. If the design truly changes ordering, acknowledgment timing, or transition behavior, make the smallest possible update in `VaultClient.swift`, `ContentView.swift`, and `AuthWindowLayoutModifier.swift`, then re-check every auth-state transition.
4. Update `ZeroPassApp.swift` only if command/menu entry points diverge from the adjusted flow.

## Todo List
- [ ] Confirm whether any approved Stitch flow gap remains after Phase 01
- [ ] Keep remaining changes view-local when possible
- [ ] Touch `VaultClient.swift` only if the sequence truly changes
- [ ] Recheck lock/recovery/unlock transitions after any stateful change

## Success Criteria
- Any real design-driven flow gap is closed with minimal state churn.
- The state machine remains simple and readable.
- Recovery, Touch ID, menu actions, focus, and resize transitions still behave predictably.

## Risk Assessment
- Risk: changing recovery or biometric sequencing introduces state bugs.  
  Mitigation: change one transition at a time and validate each auth-state hop explicitly.
- Risk: resize or focus regressions appear when moving between `locked`, `showingRecovery`, and `unlocked`.  
  Mitigation: keep `AuthWindowLayoutModifier` changes tiny and pair them with UI smoke coverage.

## Security Considerations
- Keep auth errors generic; do not reveal whether a password or recovery phrase failed in a more specific way.
- If recovery-step ordering changes, preserve the current secret-handling guarantees and avoid exposing more data earlier than necessary.

## Next Steps
- Freeze behavior once the remaining gap list is empty, then move straight into validation.
