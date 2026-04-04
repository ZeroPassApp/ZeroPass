# Audit findings traceability matrix

Source audit date: 2026-04-04

| ID | Severity | Finding | Primary evidence | Planned phases | Regression/docs owner |
|---|---|---|---|---|---|
| F-01 | Critical | Item ID path escape from `items/` | `core/vault/item/item.go`, `core/vault/importexport/import.go` | 01 | `item_test.go`, `importexport_test.go`, architecture docs |
| F-02 | High | Search index stores sensitive field values in plaintext while unlocked | `core/vault/index/index.go`, `core/vault/types/types.go` | 01, 06 | `index_test.go`, system architecture/docs |
| F-03 | High | Recovery one-time-use semantics fail open | `core/vault/store/store.go` | 01, 06 | `store_test.go`, roadmap/architecture docs |
| F-04 | Medium/High | `ValidateRecovery()` decrypts temporary key material too freely | `core/vault/store/store.go`, `core/crypto/key/recovery_key.go` | 01, 06 | `store_test.go`, architecture docs |
| F-05 | High | Sync server can start without auth and uses wildcard CORS defaults | `services/syncserver/main.go`, `core/sync/server/sync_handler.go` | 02, 05, 06 | sync/server tests, deployment docs |
| F-06 | High | Sync API key stored plaintext in `UserDefaults`/`sync.json` | `VaultClient.swift`, `bridge/sync_api.go`, `bridge/vault_api.go` | 02, 06 | bridge/macOS tests, bridge/deployment docs |
| F-07 | Medium/High | macOS clipboard and quick-search state leak secrets after lock/close | `ClipboardService.swift`, `QuickSearchPanelController.swift`, `QuickSearchView.swift`, `VaultClient.swift` | 03, 06 | macOS tests/manual checklist, README/docs |
| F-08 | Medium | Recovery UI reveals decrypted items before replacement mnemonic is acknowledged | `VaultClient.swift`, `RecoveryPhraseView.swift` | 03, 06 | macOS tests, system architecture docs |
| F-09 | Medium | CLI clipboard auto-clear likely dies with the parent process | `packages/cli/cmd/helpers.go`, `packages/cli/cmd/get.go` | 04, 06 | CLI tests, README/docs |
| F-10 | Medium | Plaintext export permissions depend on umask | `packages/cli/cmd/export_cmd.go` | 04, 06 | CLI tests, deployment/docs |
| F-11 | Medium | Sync robustness depends on client/local wall-clock assumptions and lacks request body limits | `bridge/sync_api.go`, `core/sync/client/sync_client.go`, `core/sync/server/sync_handler.go` | 05, 06 | sync/server/bridge tests, deployment/system docs |

## Notes

- Local/offline ship bar maps to: F-01, F-02, F-03, F-04, F-07, F-08, F-09, F-10 plus the local-secret-storage subset of F-06.
- Sync remains preview until the server-defaults subset of F-05, the sync-storage subset of F-06, and all of F-11 are closed.
- Every implementation PR should reference one or more `F-XX` IDs in the phase checklist and regression tests.