# Phase 5: Quick Search + Menu Bar

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/SearchResultRow.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchPanel.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/QuickSearchPanelController.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/MenuBar/MenuBarView.swift`

## Overview
- **Priority:** P2
- **Status:** Pending
- **Description:** Turn the floating quick-search panel and menu bar extra into compact premium utilities that feel aligned with the new workspace and auth styling.

## Key Insights
- `QuickSearchView` already has the right behavior and keyboard shortcuts; it mainly needs layout, density, and shell polish.
- `QuickSearchPanel` and its controller hard-code a 580×400 floating panel, so any significant visual treatment may need a matching size/position pass.
- `MenuBarView` is useful but visually plain; it should feel like a focused command surface, not a squeezed-down copy of the main app.

## Requirements
- Keep current quick-search keyboard shortcuts and expected flows (`Return`, `⌘Return`, `⇧Return`, `Esc`).
- Make the quick-search panel feel premium and native: cleaner search field, tighter results, stronger selected-state contrast, better keyboard hint strip.
- Bring the menu bar extra into the same graphite/cobalt language while preserving compactness.
- Avoid behavioral scope creep; UI-driven search-consistency cleanup is acceptable, but this phase is not a search-engine rewrite.

## Architecture
- Keep `QuickSearchPanelController` responsible for lifetime and positioning.
- Refresh SwiftUI content in `QuickSearchView` and `MenuBarView` first; touch panel sizing/position only if the new layout demands it.
- Reuse the same row/token language from workspace list styling where possible.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/SearchResultRow.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchPanel.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/QuickSearchPanelController.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/MenuBar/MenuBarView.swift`

### Files to Create
- None planned.

## Implementation Steps
1. Redesign `QuickSearchView` around a richer panel shell with tighter padding, clearer result grouping, and better selected-row readability on dark surfaces.
2. Refresh `SearchResultRow` so icon/title/subtitle/type labeling feel denser and more premium.
3. Adjust `QuickSearchPanel`/`QuickSearchPanelController` sizing or placement only if needed to support the new panel composition.
4. Rework `MenuBarView` into a compact command center with clearer lock state, more intentional quick-search emphasis, and better spacing for item actions.
5. Validate that the redesigned floating surfaces still feel fast, keyboard-first, and unmistakably macOS-native.

## Todo List
- [ ] Redesign quick-search panel chrome and row density
- [ ] Improve selected-state and keyboard hint presentation
- [ ] Refresh menu bar extra layout and emphasis
- [ ] Keep search behavior and shortcuts stable unless UI forces tiny consistency fixes

## Success Criteria
- Quick Search feels like a premium spotlight-style utility, not a generic material box.
- Menu bar interactions visually match the rest of the redesigned app.
- Floating panel behavior remains quick and predictable.

## Risk Assessment
- **Risk:** Fancy panel treatment hurts keyboard-first speed.  
  **Mitigation:** Prioritize focus behavior, row density, and predictable shortcuts over visual flourish.
- **Risk:** Panel size changes create awkward screen placement.  
  **Mitigation:** Keep controller changes minimal and validate on standard MacBook screen sizes.

## Security Considerations
- Preserve sensitive copy flows and the current clipboard timeout behavior.
- Do not add persistent previews that expose secrets more broadly than the current design.

## Next Steps
- Finish the remaining edit-heavy surfaces: item editor and the recovery utilities shown after the vault is already open.