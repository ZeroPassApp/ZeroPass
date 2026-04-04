---
title: "ZeroPass security hardening remediation"
description: "Execution-ready remediation plan for audit findings blocking a defensible password-manager security posture."
status: pending
priority: P1
effort: 18-25d eng
branch: main
tags: [security, planning, hardening, vault, sync, macos, cli]
created: 2026-04-04
---

# ZeroPass security hardening remediation

## Goal

Remediate the 2026-04-04 audit findings with the smallest set of changes that materially improves ZeroPass's real security posture. Target: defensible local/offline ship posture first; sync stays preview unless the stronger Phase 05 bar is also met.

## Source findings

- Detailed traceability matrix: [`./reports/audit-findings-traceability.md`](./reports/audit-findings-traceability.md)

## Ship bar split

**Must fix before claiming "secure enough to ship" (local/offline):** phases 01, 03, 04, 06, plus the **local-secret-storage subset** of Phase 02.  
**Must fix before claiming sync is more than a preview:** the **server-defaults subset** of Phase 02 and all of Phase 05 too.

## Phase map

| Priority | Status | Effort | Gate | Outcome | Link |
|---|---|---:|---|---|---|
| P0 | Pending | 3-4d | Must | Block item-path escape, shrink search-index exposure, make recovery fail closed | [phase-01](./phase-01-p0-core-vault-containment.md) |
| P0 | Pending | 4-5d | Local must + sync preview | Remove plaintext local sync secrets and require explicit sync-preview auth defaults | [phase-02](./phase-02-p0-sync-auth-and-secret-storage.md) |
| P1 | Pending | 3-4d | Must | Clear clipboard, invalidate quick search, gate recovery reveal, reduce selectable secrets | [phase-03](./phase-03-p1-macos-secret-surface-hardening.md) |
| P1 | Pending | 2-3d | Must | Make CLI clipboard clearing survive process exit, harden exports, reduce bridge secret lifetime | [phase-04](./phase-04-p1-cli-bridge-and-export-hardening.md) |
| P1 | Pending | 4-6d | Sync-only | Add preview-grade abuse resistance: body limits, TLS guardrails, server-authoritative sync time | [phase-05](./phase-05-p1-sync-robustness-and-abuse-resistance.md) |
| P1 | Pending | 2-3d | Must | Add regression coverage, update docs, and enforce release/security gates | [phase-06](./phase-06-p1-security-regression-tests-and-docs.md) |

## Dependency graph

```text
01 ─┬─> 02 ─┬─> 03 ─┐
    │       └─> 05 ─┤
    └─> 04 ─────────┤
                     └─> 06
```

Fastest safe order: land Phase 01 first, land Phase 02 next, then run Phases 03/04/05 in parallel only where file ownership does not overlap, then finish with Phase 06.

## File ownership matrix

| Track | Primary ownership | Parallel note |
|---|---|---|
| Vault core | `core/vault/item/*`, `core/vault/index/*`, `core/vault/store/*`, `core/vault/version/*`, `core/vault/importexport/*` | Serial first; do not overlap before Phase 01 lands |
| Sync/auth | `core/sync/**`, `services/syncserver/**`, `bridge/{sync_api.go,vault_api.go}`, `.gitignore` | Phase 05 starts after Phase 02 baseline is merged |
| macOS surface | `apps/macos/ZeroPass/ZeroPass/{Services,Views}/**` | Serialize `VaultClient.swift` between phases 02 and 03 |
| CLI/bridge | `packages/cli/cmd/**`, `apps/macos/ZeroPass/ZeroPass/Bridge/ZPBridge.swift`, `bridge/{helpers,capi_ops}.go` | Phase 04 can run after 01; auth-view edits must not race with Phase 03 |
| QA/docs | Existing `*_test.go`, `ZeroPassTests.swift`, `README.md`, `docs/*.md` | Phase 06 starts late; final pass only after code phases settle |

Shared hot spots to serialize explicitly: `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift`, `apps/macos/ZeroPass/ZeroPass/Views/Auth/{UnlockVaultView.swift,RecoveryPhraseView.swift}`, and `core/sync/server/sync_handler.go`.

## Release gates

- No item ID or import path can escape the vault `items/` directory.
- Search index never stores secret field values **or free-form notes** in plaintext by default; index rebuild/migration is covered by tests.
- Recovery unlock fails closed on rotation failure; recovery validation zeroes temporary key material.
- Local ship requires that sync API keys never live in `UserDefaults` or `sync.json`, even if sync itself remains preview.
- Sync preview requires explicit auth defaults, no wildcard CORS by default, and the Phase 05 timestamp/body-limit hardening bar.
- macOS and CLI clipboard/export surfaces meet the new hardening bar.
- Security regression suites, README, roadmap, deployment docs, and architecture docs match shipped behavior.
- If Phase 05 is incomplete, sync remains explicitly labeled **preview** everywhere.

## Cook handoff

Recommended handoff: `/ck:cook --parallel /Volumes/DATA/Developments/ZeroPass/plans/260404-password-manager-security-hardening-remediation/plan.md`