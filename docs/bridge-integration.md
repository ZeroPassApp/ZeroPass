# Bridge Integration Notes (CGO / Swift)

This repo includes a CGO-based "bridge" that exposes core vault functionality over a C ABI.
The bridge is intended to be called from a host app (e.g. SwiftUI) that owns UI + lifecycle.

## Result model + memory management

All exported bridge functions return:

- `ZPResult.code` (int) — 0 for success, non-zero for error
- `ZPResult.data` (char*) — JSON payload on success for JSON-returning functions; may be NULL
- `ZPResult.error` (char*) — error string on failure; may be NULL

The caller must free returned strings using:
- `ZPFreeResult(result)` (frees both `data` and `error`)

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
