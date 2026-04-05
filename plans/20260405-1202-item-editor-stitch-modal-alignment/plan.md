---
title: "Align macOS item editor with Stitch modal"
description: "Focused UI polish plan for making the item editor feel like a premium modal without changing item behavior."
status: in-progress
priority: P2
effort: 2h
branch: main
tags: [macos, swiftui, item-editor, ui-polish]
created: 2026-04-05
progress: "Impl: 100% | Testing: Automated ✓ | Manual QA: Pending"
---

# Item Editor Stitch Modal Alignment

## Goal
Bring the macOS item editor closer to the Stitch `Create / Edit Item Modal` target while keeping bindings, save flow, default-field logic, and generator behavior intact.

## Scope
- `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Components/PasswordStrengthBar.swift` (only if the strength treatment needs shared utility styling)

## Phase
1. [Phase 01 — align item editor modal polish](./phase-01-align-item-editor-modal-polish.md) — in-progress

## Guardrails
- No data-model, save, default-field, or validation behavior changes.
- Prefer existing `ZPTheme` surfaces/tokens over introducing new design primitives.
- Keep secret fields masked and generator actions obvious + accessible.
- Avoid broad refactors; use small private helpers inside `ItemEditorView.swift` only if the layout becomes clearer.
