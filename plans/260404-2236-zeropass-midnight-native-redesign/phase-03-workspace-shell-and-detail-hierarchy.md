# Phase 3: Workspace Shell + Detail Hierarchy

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/SidebarView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemListView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ToastOverlay.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`
- `plans/260404-1215-macos-hig-vault-screen-rewrite/plan.md`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Re-skin the unlocked vault experience around Midnight Native while keeping the existing `NavigationSplitView`, state flow, and item actions intact.

## Key Insights
- The unlocked shell is already cleanly separated into sidebar, list, and detail views; the redesign should improve hierarchy, density, and emphasis rather than rebuild navigation.
- A recent main-screen rewrite already moved search into the toolbar and restructured the detail pane; this phase should build on that work instead of undoing it.
- `ItemDetailView.swift` is still a large file and the most likely place where small helper extraction may be justified.

## Requirements
- Keep the current three-column `NavigationSplitView`.
- Make sidebar, row selection, toolbar, empty states, and detail sections feel more premium and less airy.
- Use graphite/cobalt hierarchy consistently, but avoid flooding the workspace with accent color.
- Preserve selection retention, sort behavior, context menus, quick actions, and edit/create sheet entry points.
- Reduce “rainbow” type styling so item-type accents support scanning instead of dominating the view.

## Architecture
- Keep `MainShellView` as the owner of selection, search text, and editor sheet presentation.
- Rework view styling/layout in place across `SidebarView`, `ItemListView`, and `ItemDetailView`.
- If necessary, extract 1–3 tiny subviews from `ItemDetailView.swift` for summary/actions/metadata sections, but do not create a deep component tree.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/SidebarView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemListView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ToastOverlay.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`

### Files to Create
- None planned.
- Optional: 1–3 focused detail helpers under `Views/Main/` if `ItemDetailView.swift` remains too dense after the redesign.

## Implementation Steps
1. Refresh the shell chrome in `MainShellView`: title/search/new-item rhythm, toolbar emphasis, empty-state polish, and spacing around the three columns.
2. Redesign `SidebarView` and list rows for stronger hierarchy, better selected-state contrast, tighter counts/badges, and less visual dead space.
3. Rework `ItemDetailView` into compact premium sections with clearer contrast between summary, quick actions, primary fields, notes, tags, and metadata.
4. Align `ToastOverlay` and auxiliary affordances with the Midnight Native surface rules.
5. Preserve existing behavior for selection retention, search filtering, edit/delete, copy actions, version history, and hidden sensitive fields.

## Todo List
- [ ] Restyle shell, toolbar, and empty states without changing core navigation
- [ ] Tighten sidebar/list density and selected-state polish
- [ ] Rework detail hierarchy for scan speed and premium utility feel
- [ ] Keep behavior identical for selection, copy, edit, and version history flows

## Success Criteria
- The unlocked workspace feels materially more polished while remaining recognizably native macOS.
- The detail pane is easier to scan and wastes less width.
- Sidebar and list surfaces look intentional instead of default-plus-accent.

## Risk Assessment
- **Risk:** Layout work accidentally regresses selection/search behavior.  
  **Mitigation:** Keep data flow in `MainShellView` unchanged and validate selection retention after every major visual pass.
- **Risk:** Dark styling reduces information density instead of improving it.  
  **Mitigation:** Use contrast, grouping, and row rhythm to tighten the screen rather than layering decorative containers everywhere.

## Security Considerations
- Sensitive fields must stay masked by default.
- Quick actions and destructive actions must remain explicit and readable at darker contrast levels.

## Next Steps
- Apply the same visual discipline to Settings, where forms and warnings need clearer structure but should stay natively macOS.