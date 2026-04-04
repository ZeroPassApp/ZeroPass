# Phase 01 — P0 core vault containment

## Context links

- `./plan.md`
- `core/vault/item/item.go`
- `core/vault/version/version.go`
- `core/vault/importexport/import.go`
- `core/vault/index/index.go`
- `core/vault/store/store.go`
- `docs/code-standards.md`
- `docs/system-architecture.md`

## Overview

Priority: P0  
Status: pending  
Goal: close the highest-impact local-vault issues: path escape, plaintext index leakage, recovery fail-open behavior, and excessive recovery validation key lifetime.

## Key insights

- `version.Manager` already has `validateItemID()`; `item.Manager` does not.
- 1PUX import preserves external UUIDs, so unsafe foreign IDs can flow straight to disk paths.
- Search currently indexes `joinMapValues(item.Fields, item.CustomFields)`, which pulls secret values into plaintext SQLite while unlocked.
- `UnlockWithRecovery()` currently unlocks first and swallows rotation failure.
- `ValidateRecovery()` decrypts a vault key purely to validate and should zero it immediately.

## Requirements

### Functional

- Reject or regenerate unsafe item IDs before any filesystem path join.
- Keep imports usable; unsafe external IDs should not cause data loss by default.
- Preserve search utility without indexing passwords, TOTP, API secrets, private keys, CVVs, or equivalent secret-bearing custom values.
- Make recovery unlock atomic from the user's perspective: no successful unlock state unless rotation/persistence succeeds.
- Keep `ValidateRecovery()` read-only.

### Non-functional

- No weakening knobs or runtime flags for these protections.
- Existing vaults must remain readable.
- Search-index changes must support one-time rebuild/migration rather than manual repair.
- All temporary decrypted key material must be zeroed promptly.

## Architecture

- **Shared ID validation:** move to one low-level helper reachable from `item`, `version`, and `importexport` without package cycles; prefer an existing low-level package over a new subsystem.
- **Import behavior:** regenerate unsafe or duplicate external IDs during import, rather than rejecting the whole import; preserve source data, not source path semantics.
- **Search allowlist:** index `name`, `type`, `tags`, `custom_keys`, and only a narrow allowlist of non-secret field values if explicitly approved (`username`, `email`, `url` are the likely candidates). Keep `notes` out of the default plaintext index because they are free-form and frequently secret-bearing.
- **Recovery flow:** decrypt to a temporary vault key, rotate/persist recovery material first, only then publish unlocked state and decrypt the index.
- **Validation flow:** use the temporary decrypted key only long enough to verify the recovery material; zero before returning.

## Related code files

### Modify

- `core/vault/item/item.go`
- `core/vault/version/version.go`
- `core/vault/importexport/import.go`
- `core/vault/index/index.go`
- `core/vault/store/store.go`
- `core/vault/item/item_test.go`
- `core/vault/version/version_test.go`
- `core/vault/importexport/importexport_test.go`
- `core/vault/index/index_test.go`
- `core/vault/index/index_coverage_test.go`
- `core/vault/store/store_test.go`
- `packages/cli/cmd/features_test.go` (recovery CLI regression coverage if behavior changes surface there)

### Create

- None expected; prefer existing files unless a tiny shared validator helper is required.

### Delete

- None.

## Implementation steps

1. **Unify item ID validation**
   - Extract/relocate the existing validation logic so `item.Manager`, `version.Manager`, and import paths all use the same rule.
   - Enforce validation at add/get/update/delete/list/read/decrypt boundaries, not just at the write path.
   - Decide import policy once: regenerate unsafe incoming IDs and test the mapping.

2. **Close the import traversal gap**
   - Apply shared ID rules to 1PUX and any other import paths that preserve upstream IDs.
   - Add regression tests for `/`, `\`, `..`, absolute-path-like, and empty/whitespace edge cases.

3. **Shrink plaintext search exposure**
   - Replace `joinMapValues(item.Fields, item.CustomFields)` with an allowlisted indexing strategy.
   - Rebuild the index for existing vaults once the new policy lands.
   - Verify plaintext `index.db`, `index.db-wal`, and `index.db-shm` no longer contain high-sensitivity values or note contents during unlocked operation by default.

4. **Make recovery unlock fail closed**
   - Rework `UnlockWithRecovery()` so rotation/persistence succeeds before the vault is treated as unlocked.
   - On any rotation/persist error: zero the temporary vault key, keep the vault locked, return a real error.

5. **Shorten recovery validation key lifetime**
   - Make `ValidateRecovery()` decrypt only into a short-lived temporary buffer, verify, zero, and return.
   - Add direct tests for zero-side-effect behavior: still locked, no metadata mutation, no index decrypt.

## Security regression tests to add

- `TestManagerRejectsUnsafeItemID`
- `TestImport1PUXRegeneratesUnsafeExternalID`
- `TestVersionHistoryRejectsUnsafeItemID`
- `TestIndexDoesNotStoreSecretFieldValues`
- `TestIndexDoesNotStoreNotesByDefault`
- `TestUnlockWithRecoveryFailsClosedWhenRotationPersistenceFails`
- `TestValidateRecoveryIsReadOnlyAndDoesNotUnlock`

## Verification

- `go test ./core/vault/...`
- `go test ./packages/cli/cmd/...`
- Optional spot checks during implementation: inspect temporary `index.db*` contents in a temp vault inside tests rather than by hand.

## Docs/update implications

- `docs/system-architecture.md`: update indexed search surface and recovery flow semantics.
- `docs/codebase-summary.md`: remove any implication that all field values are searchable.
- `docs/project-roadmap.md`: mark P0 vault hardening completion when merged.

## Success criteria

- No code path can persist or read an item file outside the intended `items/` directory.
- Unsafe imported IDs are neutralized without silently dropping valid imported items.
- Password/TOTP/API secret/private key/CVV values and notes are absent from plaintext index artifacts by default.
- A recovery rotation failure returns an error and leaves the vault locked.
- `ValidateRecovery()` remains side-effect free and test-covered.

## Risk assessment

- **Search usability regression:** smaller allowlist could surprise users; mitigate by documenting what remains searchable and adding focused tests rather than sneaking notes back in.
- **Import compatibility:** regenerated IDs may break assumptions in importer tests; mitigate with deterministic assertions around safe IDs and counts.
- **Recovery flow refactor:** ordering bugs can lock users out if done carelessly; mitigate with narrow, table-driven tests around success/failure paths.

## Security considerations

- Zero temporary vault keys with `crypto.ZeroBytes(...)` and preserve `runtime.KeepAlive(...)` semantics where required.
- Do not add a compatibility fallback that silently re-enables secret value indexing.
- Treat checksum validation and recovery verification as security-critical code with near-complete branch coverage.

## Todo list

- [ ] Centralize item ID validation and apply it everywhere paths are constructed.
- [ ] Neutralize unsafe preserved IDs during import.
- [ ] Replace broad field-value indexing with an allowlisted strategy and rebuild flow.
- [ ] Make recovery unlock fail closed.
- [ ] Tighten `ValidateRecovery()` key lifetime and add regression coverage.

## Next steps

- Unblocks Phase 02 by restoring confidence that local vault state is contained before sync/local-secret work builds on top.
- Enables Phase 04 CLI and bridge work to rely on safer core vault semantics.