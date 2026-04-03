---
title: "Rename CLI invocation from zeropass to zp"
description: "Lightweight plan for changing the CLI command surface from zeropass to zp with minimal blast radius."
status: completed
priority: P2
effort: 2h
branch: main
tags: [refactor, docs, cli]
created: 2026-04-03
---

# Rename CLI invocation from `zeropass` to `zp`

## Overview

Minimal user-facing rename only. Change CLI display/invocation to `zp`. Keep product/repo name `ZeroPass`, Go module path `github.com/zeropass/zeropass`, sync server naming, and on-disk vault location `~/.zeropass` unchanged unless a separate migration is explicitly approved.

## Phases

| # | Phase | Status | Effort | Notes |
|---|-------|--------|--------|-------|
| 1 | Narrow source rename | Completed | 45m | Cobra root name and user-facing usage/help examples now render `zp`. |
| 2 | Align tests + docs | Completed | 45m | Tests, completion examples, and user-facing CLI docs were updated to `zp`. |
| 3 | Validation + cleanup | Completed | 30m | Kept the rename explicit without a `zeropass` alias and verified build/help/completions/tests. |

## Recommended scope

- Update user-facing command surface only: binary name, Cobra `Use`, help/usage text, and shell completion examples.
- Update build/install/release snippets that tell users to build/run `zeropass`.
- Keep `ZeroPass` branding, repo/module/import paths, app target names, sync server service names, and `~/.zeropass` storage paths unchanged.
- Keep the rename explicit to `zp`; only add a `zeropass` compatibility alias if a follow-up requirement appears.

## Impact map

- Source: `packages/cli/cmd/root.go`, `packages/cli/cmd/completion.go`, `packages/cli/cmd/run.go`, `packages/cli/cmd/env.go`, `packages/cli/cmd/helpers.go`, `packages/cli/cmd/unlock.go`
- Tests: `packages/cli/cmd/cmd_test.go`, `packages/cli/cmd/features_test.go`
- Docs: `README.md`, `docs/codebase-summary.md`, `docs/deployment-guide.md`, `docs/project-roadmap.md`

## Validation

- Build CLI as `zp` and confirm `./zp --help` + `./zp --version` output shows `zp`.
- Verify one real help/error path that previously printed `zeropass` (for example `run` with no command or missing vault guidance).
- Verify shell completion generation still works and emits `zp`-based instructions.
- Run `go test ./packages/cli/cmd/...`.
- Verify docs/examples no longer instruct users to invoke the CLI as `zeropass`.
