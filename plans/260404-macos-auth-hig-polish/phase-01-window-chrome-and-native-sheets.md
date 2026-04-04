# Phase 1: Window Chrome + Native Sheets

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/ContentView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

## Overview
- **Priority:** P1
- **Status:** Completed
- **Description:** Remove the most obvious HIG mismatches by using standard window chrome and real macOS sheets instead of custom floating auth overlays.

## Key Insights
- `AuthWindowLayoutModifier.swift` currently hides title text, traffic lights, shadow, and uses a transparent floating window for auth states.
- `UnlockVaultView.swift` compensates with a custom top-right close button, which duplicates native window affordances.
- `WelcomeView.swift` and `UnlockVaultView.swift` present `CreateVaultView` / `OpenVaultSheet` through `authFloatingOverlay(...)`, which behaves like a custom modal card rather than a window sheet.
- `OpenVaultSheet.swift` also opens `NSOpenPanel` with `runModal()`, so the current stack is app-modal on top of a faux modal.

## Requirements
- Keep auth states compact and content-fit, but stop suppressing standard macOS window controls.
- Prefer real `.sheet` presentation for create/open flows.
- Preserve current create/open/replace-vault behavior; this phase changes framing, not business logic.
- Keep one auth window; do not introduce extra helper windows unless a sheet API absolutely requires it.

## Architecture
- Keep `AuthWindowLayoutModifier.swift` responsible for sizing, not for inventing a pseudo-panel appearance.
- Let the auth shell live inside a normal titled window with visible close affordance.
- Use SwiftUI `.sheet(item:)` / `.sheet(isPresented:)` from `WelcomeView.swift` and `UnlockVaultView.swift` instead of `authFloatingOverlay(...)`.
- If practical in-session, present `NSOpenPanel` as a sheet from the current window rather than app-modal.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/WelcomeView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/OpenVaultSheet.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`

### Files to Reuse, Not Rebuild
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift`

## Implementation Steps
1. Change auth-window appearance so auth states use system chrome (visible close control, normal shadow, non-clear window background) while still keeping a compact content size.
2. Stop fixing auth windows at `min == max == ideal`; keep a minimum and ideal size, but allow growth so larger text/accessibility settings do not clip content.
3. Remove the custom close `xmark` control from `UnlockVaultView.swift`; rely on the native window close affordance plus existing commands.
4. Replace `authFloatingOverlay(...)` in `WelcomeView.swift` and `UnlockVaultView.swift` with real sheets for create/open flows.
5. Retune `CreateVaultView.swift` and `OpenVaultSheet.swift` to read as true sheets: standard cancel/default buttons, no faux floating-card framing assumptions, safe width, keyboard cancel/default support.
6. If it fits cleanly, swap `NSOpenPanel.runModal()` to sheet-style presentation for the active window; if that risks scope, leave it as a follow-up note.

## Todo List
- [x] Restore standard auth window chrome
- [x] Allow auth windows to grow beyond the current fixed size
- [x] Remove the custom close button from unlock
- [x] Convert create/open overlays to real sheets
- [x] Retune create/open sheet sizing and button ordering
- [x] Decide whether `NSOpenPanel` sheet presentation fits the session

## Success Criteria
- Auth no longer looks like a custom borderless utility panel.
- Create/open flows use system sheet behavior instead of a fake modal layer.
- Window close, keyboard cancel/default, and focus behavior feel standard on macOS.

## Risk Assessment
- **Risk:** Standard chrome makes the auth shell feel less distinctive.  
  **Mitigation:** Keep visual identity in typography and spacing, not hidden traffic lights.
- **Risk:** Sheet conversion exposes focus regressions.  
  **Mitigation:** Land this phase before the focus polish phase and test keyboard order immediately.

## Security Considerations
- Do not change the underlying create/open/replace-vault logic.
- Do not auto-dismiss sheets on failure; errors must remain visible inside the presented sheet.

## Next Steps
- Once the window and modal behavior are native, tighten control hierarchy and focus paths in Phase 2.
