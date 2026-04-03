# Phase 3: Related Auth Screen Alignment

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Bring the rest of the auth-family screens into the same native macOS shell so the overhaul is complete rather than unlock-only.

## Key Insights
- `WelcomeView.swift` uses the same center-floating hero composition as the unlock screen.
- `RecoveryPhraseView.swift` is better than the unlock screen functionally, but it still uses a sparse full-window presentation rather than a compact guided step.
- `SecuritySettingsView.swift` shows regenerated recovery keys in a one-off sheet, which will look inconsistent unless it reuses the improved recovery presentation.
- `CreateVaultView.swift` currently remains a plain form sheet and should at least inherit the same copy hierarchy and spacing rules.

## Requirements
- Welcome, recovery display, create-vault, and recovery-regeneration surfaces should share one visual language.
- Recovery phrase display should feel like a secure document step: easy to scan, easy to copy intentionally, hard to miss.
- Button hierarchy should be consistent: one primary action, secondary contextual actions, destructive actions kept separate.
- Keep the changes scoped; reuse shared components instead of creating multiple bespoke auth shells.

## Architecture
- Reuse `AuthSceneScaffold.swift` for `WelcomeView.swift` and `RecoveryPhraseView.swift`.
- Create `RecoveryPhraseCardView.swift` so the main recovery screen and the settings regeneration sheet share the same numbered-word presentation.
- Leave `CreateVaultView.swift` as a sheet, but align its spacing, copy, validation summary, and action order with the new auth shell.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`

### Files to Create
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseCardView.swift`

## Implementation Steps
1. Rebuild `WelcomeView.swift` as a compact launch/auth screen with grouped create/open actions, clearer supporting copy, and error placement near the actions instead of near the bottom edge.
2. Update `RecoveryPhraseView.swift` to use the shared scaffold and a more readable numbered layout (for example, 3 columns × 4 rows) with copy/confirm actions that follow the new hierarchy.
3. Reuse `RecoveryPhraseCardView.swift` in `SecuritySettingsView.swift` so regenerated recovery keys do not fall back to a different ad-hoc sheet design.
4. Align `CreateVaultView.swift` with the new auth copy, spacing, and validation summaries so the create → recovery → unlock journey feels continuous.

## Todo List
- [ ] Align `WelcomeView.swift` with the shared auth shell
- [ ] Rebuild `RecoveryPhraseView.swift` around a reusable phrase card
- [ ] Reuse the recovery phrase component in `SecuritySettingsView.swift`
- [ ] Tidy `CreateVaultView.swift` so it matches the new auth-family hierarchy

## Success Criteria
- Welcome, unlock, recovery display, and recovery regeneration all look like parts of the same app.
- Recovery phrases are easier to scan and confirm.
- The overhaul does not stop at the locked screen and leave adjacent auth surfaces behind.

## Risk Assessment
- **Risk:** Welcome and recovery screens become over-designed again.  
  **Mitigation:** Keep them grouped, compact, and system-first; avoid bringing back heavy hero styling.
- **Risk:** Reusing the recovery card in settings makes the sheet too large.  
  **Mitigation:** Use the same card with a sheet-specific wrapper rather than duplicating all surrounding layout.

## Security Considerations
- Recovery display/re-generation must remain explicit about “shown once” behavior.
- Copy actions should continue honoring clipboard auto-clear behavior.
- Avoid adding decorative copy or visuals that dilute the seriousness of the recovery step.

## Next Steps
- Run focused QA on appearance, sizing, Touch ID, and command placement; then sync the docs in Phase 4.
