---
title: "Rebuild macOS unlock screen to match Stitch hierarchy"
description: "Focused fidelity pass to replace the current hero-card unlock layout with a compact Stitch-like composition."
status: in-progress
priority: P1
effort: 4h
branch: main
tags: [macos, swiftui, auth, unlock, stitch]
created: 2026-04-05
---

# Unlock screen Stitch fidelity pass

## Goal
Move the locked-state macOS unlock screen from “Stitch-inspired polish” to a much more literal Stitch composition: compact centered shell, small vault chip, nested unlock panel, and lighter footer/meta region.

## Scope
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockVaultView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockPasswordSection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/UnlockRecoverySection.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`

## Phase
1. [Phase 01 — rebuild locked-state hierarchy](./phase-01-rebuild-locked-state-hierarchy.md) — in progress (implementation landed; automated validation passed; review approved with nits; manual unlock-specific visual/accessibility regression still pending)

## Progress Snapshot
- Implementation landed in the four scoped auth files.
- Unlock UI moved from the larger hero-card layout to a tighter Stitch-like composition with compact top options, a selected-vault chip, an inner unlock panel, lighter footer metadata, tighter password/recovery chrome, and smaller locked-window sizing.
- Locked-state window title now reads `Unlock Vault`, and vault actions use an icon-only affordance.
- Automated validation passed via `xcodebuild test` for the macOS scheme.
- Independent review verdict: approve with nits; safe to keep.
- Remaining work: formal manual unlock-specific visual/accessibility regression.

## Guardrails
- Keep unlock, recovery, Touch ID, error, focus, and accessibility behavior unchanged.
- Keep zero-knowledge helper text visible, but outside or visually lighter than the main interaction panel.
- Prefer updating the existing auth files over adding new view files unless duplication becomes obvious.
