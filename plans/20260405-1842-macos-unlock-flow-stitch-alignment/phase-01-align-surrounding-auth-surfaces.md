# Phase 01 — align surrounding auth surfaces

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`

## Overview
- Priority: P1
- Status: completed
- Brief description: Carry the shipped Stitch unlock language across the rest of the auth journey so the app no longer has one polished screen surrounded by mismatched support screens.

## Key Insights
- The locked screen is already the strongest Stitch match; `WelcomeView.swift` still leans hero-style and visually precedes a different product.
- `CreateVaultView.swift` and `OpenVaultSheet.swift` are functionally sound but visually plainer than the locked state.
- `RecoveryPhraseView.swift` and the settings regeneration sheet do not fully share the same hierarchy or seriousness cues.

## Requirements
- Keep welcome, create/open, recovery display, and recovery regeneration visually continuous with the current locked state.
- Preserve native macOS sheet behavior, current folder-picking, focus restoration, and modal routing.
- Preserve recovery readability, copy safety, and one-time visibility messaging.

## Architecture
- Reuse `AuthSceneScaffold.swift` and `RecoveryPhraseCardView.swift`; add only tiny shared helpers if duplication becomes obvious.
- Prefer restyling existing views over introducing new auth containers.

## Related Code Files
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- Create: none
- Delete: none

## Implementation Steps
1. Rebuild `WelcomeView.swift` around the same compact auth hierarchy as the current unlock screen, with action-first copy and quieter footer/meta.
2. Bring `CreateVaultView.swift` and `OpenVaultSheet.swift` into the same panel/sheet language without changing validation, folder-picking, or modal routing.
3. Tighten `RecoveryPhraseView.swift` so the create → recovery → unlock path feels like one guided sequence instead of a separate full-window detour.
4. Reuse the same recovery presentation inside `SecuritySettingsView.swift` so regeneration no longer falls back to an ad-hoc sheet.

## Todo List
- [x] Align welcome with the unlock-family hierarchy
- [x] Restyle create/open sheets to match the auth panel language
- [x] Bring recovery display into the same flow
- [x] Reuse recovery presentation for settings regeneration

## Success Criteria
- Welcome, create/open, locked, and recovery surfaces look like parts of one flow.
- Recovery remains the most readable and safety-forward step in the journey.
- No sheet routing, focus, or folder-picking regressions.

## Risk Assessment
- Risk: over-styling the welcome/recovery screens reintroduces hero-style bloat.  
  Mitigation: keep the layout compact, action-led, and system-first.
- Risk: reusing recovery UI in settings makes the sheet too large.  
  Mitigation: reuse the card and action hierarchy, not necessarily the full auth-window wrapper.

## Security Considerations
- Do not weaken clipboard auto-clear behavior or one-time recovery messaging.
- Do not make recovery phrases easier to capture accidentally.

## Next Steps
- Visual continuity is now in place; next work should focus on any remaining flow/state gaps and additional manual accessibility polish if review finds them.
