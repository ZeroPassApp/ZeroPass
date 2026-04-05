---
title: "Fit macOS recovery phrase screen without default scrolling"
description: "Minimal plan to keep the recovery phrase CTA and confirmation toggle visible at the default recovery window size."
status: pending
priority: P2
effort: 1h
branch: main
tags: [macos, swiftui, auth, recovery]
created: 2026-04-05
---

# Fit macOS recovery phrase screen without default scrolling

## Goal
Keep the recovery phrase screen fully usable at the default recovery-window size, without forcing the user to scroll to reach the confirmation toggle or primary CTA.

## Scope
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/RecoveryPhraseView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/RecoveryPhraseCardView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthWindowLayoutModifier.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/AuthSceneScaffold.swift` *(read-only unless sizing + local compaction both fail)*
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`

## Phase
1. [Phase 01 — fit recovery phrase screen inside the default window](./phase-01-fit-recovery-phrase-screen-inside-default-window.md) — pending

## Root-cause summary
- `AuthSceneScaffold` always wraps auth content in a `ScrollView`, so oversized screens degrade into scrolling instead of surfacing a sizing mismatch.
- `AuthWindowLayout` gives `.showingRecovery` a compact default content size (`680x580`) even though `RecoveryPhraseView` has a much taller stack than the locked screen.
- The recovery screen height is driven by a large header, warning banner, recovery card with 12-word grid, extra helper copy/pills, action row, and a final confirmation toggle.
- Result: at the default recovery size, the lower controls fall below the fold.

## Recommended direction
Use **both**, but keep it minimal: adjust the recovery window height first, then apply only light recovery-specific compaction if the fit is still too tight. Do **not** remove the shared `ScrollView` from `AuthSceneScaffold` as the primary fix.

## Guardrails
- No recovery-flow logic changes.
- No global auth scaffold redesign.
- Keep scrolling available for manually resized/smaller windows.
- Reuse existing `ZPTheme` spacing/tokens; avoid new shared abstractions unless they clearly reduce code.

## Unresolved questions
- None at plan time; the implementation can decide the final recovery height after a quick visual check at 100% display scale.