# Phase 04 — P1 CLI, bridge, and export hardening

## Context links

- `./plan.md`
- `packages/cli/cmd/helpers.go`
- `packages/cli/cmd/get.go`
- `packages/cli/cmd/export_cmd.go`
- `packages/cli/cmd/cmd_test.go`
- `packages/cli/cmd/features_test.go`
- `packages/cli/cmd/integration_test.go`
- `apps/macos/ZeroPass/ZeroPass/Bridge/ZPBridge.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`
- `bridge/bridge_extra_test.go`

## Overview

Priority: P1  
Status: pending  
Goal: harden the remaining local secret surfaces that are smaller than core-vault work but still material: CLI clipboard clearing, plaintext export permissions, and Swift/bridge secret lifetime.

## Key insights

- CLI clipboard clearing currently relies on an in-process goroutine and `time.Sleep`, so short-lived commands can exit before the clear happens.
- Plaintext export currently uses `os.Create(...)`, which depends on caller `umask` rather than forcing `0600`.
- `ZPBridge.swift` passes passwords, mnemonics, and vault-key material through `withCString`-style calls without zeroing temporary buffers.
- Swift auth/settings views keep secret strings in `@State`, which can persist longer than necessary across success/failure flows.
- If a hidden helper subcommand is used, command registration/entrypoint files must be owned too.

## Requirements

### Functional

- Clipboard auto-clear must survive the parent CLI command exiting.
- Clipboard clear should only wipe the clipboard if the secret is still present, not clobber newer user content.
- Plaintext export files must be created with explicit owner-only permissions.
- Swift bridge calls handling passwords/mnemonics/vault-key material must use zeroable temporary buffers.
- Secret input state in auth/settings views must be cleared promptly after submit/cancel.
- Secret transport to any detached clipboard helper must avoid argv, env vars, temp files, and logs.

### Non-functional

- Keep the CLI fix pragmatic; no long-running daemon or major IPC design.
- Prefer existing files/commands over new surface area unless a hidden helper mode is the smallest robust option.
- Do not broaden exported file semantics unless required by the security fix.

## Architecture

- **CLI clipboard clear helper:** prefer a hidden helper mode/subcommand in the existing `zp` binary that can outlive the parent process, receive the copied secret only over stdin/anonymous pipe, and clear only if the clipboard still matches. Do not send raw secrets or secret-derived fingerprints through argv, env vars, or disk.
- **Export writer hardening:** switch to explicit `0600` file creation and preferably temp-write + rename for plaintext formats; keep encrypted export behavior aligned.
- **Swift bridge buffer hygiene:** add dedicated zeroing wrappers for sensitive inputs only, rather than rewriting every string bridge call. Treat this as best-effort reduction of extra transient copies, not a promise of full Swift heap scrubbing.
- **View-state hygiene:** clear `@State` secrets in `CreateVaultView`, `UnlockVaultView`, and recovery/security settings paths once the underlying action has completed or been cancelled.

## Related code files

### Modify

- `packages/cli/cmd/helpers.go`
- `packages/cli/cmd/get.go`
- `packages/cli/cmd/export_cmd.go`
- `packages/cli/cmd/root.go`
- `packages/cli/cmd/cmd_test.go`
- `packages/cli/cmd/features_test.go`
- `packages/cli/cmd/integration_test.go`
- `apps/macos/ZeroPass/ZeroPass/Bridge/ZPBridge.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/CreateVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`
- `bridge/bridge_extra_test.go`

### Create

- None preferred; use a hidden helper mode in existing CLI files before introducing a new binary or helper package.

### Delete

- None.

## Implementation steps

1. **Replace goroutine-only clipboard clearing**
   - Add a detached helper mode/subcommand that can still run after `zp get --copy` exits.
   - Pass the clear-after timeout plus the copied secret over stdin/anonymous pipe only; the helper must never expose raw secret material or secret-derived fingerprints through argv/env/disk.
   - The helper only clears if the clipboard still contains the copied secret or matching ownership state.

2. **Harden plaintext export file creation**
   - Replace `os.Create(...)` with explicit `0600` creation semantics.
   - Prefer temp-write + rename for plaintext exports if it can be done without widening scope.

3. **Zero bridge-side temporary secret buffers**
   - Introduce dedicated Swift wrappers for passwords, mnemonics, new passwords, and vault-key base64 inputs.
   - Zero the mutable buffers immediately after the C call returns.
   - Document in tests/comments that this reduces extra transient copies but does not fully zero all Swift-managed storage.

4. **Shorten Swift UI secret lifetime**
   - Clear password/mnemonic `@State` values after success, on cancellation, and when leaving the relevant flows.
   - Keep the change local to the current auth/settings views; do not refactor the entire app state model.

5. **Add regression coverage**
   - Extend CLI tests for permission mode and helper invocation behavior.
   - Extend app tests/manual verification for form-state clearing and bridge wrapper behavior where feasible.

## Security regression tests to add

- `TestExportPlaintextUses0600Permissions`
- `TestClipboardCopySpawnsDetachedClearHelper`
- `TestClipboardHelperDoesNotClearReplacedClipboardContents`
- `ZPBridge zeroes sensitive temporary buffers after call`
- `Auth/settings views clear password and mnemonic state after use`

## Verification

- `go test ./packages/cli/cmd/...`
- `go test ./bridge/...`
- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- Manual verification: run a short-lived `zp get --copy`, exit immediately, and confirm the clipboard still clears on schedule.

## Docs/update implications

- `README.md`: note clipboard clearing is now robust against short-lived CLI invocation.
- `docs/codebase-summary.md`: update export/clipboard behavior and bridge secret-handling notes.
- `docs/system-architecture.md`: mention explicit plaintext export permissions and helper-based clipboard clearing if documented there.

## Success criteria

- Clipboard clearing still happens even when the CLI command exits immediately after copying.
- Clipboard clearing does not erase unrelated content copied later by the user.
- Plaintext export files are created with `0600` permissions regardless of `umask`.
- Swift bridge temporary buffers for secret inputs are zeroed after use and the plan/docs describe this as best-effort copy reduction, not total memory erasure.
- Auth/settings secret strings do not linger in `@State` beyond their immediate flows.

## Risk assessment

- **Cross-platform clipboard behavior:** helper process semantics can vary; mitigate with platform-aware tests, stdin/pipe transport only, and minimal clipboard assumptions.
- **UX regressions in forms:** aggressive field clearing can frustrate retries; mitigate by clearing on success/cancel and intentional view exits, not every validation error.
- **Over-scoping:** avoid turning this into a full secure-text architecture rewrite.

## Security considerations

- Never pass raw secret material or secret-derived fingerprints to helper argv/env/logs/temp files/process titles.
- Keep the clipboard helper idempotent and comparison-based to avoid destructive clearing.
- Treat export hardening as a ship blocker for plaintext export formats, even though encrypted export is lower risk.

## Todo list

- [ ] Replace in-process clipboard clearing with a detached helper path.
- [ ] Force owner-only plaintext export permissions.
- [ ] Zero sensitive temporary buffers in `ZPBridge.swift`.
- [ ] Clear secret `@State` values after use in auth/settings flows.
- [ ] Add CLI/bridge regression coverage and manual clipboard validation.

## Next steps

- Can run in parallel with Phase 05 once Phase 01 is merged.
- Coordinate auth-view edits with Phase 03 to avoid file conflicts in `UnlockVaultView.swift` and related screens.