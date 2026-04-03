# Phase 02 — P1 sync contract alignment

## Overview

Priority: P1  
Status: completed  
Goal: align product/docs/code on what sync and secret references actually are, then only expand code where the chosen contract is worth shipping.

## Outcome

- Aligned public docs around the currently implemented `zp://item/field` secret reference shape.
- Removed stale version-vector claims from public sync documentation.
- Reframed sync as a preview/transport layer rather than a fully applied generic merge pipeline.

## Decision applied

The current implementation was treated as the authoritative ship contract, and docs were updated in one pass so public materials no longer describe a different system.

## Concrete tasks

1. **Decide sync truth model**
   - Treat current implementation as the source of truth unless there is a strong release requirement to upgrade it now.
   - Current verified implementation:
     - `protocol.SyncItem` uses `Version int` + `Timestamp int64`
     - conflict resolver uses timestamp, then integer version, then remote/server preference
   - If keeping current implementation for ship:
     - rewrite docs/specs away from version vectors
     - reposition sync as simple deterministic timestamp/version sync
   - If upgrading code to version vectors instead:
     - create a separate scoped feature plan; do **not** sneak it into ship-readiness.

2. **Align secret reference format**
   - Decide whether ship target is:
     - `zp://item/field` (matches code today), or
     - `zp://vault/item/field` (matches PRD/docs)
   - Recommended for ship-readiness: keep `zp://item/field` now, update docs to current reality, and optionally add backwards-compatible parsing later if multi-vault addressing becomes necessary.
   - Update CLI help, README, PRD, tests, examples, and deployment docs together.

3. **Audit sync surface claims**
   - Review and correct these docs first:
     - `docs/system-architecture.md`
     - `docs/codebase-summary.md`
     - `docs/project-roadmap.md`
     - `docs/project-overview-pdr.md`
     - `prd.md`
   - Remove claims for:
     - version vectors/logical clocks if not implemented
     - stronger conflict guarantees than current tests prove
     - production-ready sync if deployment/testing are incomplete

4. **Add contract-level tests where mismatch risk is highest**
   - For secret refs: parser tests covering the chosen supported syntax.
   - For sync: tests/documentation proving current conflict rules:
     - later timestamp wins
     - equal timestamp => higher integer version wins
     - full tie => remote/server wins

## Docs vs code fixes in this phase

| Mismatch | Preferred action | Notes |
|---|---|---|
| Docs/spec say version vectors; code uses timestamps + integer versions | **Fix docs to code now** | Lowest-risk route to ship; reserve version-vector work for a future sync-hardening track |
| PRD/docs say `zp://vault/item/field`; code supports `zp://item/field` | **Fix docs to code now** | Only change code if multi-vault secret addressing is a hard near-term requirement |
| Sync described as fully conflict-safe across 3+ devices | **Fix docs wording** to "partial / preview / deterministic conflict handling" | Keep claims bounded to tested behavior |

## Definition of done

- One documented sync contract exists across code comments, docs, README, and PRD.
- One documented `zp://...` reference format exists across CLI help, tests, and examples.
- No public doc mentions version vectors unless code actually ships them.
- No public doc implies interactive TUI or recovery PDF export is in the current release scope.
- High-risk parser/conflict rules are covered by focused tests.

## Quick wins (< 0.5 day)

- Update `packages/cli/cmd/run.go` help text and README examples to the chosen `zp://` format.
- Rewrite the conflict-resolution bullets in `docs/codebase-summary.md` and `docs/system-architecture.md`.
- Add a short sync-status note: "implemented but not production-ready".

## Risks

- Touching `zp://` syntax may create backward-compat questions for existing examples/tests.
- If stakeholders really want multi-vault refs now, code changes can spill beyond ship-readiness into feature work.
