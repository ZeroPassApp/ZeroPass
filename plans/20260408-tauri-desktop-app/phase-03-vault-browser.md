# Phase 03: Vault Browser

## Context Links
- [plan.md](plan.md) | [phase-02](phase-02-auth-flow.md)
- [macOS MainShellView](../../apps/macos/ZeroPass/ZeroPass/Views/Main/MainShellView.swift)
- [macOS ItemEditorView](../../apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift)

## Overview
- **Priority**: P0 — Primary feature: browse, create, edit, delete vault items
- **Status**: pending
- **Description**: 3-column layout (Sidebar + ItemList + ItemDetail), CRUD for all 8 item types

## Implementation Steps

- [ ] 1. Vault store (Zustand): items, selectedItemId, filter, sort, CRUD actions
- [ ] 2. Tauri commands: `commands/items.rs` — list, get, create, update, delete
- [ ] 3. Tauri commands: `commands/crypto.rs` — generatePassword, generatePassphrase, scorePassword
- [ ] 4. IPC bindings for items + crypto in commands.ts
- [ ] 5. VaultPage — 3-column CSS Grid (sidebar 220px | list 300px | detail 1fr)
- [ ] 6. Sidebar — Library (All, Favorites), Categories (8 types), Tags
- [ ] 7. ItemList + ItemListRow — sortable (name, date, type), filterable by category/tag
- [ ] 8. ItemDetail — field display with copy buttons, quick actions (edit, delete, favorite)
- [ ] 9. ItemEditor — full CRUD form supporting all 8 types (login, apikey, sshkey, note, creditcard, identity, passkey, custom)
- [ ] 10. Components: FieldRow, FlowLayout, QuickActionButton, EmptyState
- [ ] 11. Password generator in editor (length, options, passphrase mode)
- [ ] 12. Tests: vault-store, component tests, Rust command tests

## Success Criteria
- 3-column layout renders correctly
- All 8 item types can be created, edited, deleted
- Filtering by type, favorites, tags works
- Sorting by name/date works
- Password generator produces strong passwords
