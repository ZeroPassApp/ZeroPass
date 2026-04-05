---
title: "ZeroPass Midnight Native macOS Redesign"
description: "Implementation plan to restyle the existing SwiftUI macOS app around a premium native dark utility direction."
status: in-progress
priority: P1
effort: 10d
branch: main
tags: [macos, swiftui, redesign, ui, midnight-native]
created: 2026-04-04
---

# ZeroPass Midnight Native macOS Redesign

## Outcome
- Redesign the existing macOS SwiftUI app to match the chosen **ZeroPass Midnight Native** direction: native macOS first, dark graphite surfaces, cobalt accent, stronger hierarchy, less dead space, premium utility feel.
- Use the existing SwiftUI/AppKit structure as the implementation base; Stitch concepts for welcome, unlock, workspace, settings, quick search, recovery phrase, and item editor are reference inputs, not a new architecture.

## Delivery Rules
- Prefer updating existing files over introducing new abstractions.
- Keep core composition intact: `ContentView` auth state switching, `NavigationSplitView`, settings tabs, quick-search panel, and current `VaultClient` flows.
- Only extract small helper views if it is the cleanest way to keep `ItemDetailView.swift` or `ItemEditorView.swift` maintainable.

## Phases

| # | Phase | Effort | Purpose | Link |
|---|-------|--------|---------|------|
| 1 | Midnight Token System + Shared Chrome | 1d | Establish graphite/cobalt tokens and shared surface rules | [phase-01](./phase-01-midnight-token-system.md) |
| 2 | Auth Screens | 1.5d | Welcome, unlock, create/open vault, and auth recovery presentation | [phase-02](./phase-02-auth-screens.md) |
| 3 | Workspace Shell + Detail Hierarchy | 2d | Main shell, sidebar, list, detail, toolbar, and toast polish | [phase-03](./phase-03-workspace-shell-and-detail-hierarchy.md) |
| 4 | Settings Redesign | 1.5d | General, Security, Sync, About, and settings-adjacent sheets | [phase-04](./phase-04-settings-redesign.md) |
| 5 | Quick Search + Menu Bar | 1d | Floating search panel and menu bar extra refresh | [phase-05](./phase-05-quick-search-and-menu-bar.md) |
| 6 | Item Editor + Recovery Utilities | 1.5d | Editor polish plus in-app recovery regeneration surfaces | [phase-06](./phase-06-item-editor-and-recovery-utilities.md) |
| 7 | Validation + Regression | 1d | UI tests, manual QA, and repo-level verification commands | [phase-07](./phase-07-validation-and-regression.md) |
| 8 | Docs Sync + Follow-Through | 0.5d | Changelog, roadmap, and superseded plan/doc alignment | [phase-08](./phase-08-docs-sync-and-follow-through.md) |

## Primary Files Likely Touched
- `apps/macos/ZeroPass/ZeroPass/Views/Components/ZeroPassTheme.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Auth/*`
- `apps/macos/ZeroPass/ZeroPass/Views/Main/*`
- `apps/macos/ZeroPass/ZeroPass/Views/Settings/*`
- `apps/macos/ZeroPass/ZeroPass/Views/Search/*`
- `apps/macos/ZeroPass/ZeroPass/Views/MenuBar/MenuBarView.swift`
- `apps/macos/ZeroPass/ZeroPass/Views/Editors/ItemEditorView.swift`
- `apps/macos/ZeroPass/ZeroPass/Models/VaultItemType+UI.swift`
- `apps/macos/ZeroPass/ZeroPassTests/*`
- `apps/macos/ZeroPass/ZeroPassUITests/*`

## Validation Commands
- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- `go test ./...`
- `go vet ./...`
- `go build -o zp ./packages/cli/`
- `go build -o syncserver ./services/syncserver/`
- `bash services/syncserver/smoke-test.sh both`

## Success Shape
- The redesign reads as one coherent product, not a mix of old system defaults and new premium surfaces.
- Auth, workspace, settings, quick search, menu bar, editor, and recovery utilities all share the same visual hierarchy and accent language.
- No vault behavior, security posture, or existing repo validation flows regress while the UI changes land.