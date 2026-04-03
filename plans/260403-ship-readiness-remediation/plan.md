---
title: "ZeroPass ship-readiness remediation roadmap"
description: "Practical 3-phase plan to move the verified repo state to ship-ready without overbuilding."
status: completed
priority: P1
effort: 2-3 weeks
branch: main
tags: [planning, release, docs, sync, testing]
created: 2026-04-03
---

# ZeroPass ship-readiness remediation roadmap

## Overview

Goal: move ZeroPass from "Phase 1 verified, Phase 2 partially real" to a ship-ready repo by first restoring repo truth, then aligning the sync contract, then closing release/deployment gaps.

## Verified baseline

- Phase 1 local vault remains functionally complete.
- macOS tests pass via `xcodebuild test`.
- `go test ./...` now passes after restoring the missing bridge helper wrappers.
- Sync docs/specs were narrowed to the current preview implementation (timestamp-first server admission, no version vectors in the live protocol).
- Public docs now consistently describe `zp://item/field`-style secret references.
- Interactive TUI mode and recovery PDF export remain explicitly deferred.
- Sync server now has an in-repo Dockerfile, runs as non-root, and has a repeatable smoke-test script for binary and Docker operator paths.

## Phase map

| Priority | Status | Focus | Outcome | Link |
|---|---|---|---|---|
| P0 | ✅ Completed | Repo truth + green validation | `go test ./...` is green and repo/docs now reflect the verified state | [phase-01](./phase-01-p0-repo-truth-and-green-builds.md) |
| P1 | ✅ Completed | Sync contract alignment | Product/docs/code agree on preview sync semantics and secret reference format | [phase-02](./phase-02-p1-sync-contract-alignment.md) |
| P2 | ✅ Completed | Production readiness | Sync server is containerized, rootless, smoke-tested, and release-gated as a preview path | [phase-03](./phase-03-p2-production-readiness-and-release-gates.md) |

## Execution order

1. Fix broken repo signals first (`go test`, misleading docs, stale roadmap claims).
2. Decide contract direction once, then update docs/code/tests together.
3. Only then spend time on packaging, deployment, and release gates.

## Guardrails

- Keep deferred items deferred: no TUI mode, no recovery PDF export in this roadmap.
- Prefer docs correction over code expansion unless a mismatch blocks a promised ship target.
- Treat sync as "beta/self-hosted preview" until deployment, conflict testing, and operator docs are real.

## Quick wins

- Add/restore missing bridge test helper wrappers for sync + crypto C-API tests.
- Correct `run` command help text + README/PRD reference format mismatch.
- Mark sync docs as timestamp/version-based today, not version-vector based.
- Add a real `Dockerfile` matching `services/syncserver/main.go` flags.

## Success bar

- Repo status is truthful.
- Default validation commands are green.
- Docs match behavior.
- Sync server can be built, containerized, smoke-tested, and positioned honestly for release.
