# Phase 01: Fit recovery phrase screen inside the default window

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`

## Overview
- **Priority:** P2
- **Status:** Pending
- **Description:** Remove the need to scroll on the recovery phrase screen at the default recovery-window size with the smallest safe UI change set.

## Key Insights
- The screen already has a safe fallback (`ScrollView`), so the bug is usability/polish, not a broken flow.
- The tight constraint is vertical, not horizontal; the known problematic default is `680x580` content.
- The global scaffold is shared by multiple auth screens, so changing its scrolling model has wider blast radius than changing the recovery window size or compacting recovery-only content.
- The recovery view contains a couple of easy compaction candidates: duplicate helper pills and generous spacing between sections.

## Requirements
- At the default recovery window size, the confirmation toggle and Continue button must be visible without user scrolling.
- Keep the recovery phrase readable and selectable.
- Preserve existing clipboard, continue, accessibility-identifier, and confirmation-toggle behavior.
- Keep the screen resilient for smaller manually resized windows by retaining scroll fallback.

## Architecture
- Primary fix in `AuthWindowLayoutModifier.swift`: give `.showingRecovery` a taller default content height; only widen if the visual pass proves width contributes to wrapping.
- Secondary fix in `RecoveryPhraseView.swift` and/or `RecoveryPhraseCardView.swift`: trim only recovery-specific vertical weight if sizing alone still leaves the footer too close to the fold.
- Leave `AuthSceneScaffold.swift` unchanged unless both of the above fail; it is shared infrastructure and not the root mismatch.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`

### Files to Avoid Unless Necessary
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`

## Implementation Steps
1. Increase the `.showingRecovery` default content height in `AuthWindowLayoutModifier.swift` (start with a modest bump, roughly +60 to +80pt) and keep the change isolated to the recovery state.
2. Re-check the recovery screen at the default size; if the CTA/toggle still sit too low, compact only the recovery-specific layout by removing redundant helper chrome or tightening recovery-only spacing/order.
3. Keep `AuthSceneScaffold` scroll behavior intact so smaller windows still degrade gracefully instead of clipping.
4. Extend the macOS UI test flow to assert the recovery confirmation toggle and Continue button are both visible/hittable before interaction, then run the existing create-vault recovery path as the regression gate.

## Todo List
- [ ] Tune recovery-state default window height only
- [ ] Apply light recovery-only compaction if sizing alone is still insufficient
- [ ] Add/adjust UI verification for on-screen recovery controls
- [ ] Run targeted UI validation + manual default-size spot check

## Success Criteria
- Recovery toggle visible at default recovery window size.
- Continue button visible at default recovery window size.
- No scroll required on a normal create-vault recovery pass at default size.
- Locked/no-vault/unlocked window behavior unchanged.

## Risk Assessment
- **Risk:** Over-correcting the window size makes the auth window feel oversized.  
  **Mitigation:** Change recovery height only; keep width unchanged unless proven necessary.
- **Risk:** Compacting the recovery view harms readability.  
  **Mitigation:** Remove redundancy first; avoid shrinking typography or word-cell tap targets.
- **Risk:** UI tests pass on element existence but miss offscreen placement.  
  **Mitigation:** Assert `isHittable`/interactive visibility before clicking.

## Security Considerations
- Do not change mnemonic handling, text-selection behavior, or clipboard semantics.
- Do not add logging around recovery phrase content.
- Preserve the explicit confirmation gate before continuing.

## Next Steps
- If this screen still feels crowded after the minimal fix, consider a later auth-polish pass to consolidate helper copy/pills across auth surfaces.

## Unresolved questions
- Whether the final recovery height lands closer to `640` or `660` should be decided by the implementation pass after a quick visual verification.