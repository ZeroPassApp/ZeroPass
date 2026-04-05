# Phase 01 — align unlock screen polish

## Context Links
- Review summary from request
- Focused follow-up plan: `plans/20260405-1230-unlock-vault-stitch-fidelity-pass/`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`

## Overview
- Priority: P2
- Status: in progress
- Brief description: Original locked-state polish slice. The later focused Stitch fidelity pass deepened the layout work without changing unlock behavior; manual unlock-specific visual/accessibility regression is still pending.

## Key Insights
- Current `ScrollView` top-aligns the card, so the hero content never feels vertically centered.
- The vault menu trigger uses a generic `Options` label instead of a vault-specific affordance.
- The password visibility control is text-led; Stitch suggests a lighter icon-first affordance.
- Existing behavior to preserve: focus restoration, Return-to-unlock, busy/disabled states, error announcement/shake, and Touch ID path.

## Requirements
- Center the hero card vertically when content fits, while still allowing scroll when height is constrained.
- Make the vault menu trigger feel more like Stitch: icon-first and more specific to vault actions.
- Make password reveal feel closer to Stitch and keep it accessible.
- Avoid behavior regressions.

## Architecture
- `UnlockVaultView`: adjust the outer layout so content can center inside the available height without losing overflow safety; refresh the vault menu trigger styling.
- `UnlockPasswordSection`: refine the trailing reveal/hide affordance into an icon-based toggle while preserving focus and submit behavior.
- `AuthWindowLayoutModifier`: only make a small locked-state size tweak if the centered layout still feels cramped.
- `UnlockRecoverySection`: touch only if shared field chrome/spacing needs parity after the password-field polish.

## Related Code Files
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- Optional: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`

## Implementation Steps
1. Update `UnlockVaultView` so the card sits visually centered in the locked window when there is enough height, but still scrolls on smaller windows.
2. Restyle the vault menu trigger to be icon-first and vault-specific while keeping the same menu actions and busy-state behavior.
3. Replace or restyle the text `Show/Hide` password affordance with an accessible icon toggle that preserves current keyboard, focus, and secure-entry behavior.
4. Smoke-check locked-window proportions and make only a minimal `AuthWindowLayoutModifier` size adjustment if needed.
5. Only update `UnlockRecoverySection` if matching spacing/border treatment is necessary for consistency.

## Todo List
- [x] Center the unlock card without removing overflow safety
- [x] Update the vault menu trigger styling/copy without changing actions
- [x] Refine the password visibility affordance and preserve accessibility/focus
- [x] Validate locked-state sizing and touch recovery section only if needed
- [ ] Run the remaining formal manual unlock-specific visual/accessibility regression pass

## Success Criteria
- Unlock card feels vertically centered at the normal locked-window size.
- Vault menu trigger feels closer to Stitch and remains clear to keyboard and VoiceOver users.
- Password reveal is icon-first, accessible, and does not break submit/focus.
- Password, recovery, Touch ID, error, and menu behaviors continue to work.

## Risk Assessment
- Removing the scroll container or centering incorrectly could clip content in smaller windows or larger text settings.
- Changing the password field control incorrectly could break `SecureField` semantics, Return-to-unlock, focus restore, or `privacySensitive()` behavior.
- Over-tuning `AuthWindowLayoutModifier` could make window transitions feel jumpy.
- An icon-only menu/toggle can regress accessibility if labels and hints are not preserved.

## Security Considerations
- Preserve `privacySensitive()` and current secure-entry behavior.
- Do not log, expose, or transform password/recovery input beyond existing behavior.

## Next Steps
- Automated validation completed: `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""` passed after the final polish changes.
- Follow-on hierarchy work landed under `plans/20260405-1230-unlock-vault-stitch-fidelity-pass/`; independent review approved it with nits and deemed it safe to keep.
- Finish with the still-pending manual unlock-specific visual/accessibility regression pass before formally closing this polish slice.
