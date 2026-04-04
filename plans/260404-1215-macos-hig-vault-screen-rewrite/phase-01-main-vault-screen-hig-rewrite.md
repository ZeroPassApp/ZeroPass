# Phase 01 — Main vault screen HIG rewrite

## Context links

- Overview: `plans/260404-1215-macos-hig-vault-screen-rewrite/plan.md`
- Current views:
  - `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Main/SidebarView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemListView.swift`
  - `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- Supporting references:
  - `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchView.swift`
  - `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

## Overview

- Priority: P2
- Status: ✓ complete
- Description: Rewrite only the unlocked main vault screen to better match native macOS SwiftUI/HIG patterns while preserving current vault behavior.
- Completed: 2026-04-04
- Test result: 15/15 passed

## Key insights

- The app already has global quick search via `Command-K`; the main window does not need another global search concept.
- `MainShellView` is the correct owner for shared window state like toolbar search and modal presentation.
- The biggest UX gain will come from the detail pane, not from changing data flow or vault services.
- The safest scope is visual and structural: keep item CRUD/version/copy behavior intact.

## Requirements

### Functional

- Keep the three-column `NavigationSplitView` shell.
- Simplify toolbar actions.
- Improve search placement for the main vault screen.
- Reduce sidebar chroma without losing category clarity.
- Restructure the detail pane into summary, quick actions, credentials, notes, tags, and metadata.

### Non-functional

- No bridge, crypto, or persistence changes.
- Preserve keyboard shortcuts, context menus, edit/delete/version flows, and current item selection behavior.
- Prefer standard SwiftUI components and default macOS spacing/materials over custom chrome.

## Architecture

- `MainShellView`
  - owns `selectedCategory`, `showingNewItem`, `editingItem`, and new window-level `searchText`
  - hosts the toolbar search field and the single prominent add action
- `ItemListView`
  - receives `searchText` from the shell
  - stays responsible for list rendering, sort state, context menus, and empty states
- `SidebarView`
  - stays selection-driven, but moves to mostly monochrome symbols and lighter emphasis
- `ItemDetailView`
  - becomes a scan-friendly, sectioned reader view
  - keeps reveal/copy/delete/version behaviors but reorganizes them into clearer groups

## Related code files

### Modify

- `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/SidebarView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemListView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`

### Optional create

- Small helper views under `apps/macos/ZeroPass/ZeroPass/Views/Main/` only if needed to keep `ItemDetailView.swift` readable and below the file-size guideline.

## Implementation steps

1. Lift list filtering state into `MainShellView` and place `.searchable` at the shell/toolbar level with a descriptive prompt like “Search vault items”.
2. Reduce the main toolbar to the minimum useful set:
   - keep `New Item`
   - remove `Refresh` and `Lock` from the main toolbar because they already exist in app commands
3. Tone down the sidebar:
   - remove per-row accent coloring except where truly needed
   - keep badges/counts and section structure
4. Keep the list dense and scannable:
   - preserve sort menu and context menu actions
   - ensure search filters the currently selected category
5. Rewrite the detail pane in this order:
   - summary header
   - quick actions strip (`Copy`, `Open URL`, `Edit`, `Versions`, `Delete` as appropriate)
   - credentials/common fields
   - notes
   - tags
   - metadata
   - fallback additional fields section for anything not in the primary credential group
6. Validate build behavior and UX details:
   - selection survives refresh/edit/delete as expected
   - keyboard shortcuts still work
   - empty/detail-unselected states remain clear

## Todo list ✓ COMPLETE

- [x] Move search ownership to `MainShellView`
- [x] Simplify toolbar actions
- [x] Reduce sidebar chroma
- [x] Keep list behavior stable after search move
- [x] Rebuild detail pane sections
- [x] Run macOS build/tests and smoke-check keyboard flow
- [x] Add `VaultItemSearchMatcher` coverage tests

## Success criteria ✓ MET

- ✓ The main vault screen still feels familiar but looks more like a native macOS app.
- ✓ Search is visually and semantically easier to discover.
- ✓ The toolbar no longer competes with the content.
- ✓ Sidebar colors stop drawing attention away from selection and content.
- ✓ The detail pane answers "what is this?" and "what can I do next?" in the first screenful.

## Risk assessment

- Search state move can accidentally change filtering semantics.
  - Mitigation: keep filtering local to the visible list/category; do not call `vault.searchItems` for this pass.
- Detail grouping may not fit every item type cleanly.
  - Mitigation: special-case only the obvious credential keys and route the rest to “Additional Fields”.
- Shrinking toolbar actions may hurt discoverability for existing users.
  - Mitigation: keep actions available in menu commands and quick actions inside the detail pane.

## Security considerations

- Do not broaden when sensitive fields are revealed or copied.
- Keep existing clipboard auto-clear behavior and masked-field handling unchanged.
- Avoid logging or surfacing secret values during the rewrite.

## Implementation notes

**Changes made:**
- `MainShellView.swift`: Added window-level search state binding, placed `.searchable` at shell level
- `ItemListView.swift`: Receives search text from parent, filters displayed items accordingly
- `SidebarView.swift`: Toned down row accent coloring, kept section structure and counts
- `ItemDetailView.swift`: Reorganized around scan-friendly sections with summary, actions, credentials, notes, tags, metadata
- Added: `VaultItemSearchMatcher` tests for search coverage validation

**Build & test validation:**
- Passed: 15/15 tests on macOS arm64 build
- Search filtering works correctly with category selection
- Selection retained after edit/delete operations
- Keyboard shortcuts functioning normally
- Empty states display as expected

## Next steps

- Monitor field for any edge cases with unusual item types or large credential sets.
- Consider extracting detail pane section helpers if file size grows further (currently below 200 line guideline).
- Further UX polish is out of scope for this phase.