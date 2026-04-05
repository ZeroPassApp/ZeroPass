# Phase 01 — align item editor modal polish

## Context Links
- User request summary
- `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`

## Overview
- Priority: P2
- Status: in-progress
- Brief description: Small visual hierarchy pass only; preserve existing item editor behavior.
- **Latest update:** Implementation complete. Compile regression fixed (FlowLayout visibility, frame modifier arguments). Automated validation passed (macOS unit + UI tests).

## Key Insights
- Current `ItemEditorView` is functional but the grouped `Form` makes it read like Settings instead of a premium modal.
- Dynamic field rows currently mix editing, generation, and removal affordances in a utilitarian row shape.
- `CreateVaultView.swift` already shows the card/inset treatment that better matches the app’s current premium direction.
- The specific Stitch screen artifact is not present in the workspace, so implementation should use the provided screen reference as the visual target and keep the current logic as the source of truth.

## Requirements
- Strengthen hierarchy: hero header, clearer section separation, tighter spacing, and a more intentional footer action row.
- Keep field ordering, bindings, save/cancel behavior, default-field logic, and secret masking unchanged.
- Improve field-row affordances for generator/remove/add without making the sheet feel busy.
- Optionally upgrade password-strength presentation if it helps the modal feel more polished.

## Architecture
- `ItemEditorView.swift`: replace the generic grouped-form presentation with a modal-like scroll layout using existing `ZPTheme` surfaces and spacing.
- Keep layout helpers private to the same file if needed (header, general section, field row, footer) rather than extracting a mini framework.
- `PasswordStrengthBar.swift`: optional visual parity update only if the current plain summary text feels out of place after the layout refresh.

## Related Code Files
- Modify: `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- Optional: `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift`

## Implementation Steps
1. Rebuild `ItemEditorView` around a modal-style header, scrollable card sections, and a stable footer action row instead of the current grouped `Form` feel.
2. Rework the general/details hierarchy so type, name, favorite, notes, and section labels scan quickly and feel denser without reducing readability.
3. Tighten dynamic field rows with better row chrome, consistent action alignment, subtler destructive affordance, and a clearer generator trigger for secret fields.
4. Refresh the add-field composer and the embedded password generator popover so they read as intentional utilities, not spare controls.
5. If needed, replace the plain password-strength text treatment with `PasswordStrengthBar` plus a short caption while preserving the same score/update logic.
6. Verify new-item defaults, edit-item retention, add/remove field flow, generator insertion, cancel, and save still behave exactly as before.

## Todo List
- [x] Restructure the editor into a premium modal hierarchy
- [x] Refine field rows, add-field composer, and generator affordances
- [x] Decide whether password strength needs the shared bar treatment
- [x] Run build/tests and automated validation
- [ ] Manual visual QA pass of item editor modal (create, edit, field operations, generator)

## Success Criteria
- The editor feels like a premium modal, not a grouped settings form.
- Existing item data flows behave the same before and after the visual refresh.
- Secret-field generation remains obvious, accessible, and visually integrated.

## Risk Assessment
- Swapping out `Form` can accidentally change focus, scrolling, or keyboard behavior.
- Over-tight spacing can hurt readability for long secrets, notes, or custom-field values.
- Making generator/remove affordances too subtle can reduce discoverability or accessibility.
- Pulling in new theme tokens for one sheet can create avoidable style churn.

## Security Considerations
- Keep secret fields masked by default everywhere they are masked today.
- Do not change save timing, validation conditions, or generator insertion behavior.
- Avoid logging or surfacing generated secrets outside the current visible UI path.

## Next Steps
- Recommended validation command: `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- Recommended manual pass: create item, edit existing item, generate secret, add/remove custom field, resize sheet, verify VoiceOver labels on generator/remove actions.
