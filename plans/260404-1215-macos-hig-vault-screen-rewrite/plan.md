---
title: "Direct macOS HIG rewrite for main vault screen"
description: "Small implementation plan for a native macOS SwiftUI rewrite of the unlocked vault screen."
status: complete
priority: P2
effort: 6h
actual: 6h
branch: main
tags: [macos, swiftui, hig, vault-screen]
created: 2026-04-04
completed: 2026-04-04
---

## Goal

- Keep the current three-column `NavigationSplitView`.
- Make the main window feel more native macOS: quieter sidebar, lighter toolbar, search in the right place, and a detail pane organized for scan speed.

## Current findings

- `MainShellView.swift` already owns selection, sheets, and split view layout; it is the right place to own window-level search state.
- `ItemListView.swift` keeps search local and uses `.searchable(..., placement: .sidebar)`, which reads like sidebar filtering rather than main-content search.
- `SidebarView.swift` applies persistent accent colors plus a yellow favorites row, creating more chroma than a standard macOS sidebar.
- `ItemDetailView.swift` mixes header, actions, fields, tags, metadata, toast, version sheet, and layout helpers in one large file.

## Phase breakdown

1. **Shell + search pass** — lift search state to `MainShellView`, move search to a trailing toolbar search field, reduce toolbar to search + New Item, and leave Lock/Refresh in menu commands.
2. **Sidebar + list polish** — keep the three-column layout, tone down sidebar colors, preserve counts and sort/context menus, and tighten row hierarchy.
3. **Detail pane rewrite** — rebuild `ItemDetailView` around summary, quick actions, credentials, notes, tags, and metadata; extract only tiny helpers if the file grows further.
4. **Validation** — verify keyboard flow, empty states, selection retention after edit/delete, and macOS build/test health.

## Target files

- `apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/SidebarView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemListView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- Optional: 1–3 small detail subviews only if needed to keep file size and responsibilities sane.

## Out of scope

- `VaultClient` data/API changes
- `QuickSearchView` / `Command-K` redesign
- `ItemEditorView` rewrite
- sync, settings, or auth flow changes

## Main risks

- Two search concepts (`Command-K` global quick search vs in-window filter) can blur. Keep quick search unchanged; toolbar search filters the visible list only.
- Field layouts vary by `VaultItemType`. Group common credential keys first, then fall back to an “Additional Fields” bucket instead of a type-by-type redesign.
- The detail view is already oversized. If the rewrite grows, extract only the summary/actions/section helpers needed to stay maintainable.

## Success check ✓ PASSED

- ✓ Main screen still uses a three-column `NavigationSplitView`.
- ✓ Toolbar feels lighter and more native.
- ✓ Search is easier to find and semantically correct.
- ✓ Sidebar reads quieter.
- ✓ Detail pane is faster to scan without changing vault behavior.

## Completion summary

**Implementation (Complete 2026-04-04)**
- Search moved into toolbar with `.searchable` placement at `MainShellView` level
- Toolbar simplified: New Item + search only; Lock/Refresh moved to menu commands
- Sidebar toned down: neutral icons, reduced accent coloring, preserved sections and counts
- Detail pane rebuilt: summary → quick actions → credentials → notes → tags → metadata sections
- External selection synced into visible list scope
- `VaultItemSearchMatcher` coverage tests added

**Validation (Passed 15/15)**
```
xcodebuild test \
  -project apps/macos/ZeroPass/ZeroPass.xcodeproj \
  -scheme ZeroPass \
  -destination 'platform=macOS,arch=arm64' \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""
✓ All 15 tests passed
```