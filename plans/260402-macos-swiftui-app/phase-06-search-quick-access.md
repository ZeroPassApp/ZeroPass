# Phase 6: Search & Quick Access

## Context
- Depends on: [Phase 5 — Item CRUD Views](./phase-05-item-crud-views.md)
- **Design Reference**: [UI/UX Guideline §11 Quick Search](./reports/ui-ux-design-guideline.md#11-quick-search-spotlight-style)

## Overview
- **Priority:** P0
- **Status:** Pending
- **Description:** Build the spotlight-style quick search window (⌘K) and integrate full-text search throughout the app. This is a key UX differentiator per PRD.

## Key Insights
- 1Password uses ⌘⇧Space, Raycast uses ⌥Space — ZeroPass can use configurable global hotkey
- Quick search is a separate `Window` scene with `.hiddenTitleBar` style
- FTS5 search in Go backend returns item IDs → fetch details in batch
- Search should be < 50ms per PRD performance requirements
- Debounce keystrokes at 150ms to avoid excessive bridge calls

## Design Specs (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

### Panel Specifications (Raycast-inspired)
- **Window**: `NSPanel`, `.nonactivatingPanel`, `.floating`
- **Size**: 580×400pt (max 10-12 results visible)
- **Position**: Center-top of screen, 200pt from top (like Spotlight)
- **Background**: `.ultraThickMaterial` (opaque frosted glass)
- **Corner radius**: 12pt
- **Shadow**: `.shadow(color: .black.opacity(0.3), radius: 20)`
- **Search field**: 15pt, no border, full width, auto-focus on open

### Result Rows
- **Row height**: 36pt
- **Icon**: 16pt, type-colored (see [Phase 4 color coding](./phase-04-main-ui-layout.md))
- **Title**: 13pt semibold
- **Subtitle**: 11pt `.secondary` (username/URL)
- **Type badge**: 10pt, right-aligned, `.quaternary`
- **Selected**: System accent color background

### Animations
- **Open**: `.spring(response: 0.25, dampingFraction: 0.8)`, scale 0.95→1.0
- **Close**: Fade out 150ms `easeIn`
- **Results stagger**: Each result fades in +30ms after previous (Raycast pattern)

### Footer Keyboard Hints
- **Text**: 10pt `.quaternary`, separator line above
- **Content**: `⏎ Copy password  ⌘⏎ Open  ⇧⏎ Open URL  ⌘C Copy`

### Empty Query State
- Show **last 5 accessed items** with "RECENT" header (10pt uppercase `.tertiary`)

### No Results State
- "No items matching 'query'" + "Create new item" link

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Views/Search/QuickSearchView.swift` | Spotlight-style floating search (SwiftUI content) |
| `macos/ZeroPass/Views/Search/QuickSearchPanel.swift` | **NSPanel wrapper** — `.nonactivatingPanel` + `.floating` for true spotlight behavior |
| `macos/ZeroPass/Views/Search/SearchResultRow.swift` | Compact result row (icon, name, type, subtitle) |
| `macos/ZeroPass/Services/HotkeyService.swift` | Global keyboard shortcut registration |

## Implementation Steps

### QuickSearchView
1. **Use `NSPanel` (NOT SwiftUI `Window` scene)** — SwiftUI `Window` with `.hiddenTitleBar` still shows in Dock/window list and can't do true floating panel behavior
2. `QuickSearchPanel.swift`: subclass `NSPanel` with `styleMask: [.nonactivatingPanel, .titled, .fullSizeContentView]`, `level: .floating`, `isMovableByWindowBackground: true`
3. Host SwiftUI content via `NSHostingView(rootView: QuickSearchView())`
4. Large search field at top (⌘K activated)
5. Results list below (max 10-15 visible)
6. Each result shows: type icon, name, username/URL preview
7. Keyboard navigation: ↑↓ to select, ⏎ to action
8. Actions per result:
   - ⏎ Copy password (for logins/apikeys)
   - ⌘⏎ Open in main window
   - ⌘C Copy username
   - ⇧⏎ Open URL in browser
9. Escape to close
10. Auto-focus search field on open
11. Show "No results" when query returns empty
12. Show recent items when query is empty

### HotkeyService (Global Shortcut)
1. Use `NSEvent.addGlobalMonitorForEvents(matching: .keyDown)` for basic detection
2. Default shortcut: ⌘⇧P (configurable in settings)
3. On trigger: `NSApp.activate(ignoringOtherApps: true)` THEN show/hide `QuickSearchPanel`
4. Recommended: use **HotKey library** (soffes/HotKey) for user-configurable shortcuts — handles key code + modifier registration properly

### In-App Search Enhancement
1. `.searchable()` modifier on NavigationSplitView (Phase 4) already provides basic filtering
2. Enhance: debounce 150ms, call `ZPSearch()` for FTS5 results
3. Highlight matching text in results (if possible)
4. Recent searches history (in-memory only, not persisted)

## Success Criteria
- [ ] ⌘K opens floating quick search panel (580×400pt, center-top, frosted glass)
- [ ] Search results appear within 50ms of typing (debounced 150ms)
- [ ] Results: 36pt rows with type-colored icons, stagger fade-in animation
- [ ] Keyboard navigation (↑↓⏎ Esc) works correctly
- [ ] Copying password from quick search works with auto-clear toast
- [ ] Global hotkey (⌘⇧P default, configurable) triggers quick search from any app
- [ ] Empty query shows last 5 recent items with "RECENT" header
- [ ] No results shows message + "Create new item" link
- [ ] Escape closes panel with fade-out animation (150ms)
- [ ] Footer shows keyboard hint bar (10pt `.quaternary`)
- [ ] Reduced motion: skip stagger animation, use instant display
- [ ] VoiceOver: results navigable, actions announced
