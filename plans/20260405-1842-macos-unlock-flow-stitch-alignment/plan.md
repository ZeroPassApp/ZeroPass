---
title: "Close remaining macOS unlock-flow gaps vs Stitch"
description: "Keep the shipped locked screen as the baseline, align adjacent auth surfaces, and only change flow/state where the approved design truly requires it."
status: pending
priority: P1
effort: 6h
branch: main
tags: [macos, swiftui, auth, unlock, stitch]
created: 2026-04-05
---

# macOS unlock-flow Stitch alignment

The locked-state Stitch fidelity pass is mostly done already. The remaining work is the flow around it: `welcome/create/open → locked → showingRecovery → unlocked`, plus any small state changes the approved Stitch flow still requires.

## Goals
- Treat `plans/20260405-1230-unlock-vault-stitch-fidelity-pass/` as the locked-screen baseline, not a fresh rewrite target.
- Make welcome, create/open, recovery display, and recovery regeneration feel like one auth family.
- Keep flow/state changes conditional and minimal; no auth-state rewrite unless the design cannot fit the current state machine.
- Close the remaining validation gap: recovery unlock, Touch ID, keyboard flow, resize cycle, and VoiceOver.

## Non-goals
- No crypto, bridge, CLI, sync, or main-shell feature work.
- No speculative new auth architecture.
- No broad token/theme refactor unless a tiny shared helper clearly removes duplication.

## Likely files
- Visual/auth surfaces: `Views/Auth/WelcomeView.swift`, `CreateVaultView.swift`, `OpenVaultSheet.swift`, `RecoveryPhraseView.swift`, `AuthSceneScaffold.swift`, `Views/Components/RecoveryPhraseCardView.swift`, `Views/Settings/SecuritySettingsView.swift`
- Flow/state only if needed: `Views/Auth/UnlockVaultView.swift`, `Services/VaultClient.swift`, `ContentView.swift`, `Views/Auth/AuthWindowLayoutModifier.swift`, optional `ZeroPassApp.swift`
- Validation: `ZeroPassUITests/ZeroPassUITests.swift`, optional `ZeroPassTests/RecoveryPhraseSupportTests.swift`

## Phases
| Phase | Focus | Effort | Notes | Link |
|---|---|---:|---|---|
| 1 | Align surrounding auth surfaces | 3h | Welcome, create/open, recovery, regen sheet | [phase-01](./phase-01-align-surrounding-auth-surfaces.md) |
| 2 | Close flow/state gaps only where Stitch demands | 1.5h | Touch risky state files last, only if the visual diff proves a real gap | [phase-02](./phase-02-close-conditional-flow-gaps.md) |
| 3 | Validate and sync | 1.5h | UI tests, manual QA, docs/changelog only if behavior changes | [phase-03](./phase-03-validate-unlock-flow.md) |

## Sequencing guardrails
- Do **not** start by reworking `UnlockVaultView.swift` again; first align the adjacent surfaces that still drift visually.
- Touch `VaultClient.swift`, `ContentView.swift`, or `ZeroPassApp.swift` only after the approved Stitch flow proves a real sequence gap.
- Keep recovery readability, copy safety, and one-time-use seriousness above visual fidelity.

## Dependencies
- Latest approved Stitch frames for welcome/open/create/recovery in addition to the already-landed locked screen.
- Access to Touch ID hardware for the final manual pass.

## Unresolved questions
- Does the approved Stitch flow still want a segmented password/recovery switch, or a password-first path with a secondary recovery action?
- Should recovery regeneration in settings mirror the auth recovery screen exactly, or only reuse the phrase card and action hierarchy?
- Are there newer Stitch revisions beyond the locked-screen pass already implemented?
