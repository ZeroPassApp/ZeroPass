# Phase 4: QA + Doc Sync

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `plans/260402-macos-swiftui-app/phase-03-authentication-views.md`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`
- `docs/code-standards.md`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Validate the auth overhaul end-to-end and update any design docs that still describe the old hero/full-window unlock direction.

## Key Insights
- The existing macOS auth design docs still point toward a centered hero/full-window unlock experience.
- The biggest regression risk is not code correctness; it is polish drift: bad window transitions, weak keyboard flow, or inconsistent light/dark behavior.
- `docs/code-standards.md` is repo-wide and likely unaffected by this UI-only change.

## Requirements
- Validate appearance in system, light, and dark modes.
- Validate Full Keyboard Access, focus order, VoiceOver labels, and reduced-motion fallbacks.
- Validate auth window sizes on first launch, after lock, after recovery display, and after returning to the unlocked main shell.
- Update any canonical auth planning docs that would otherwise send future work back toward the old hero design.

## Architecture
- Treat QA as a matrix across auth states, not a single happy path.
- Keep doc updates limited to design direction and implementation notes; avoid editing broad standards docs unless a real convention changed.

## Related Code Files

### Files to Edit
- `plans/260402-macos-swiftui-app/phase-03-authentication-views.md`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`

### Files Likely Unchanged
- `docs/code-standards.md`

## Implementation Steps
1. Run a manual QA pass for `noVault → create/open → locked → Touch ID/password/recovery → showingRecovery → unlocked → lock again`.
2. Verify the auth window does not reopen at `1000×680` with tiny centered content.
3. Verify Touch ID visibility, mode switching, command placement, and destructive confirmations.
4. Update the older macOS plan/design docs so they describe the new native auth shell and state-aware sizing instead of the old hero screen.
5. Capture any follow-up bugs separately rather than sneaking partial fixes into the overhaul.

## Todo List
- [ ] Run auth-state QA matrix
- [ ] Validate accessibility and reduced-motion behavior
- [ ] Validate state-aware window sizes and transitions
- [ ] Update outdated macOS auth planning/design docs

## Success Criteria
- The overhaul is complete, not just visually improved on one screen.
- The old hero/full-window auth guidance is no longer the documented default.
- Future contributors can follow the updated plan without reintroducing the current issues.

## Risk Assessment
- **Risk:** Old docs keep reintroducing the previous direction.  
  **Mitigation:** Update the canonical phase/design docs in the same change window.
- **Risk:** Window sizing behaves correctly only on the happy path.  
  **Mitigation:** Test transitions from every auth state, including lock after unlock and recovery display.

## Security Considerations
- Ensure QA covers recovery phrase visibility rules, clipboard behavior, and generic auth error messaging.
- Do not add screenshots of live recovery phrases to docs or test artifacts.

## Next Steps
- Hand the plan to implementation with the auth-shell phase first; do not start by tweaking the current unlock screen in place.
