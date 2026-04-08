# Research Report: Tauri v2 + React 19 Integration

## Executive Summary

Tauri v2 (v2.10+) is production-ready for desktop apps with strong security foundation (Rust backend) and minimal bundle size (<600KB). React 19 introduces `use()` hook for async patterns, ideal for IPC calls. **Recommendation:** Use typed IPC commands over events; leverage React Suspense + ErrorBoundary for async operations; scaffold via `create-tauri-app`, skip legacy v1 patterns.

## Key Findings

### 1. Tauri v2 Scaffolding & Project Structure

**Correct command:**
```bash
npm create tauri-app@latest -- --template react --typescript
cd src-tauri && cargo add serde serde_json
```

**Project layout:**
```
├── src/                  # React frontend (src/main.tsx, components/)
├── src-tauri/            # Rust backend
│   ├── src/lib.rs        # Entry point (tauri::command macros)
│   ├── src/commands.rs   # Modular command definitions
│   ├── Cargo.toml        # Rust deps (serde, tokio, thiserror)
│   └── tauri.conf.json   # Window, bundle, security config
├── vite.config.ts        # Dev server (devUrl: http://localhost:5173)
└── tsconfig.json
```

**tauri.conf.json v2 essentials:**
- `build.devUrl`: Points to Vite dev server (http://localhost:5173)
- `build.beforeDevCommand`: `npm run dev`
- `app.windows[0].label`: "main" (for WebviewWindow access)
- `security.csp`: Enable for XSS protection; disable in dev

### 2. Tauri v2 IPC Patterns (Type-Safe)

**Commands (preferred for type safety):**
```rust
// src-tauri/src/commands.rs
use serde::Deserialize;

#[derive(Deserialize)]
pub struct LoginPayload {
    pub username: String,
    pub password: String,
}

#[tauri::command]
pub async fn login(payload: LoginPayload) -> Result<String, String> {
    // Validate credentials against Go bridge or local vault
    if payload.username == "user" && payload.password == "pass" {
        Ok("session_token_here".to_string())
    } else {
        Err("Invalid credentials".to_string())
    }
}
```

**React 19 invocation with Suspense:**
```typescript
// src/hooks/useLogin.ts
import { invoke } from '@tauri-apps/api/core';
import { use, Suspense } from 'react';

export function useLogin(username: string, password: string) {
    const promise = invoke<string>('login', { username, password });
    return use(promise); // Works with Suspense + ErrorBoundary
}

// src/components/LoginForm.tsx
import { Suspense } from 'react';

function LoginComponent() {
    const token = useLogin('user', 'pass');
    return <p>Logged in: {token}</p>;
}

export default function LoginPage() {
    return (
        <ErrorBoundary fallback={<p>Login failed</p>}>
            <Suspense fallback={<p>Authenticating...</p>}>
                <LoginComponent />
            </Suspense>
        </ErrorBoundary>
    );
}
```

**Events (use for real-time streams, NOT commands):**
- Commands: 1-request/1-response, type-safe, preferred
- Events: pub/sub, JSON-only, for progress updates, file watches
- **Do NOT mix:** Each IPC pattern has clear boundary

**Error handling pattern:**
```rust
#[derive(thiserror::Error, Debug, serde::Serialize)]
#[serde(tag = "kind", content = "message")]
pub enum CommandError {
    #[error("vault not found")]
    VaultNotFound,
    #[error("invalid password")]
    InvalidPassword,
}

#[tauri::command]
pub fn unlock_vault(password: String) -> Result<Vec<u8>, CommandError> {
    // Returns { kind: "invalid_password", message: "..." } on error
}
```

### 3. Tauri v2 Breaking Changes from v1

- **Capability system:** `/src-tauri/capabilities/default.json` required; define permissions (file, fs, http)
- **Window API:** `WebviewWindow` replaces `Window`; access via `tauri::WebviewWindow` in commands
- **State management:** `tauri::State<T>` still works; use `.manage()` on builder
- **AppHandle:** Required for global shortcuts, menus; pass to commands explicitly

### 4. React 19 Desktop-Specific Features

**`use()` hook with IPC:**
- Suspends component while Promise resolves
- Works with Tauri `invoke()` promises
- Can be called in conditionals (unlike older hooks)

**Concurrent rendering implications:**
- Not blocked by IPC; React reschedules during network wait
- Use ErrorBoundary + Suspense for loading states (no try/catch)

**Routing recommendation:**
- TanStack Router: File-based, better error handling
- React Router v7: More stable, broader ecosystem
- Both work; TanStack Router newer

### 5. Build Configuration

**Cargo.toml essentials:**
```toml
[dependencies]
tauri = { version = "2.10", features = ["protocol-asset"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
tokio = { version = "1", features = ["full"] }
thiserror = "1"

[dev-dependencies]
tauri-build = "2.5"  # Compile-time config generation
```

**Vite config for mobile-safe dev server:**
```typescript
export default defineConfig({
    server: {
        host: process.env.TAURI_DEV_HOST || false,
        port: 1420,
        hmr: process.env.TAURI_DEV_HOST 
            ? { host: process.env.TAURI_DEV_HOST, port: 1421 }
            : undefined,
    },
});
```

**Cross-compilation:** Uses Rust's target system (`cargo build --target aarch64-apple-darwin` for M1 macOS). More info: [Tauri bundler docs](https://tauri.app/develop/).

## Recommendations for ZeroPass

1. **IPC strategy:** Go bridge (via CGO to libzeropass.a) ↔ Tauri commands → React. **NOT** direct Go-to-React.
2. **Type safety:** Use `#[derive(serde::Serialize, serde::Deserialize)]` on all payloads; generate TypeScript types via `specta` crate (optional advanced pattern).
3. **Async ops:** Vault unlock, item decrypt → async commands; return `Result<T, Error>`.
4. **Security:** No plaintext in logs; zero sensitive data after use (Rust's `defer_zero_bytes()`); enable CSP in prod.
5. **State:** Use Tauri `State<VaultSession>` for unlocked vault; NOT React state.

## Unresolved Questions

- Specta for auto-generating TypeScript from Rust return types?
- CGO bridge performance implications for high-frequency IPC calls?

---
**Sources:** tauri.app v2 docs (Feb 2026), react.dev v19.2, Tauri GitHub examples  
**Date:** April 8, 2026  
**Confidence:** High (production-ready patterns)
