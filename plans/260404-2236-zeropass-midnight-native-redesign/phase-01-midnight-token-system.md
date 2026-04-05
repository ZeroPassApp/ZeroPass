# Phase 1: Midnight Token System + Shared Chrome

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ToastOverlay.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchPanel.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/SearchResultRow.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Define the graphite/cobalt foundation once so every later screen update can reuse the same surface, text, stroke, and accent decisions.

## Key Insights
- `ZPTheme` exists, but it is narrow and auth-oriented; most unlocked surfaces still style themselves ad hoc.
- The new direction needs a darker, premium baseline without abandoning native SwiftUI/AppKit materials and semantic colors.
- Per-type colors in `VaultItemType+UI.swift` are currently vivid enough to look noisy against a graphite-first UI.

## Requirements
- Keep theming centralized in the existing `ZeroPassTheme.swift` file unless it becomes unreasonably large.
- Introduce semantic roles for background layers, elevated panels, row selection, separators, primary/secondary text, accent, warning, success, and destructive states.
- Support the existing `appearanceMode` preference while making dark mode the best-looking version of the product.
- Define consistent spacing, corner radius, and emphasis rules that later phases can apply without inventing one-off styling.

## Architecture
- Extend `ZPTheme` rather than creating a new design system package.
- Add only tiny reusable helpers if the same card/pill/elevation treatment repeats across 3+ surfaces.
- Keep root color-scheme control in `ZeroPassApp.swift`; screen-level views should consume tokens, not force their own color logic.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ToastOverlay.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/SearchResultRow.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchPanel.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`

### Files to Create
- None planned.
- Optional: one small helper under `Views/Components/` only if repeated card/chip styling becomes too noisy inline.

## Implementation Steps
1. Audit current style calls that bypass `ZPTheme` (`Color.accentColor`, `.quaternary`, hard-coded icon colors, search/menu materials).
2. Expand `ZeroPassTheme.swift` with Midnight Native semantic tokens: graphite layers, cobalt accent roles, subdued borders, selected-row fill, muted copy, panel elevation, and badge/chip treatments.
3. Tune `VaultItemType+UI.swift` so type colors still differentiate categories without overpowering the new darker chrome.
4. Apply shared panel/toast/search-row shell styles where repetition is already obvious, but avoid large-scale file churn before screen-specific phases.
5. Freeze token names before screen work starts so later phases are layout/content passes instead of palette re-litigation.

## Todo List
- [ ] Extend `ZPTheme` for Midnight Native semantic tokens
- [ ] Normalize shared chrome patterns (panel, selection, toast, chip)
- [ ] Reduce type-color noise for graphite surfaces
- [ ] Keep new helpers minimal and justified

## Success Criteria
- One token source can style auth, workspace, settings, search, menu bar, and editor surfaces.
- Graphite/cobalt choices are visible without turning the app into a custom-painted web dashboard.
- Later phases can mostly consume tokens rather than invent new color rules.

## Risk Assessment
- **Risk:** Over-custom styling makes the app feel less native.  
  **Mitigation:** Use semantic macOS colors/materials as the base and apply Midnight Native through role mapping, contrast, and spacing.
- **Risk:** Type accents become visually loud on dark surfaces.  
  **Mitigation:** Desaturate/mute category colors and reserve cobalt for product-level emphasis.

## Security Considerations
- Do not encode security state purely in color; lock/destructive/recovery actions still need explicit labels and symbols.
- Ensure darker surfaces preserve contrast for warnings and destructive confirmations.

## Next Steps
- Apply the finalized tokens to the auth flow first, because that is the narrowest surface set and the cleanest place to validate the direction.