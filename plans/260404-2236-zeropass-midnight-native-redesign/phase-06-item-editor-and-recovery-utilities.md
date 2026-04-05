# Phase 6: Item Editor + Recovery Utilities

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Finish the redesign on the editor and in-app recovery utilities so creation/editing flows feel as premium and deliberate as browsing and unlocking.

## Key Insights
- `ItemEditorView.swift` is one of the biggest single UI files in the app and currently mixes structure, field rendering, password generation, and footer state in one place.
- The editor is functionally strong already; it mostly needs hierarchy, grouping, and reduced dead space.
- Recovery phrase presentation is shared across auth and settings, so the unlocked regeneration sheet needs to align with the auth-side work from Phase 2.

## Requirements
- Make new/edit item flows feel denser, premium, and easier to scan without changing the data model or save behavior.
- Keep secret fields masked appropriately and preserve generator affordances.
- Match the recovery mnemonic sheet and card to the same Midnight Native language used in auth recovery.
- Prefer editing `ItemEditorView.swift` in place; only extract small helpers if the file becomes harder to maintain.

## Architecture
- Keep `ItemEditorView` as the entry point for create/edit sheets.
- Reuse tokenized section/chip/field styles from earlier phases.
- If decomposition is needed, extract at most a few obvious helpers (for example: general section, dynamic fields section, generator popover host), not a new editor framework.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`

### Files to Create
- None planned.
- Optional: 1–2 small editor helpers under `Views/Editors/` if the file remains too large after cleanup.

## Implementation Steps
1. Redesign `ItemEditorView` around clearer general/details/notes/action grouping with less vertical sprawl.
2. Improve the dynamic field rows so secret inputs, generator buttons, remove actions, and custom field creation feel more intentional and visually aligned.
3. Refresh `PasswordStrengthBar` and password-generator popover styling to match the Midnight Native utility aesthetic.
4. Align the recovery mnemonic sheet inside `SecuritySettingsView` with the new recovery card treatment and one-time visibility messaging already used in auth.
5. Keep save/cancel behavior, field bindings, default-field logic, and validation untouched unless a visual change requires a tiny supporting cleanup.

## Todo List
- [ ] Rework item editor hierarchy and spacing
- [ ] Improve dynamic field rows and generator affordances
- [ ] Bring password strength/generator UI into the shared visual language
- [ ] Align in-app recovery regeneration sheet with auth recovery styling

## Success Criteria
- The editor feels premium and efficient instead of long and form-heavy.
- Secret fields, copy-sensitive actions, and password generation remain obvious and safe.
- Recovery regeneration no longer looks like a separate design era from the auth flow.

## Risk Assessment
- **Risk:** Editor cleanup accidentally changes save/default-field behavior.  
  **Mitigation:** Keep logic intact and confine work to structure/styling unless a tiny cleanup is clearly required.
- **Risk:** Smaller spacing hurts readability for long secret values and notes.  
  **Mitigation:** Tighten chrome and grouping, not the actual content legibility.

## Security Considerations
- Secret fields must remain masked by default where they are masked today.
- Recovery phrase views must continue to emphasize one-time visibility and avoid accidental persistence beyond current behavior.

## Next Steps
- Run the full validation matrix so the new styling does not quietly regress tests, accessibility, or the Go-side build health.