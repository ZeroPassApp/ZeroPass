# Phase 02 — P0 sync auth and secret storage

## Context links

- `./plan.md`
- `services/syncserver/main.go`
- `core/sync/server/sync_handler.go`
- `core/sync/client/sync_client.go`
- `bridge/sync_api.go`
- `bridge/vault_api.go`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/KeychainService.swift`
- `.gitignore`
- `docs/deployment-guide.md`

## Overview

Priority: P0  
Status: pending  
Goal: set a minimum sane sync baseline: no anonymous-by-accident server startup, no wildcard CORS by default, and no plaintext local storage of sync API keys. This phase contains both a **local ship blocker subset** (plaintext local secret removal) and a **sync-preview subset** (server startup/CORS defaults).

## Key insights

- `services/syncserver/main.go` currently treats empty `--api-key` as no auth.
- `core/sync/server/sync_handler.go` always emits `Access-Control-Allow-Origin: *`, even though the shipped client is native.
- `bridge/sync_api.go` persists `sync.json`, and `bridge/vault_api.go` defines `api_key` inside that file.
- `VaultClient.swift` also stores `syncAPIKey` in `UserDefaults`, leaving a second plaintext copy that is not cleared on disable.
- Swift UI/config compatibility also flows through `ZPBridge.swift` and `SyncSettingsView.swift`, so the phase must own those files too.
- `.gitignore` ignores `vault.json` and `items/` but not `sync.json`.

## Requirements

### Functional

- Server startup must require explicit auth by default.
- Unsafe dev/no-auth mode must be opt-in and obviously unsafe.
- Default sync server responses must not emit wildcard browser CORS.
- `sync.json` must stop storing API keys.
- macOS app must store the sync API key in Keychain only and clear legacy plaintext copies.
- Disabling sync must clear both on-disk config and secure-secret references.
- Swift-side sync settings and bridge models must stay compatible after `api_key` is removed from persisted config and preview wording remains honest.

### Non-functional

- Keep sync labeled preview after this phase; this is baseline hardening, not a production-sync claim.
- Preserve existing device/server URL/last-sync persistence.
- Avoid new dependencies if `KeychainService.swift` can be reused.

## Architecture

- **Auth gating:** replace implicit empty-key behavior with explicit `--unsafe-no-auth` (name can vary, but it must read as dangerous).
- **CORS posture:** disable CORS by default; only add allowlisted origins if a real browser client exists later.
- **Secret storage split:** keep non-secret sync metadata in `sync.json`; store API key in Keychain keyed by a stable per-vault namespace (vault path/account/service), not a mutable server URL.
- **Migration:** on first load after upgrade, move any `UserDefaults`/`sync.json` API key into Keychain, then erase plaintext copies.
- **Gate split:** treat local-secret-storage cleanup as a local/offline ship blocker; treat server auth/CORS defaults as required before any stronger sync-preview posture claim.

## Related code files

### Modify

- `services/syncserver/main.go`
- `services/syncserver/main_test.go`
- `services/syncserver/server_test.go`
- `core/sync/server/sync_handler.go`
- `core/sync/server/sync_handler_test.go`
- `bridge/sync_api.go`
- `bridge/vault_api.go`
- `bridge/bridge_extra_test.go`
- `apps/macos/ZeroPass/ZeroPass/Bridge/ZPBridge.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`
- `apps/macos/ZeroPass/ZeroPass/Services/KeychainService.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/SyncSettingsView.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`
- `.gitignore`

### Create

- None expected; keep the migration and keychain logic inside current files unless a tiny helper method is unavoidable.

### Delete

- None.

## Implementation steps

1. **Require explicit server auth**
   - Change startup validation so empty API key is rejected unless an explicit unsafe/dev flag is supplied.
   - Update help text, tests, and smoke expectations accordingly.

2. **Tighten default CORS**
   - Remove `Access-Control-Allow-Origin: *` as the default.
   - Keep native-client traffic working without browser headers.
   - If future web use is needed, make origin allowlisting explicit instead of implicit.

3. **Stop writing API keys to disk**
   - Remove `api_key` from persisted `sync.json`.
   - Keep only `server_url`, `device_id`, and `last_sync_time` in bridge-managed config.
   - Update `ZPBridge.SyncConfig` and UI decode/save paths so Swift still loads metadata-only config cleanly.

4. **Move macOS sync API keys to Keychain**
   - Reuse `KeychainService.swift` instead of inventing a new secret store.
   - Migrate legacy `UserDefaults` values on load, then delete the old key.
   - When sync is disabled, clear both `sync.json` and the Keychain entry.
   - Define the reopen path explicitly: when a vault/app session restarts, rehydrate the sync API key from Keychain back into the live bridge/session before register/push/full-sync calls.

5. **Keep preview labeling honest in the app UI and bridge surface**
   - Ensure settings/help text keeps sync positioned as preview after the config/storage migration.
   - Avoid any UI copy that implies production-grade sync security just because plaintext key storage is removed.

6. **Ignore sync config in git**
   - Add `sync.json` to `.gitignore` so preview sync metadata is not accidentally committed from example/demo vaults.

## Security regression tests to add

- `TestRunRejectsEmptyAPIKeyWithoutUnsafeFlag`
- `TestHandlerNoLongerReturnsWildcardCORSByDefault`
- `TestBridgeSyncSetupPersistsConfigWithoutAPIKey`
- `TestLoadSyncClientReadsMetadataWithoutSecret`
- `TestVaultClientMigratesLegacySyncAPIKeyToKeychain`
- `TestVaultClientDisableSyncClearsKeychainAndDisk`
- `TestSwiftSyncConfigDecodeWithoutPersistedAPIKey`
- `TestVaultReopenRehydratesSyncSecretFromKeychain`

## Verification

- `go test ./core/sync/... ./services/syncserver/...`
- `go test ./bridge/...`
- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`

## Docs/update implications

- `README.md`: keep sync positioned as preview and document explicit auth expectation.
- `docs/deployment-guide.md`: replace "optional auth" language with explicit unsafe/dev opt-in wording.
- `docs/codebase-summary.md`: update sync config persistence description so `sync.json` is metadata-only.
- `docs/bridge-integration.md`: update bridge-side sync config behavior and secret-storage boundaries.

## Success criteria

- Server no longer starts anonymously by accident.
- Default sync HTTP responses no longer advertise wildcard cross-origin access.
- `sync.json` never contains `api_key`.
- `UserDefaults` no longer holds the sync API key after migration.
- Disabling sync removes both the on-disk config file and secure stored secret.
- Swift bridge/config UI continues to function with metadata-only persisted config.
- `.gitignore` covers `sync.json`.

## Risk assessment

- **Migration bugs:** legacy users could lose sync access if migration is wrong; mitigate with explicit migration tests and one-time fallback logging that does not expose secrets.
- **Shared file contention:** `VaultClient.swift` overlaps Phase 03; serialize those changes intentionally.
- **Developer friction:** local test setups may depend on empty auth; mitigate with an explicit unsafe/dev flag rather than silent bypass.

## Security considerations

- Do not replace `UserDefaults` with another plaintext file or environment variable.
- Keep preview semantics honest: this phase raises the floor, not the ceiling.
- Ensure migration code clears plaintext copies even on partial failure paths where possible.

## Todo list

- [ ] Make server auth opt-out explicit instead of implicit.
- [ ] Remove wildcard CORS default.
- [ ] Strip `api_key` from persisted sync config.
- [ ] Move macOS sync secrets into Keychain and erase plaintext legacy copies.
- [ ] Ignore `sync.json` in git and add regression coverage.

## Next steps

- Required before Phase 05 can harden sync transport/time semantics without building on top of unsafe defaults.
- Must land before Phase 03 touches `VaultClient.swift` lock/close behavior.