---
title: "Align macOS unlock screen with Stitch redesign"
description: "Small UI polish plan for centering the unlock card and refining vault/password controls without behavior changes."
status: in-progress
priority: P2
effort: 2h
branch: main
tags: [macos, swiftui, auth, ui-polish]
created: 2026-04-05
---

# Unlock Vault Redesign Polish

## Goal
Bring the locked-state macOS unlock screen closer to the Stitch `Unlock Vault Redesign` mock with minimal, low-risk UI-only changes.

## Scope
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift` (only if shared styling needs parity)

## Phase
1. [Phase 01 — align unlock screen polish](./phase-01-align-unlock-screen-polish.md) — in progress (original polish landed, then extended by the focused Stitch fidelity pass; automated validation done; manual unlock-specific visual/accessibility regression still pending)

## Tracking Note
- The deeper layout/hierarchy follow-up now lives in `plans/20260405-1230-unlock-vault-stitch-fidelity-pass/`.
- Keep this plan open only as historical polish context until the shared manual unlock-specific regression is formally completed.

## Guardrails
- Keep unlock, recovery, Touch ID, focus, and accessibility behavior unchanged.
- Preserve overflow handling for smaller auth window sizes.
- Prefer targeted layout/styling updates over structural redesign.
