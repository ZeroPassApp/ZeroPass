# Phase 06 — P1 security regression tests and docs

## Context links

- `./plan.md`
- `./reports/audit-findings-traceability.md`
- `README.md`
- `docs/project-roadmap.md`
- `docs/deployment-guide.md`
- `docs/system-architecture.md`
- `docs/bridge-integration.md`
- `docs/project-overview-pdr.md`
- `docs/codebase-summary.md`
- `.github/copilot-instructions.md`
- `core/vault/**/*_test.go`
- `core/sync/**/*_test.go`
- `services/syncserver/**/*_test.go`
- `packages/cli/cmd/*_test.go`
- `bridge/*_test.go`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`

## Overview

Priority: P1  
Status: pending  
Goal: lock the hardening work in with regression coverage, update repo truth, and define the exact release/security gates for local ship vs preview sync.

## Key insights

- The audit findings span Go core, sync server, bridge, CLI, and Swift UI; a single package test run is not enough.
- Existing repo docs already separate some sync-preview language, but they still need to reflect whichever subset of Phases 02-05 actually lands.
- The repo has `project-roadmap.md` but no dedicated `project-changelog.md`; update existing docs rather than inventing a new doc unless maintainers explicitly ask for one.

## Requirements

### Functional

- Every confirmed audit finding must end with either an automated regression test or a documented manual validation step if automation is impractical.
- A finding-to-phase-to-test-to-doc traceability matrix must exist and stay current as phases land.
- README and docs must state the same security posture the code actually provides.
- Release gates must separate local/offline ship readiness from sync-preview hardening.
- Default validation commands must be documented and green before sign-off.

### Non-functional

- Prefer updating existing test files/docs over creating a new parallel documentation tree.
- Keep hidden-test risk low by adding direct regression coverage around failure paths, not only happy paths.

## Architecture

- **Test layering:** keep unit tests close to the touched packages, use existing integration suites for cross-package behavior, and use `xcodebuild test` plus a short manual checklist for UI-specific hygiene.
- **Docs truth source:** update the existing README + docs set instead of creating a standalone “security hardening” manual that will drift.
- **Release gate split:** define one gate for local/offline password-manager posture and a stricter optional gate for sync-preview posture.

## Related code files

### Modify

- `reports/audit-findings-traceability.md`
- `README.md`
- `docs/project-roadmap.md`
- `docs/deployment-guide.md`
- `docs/system-architecture.md`
- `docs/bridge-integration.md`
- `docs/project-overview-pdr.md`
- `docs/codebase-summary.md`
- `.github/copilot-instructions.md`
- `core/vault/store/store_test.go`
- `core/vault/item/item_test.go`
- `core/vault/index/index_test.go`
- `core/vault/index/index_coverage_test.go`
- `core/vault/importexport/importexport_test.go`
- `core/vault/version/version_test.go`
- `core/sync/server/sync_handler_test.go`
- `core/sync/client/sync_client_test.go`
- `services/syncserver/server_test.go`
- `services/syncserver/storage_test.go`
- `services/syncserver/main_test.go`
- `services/syncserver/smoke_test.go`
- `packages/cli/cmd/features_test.go`
- `packages/cli/cmd/integration_test.go`
- `bridge/bridge_extra_test.go`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`

### Create

- None preferred; keep coverage in current suites unless a very small missing test file is truly required.

### Delete

- None.

## Implementation steps

1. **Fill regression gaps per finding**
   - Add or expand tests for path traversal, search-index plaintext leakage, recovery fail-closed semantics, sync auth/storage behavior, CLI export permissions, clipboard clearing, and sync abuse controls.
   - Keep the `F-XX` finding IDs from `reports/audit-findings-traceability.md` visible in PR/checklist language.

2. **Run the full verification matrix**
   - `go test ./core/vault/... ./core/sync/... ./services/syncserver/... ./packages/cli/...`
   - `go test ./...`
   - `go vet ./...`
   - `bash services/syncserver/smoke-test.sh both`
   - `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`

3. **Update repo truth**
   - Refresh README and docs to match the implemented behavior, not the planned or aspirational behavior.
   - Update `.github/copilot-instructions.md` only where the verified build/test/posture guidance has materially changed.
   - Update `docs/bridge-integration.md` for sync-config and bridge secret-handling changes from Phases 02 and 04.

4. **Publish release gates**
   - Local/offline ship gate: phases 01-04 complete, phase 06 complete, tests green, docs updated, sync still clearly preview if Phase 05 is not done.
   - Hardened preview-sync gate: phases 01-06 complete, including size limits, transport guardrails, and server-authoritative time.

5. **Maintain the audit traceability matrix**
   - Update `reports/audit-findings-traceability.md` with phase status, test status, and docs status as each finding closes.
   - Refuse sign-off if any `F-XX` item has no owner, no verification step, or stale docs mapping.

6. **Capture manual validation where automation is thin**
   - Clipboard cleared on lock/close and post-CLI copy.
   - Quick-search stale state absent after lock/close.
   - Recovery confirmation gate blocks item hydration until acknowledged.

## Detailed release gate checklist

- Local vault CRUD/search/recovery flows pass with no known traversal or fail-open issues.
- No plaintext sync API key exists in `UserDefaults`, `sync.json`, or tracked files.
- Plaintext export produces owner-only files.
- Clipboard clearing works for macOS app and CLI copy flows.
- `go test ./...`, `go vet ./...`, sync smoke tests, and macOS tests are all green.
- README/docs consistently state whether sync is still preview.
- If Phase 05 is incomplete, no release note or doc claims sync is secure enough for general deployment.

## Security regression tests to add

- Cross-package regression cases for each confirmed audit finding.
- Manual validation checklist committed to docs if a UI behavior cannot be asserted deterministically in tests.
- One focused “ship bar” checklist in docs/README so future regressions are visible in review.

## Verification

- `go test ./core/vault/... ./core/sync/... ./services/syncserver/... ./packages/cli/...`
- `go test ./...`
- `go vet ./...`
- `bash services/syncserver/smoke-test.sh both`
- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`

## Docs/update implications

- `README.md`: top-level product posture and preview wording.
- `docs/project-roadmap.md`: progress and phase completion status.
- `docs/deployment-guide.md`: operator expectations for auth, TLS/unsafe mode, and preview boundaries.
- `docs/system-architecture.md`: updated vault recovery/search/sync flows.
- `docs/bridge-integration.md`: bridge-side sync config persistence and secret-handling boundaries.
- `docs/project-overview-pdr.md` and `docs/codebase-summary.md`: remove stale claims and reflect verified behavior.

## Success criteria

- Regression coverage exists for every confirmed finding or there is a deliberate manual validation note for the few UI cases that cannot be automated reasonably.
- All verification commands are green on the repo’s supported paths.
- Docs and release messaging tell the same truth as the code.
- `reports/audit-findings-traceability.md` is up to date and no finding is orphaned.
- The team has an explicit checklist for when it is safe to say “local/offline password-manager-grade posture achieved.”

## Risk assessment

- **Docs drift:** partial updates can be as dangerous as no updates; mitigate with one owner for the full doc pass.
- **False confidence:** green tests without release-gate wording can still mislead; mitigate by documenting the gate explicitly.
- **Over-testing the wrong thing:** focus on the exact security behaviors identified by the audit, not cosmetic coverage inflation.

## Security considerations

- Do not soften preview wording to make the release narrative cleaner.
- Hidden tests will likely target failure ordering and negative cases; keep those branches covered.
- Treat documentation as part of the security surface for this release because user/operator expectations matter.

## Todo list

- [ ] Add or extend regression coverage for every confirmed audit finding.
- [ ] Run the full Go, sync smoke, vet, and macOS verification matrix.
- [ ] Update README and existing docs to match actual behavior.
- [ ] Publish explicit local-ship vs preview-sync release gates.
- [ ] Record manual validation steps for UI-only security behaviors.

## Next steps

- Final phase before any security-posture claim.
- Once complete, hand off the plan via `/ck:cook --parallel /Volumes/DATA/Developments/ZeroPass/plans/260404-password-manager-security-hardening-remediation/plan.md`.