# Phase 04: Search + Quick Search

## Context Links
- [plan.md](plan.md) | [phase-03](phase-03-vault-browser.md)
- [macOS QuickSearchPanel](../../apps/macos/ZeroPass/ZeroPass/Views/Search/QuickSearchPanel.swift)

## Overview
- **Priority**: P1
- **Status**: pending
- **Description**: Toolbar search in ItemList + Cmd+K global quick search panel

## Implementation Steps

- [ ] 1. Search store (Zustand): query, results, loading
- [ ] 2. Tauri commands: `commands/search.rs` — search, rebuildIndex
- [ ] 3. IPC bindings for search in commands.ts
- [ ] 4. Toolbar search input in ItemList header
- [ ] 5. QuickSearch floating panel (Cmd+K) — overlay with input + results
- [ ] 6. SearchResultRow — item type icon, name, highlight match
- [ ] 7. Global shortcut: register CommandOrControl+K via tauri-plugin-global-shortcut
- [ ] 8. Tauri event: emit `cmd-k-pressed` from Rust → listen in React
- [ ] 9. Tests: search-store, QuickSearch component, Rust search command

## Success Criteria
- Toolbar search filters items in real-time
- Cmd+K opens floating search panel from anywhere
- Results show with type icons and name highlighting
- Selecting a result navigates to item detail
