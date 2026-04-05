# Phase 01 — rebuild locked-state hierarchy

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- Existing superseded polish plan: `plans/20260405-1116-unlock-vault-redesign-polish/`

## Overview
- Priority: P1
- Status: in progress
- Brief description: Replace the current single hero-card composition with a compact Stitch-like locked-state hierarchy. Implementation landed; formal manual unlock-specific visual/accessibility regression is still pending.

## Key Insights
- `UnlockVaultView` currently places header, vault context, method switcher, fields, CTAs, badges, and helper copy inside one elevated surface.
- `vaultMetaCard` reads like a second content block, while Stitch uses a much smaller selected-vault chip near the top.
- The badges/helper region is visually too heavy because it lives inside the primary card.
- `AuthWindowLayoutModifier` still sizes locked auth like a roomy hero state, not a tighter unlock moment.

## Requirements
- Compact centered auth window with less vertical sprawl.
- Top row title plus subtle vault/options affordance.
- Small vault chip/selected-vault context near the top.
- Inner unlock panel containing segmented control, active input section, main CTA, and Touch ID secondary action.
- Bottom badges row and zero-knowledge helper outside or visually lighter than the main interaction panel.
- No logic or security behavior changes.

## Architecture
- `UnlockVaultView`: split into a lighter outer auth shell with three zones: header/options, compact vault context + inner unlock panel, and lighter footer meta.
- `UnlockPasswordSection` / `UnlockRecoverySection`: tighten spacing and copy rhythm so both fit naturally inside the new inner panel without changing focus or submit behavior.
- `AuthWindowLayoutModifier`: shrink locked and recovery sizing only as much as needed to support the new compact composition.

## Related Code Files
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`

## Implementation Steps
1. Recompose `UnlockVaultView` into a tighter three-zone layout: top header/options row, middle unlock stack (vault chip + inner panel), bottom light footer for badges/helper text.
2. Replace the current full-width `vaultMetaCard` with a compact selected-vault chip that preserves vault name/path context and existing menu actions without reading like a separate card.
3. Move the segmented control, method-specific content, error message, Unlock CTA, and Touch ID action into a dedicated inner panel; remove heavy divider treatment and tighten spacing to match Stitch.
4. Update `UnlockPasswordSection` and `UnlockRecoverySection` so their field chrome, explanatory copy, and vertical rhythm fit the denser panel while preserving secure entry, focus restore, accessibility labels, and submit behavior.
5. Reduce locked/showing-recovery window sizing in `AuthWindowLayoutModifier`, then run a manual regression pass on resize behavior, password/recovery flows, Touch ID, error shake, and VoiceOver labels.

## Todo List
- [x] Recompose the locked-state hierarchy in `UnlockVaultView`
- [x] Convert the vault context block into a compact chip + subtle options affordance
- [x] Tighten password/recovery sections for the new inner panel
- [x] Adjust locked/recovery window sizing to fit the compact composition
- [ ] Manually regression-check auth behaviors and accessibility

## Validation Snapshot
- Changed files:
	- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
	- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
	- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
	- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- Observed UI details now in code: compact selected-vault chip, inner unlock panel, lighter footer/badges, icon-only vault options affordance, tighter password/recovery chrome, smaller locked-state sizing, and locked window title `Unlock Vault`
- Automated validation passed:
	- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=''`
- Independent review verdict: approve with nits; safe to keep.
- Remaining gate: formal manual unlock-specific visual/accessibility regression for password, recovery phrase, Touch ID, window resize, keyboard flow, and VoiceOver labeling.

## Success Criteria
- Locked state reads as a compact, centered unlock experience instead of one large hero card.
- Vault context is visibly smaller and secondary to the unlock action.
- The main interaction area is the inner unlock panel; footer metadata reads lighter and separate.
- Password, recovery, Touch ID, focus, error, and accessibility behaviors remain intact.

## Risk Assessment
- Over-compressing the layout can cause clipping at smaller window heights or larger accessibility text sizes.
- Moving controls across surfaces can accidentally weaken focus order, default-button behavior, or VoiceOver clarity.
- Window downsizing that is too aggressive may make recovery mode feel cramped.

## Security Considerations
- Preserve `privacySensitive()` behavior and existing secure-entry controls.
- Do not change unlock logic, recovery normalization, or biometric flows.

## Next Steps
- Track this as implementation-complete but not formally closed until the unlock-specific manual visual/accessibility regression is run.
- Keep the earlier polish-only plan as historical context; this focused plan is now the source of truth for the deeper Stitch hierarchy pass.
