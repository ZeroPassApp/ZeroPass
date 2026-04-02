# Phase 4: Main UI Layout

## Context
- Depends on: [Phase 3 — Authentication Views](./phase-03-authentication-views.md)
- **Design Reference**: [UI/UX Design Guideline §9 Main Layout](./reports/ui-ux-design-guideline.md#9-main-layout-3-column)

## Overview
- **Priority:** P0
- **Status:** Pending
- **Description:** Build the main 3-column NavigationSplitView layout — the core shell of the app. Sidebar (categories/tags), item list (filterable, sortable), and detail pane (placeholder for Phase 5).

## Key Insights
- `NavigationSplitView` with 3 columns is THE standard for macOS password managers (1Password, KeePass)
- macOS `Table` view gives native sortable columns — use for item list
- `.searchable()` modifier provides native search bar in toolbar
- System accent colors + `.regularMaterial` background = native macOS feel
- Sidebar should show dynamic tag list from vault items

## Design Specs (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

### Window Sizing
- **Default**: 1024×680pt
- **Minimum**: 800×500pt
- **Maximum**: 1800×1200pt (or screen size)
- Use `@SceneStorage` to persist window position, size, column widths, selected category, sort order

### Column Widths
```swift
NavigationSplitView(columnVisibility: $columnVisibility) {
    SidebarView()
        .navigationSplitViewColumnWidth(min: 200, ideal: 220, max: 280)
} content: {
    ItemListView()
        .navigationSplitViewColumnWidth(min: 280, ideal: 350, max: 450)
} detail: {
    DetailView()
        .navigationSplitViewColumnWidth(min: 320, ideal: 400, max: .infinity)
}
.navigationSplitViewStyle(.balanced)
```

### Responsive Behavior
- Window < 1000pt → Hide detail pane (show on selection as overlay)
- Window < 700pt → Hide sidebar (toggle via button)

### Sidebar Layout Spec
- **Background**: `.regularMaterial` (translucent)
- **Row height**: 28pt
- **Icon size**: 14pt SF Symbols
- **Text**: 13pt `.body`
- **Count badge**: 11pt, right-aligned, `.secondary` color
- **Selected row**: System accent color background, `.primary` text
- **Section header**: 10pt, `.tertiary`, uppercase
- **Bottom status**: 10pt, `.quaternary` (item count + last sync)

### Item Type Color Coding
| Type | Color | SF Symbol |
|------|-------|-----------|
| Login | `.blue` | `key.fill` |
| API Key | `.purple` | `wrench.and.screwdriver.fill` |
| SSH Key | `.green` | `terminal.fill` |
| Secure Note | `.yellow` | `note.text` |
| Credit Card | `.orange` | `creditcard.fill` |
| Identity | `.teal` | `person.crop.circle.fill` |
| Passkey | `.indigo` | `person.badge.key.fill` |
| Custom | `.gray` | `square.grid.2x2.fill` |

### Table View Spec
- **Row height**: 24pt (compact macOS density)
- **Column headers**: 11pt, `.secondary`, clickable for sort
- **Name column**: Icon (12pt) + text (13pt), min 150pt
- **Type column**: Text only (11pt), width 80pt
- **Username column**: Truncated with `...`, 13pt, width 120pt
- **Updated column**: Relative date (11pt), right-aligned, width 80pt
- **Multi-select**: ⌘+click for multiple, ⇧+click for range

### macOS Menu Bar (REQUIRED)
Full menu structure defined in [Guideline §15](./reports/ui-ux-design-guideline.md#15-keyboard-shortcuts). Must implement:
- **File**: New Item (⌘N), Import, Export, Close Window
- **Edit**: Undo/Redo, Cut/Copy/Paste, Find (⌘F)
- **View**: Toggle Sidebar (⌘⌃S), Sort options, Full Screen
- **Item**: Edit, Copy Password (⌘⇧C), Copy Username, Open URL, Delete
- **Window**: Minimize, Zoom, Bring All to Front

### Empty State
- Use `ContentUnavailableView` (macOS 14+)
- Icon: 48pt SF Symbol, `.tertiary`
- Message: 13pt `.secondary`
- Action button: Accent color, `.borderedProminent`

### Context Menu (Right-Click on Table Row)
```
┌───────────────────────────┐
│ Copy Password             │
│ Copy Username             │
│ ───────────────────────── │
│ Open URL in Browser       │
│ ───────────────────────── │
│ Edit                      │
│ Duplicate                 │
│ Toggle Favorite           │
│ ───────────────────────── │
│ Delete (red)              │
└───────────────────────────┘
```

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Views/Main/MainView.swift` | Root NavigationSplitView (3-column) |
| `macos/ZeroPass/Views/Sidebar/SidebarView.swift` | Categories + tags sidebar |
| `macos/ZeroPass/Views/Sidebar/SidebarCategory.swift` | Category enum with icons |
| `macos/ZeroPass/Views/ItemList/ItemListView.swift` | Filterable/sortable item table |
| `macos/ZeroPass/Views/ItemList/ItemRowView.swift` | Compact row for list mode |
| `macos/ZeroPass/Views/ItemDetail/ItemDetailPlaceholder.swift` | "Select an item" empty state |
| `macos/ZeroPass/Views/Components/ToolbarView.swift` | Main toolbar (add, lock, sync) |
| `macos/ZeroPass/Views/Components/EmptyStateView.swift` | "No items" empty state |

## Implementation Steps

### MainView (NavigationSplitView)
1. 3-column layout: sidebar → content → detail
2. `@State` for selectedCategory, selectedItemID, columnVisibility
3. `.searchable()` bound to VaultManager.searchText
4. `.navigationTitle("ZeroPass")`
5. `.toolbar` with MainToolbar

### SidebarView
1. Section "Categories":
   - All Items (tray.full) — with item count badge
   - Favorites (star.fill)
   - Logins (key.fill)
   - API Keys (wrench.and.screwdriver)
   - SSH Keys (terminal)
   - Secure Notes (note.text)
   - Credit Cards (creditcard)
   - Identities (person.crop.circle)
   - Passkeys (person.badge.key)
2. Section "Tags" (dynamic from vault):
   - Iterate `VaultManager.allTags`
   - Each tag as selectable item with `#` prefix
3. Bottom: vault status indicator (items count, last sync time)

### ItemListView (Table)
1. Native macOS `Table` with sortable columns:
   - Name (primary, with type icon)
   - Type
   - Username (from fields["username"])
   - Updated (relative date)
2. `selection: $selectedItemID` for single selection
3. `.contextMenu` for right-click actions:
   - Copy Password / Copy Username
   - Open URL in browser
   - Edit
   - View Version History
   - Delete (with confirmation)
4. Filter by selectedCategory from sidebar
5. Filter by searchText from toolbar
6. Empty state when no items match filter

### ToolbarView
1. Leading: sidebar toggle (`.toolbar { ... }`)
2. Center: search field (via `.searchable`)
3. Trailing:
   - "+" Add Item button (⌘N)
   - Lock button (⌘⇧L)
   - Sync button (if sync configured)
4. Custom menu commands:
   - Vault menu: Lock, Unlock, Quick Search
   - Item menu: New, Edit, Delete, Copy Password

## Success Criteria
- [ ] 3-column layout renders correctly on various window sizes
- [ ] Default window size 1024×680pt, minimum 800×500pt
- [ ] Column widths match guideline specs and persist between launches
- [ ] Responsive: detail hides < 1000pt, sidebar hides < 700pt
- [ ] Sidebar categories filter item list with type-colored icons
- [ ] Tags section populated dynamically, sorted by count descending
- [ ] Table rows 24pt height with sortable columns (name, type, username, updated)
- [ ] Table supports multi-select (⌘+click, ⇧+click range)
- [ ] Search filters items in real-time (debounced 150ms)
- [ ] Context menu on table rows with Copy Password/Username, Edit, Delete
- [ ] Toolbar buttons functional (add → sheet, lock → UnlockView)
- [ ] Full macOS menu bar implemented (File, Edit, View, Item, Window, Help)
- [ ] All keyboard shortcuts working (⌘N, ⌘L, ⌘F, ⌘⇧C, ⌘⌃S, etc.)
- [ ] Keyboard navigation: ↑↓ in list, Tab between columns, ⏎ to open
- [ ] Empty states use `ContentUnavailableView` with SF Symbol + action button
- [ ] Support drag-and-drop items between categories/tags
- [ ] Window position + size persisted with `@SceneStorage`

## Critical Notes

> **Auto-Lock State Observation:** MainView MUST observe `VaultManager.isLocked` and immediately transition to the UnlockView overlay when auto-lock fires (either Swift-side `AutoLockService` or Go-side timer). Without this, the user sees stale decrypted data in the UI after the vault key has been zeroed — any subsequent operation will crash or return errors. Use `.onChange(of: vaultManager.isLocked)` to trigger navigation back to UnlockView.

## macOS Design Patterns
- Use SF Symbols for all icons (see [Guideline §6 Iconography](./reports/ui-ux-design-guideline.md#6-iconography) for full symbol map)
- Semantic colors: `.primary`, `.secondary`, `.accentColor` — NEVER hard-code RGB
- `.regularMaterial` for sidebar background
- `.contentUnavailableView` for empty states (macOS 14+)
- Respect user's system appearance (dark/light auto)
- **Animations**: Sidebar toggle 250ms `easeInOut`, navigation transitions 200ms `easeInOut`
- **Spacing**: 8pt grid system — see [Guideline Appendix B](./reports/ui-ux-design-guideline.md#appendix-b-design-token-summary)
- **Drag-and-drop**: Support dragging items between categories/tags (highlight drop zone)
- **VoiceOver**: Label sidebar sections, table columns, and all interactive elements
- **Reduced motion**: Check `@Environment(\.accessibilityReduceMotion)` for sidebar toggle animation
