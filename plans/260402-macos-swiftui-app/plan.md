---
title: "ZeroPass macOS SwiftUI Native App"
description: "Build native macOS app using SwiftUI + Go c-archive bridge for the ZeroPass credential manager"
status: pending
priority: P1
effort: 28d
branch: feat/macos-app
tags: [macos, swiftui, frontend, security, ffi]
created: 2026-04-02
---

# ZeroPass macOS SwiftUI Native App

## Overview

Build a **native macOS application** using SwiftUI that serves as the graphical interface for ZeroPass. The app reuses 100% of the existing Go core (crypto, vault, index, sync) via a **Go c-archive static library** with JSON-over-FFI bridge — the same proven pattern used by WireGuard (Go+Swift) and analogous to 1Password 8 (Rust+Swift).

**Key architectural decisions (from research):**
- **Go `c-archive`** (static library) — not c-shared, not subprocess, not daemon
- **JSON-over-FFI** protocol — all complex data crosses boundary as JSON
- **Handle-based API** — Go-side `map[int64]*Vault`, Swift gets opaque handles
- **Direct distribution** (DMG + Homebrew cask) — NOT Mac App Store (sandbox blocks vault access)
- **@Observable + NavigationSplitView** — modern SwiftUI, no legacy patterns
- **Shared vault** — CLI and GUI operate on same `~/.zeropass/` directory

## Research Reports

- [SwiftUI macOS Research](./research/swiftui-macos-research.md)
- [Go↔Swift Interop Research](./research/go-swift-interop-research.md)
- [Gap Analysis](./reports/gap-analysis-260402.md)
- [**UI/UX Design Guideline**](./reports/ui-ux-design-guideline.md) — comprehensive design standards based on Apple HIG, 1Password, Things 3, Raycast

## UI/UX Design Standards

> **MANDATORY**: All UI phases (3-8) MUST follow the [UI/UX Design Guideline](./reports/ui-ux-design-guideline.md). The guideline defines exact specs for typography, colors, spacing, animations, keyboard shortcuts, accessibility, and security UX patterns.

### Key Design Rules (Quick Reference)

| Rule | Spec | Guideline Section |
|------|------|-------------------|
| Body text | **13pt** (NOT 17pt iOS) | §4 Typography |
| Table row height | **24pt** compact | §9 Main Layout |
| Sidebar width | **200-250pt**, collapsible | §3 Window & Layout |
| Default window | **1024×680pt** (min 800×500) | §3 Window & Layout |
| All icons | **SF Symbols only** (no emojis) | §6 Iconography |
| All colors | **Semantic only** (`.primary`, `.secondary`) | §5 Color System |
| Sidebar material | **`.regularMaterial`** | §5 Color System |
| Animation max | **300ms** (macOS faster than iOS) | §14 Animation |
| Keyboard shortcuts | **Full menu bar + ⌘ shortcuts** for every action | §15 Keyboard Shortcuts |
| Accessibility | **VoiceOver labels**, 4.5:1 contrast, reduced motion | §16 Accessibility |
| Item type colors | Blue/purple/green/yellow/orange/teal/indigo/gray | §5 Color System |
| Grid system | **8pt grid** | Appendix B Design Tokens |

### macOS Menu Bar (REQUIRED — Phase 2+)

```
ZeroPass  File  Edit  View  Item  Window  Help
```

Full menu structure defined in [Guideline §15](./reports/ui-ux-design-guideline.md#15-keyboard-shortcuts). Every toolbar action MUST also appear in menu bar.

## Phases

## Pre-Implementation Requirements (from Gap Analysis)

These Go core changes MUST be completed before Phase 1:

1. **Implement `ChangeMasterPassword(oldPassword, newPassword string) error`** in `core/vault/store/store.go` — needed by Phase 8 Settings
2. **Implement `UnlockWithKey(vaultKey []byte) error`** in `core/vault/store/store.go` — needed by Phase 3 TouchID (bypasses KDF)
3. **Update PRD deployment target** from macOS 13+ to **macOS 14+** (required for @Observable)

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 0 | Go Core Prereqs (above) | Pending | 1d | — |
| 1 | Go Bridge Layer | Pending | 4d | [phase-01](./phase-01-go-bridge-layer.md) |
| 2 | Xcode Project + Swift Bridge | Pending | 2d | [phase-02](./phase-02-xcode-project-setup.md) |
| 3 | Authentication Views | Pending | 2d | [phase-03](./phase-03-authentication-views.md) |
| 4 | Main UI Layout | Pending | 3d | [phase-04](./phase-04-main-ui-layout.md) |
| 5 | Item CRUD Views | Pending | 4d | [phase-05](./phase-05-item-crud-views.md) |
| 6 | Search & Quick Access | Pending | 2.5d | [phase-06](./phase-06-search-quick-access.md) |
| 7 | Platform Integration | Pending | 3d | [phase-07](./phase-07-platform-integration.md) |
| 8 | Settings & Preferences | Pending | 1.5d | [phase-08](./phase-08-settings-preferences.md) |
| 9 | Build & Distribution Pipeline | Pending | 2d | [phase-09](./phase-09-build-distribution.md) |
| 10 | Testing & QA | Pending | 2d | [phase-10](./phase-10-testing.md) |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  ZeroPass.app (SwiftUI)                             │
│  ┌───────────────────────────────────────────────┐  │
│  │  UI Layer (SwiftUI Views)                     │  │
│  │  NavigationSplitView, Table, Forms, MenuBar   │  │
│  └──────────────────┬────────────────────────────┘  │
│                     │                                │
│  ┌──────────────────▼────────────────────────────┐  │
│  │  VaultManager (@Observable)                   │  │
│  │  State management, UI logic, async dispatch   │  │
│  └──────────────────┬────────────────────────────┘  │
│                     │                                │
│  ┌──────────────────▼────────────────────────────┐  │
│  │  GoBridge (Swift wrapper)                     │  │
│  │  JSON encode/decode, C function calls, memory │  │
│  └──────────────────┬────────────────────────────┘  │
│                     │ C ABI (static link)            │
│  ┌──────────────────▼────────────────────────────┐  │
│  │  libzeropass.a (Go c-archive)                 │  │
│  │  All crypto, vault, search, sync, import      │  │
│  │  Existing Go core — ZERO rewrite              │  │
│  └───────────────────────────────────────────────┘  │
│                                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │  Native macOS Services (Swift)                │  │
│  │  Keychain · TouchID · Clipboard · MenuBar     │  │
│  │  Global Hotkey · Notifications · Sparkle      │  │
│  └───────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
         │
         ▼
  ~/.zeropass/vaults/default/  (shared with CLI)
```

## File Ownership Matrix

/Volumes/DATA/Developments/ZeroPass/apps/macos/ZeroPass

```
bridge/                     → Phase 1 (Go)
macos/ZeroPass/Bridge/      → Phase 2 (Swift bridge)
macos/ZeroPass/Views/Auth/  → Phase 3
macos/ZeroPass/Views/Main/  → Phase 4
macos/ZeroPass/Views/Items/ → Phase 5
macos/ZeroPass/Views/Search/→ Phase 6
macos/ZeroPass/Services/    → Phase 7
macos/ZeroPass/Views/Settings/ → Phase 8
macos/scripts/              → Phase 9
macos/ZeroPassTests/        → Phase 10
```

## Dependencies

- Go 1.26+ (existing)
- Xcode 16+ (macOS 15 SDK)
- macOS 14+ deployment target (Sonoma — for @Observable, NavigationSplitView maturity)
- Swift Package Manager (Sparkle, LaunchAtLogin)
- Apple Developer ID certificate (for signing/notarization)

## Design References

| App | Pattern to Follow |
|-----|-------------------|
| **1Password 8** | Vault-colored accents, reveal animations, Watchtower security UI |
| **Apple Passwords** | Minimal chrome, modal-first password display, system integration |
| **Things 3** | Perfect sidebar + list + detail layout, checkbox animations |
| **Raycast** | Spotlight search UX, stagger animations, keyboard-first |
| **Bear** | Tag-based organization, inline editing |
| **Fantastical** | Information density, natural language input |

## Decisions (Confirmed)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Bridge approach | **c-archive (static library)** | WireGuard precedent, no dylib signing complexity |
| Deployment target | **macOS 14 (Sonoma)** | @Observable + mature NavigationSplitView |
| Universal binary | **Yes (arm64 + x86_64)** | Cover all Mac users, ~2x lib size acceptable |
| TouchID phase | **Phase 3 (with Auth views)** | Core UX feature, integrate early |
| Concurrent access | **flock() file locking** | CLI + GUI may open same vault |

## Key Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Go c-archive + Hardened Runtime | May need `allow-unsigned-executable-memory` entitlement | WireGuard precedent; test in Phase 2 (not Phase 9!) |
| CLI/GUI concurrent vault access | Vault data corruption | `flock()` advisory lock on `vault.lock` file — Phase 1 |
| Password memory in Swift | Swift strings are immutable, may be cached | Use `Data` + `UnsafeMutableBufferPointer` for sensitive values |
| Universal binary size | ~15MB for Go static lib | Acceptable for desktop app |
| Go auto-lock vs Swift auto-lock | Double-lock conflict, unexpected key zeroing | Disable Go-side auto-lock in bridge (set `AutoLockTimeout=0`); Swift handles auto-lock |
| ZPCreateVault mnemonic in C-heap | Recovery mnemonic lingers in malloc'd memory | Swift must `ZPFreeResult` immediately + zero mnemonic copy |
| No `ChangeMasterPassword` in Go core | Phase 8 Settings blocked | Prereq: implement in `store.go` before Phase 1 |
