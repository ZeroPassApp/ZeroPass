# Phase 01 — P0 repo truth and green builds

## Overview

Priority: P0  
Status: completed  
Goal: remove false negatives and false positives so the repo has one credible status signal before any release push.

## Outcome

- Restored the missing bridge helper wrappers referenced by the extra bridge tests.
- Re-validated the repo with `go test ./...`.
- Updated release-facing docs so Phase 1 is described as verified and Phase 2 as preview/incomplete.

## Why first

Initial repo state said "Phase 1 complete" and macOS was green, but `go test ./...` still failed. That blocked CI trust and made every later plan noisy.

## Concrete tasks

1. **Unblock Go test suite**
   - Fix `bridge/bridge_extra_test.go` undefined helper symbol failures for:
     - `zpCSyncSetup`
     - `zpCGeneratePassword`
     - `zpCGeneratePassphrase`
     - `zpCScorePassword`
   - Choose one approach and use it consistently:
     - add missing Go-side test helper wrappers, or
     - rename tests to use the existing wrapper names/API surface.
   - Re-run targeted bridge tests, then `go test ./...`.

2. **Normalize repo health signals**
   - Document the verified default validation commands in repo docs/CI notes:
     - `go test ./...`
     - macOS `xcodebuild test ...`
   - Remove or downgrade any stale claim that Phase 2 is more complete than the validated code proves.

3. **Correct roadmap/status documents to current truth**
   - Update `docs/project-roadmap.md` and `docs/project-overview-pdr.md` to reflect:
     - Phase 1: complete/verified
     - Phase 2: partial implementation, not production-ready
     - Sync deployment/testing: incomplete
   - Keep deferred items clearly labeled as deferred, not missing bugs.

4. **Tighten release-entry criteria**
   - Define the minimum ship gate for the current release train:
     - local vault + CLI + macOS app are shippable now
     - sync server is preview/beta until P2 closes

## Docs vs code fixes in this phase

| Mismatch | Fix docs | Fix code |
|---|---|---|
| `go test ./...` fails while Phase 1 is presented as complete | Add explicit note that Phase 1 is feature-complete but repo validation must be green to claim ship-ready | Restore missing bridge test helper coverage so the test command actually passes |
| TUI mode still mentioned in product docs as CLI behavior | Mark as deferred/non-goal for current release | No code change |
| Recovery PDF export appears as spec expectation | Mark as deferred/non-goal for current release | No code change |

## Definition of done

- `go test ./...` passes on the main branch state.
- `xcodebuild test` remains green.
- `docs/project-roadmap.md` and `docs/project-overview-pdr.md` describe the present state honestly.
- Release scope explicitly excludes TUI mode and recovery PDF export.
- Team can answer "what is shippable today?" in one sentence without caveats.

## Quick wins (< 0.5 day)

- Add the four missing bridge helper wrappers or rename the failing test calls.
- Update roadmap percentages/status text.
- Add one short "current release scope" note to README/docs.

## Risks

- Bridge tests may expose a real missing exported API instead of just missing wrappers.
- Fixing test helpers may reveal additional latent CI failures hidden behind the first compile stop.
