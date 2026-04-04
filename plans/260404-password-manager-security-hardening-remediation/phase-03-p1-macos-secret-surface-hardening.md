# Phase 03 — P1 macOS secret surface hardening

## Context links

- `./plan.md`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/ClipboardService.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/QuickSearchPanelController.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`

## Overview

Priority: P1  
Status: pending  
Goal: reduce passive secret exposure on macOS by clearing clipboard and panel state on lock/close, removing selectable secret surfaces, and not preloading decrypted items before the user acknowledges a new recovery phrase.

## Key insights

- `ClipboardService` can clear, but `VaultClient.lock()` and `closeVault()` do not call it.
- Quick search uses a cached `NSPanel`; closing it hides the panel but does not invalidate embedded state.
- `VaultClient.unlockWithRecovery()` currently calls `refreshItems()` before the user confirms they saved the replacement mnemonic.
- `RecoveryPhraseCardView` and secret-bearing parts of `ItemDetailView` use `.textSelection(.enabled)`.
- `SearchResultRow` falls back to `notes`, which is too revealing for a global hotkey surface.
- Quick-search copy actions close the panel but do not guarantee state invalidation on later lock/close events.

## Requirements

### Functional

- Clear clipboard on vault lock and close only if ZeroPass still owns the copied value; never wipe unrelated newer clipboard content.
- Invalidate quick-search UI state on vault lock and close, not just visually hide the panel.
- Delay decrypted item refresh and keychain persistence after recovery unlock until user confirmation.
- Remove text-selection affordances from recovery phrase and secret-bearing item surfaces; preserve explicit copy actions.
- Quick-search result rows must not fall back to note content or other free-form secret-bearing text.
- Keep normal non-secret browsing flows usable.

### Non-functional

- Keep SwiftUI/AppKit changes small; avoid rewriting window/panel architecture.
- Preserve keyboard-driven quick-search UX after reopening the panel.
- Use existing services instead of creating parallel state managers.

## Architecture

- **Lock/close hygiene hook:** `VaultClient` owns the cleanup trigger because it already owns lock/close state transitions.
- **Clipboard ownership:** track the last ZeroPass-owned pasteboard content or generation token so lock/close can clear safely without clobbering unrelated content.
- **Panel invalidation:** add an explicit controller method that both closes and discards cached panel/view state so stale results are not retained.
- **Recovery gate:** `unlockWithRecovery()` may unlock the underlying vault session, but UI should not hydrate items or persist biometric material until `acceptRecoveryPhrase()`.
- **Selectable surface reduction:** default to non-selectable secret-bearing text; keep copy buttons/toasts as the sanctioned path and add an explicit note-copy path where selection is removed.
- **Quick-search display minimization:** show only non-secret summary metadata; never use notes as fallback subtitle text.

## Related code files

### Modify

- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/ClipboardService.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/QuickSearchPanelController.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/SearchResultRow.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/ItemDetailView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`

### Create

- None expected; extend existing tests rather than adding a parallel UI test suite.

### Delete

- None.

## Implementation steps

1. **Add lock/close cleanup**
   - Call clipboard clear from `lock()` and `closeVault()` only when ZeroPass still owns the pasteboard value or generation.
   - Cancel any pending clipboard timer on lock/close.
   - Apply the same cleanup when `replaceVault(with:)` switches to another vault without going through the normal close path.
   - Close and invalidate quick search whenever the vault leaves an unlocked state.

2. **Make quick-search state disposable**
   - Add an explicit `invalidate()` or equivalent to `QuickSearchPanelController`.
   - Ensure reopening the panel recreates a clean `QuickSearchView` with empty query/results state.

3. **Gate recovery-unlock item hydration**
   - Remove `refreshItems()` and biometric/key persistence from the pre-confirmation path.
   - Move item refresh/persistence to `acceptRecoveryPhrase()` after the user confirms they saved the new mnemonic.

4. **Reduce passive text exposure**
   - Remove `.textSelection(.enabled)` from `RecoveryPhraseCardView`.
   - Remove selection from sensitive field rows and notes where secrets may appear.
   - Keep explicit copy controls for legitimate copy workflows.
   - Update `SearchResultRow` so quick-search subtitles never fall back to note text.
   - Add an explicit note-copy affordance if note selection is removed from `ItemDetailView`.

5. **Add focused app regression checks**
   - Extend existing unit tests where feasible.
   - Add a short manual checklist for panel invalidation and clipboard clearing if a fully automated assertion is impractical.

## Security regression tests to add

- `VaultClient lock clears clipboard and quick-search state`
- `VaultClient closeVault clears clipboard and quick-search state`
- `Recovery unlock does not refresh items before confirmation`
- `Recovery acceptance triggers first refresh and optional biometric persistence`
- `Recovery phrase view exposes copy action but no selectable phrase surface`
- `Quick-search subtitles never expose notes`

## Verification

- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- Manual verification: copy a secret, lock the vault, confirm clipboard is empty; open quick search, lock, reopen, confirm stale results are gone.
- Manual verification: copy a ZeroPass secret, then overwrite the clipboard with unrelated content, lock the vault, confirm the unrelated clipboard content is preserved.

## Docs/update implications

- `README.md`: mention clipboard/quick-search cleanup as part of the local security posture.
- `docs/codebase-summary.md`: update recovery unlock flow to reflect confirmation before item hydration.
- `docs/system-architecture.md`: update UI-side recovery flow and clipboard hygiene notes.

## Success criteria

- Clipboard contents are cleared on lock and close only when ZeroPass still owns the copied value.
- Quick search cannot reveal stale results after lock or close.
- Recovery unlock does not surface decrypted items until the user confirms they saved the new mnemonic.
- Recovery phrases and secret-bearing item text are no longer passively selectable.
- Quick-search result rows never reveal note content.
- Existing quick-search usability remains intact after reopen.

## Risk assessment

- **Usability pushback:** removing text selection may annoy power users; mitigate with reliable copy actions, visible feedback, and an explicit note-copy path.
- **Panel lifecycle bugs:** invalidating the panel could break focus/hotkey behavior; mitigate with reopen tests and manual hotkey checks.
- **Recovery flow race conditions:** moving refresh timing could surface state bugs; mitigate with explicit tests around `showingRecovery` → `unlocked`.

## Security considerations

- Treat notes as potentially secret-bearing unless the UI can distinguish them safely.
- Do not leave a hidden but still live quick-search panel in memory after lock/close.
- Avoid duplicating decrypted state in multiple SwiftUI stores during the recovery-confirmation step.

## Todo list

- [ ] Clear clipboard on lock and close.
- [ ] Invalidate quick-search state on lock and close.
- [ ] Delay item refresh/persistence until recovery confirmation.
- [ ] Remove passive text selection from recovery and secret-bearing views.
- [ ] Add macOS regression coverage and manual verification notes.

## Next steps

- Coordinate file ownership with Phase 02 because both touch `VaultClient.swift`.
- Phase 04 can follow on auth-view memory hygiene once this recovery/clipboard work is stable.