# Phase 4: Settings Redesign

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/GeneralSettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SyncSettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/AboutView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/HotkeyRecorderView.swift`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Rework the preferences experience so it matches Midnight Native while keeping familiar macOS settings patterns and the current bindings intact.

## Key Insights
- Settings are already split sensibly by tab and concern, so this is not a structural rewrite.
- The current forms are functional but visually flat: wide spacing, weak grouping, and little distinction between informational, destructive, and high-value settings.
- Security and sync sections carry the highest UX risk because warnings, export actions, and secret inputs must remain unmistakable.

## Requirements
- Keep the current `SettingsView` container unless the redesign proves a different container is clearly worth the migration cost.
- Make each tab feel denser and more purposeful through better grouping, labels, helper text, and status presentation.
- Bring sheets/popovers such as the hotkey recorder, change-password flow, and recovery mnemonic sheet into the same visual system.
- Preserve current preferences bindings, validation, disabled states, and warnings.
- Ensure export/import and sync preview messaging stay explicit and not visually buried.

## Architecture
- Keep `TabView` and `Form` as the primary structure for settings.
- Restyle each settings view in place, adding only lightweight helper views if repeated row/panel patterns become obvious.
- Reuse the theme/token layer rather than building a parallel settings-only visual system.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/GeneralSettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SyncSettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/AboutView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/HotkeyRecorderView.swift`

### Files to Create
- None planned.
- Optional: one small settings helper for repeated section headers/status cards if repetition becomes obvious.

## Implementation Steps
1. Update the overall settings shell so tabs, padding, and section rhythm match Midnight Native without breaking native familiarity.
2. Redesign `GeneralSettingsView` and `SyncSettingsView` to improve hierarchy for path, hotkey, sync metadata, and status copy.
3. Rework `SecuritySettingsView` so biometrics, auto-lock, clipboard, export/import, change-password, and recovery actions scan cleanly and dangerous actions stay unmistakable.
4. Bring `HotkeyRecorderView`, change-password UI, and recovery mnemonic sheet chrome into the same premium panel language.
5. Refresh `AboutView` so it feels like part of the redesigned app, not a leftover template screen.

## Todo List
- [ ] Tighten the settings shell while keeping native tabbed preferences
- [ ] Improve hierarchy in General, Security, Sync, and About tabs
- [ ] Unify settings-adjacent sheets/popovers with the new visual language
- [ ] Preserve all validation, disabled states, and warning semantics

## Success Criteria
- Settings feel like a first-class part of the redesigned app, not a functional afterthought.
- Security/export/sync sections are easier to scan and harder to misuse.
- The hotkey recorder and recovery/password sheets no longer feel disconnected from the rest of the app.

## Risk Assessment
- **Risk:** Over-styling makes Settings feel less like macOS.  
  **Mitigation:** Keep `TabView` + `Form` semantics and use Midnight Native mainly through grouping, spacing, and tokenized emphasis.
- **Risk:** Important warnings lose prominence inside richer panels.  
  **Mitigation:** Reserve stronger contrast and iconography for dangerous/export/recovery messaging.

## Security Considerations
- Keep API keys in `SecureField`, destructive exports behind explicit confirmation, and recovery regeneration behind current gating.
- Do not reduce clarity around clipboard auto-clear or biometric availability/disabled states.

## Next Steps
- Move into the floating utility surfaces where the Midnight Native direction should feel most premium: quick search and the menu bar extra.