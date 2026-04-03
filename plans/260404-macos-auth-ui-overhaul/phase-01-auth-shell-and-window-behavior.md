# Phase 1: Auth Shell + Window Behavior

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `plans/260402-macos-swiftui-app/phase-03-authentication-views.md`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Replace the dark hero-shell foundation with a shared, adaptive auth scaffold and window behavior that fits auth content.

## Key Insights
- `UnlockVaultView.swift` hard-forces `.preferredColorScheme(.dark)` and uses dark-only RGB tokens from `ZeroPassTheme.swift`.
- `ZeroPassApp.swift` opens the main window at `1000×680`, which makes `noVault`, `locked`, and `showingRecovery` states feel empty and oversized.
- The current auth views all use centered spacers instead of a reusable native macOS composition pattern.

## Requirements
- Support system/light/dark appearance with semantic text/background colors and system materials.
- Introduce one shared auth scaffold so welcome, unlock, and recovery screens stop diverging.
- Use state-aware window sizing:
  - `.noVault`: ideal `560×420`
  - `.locked`: ideal `560×460`
  - `.showingRecovery`: ideal `640×520`
  - `.unlocked`: ideal `1000×680`, min `800×500`
- Keep auth-specific code modular so no single auth file stays in the current 400+ line range.

## Architecture
- Create `AuthSceneScaffold.swift` for shared header/body/footer structure, top-right secondary actions, and standard spacing.
- Create `AuthWindowLayoutModifier.swift` to access the backing `NSWindow` and update min/ideal size when `vault.state` changes.
- Keep `ZeroPassApp.swift` responsible only for app-level appearance preference; individual auth views must stop overriding color scheme.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

### Files to Create
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`

## Implementation Steps
1. Replace dark-only auth tokens in `ZeroPassTheme.swift` with adaptive auth tokens built from semantic `NSColor`/SwiftUI materials.
2. Build `AuthSceneScaffold` with slots for title, supporting copy, body, footer, and a secondary action area for vault-level commands.
3. Apply `AuthWindowLayoutModifier` from `ContentView.swift` so auth states resize to their ideal content size and unlocked state restores the main window footprint.
4. Change `WindowGroup.defaultSize` in `ZeroPassApp.swift` to an auth-friendly startup size and rely on the modifier to grow when the main shell appears.
5. Verify light mode, dark mode, increase contrast, and reduce transparency before moving on.

## Todo List
- [ ] Add adaptive auth tokens to `ZeroPassTheme.swift`
- [ ] Create shared `AuthSceneScaffold.swift`
- [ ] Create `AuthWindowLayoutModifier.swift`
- [ ] Wire state-aware sizing through `ContentView.swift` and `ZeroPassApp.swift`

## Success Criteria
- Auth states no longer open inside a huge mostly-empty window.
- Forced dark mode is gone from the unlock/auth flow.
- Welcome, unlock, and recovery views all have one shared structural shell.

## Risk Assessment
- **Risk:** Window resizing jitters during auth ↔ unlocked transitions.  
  **Mitigation:** Resize only on real state changes and animate conservatively.
- **Risk:** Materials become unreadable in high-contrast settings.  
  **Mitigation:** Prefer semantic colors first and use materials as accents, not as the only contrast source.

## Security Considerations
- Keep screenshot/privacy behavior separate from general scaffold code so recovery-specific protections remain explicit.
- Do not move auth decisions into the theme layer; the scaffold should stay visual only.

## Next Steps
- Feed the new scaffold and sizing behavior into the unlock screen rewrite in Phase 2.
