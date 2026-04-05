# Phase 2: Auth Screens

## Context Links
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`

## Overview
- **Priority:** P1
- **Status:** In Progress
- **Description:** Bring welcome, unlock, create/open vault, and auth recovery presentation into a tighter, premium native flow with less dead space and clearer action hierarchy.

## Key Insights
- The auth route is already structurally organized around `AuthSceneScaffold` and state-aware `ContentView` switching; this phase is mostly a design-system-and-layout pass, not a flow rewrite.
- Stitch concepts exist for welcome, unlock, and recovery phrase, so this phase has strong visual guidance without needing to change app architecture.
- `UnlockVaultView` already carries the right behaviors (password/recovery split, biometrics, vault menu, error handling); the redesign should make those behaviors feel more intentional and discoverable.
- A focused unlock-screen polish slice has landed: hero-card centering, icon-first vault menu trigger, eye-based password visibility control, contextual Touch ID cues, and stronger error-highlight semantics.

## Requirements
- Welcome and unlock views must immediately communicate “premium utility,” not sparse hero screen.
- Cobalt should highlight the primary action and key affordances; vault identity, helper text, and destructive actions should sit in quieter graphite layers.
- Create/open vault sheets should match the new auth chrome and tighten vertical spacing.
- Recovery phrase presentation must stay highly readable, obviously one-time, and safe for copy/write-down workflows.
- Preserve current modal routing, focus restoration, auth state transitions, and `authWindowLayout` behavior.

## Architecture
- Reuse `AuthSceneScaffold` and the existing section subviews; avoid a new auth-specific component tree.
- Treat the Stitch concepts as visual targets mapped onto the current state machine.
- If `UnlockVaultView.swift` grows further, split only visually isolated regions, not the unlock logic itself.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`

### Files to Create
- None planned.
- Optional: one small auth-only helper if it removes repeated header/action chrome from multiple auth files.

## Implementation Steps
1. Map the Stitch welcome/unlock/recovery concepts onto the current auth states and identify which ideas can be expressed with native SwiftUI controls.
2. Restyle `WelcomeView` so the app introduction, primary CTA, secondary CTA, and version/footer feel denser and more premium without losing clarity.
3. Rework `UnlockVaultView` plus its password/recovery subviews so vault identity, unlock mode, Touch ID, helper text, and the vault options menu read with stronger hierarchy and less visual drift.
4. Update `CreateVaultView` and `OpenVaultSheet` to match the Midnight Native panel treatment while keeping their current validation and focus behavior.
5. Refresh `RecoveryPhraseView` and `RecoveryPhraseCardView` to align with the chosen visual direction while preserving copy safety and one-time visibility messaging.

## Todo List
- [ ] Apply Midnight Native layout and emphasis to welcome screen
- [ ] Redesign unlock screen around clearer mode selection and biometric prominence
- [ ] Bring create/open vault sheets into the same panel language
- [ ] Restyle auth recovery phrase presentation without reducing readability

## Success Criteria
- Auth views look like one deliberate product family.
- The unlock screen no longer feels sparse or generic; Touch ID and vault identity are obvious.
- Create/open/recovery surfaces match the direction without regressing focus or keyboard flow.

## Risk Assessment
- **Risk:** Premium styling adds too much chrome to a security-critical flow.  
  **Mitigation:** Keep the structure native and simple; use hierarchy, density, and color restraint instead of ornamental effects.
- **Risk:** Recovery phrase styling prioritizes looks over legibility.  
  **Mitigation:** Preserve large text, generous contrast, and a print-like card treatment for the mnemonic itself.

## Security Considerations
- Do not weaken current clipboard, recovery, or generic auth error behavior.
- Keep destructive vault actions separated and clearly labeled even if they move visually.

## Next Steps
- Finish the remaining auth-screen work: welcome, create/open vault, and recovery presentation alignment.
- Then move into the unlocked workspace, where the same token system needs to scale across sidebar, list, detail, and toolbar surfaces.