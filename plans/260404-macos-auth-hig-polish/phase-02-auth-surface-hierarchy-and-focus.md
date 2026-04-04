# Phase 2: Auth Surface Hierarchy + Focus

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`

## Overview
- **Priority:** P1
- **Status:** Completed
- **Description:** Simplify visual hierarchy and make keyboard-first interaction paths feel unmistakably macOS-native without changing auth behavior.

## Key Insights
- The scaffold is already shared, but it still reads as a custom floating card: large shadow, clear background, decorative icon tile, generous padding.
- `WelcomeView.swift` still leans toward brand/marketing copy instead of a task-first launch screen.
- `UnlockVaultView.swift` mixes segmented auth mode, accessory actions, error state, biometrics, and primary action in one vertical stack; functionally sound, but the hierarchy can be calmer.
- `CreateVaultView.swift` and `RecoveryPhraseView.swift` are closer to native than before, but their spacing and copy should align more tightly with a macOS sheet/document step than a stylized overlay card.

## Requirements
- One primary action per surface; secondary actions grouped and visually quieter.
- Task-first copy: short titles, concise supporting text, platform-standard button labels where possible.
- Strong keyboard path: first field focused, predictable Tab order, Return activates the default action, Esc cancels sheets.
- Keep existing password/recovery/biometric/recovery-display flows intact.

## Architecture
- Keep `UnlockVaultView.swift` as the coordinator for async actions and errors.
- Keep `UnlockPasswordSection.swift` and `UnlockRecoverySection.swift` mostly presentational; only move focus helpers if needed.
- Reuse `RecoveryPhraseCardView.swift`; do not create a second recovery presentation component.
- Keep `ZPTheme` changes small and semantic; prefer removing decoration over adding new tokens.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`

## Implementation Steps
1. Tone down `AuthSceneScaffold.swift`: reduce card-like theatrics, keep semantic backgrounds, and bias layout toward top-aligned grouped content instead of “floating hero” composition.
2. Reword `WelcomeView.swift` toward action-first copy; keep “Create Vault” primary, “Open Existing Vault” secondary, and move version text to the quietest visual tier.
3. Keep the unlock segmented control, but simplify the accessory/action area so vault actions are discoverable without competing with the main task.
4. Keep Touch ID visible as a secondary button, but ensure helper text is subordinate and does not visually crowd the primary password action.
5. Tighten `CreateVaultView.swift` button ordering, spacing, and field grouping to look like a native sheet; make folder selection, password, and confirmation feel like one form, not stacked cards.
6. Keep `RecoveryPhraseView.swift` explicit and serious: concise warning copy, obvious primary confirmation, and no decorative extras beyond the shared recovery grid.
7. If focus bugs remain, centralize focus triggers so returning from open/create sheets reliably lands on the expected field.

## Todo List
- [x] Reduce custom floating-card styling in the shared scaffold
- [x] Make the welcome screen more task-first and less promotional
- [x] Calm the unlock action hierarchy without changing behavior
- [x] Retune create-vault spacing and grouping for a real sheet
- [x] Tighten recovery-screen copy and action emphasis
- [x] Fix any remaining sheet-return focus problems

## Success Criteria
- Welcome, unlock, create, and recovery surfaces feel like parts of the same Mac app.
- The primary action is visually obvious on every auth surface.
- Keyboard users can move through auth without ambiguous focus jumps.

## Risk Assessment
- **Risk:** Over-correcting toward austerity removes too much helpful context.  
  **Mitigation:** Keep one line of support copy per surface; remove decoration before removing guidance.
- **Risk:** Focus fixes become view-state hacks.  
  **Mitigation:** Prefer one coordinator-owned focus request path over multiple ad-hoc `DispatchQueue.main.async` calls.

## Security Considerations
- Keep auth error copy generic and non-leaky.
- Do not weaken recovery warnings or the “shown once” seriousness of the recovery phrase step.

## Next Steps
- After hierarchy/focus polish lands, run the full accessibility and regression sweep in Phase 3.
