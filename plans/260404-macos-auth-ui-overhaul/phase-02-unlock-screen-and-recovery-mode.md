# Phase 2: Unlock Screen + Recovery Mode

## Context Links
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`

## Overview
- **Priority:** P1
- **Status:** Complete (verified 2026-04-04)
- **Description:** Rewrite the locked/unlock experience so it feels like a native macOS auth window, not a sparse hero screen.

## Key Insights
- `UnlockVaultView.swift` is currently a single large file with layout, motion, validation, and action logic all mixed together.
- Touch ID is hidden as a trailing icon inside the password field, which makes the capability easy to miss.
- Recovery mode is toggled by a checkbox and uses a single-line text field for a 12-word phrase.
- “Close Vault” lives at the bottom edge of the window, exactly where the HIG says not to put critical actions.

## Requirements
- Replace the checkbox with a clearly labeled auth-mode switch (`Password` / `Recovery Phrase`).
- Make Touch ID a first-class secondary action with explanatory copy.
- Move vault-closing/switching to a safer top-right menu and mirror it in app commands.
- Improve recovery entry to handle pasted phrases, whitespace, line breaks, and numbered lists.
- Preserve caps-lock warnings, inline errors, reduced-motion behavior, and keyboard-first flow.

## Architecture
- Keep `UnlockVaultView.swift` as the coordinator only.
- Split method-specific UI into `UnlockPasswordSection.swift` and `UnlockRecoverySection.swift`.
- If automatic Touch ID prompting is retained, gate it behind a one-shot per-appearance rule in `VaultClient.swift` or view-local state so the window does not repeatedly nag.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` *(only if one-shot biometric orchestration is added)*

### Files to Create
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`

## Implementation Steps
1. Rebuild the unlock screen inside `AuthSceneScaffold` with a compact header, vault name/path context, and standard action layout.
2. Replace the recovery checkbox with a segmented `Picker` or equivalent native control-group switch.
3. Move Touch ID out of the text field into a visible button row (`Unlock with Touch ID`) and add helper copy for the disabled-but-available state.
4. Replace the single-line mnemonic field with a multiline paste area that normalizes input and shows live recovery-phrase validation feedback.
5. Remove the bottom footer `Close Vault` button; replace it with a top-right secondary menu (`Choose Different Vault…`, `Close Vault…`) and add the matching command in `ZeroPassApp.swift`.
6. Keep inline error handling, caps-lock state, focus restoration, and reduced-motion shake fallback intact.

## Todo List
- [x] Split `UnlockVaultView.swift` into coordinator + method sections
- [x] Add native mode-switch control for password vs recovery
- [x] Promote Touch ID to a visible secondary action
- [x] Replace bottom `Close Vault` action with top/menu command placement
- [x] Add recovery phrase normalization and validation UX

## Success Criteria
- No checkbox remains in the unlock flow.
- Touch ID is visible without having to inspect the password field chrome.
- Recovery entry is usable for real-world pasted phrases, not just perfect single-line input.
- Vault-closing/switching is safer and feels like a normal macOS command.

## Risk Assessment
- **Risk:** Auto-prompted Touch ID becomes intrusive.  
  **Mitigation:** Prefer a visible button first; only auto-prompt once when a stored key exists and there has been no recent failure.
- **Risk:** Phrase normalization rejects valid pasted content.  
  **Mitigation:** Normalize whitespace/punctuation conservatively and validate against the actual expected word count before blocking submit.
- **Risk:** Splitting the file scatters logic.  
  **Mitigation:** Keep unlock state/actions in the coordinator; make child files purely presentational.

## Security Considerations
- Never log or persist normalized mnemonic text.
- Keep error copy generic; do not leak whether the password or recovery phrase failed in a way that exposes vault state.
- Do not auto-trigger biometric prompts after a user explicitly switched to recovery mode.

## Next Steps
- Reuse the same shell and recovery-phrase presentation style across welcome/create/recovery surfaces in Phase 3.
