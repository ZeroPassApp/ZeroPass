# Phase 01: Project Scaffold + Rust Bridge

## Context Links
- [plan.md](plan.md)
- [Rust FFI Report](../reports/researcher-rust-ffi-report.md)
- [Tauri React Report](../reports/researcher-tauri-react-report.md)
- [Tauri Plugins Report](../reports/researcher-tauri-plugins-report.md)
- [bridge/zp_bridge.h](../../bridge/zp_bridge.h) — ZPResult struct
- [bridge/libzeropass.h](../../bridge/libzeropass.h) — 47 exported C functions

## Overview
- **Priority**: P0 — Foundation for all subsequent phases
- **Status**: pending
- **Description**: Scaffold Tauri v2 project with React 19 + TypeScript, write Rust FFI bridge wrapping all Go core C functions

## Key Insights
- `libzeropass_<arch>.a` already built per-arch (arm64/amd64)
- ZPResult: `{data: *mut char, error: *mut char, code: i32}` — must free with `ZPFreeResult`
- Use RAII `ZPGuard` with `impl Drop` for automatic cleanup
- CString: Rust owns, C borrows via `as_ptr() as *mut _`
- Go runtime auto-initializes when linked as C-archive
- Async Tauri commands: use `tokio::task::spawn_blocking` for FFI calls

## Requirements
- Tauri v2 project compiles and links libzeropass successfully
- All 47 C functions wrapped in safe Rust bridge
- Structured error types (thiserror + Serialize)
- Tauri plugins registered (clipboard, dialog, notification, shortcut, autostart, store, os)
- `cargo build` passes; `npm run tauri dev` launches window

## Architecture
```
apps/desktop/
├── src/                    # React (empty shell for now)
│   ├── main.tsx
│   └── App.tsx
├── src-tauri/
│   ├── Cargo.toml
│   ├── tauri.conf.json
│   ├── capabilities/default.json
│   ├── build.rs            # Links libzeropass_<arch>.a
│   └── src/
│       ├── main.rs
│       ├── lib.rs
│       ├── bridge.rs       # Safe FFI wrappers (47 functions)
│       ├── error.rs        # BridgeError enum
│       └── state.rs        # VaultState managed state
├── package.json
├── vite.config.ts
├── tsconfig.json
└── index.html
```

## Related Code Files
- **Create**: All files under `apps/desktop/`
- **Read**: `bridge/zp_bridge.h`, `bridge/libzeropass.h`, `bridge/Makefile`

## Implementation Steps

- [ ] 1. Create feature branch `feat/desktop-app` from `main`
- [ ] 2. Scaffold Tauri v2 project: `npm create tauri-app` with React 19 + TypeScript
- [ ] 3. Configure `package.json` — add deps (zustand, react-router, etc.)
- [ ] 4. Configure `vite.config.ts` — dev server port 1420
- [ ] 5. Configure `tsconfig.json` — strict mode, paths
- [ ] 6. Configure `tauri.conf.json` — window 1200x800, title ZeroPass, CSP, bundler
- [ ] 7. Configure `capabilities/default.json` — permissions for all plugins
- [ ] 8. Write `build.rs` — per-arch linking of libzeropass, system libs, tauri_build
- [ ] 9. Write `Cargo.toml` — deps (serde, thiserror, keyring, tauri plugins)
- [ ] 10. Write `src-tauri/src/error.rs` — BridgeError enum with thiserror + Serialize
- [ ] 11. Write `src-tauri/src/bridge.rs` — safe FFI wrappers for all 47 C functions with ZPGuard
- [ ] 12. Write `src-tauri/src/state.rs` — VaultState (handle, lock status)
- [ ] 13. Write `src-tauri/src/lib.rs` — module declarations + Tauri builder with plugins
- [ ] 14. Write `src-tauri/src/main.rs` — entry point  
- [ ] 15. Write minimal React shell: `main.tsx`, `App.tsx`, `index.html`
- [ ] 16. Verify `cargo build` compiles and links — fix any linker errors
- [ ] 17. Verify `npm run tauri dev` launches window with React shell

## Success Criteria
- `cargo build` succeeds with libzeropass linked
- `npm run tauri dev` opens a window showing React shell
- All 47 bridge functions have safe Rust wrappers
- No unsafe code outside `bridge.rs`
- Error types serialize for IPC transport

## Risk Assessment
- **Linker errors**: Go C-archive may need additional system libs → check `go build -v` output
- **Arch mismatch**: Must match CARGO_CFG_TARGET_ARCH to Go lib suffix
- **Go runtime conflicts**: Tauri uses Tokio; Go has its own scheduler → test under load

## Security Considerations
- ZPResult memory freed via RAII Drop — no leaks
- CString zeroed where sensitive (passwords) before drop
- No plaintext logged from bridge calls
