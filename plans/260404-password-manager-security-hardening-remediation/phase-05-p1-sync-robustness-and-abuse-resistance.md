# Phase 05 — P1 sync robustness and abuse resistance

## Context links

- `./plan.md`
- `services/syncserver/main.go`
- `services/syncserver/main_test.go`
- `services/syncserver/server_test.go`
- `services/syncserver/storage.go`
- `services/syncserver/postgres_storage.go`
- `services/syncserver/smoke-test.sh`
- `services/syncserver/smoke_test.go`
- `bridge/sync_api.go`
- `bridge/bridge_extra_test.go`
- `core/sync/conflict/resolver.go`
- `core/sync/conflict/resolver_test.go`
- `core/sync/server/sync_handler.go`
- `core/sync/server/sync_handler_test.go`
- `core/sync/client/sync_client.go`
- `core/sync/client/sync_client_test.go`
- `core/sync/protocol/types.go`
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`

## Overview

Priority: P1  
Status: pending  
Goal: raise sync from obviously under-hardened preview to a defensible preview by adding abuse controls and removing the most fragile timestamp trust assumptions. This phase is still not a promise of production-grade sync.

## Key insights

- Request bodies are not size-limited today.
- The server currently trusts client-provided item timestamps for admission decisions and uses local `time.Now()` on both client and server for sync progress.
- The bridge currently collects local sync candidates using filesystem `modTime > last_sync_time`; that breaks if `last_sync_time` becomes server-authoritative without changing local dirty detection.
- There is no TLS enforcement; operators can run insecure HTTP without an explicit danger acknowledgment.
- Phase 02 fixes auth/default secret handling, but it does not harden transport or abuse resistance by itself.

## Requirements

### Functional

- Reject oversized sync/register requests with a predictable error code.
- Require TLS or an explicit unsafe HTTP opt-in at startup.
- Local dirty detection must not depend on comparing local file mtimes against server-authoritative sync time.
- Move sync progress tracking to server-authoritative time values returned in responses.
- Reduce admission reliance on raw client clocks by server-stamping accepted writes.
- Keep sync clearly labeled preview until this phase and Phase 06 are complete.

### Non-functional

- Avoid a major protocol rewrite or version-vector redesign in this remediation.
- Preserve compatibility where possible; if protocol behavior changes, document it clearly.
- Keep the storage layer changes incremental for both SQLite and Postgres backends.
- Prefer correctness over incremental-efficiency tricks for preview-scale vaults; do not skip legitimate local changes to save a file scan.

## Architecture

- **Body limits:** wrap request bodies with `http.MaxBytesReader` (or equivalent) using a safe default and testable limit.
- **Transport guardrail:** require either `--tls-cert/--tls-key` or an explicit `--unsafe-http`/equivalent acknowledgment; if unsafe mode remains, it must read as dangerous and stay preview-only.
- **Dirty-state detection:** replace bridge-side `modTime > last_sync_time` gating with a safer mechanism before server-time adoption. Acceptable pragmatic options: full item enumeration plus manifest diff on `{item_id, version, checksum, deleted}`, or another checksum/version-based snapshot that does not rely on wall-clock comparisons.
- **Conflict/admission rule:** once server-stamped time exists, stop treating client timestamps as the primary trust anchor. For preview remediation, use `version` plus persisted server-stamped time for ordering/conflict behavior: lower version loses, higher version may replace, equal-version cross-device writes conflict, and full ties remain server-wins. Client timestamps become advisory/diagnostic only, not the sole admission gate.
- **Server-authoritative time:** after dirty detection no longer depends on local-vs-server clock comparison, overwrite accepted persisted timestamps with server time and return `ServerTime` for clients to persist as pull/push progress.
- **Scope control:** leave rate limiting, account isolation, and audit logging out of this remediation; they remain future work and part of why sync is still preview.

## Related code files

### Modify

- `services/syncserver/main.go`
- `services/syncserver/main_test.go`
- `services/syncserver/server_test.go`
- `services/syncserver/storage.go`
- `services/syncserver/postgres_storage.go`
- `services/syncserver/storage_test.go`
- `services/syncserver/postgres_storage_test.go`
- `services/syncserver/smoke-test.sh`
- `services/syncserver/smoke_test.go`
- `bridge/sync_api.go`
- `bridge/bridge_extra_test.go`
- `core/sync/conflict/resolver.go`
- `core/sync/conflict/resolver_test.go`
- `core/sync/server/sync_handler.go`
- `core/sync/server/sync_handler_test.go`
- `core/sync/client/sync_client.go`
- `core/sync/client/sync_client_test.go`
- `core/sync/protocol/types.go` (only if response/field semantics need tightening)
- `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`

### Create

- None expected; use current server/client packages instead of introducing a second sync protocol path.

### Delete

- None.

## Implementation steps

1. **Add request size limits**
   - Apply explicit byte caps to `/sync/push` and `/devices/register`.
   - Return `413`/structured JSON errors and add tests for truncated/oversized bodies.

2. **Add transport guardrails**
   - Refuse default insecure HTTP startup unless an explicit unsafe flag is provided.
   - Prefer built-in TLS flags or a strongly documented reverse-proxy/TLS requirement, but keep the operator path simple.

3. **Decouple local dirty detection from wall-clock time**
   - Update bridge-side sync candidate collection so it no longer compares local file mtimes against `last_sync_time`.
   - Prefer correctness-first detection for preview scale, even if it means a broader file scan.

4. **Make server time authoritative**
   - Stamp accepted items with server time before persistence.
   - Update pull/push responses and client last-sync handling so progress is based on `ServerTime`, not client wall clock.
   - Update Swift/bridge sync state handling so the new server time semantics do not suppress legitimate local changes.
   - Persist updated server sync time on standalone push paths too, not only on full-sync flows.

5. **Redefine conflict and admission semantics explicitly**
   - Update `core/sync/server/sync_handler.go` and `core/sync/conflict/resolver.go` together so preview conflict handling matches the new rule: client timestamps are no longer the primary trust input.
   - Add negative tests for forged future timestamps, equal-version cross-device ties, and server-stamped-vs-client-stamped comparisons.

6. **Rebaseline tests and smoke flows**
   - Update integration/smoke tests for new startup rules, size limits, and timestamp behavior.
   - Cover both SQLite and Postgres persistence paths where timestamp behavior changes.
   - Add direct bridge/client regressions for local-change detection under clock skew or server-time advancement.

7. **Keep preview positioning honest**
   - Document that this phase improves the preview but does not turn sync into a production/enterprise claim.

## Security regression tests to add

- `TestHandlerRejectsOversizedPushBody`
- `TestRegisterRejectsOversizedBody`
- `TestRunRequiresTLSOrExplicitUnsafeHTTP`
- `TestCollectLocalSyncItemsDoesNotSkipChangesAfterServerTimeAdvance`
- `TestPushPersistsServerTimestampInsteadOfClientTimestamp`
- `TestSyncClientPersistsLastSyncTimeFromServerResponse`
- `TestResolverIgnoresForgedFutureClientTimestamp`
- `TestResolverConflictsOnEqualVersionCrossDeviceTie`
- `Smoke test covers explicit auth + unsafe/TLS startup expectations`

## Verification

- `go test ./core/sync/... ./services/syncserver/...`
- `bash services/syncserver/smoke-test.sh both`
- Optional targeted reruns for timestamp-sensitive tests in `core/sync/client` and `services/syncserver`.

## Docs/update implications

- `README.md`: sync remains preview even after this phase unless Phase 06 docs/tests are also done.
- `docs/deployment-guide.md`: update operator guidance for TLS/unsafe mode and request-limit expectations.
- `docs/system-architecture.md`: update sync flow to show server-authoritative time.
- `docs/bridge-integration.md`: update bridge-side sync state semantics after local dirty-detection and server-time changes.
- `docs/project-roadmap.md`: reflect that sync is hardened preview, not GA.

## Success criteria

- Oversized sync requests are rejected predictably.
- The server will not run insecurely without an explicit unsafe acknowledgment.
- No legitimate local change is skipped because bridge/client code compared local file mtimes against server sync time.
- Client-controlled timestamps are no longer the primary admission/conflict trust input.
- Accepted sync writes carry server-assigned timestamps.
- Clients persist `lastSyncTime` from server responses, reducing clock-skew sensitivity.
- Smoke/integration tests cover the new preview-grade constraints.

## Risk assessment

- **Behavioral compatibility:** timestamp changes can affect conflict expectations; mitigate with explicit tests and doc updates.
- **Dirty-detection correctness:** the bridge scan change is the real hazard; mitigate by making it an explicit sub-scope, not an implied afterthought.
- **Operator friction:** TLS/unsafe-mode requirements may surprise existing preview users; mitigate with clear startup errors and docs.
- **Scope creep:** full production hardening would need rate limits, authz, and auditability; keep those out of this phase.

## Security considerations

- Do not call sync “production ready” at the end of this phase.
- Keep body-limit defaults tight enough to matter, but large enough for realistic encrypted payload batches.
- Make unsafe transport mode loudly dangerous in flags, logs, and docs.

## Todo list

- [ ] Add request size limits and predictable error handling.
- [ ] Require TLS or explicit unsafe HTTP opt-in.
- [ ] Make server time authoritative for accepted writes and sync progress.
- [ ] Update sync tests and smoke flows for the new rules.
- [ ] Preserve honest preview positioning in code and docs.

## Next steps

- This phase is only required for a hardened sync-preview claim; local/offline ship can still proceed without it if preview wording remains strict.
- Feeds directly into Phase 06 documentation and release-gate updates.