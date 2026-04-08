# ZeroPass Tauri Desktop App

**Status**: In Progress  
**Path**: `apps/desktop/`  
**Stack**: Tauri v2 + React 19 + TypeScript + Vite + Rust + Go core (C FFI)  
**Branch**: `feat/desktop-app`

## Phases

| # | Phase | Status | File |
|---|-------|--------|------|
| 01 | Project Scaffold + Rust Bridge | pending | [phase-01](phase-01-project-scaffold-rust-bridge.md) |
| 02 | Auth Flow | pending | [phase-02](phase-02-auth-flow.md) |
| 03 | Vault Browser | pending | [phase-03](phase-03-vault-browser.md) |
| 04 | Search + Quick Search | pending | [phase-04](phase-04-search-quick-search.md) |
| 05 | Settings + Platform Features | pending | [phase-05](phase-05-settings-platform-features.md) |
| 06 | Health + Sync + Polish + E2E | pending | [phase-06](phase-06-health-sync-polish-e2e.md) |

## Key Decisions

- **FFI**: Reuse existing `libzeropass_<arch>.a` — Rust wraps C functions via safe bridge
- **Biometric**: `keyring` crate v3.6.3 — OS auto-prompts biometric on keychain access
- **E2E**: Playwright (WebDriverIO deprecated for Tauri)
- **Memory**: RAII `ZPGuard` Drop wrapper for ZPResult cleanup
- **Errors**: `thiserror` + `serde::Serialize` for structured IPC errors
- **Build**: Per-arch, manual (`npm run tauri build`)

## Dependencies

- Go core bridge: `bridge/libzeropass_arm64.a`, `bridge/libzeropass_amd64.a`
- macOS app: `apps/macos/` kept as reference (untouched)

## Reports

- [Tauri + React](../reports/researcher-tauri-react-report.md)
- [Rust FFI](../reports/researcher-rust-ffi-report.md)
- [Keyring Biometric](../reports/researcher-keyring-biometric-report.md)
- [Tauri Plugins](../reports/researcher-tauri-plugins-report.md)
- [Testing Strategy](../reports/researcher-testing-strategy-report.md)
