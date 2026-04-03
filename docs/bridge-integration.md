# Bridge Integration Notes (CGO / Swift)

This repo includes a CGO-based "bridge" that exposes core vault functionality over a C ABI.
The bridge is intended to be called from a host app (e.g. SwiftUI) that owns UI + lifecycle.

## Building the bridge static library (macOS)

Prereqs: Go, Xcode command line tools (for `clang`/`lipo`), and `make`.

```bash
# Universal (arm64 + x86_64) static library + header
make -C bridge build-universal
```

Outputs (by default in `bridge/`):
- `bridge/libzeropass.a` (universal)
- `bridge/libzeropass.h` (generated; includes bridge exports)
- `bridge/zp_bridge.h` (defines `ZPResult`)

Host app note (SwiftUI):
- The macOS Xcode target runs `make -C bridge build-universal OUT_DIR="$(TARGET_TEMP_DIR)/zeropass_bridge"` automatically during builds.
- The build phase also normalizes `PATH` for Xcode GUI builds so Homebrew Go (e.g. `/opt/homebrew/bin/go`) is discoverable.

## Result model + memory management

All exported bridge functions return:

- `ZPResult.code` (int) — 0 for success, non-zero for error
- `ZPResult.data` (char*) — JSON payload on success for JSON-returning functions; may be NULL
- `ZPResult.error` (char*) — error string on failure; may be NULL

The caller must free returned strings using:
- `ZPFreeResult(result)` — frees both `data` and `error`
- `ZPFreeResultPtr(&result)` — same, but also NULLs fields and resets `code`

`ZPFree*` wipes the C string contents (via `memset`) before freeing to reduce secret remanence.

## Error codes

Defined in `bridge/helpers.go`:

| Code | Name | Meaning | Typical handling |
|---:|---|---|---|
| 0 | OK | Success | Continue |
| 1 | Locked | Vault is locked | Prompt to unlock; call `ZPUnlock*` |
| 2 | NotFound | Missing handle / missing file | Treat as “not found”; refresh state |
| 3 | AuthFailed | Bad password/key or unauthorized sync request | Re-prompt credentials / re-auth |
| 4 | Internal | Any other error | Show error; log details |
| 5 | Busy | Vault already open elsewhere | Retry/backoff or ask user to close other instance |

## Advisory locking (`vault.lock`) and Busy

On `ZPCreateVault()` and `ZPOpenVault()`, the bridge acquires a **non-blocking exclusive advisory lock**
using `flock()` on:

`{vaultPath}/vault.lock`

If the lock cannot be acquired, the call fails with **code=5 (Busy)**.

Operational notes:
- The lock is held for the lifetime of the session and is released on `ZPCloseVault()` (or process exit).
- The lock is advisory; other tools must also respect it to avoid concurrent mutation.

Recommended host behavior on Busy:
- Treat as retryable.
- Use exponential backoff with a max wait, then prompt the user.

## Auto-lock behavior in the bridge

The core vault supports an internal auto-lock timer, but the bridge disables it at runtime by calling:
`DisableAutoLock()` during create/open.

This is **runtime-only** (not persisted). The host app should implement inactivity tracking and call:
- `ZPLock(handle)` to lock, and/or
- `ZPCloseVault(handle)` to fully close and release `vault.lock`

## Version history at rest

Item version history is stored at:
`items/{item_id}.versions.json`

Important details:
- The on-disk history is **encrypted at rest** (per-item key derived from vault key + item ID).
- Legacy plaintext history files are migrated in-place when read.

## Concurrency model (host-side)

Bridge calls are serialized per vault handle via a mutex inside the session.
Avoid long-running calls on the UI thread; dispatch to a background queue.

## macOS app (SwiftUI) integration

### Import sources

The macOS app supports importing from all 9 sources via a Settings dropdown
(`SecuritySettingsView`). Each source maps to a bridge function:

| Source | Bridge function | File format |
|--------|----------------|-------------|
| Chrome | `ZPImportChrome` | CSV |
| Firefox | `ZPImportFirefox` | CSV |
| Safari | `ZPImportSafari` | CSV |
| 1Password | `ZPImport1Password` | CSV |
| 1PUX | `ZPImport1PUX` | 1PUX archive |
| Bitwarden | `ZPImportBitwarden` | CSV/JSON |
| LastPass | `ZPImportLastPass` | CSV |
| KeePass | `ZPImportKeePass` | CSV |
| Generic CSV | `ZPImportCSV` | CSV |

The Swift-side enum `VaultClient.ImportSource` dispatches to the correct bridge call.

### Export

CSV and JSON exports require the user to type **EXPORT** in a confirmation dialog
(`SecuritySettingsView`). This prevents accidental plaintext credential exposure.

### Accessibility

The macOS app includes:
- **VoiceOver labels** — `.accessibilityLabel()` on all interactive elements (sidebar items, toolbar buttons, item rows, field actions)
- **Accessibility hints** — `.accessibilityHint()` on list items and search results
- **Keyboard navigation** — `@FocusState` on `CreateVaultView`, `UnlockVaultView`, `ItemEditorView`, `QuickSearchView`
- **High contrast mode** — Toggle in General Settings (`@AppStorage("highContrastMode")`) applies `.contrast(0.15)` and `.environment(\.legibilityWeight, .bold)` app-wide
