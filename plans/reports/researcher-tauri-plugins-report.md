# Research Report: Tauri v2 Plugin Ecosystem for ZeroPass Desktop

## Executive Summary

All 7 required plugins are production-ready for Tauri v2 with versions **2.0.0**. Desktop platforms (macOS, Windows, Linux) are fully supported. **Clipboard auto-clear requires custom timer logic** (plugin handles read/write only). Global shortcuts use platform-independent API (Cmd→CommandOrControl on macOS). Plugins register via single `.plugin()` chain in Builder, sharing event system via Emitter.

---

## Plugin Inventory

| Plugin | Cargo Crate | NPM Package | v2 Version | macOS | Win | Linux |
|--------|-------------|-------------|------------|-------|-----|-------|
| **clipboard-manager** | `tauri-plugin-clipboard-manager` | `@tauri-apps/plugin-clipboard-manager` | 2.0.0 | ✓ | ✓ | ✓ |
| **global-shortcut** | `tauri-plugin-global-shortcut` | `@tauri-apps/plugin-global-shortcut` | 2.0.0 | ✓ | ✓ | ✓ |
| **notification** | `tauri-plugin-notification` | `@tauri-apps/plugin-notification` | 2.0.0 | ✓ | ✓ | ✓ |
| **dialog** | `tauri-plugin-dialog` | `@tauri-apps/plugin-dialog` | 2.0.0 | ✓ | ✓ | ✓ |
| **autostart** | `tauri-plugin-autostart` | `@tauri-apps/plugin-autostart` | 2.0.0 | ✓ | ✓ | ✓ |
| **store** | `tauri-plugin-store` | `@tauri-apps/plugin-store` | 2.0.0 | ✓ | ✓ | ✓ |
| **os** | `tauri-plugin-os` | `@tauri-apps/plugin-os` | 2.0.0 | ✓ | ✓ | ✓ |

---

## Key Findings

### Clipboard Manager
- **API**: `readText()`, `writeText()`, `writeHtml()`, `clear()`
- **Auto-clear**: No built-in auto-clear. Implement custom logic: `writeText() → setTimeout(clear, 5000)`
- **No capabilities** required (read/write always allowed)

### Global Shortcut
- **Register/Unregister**: `register('CommandOrControl+K', handler)`, `unregister()`
- **Modifiers**: `CommandOrControl` (Cmd on macOS, Ctrl elsewhere), `Shift`, `Alt`
- **Event State**: `ShortcutState.Pressed | Released`
- **No mobile** support (Android/iOS marked ✗)

### Notification
- **API**: `sendNotification({ title, body, sound })`, `isPermissionGranted()`, `requestPermission()`
- **Permission**: Requires `"notification:default"` in `capabilities/main.json`
- **Sound**: Platform-specific (macOS: system sounds via name, Windows/Linux: file paths)

### Dialog
- **APIs**: `open()`, `save()`, `message()`, `ask()`, `confirm()`
- **Filters**: `{ name, extensions: ['svg', 'png'] }` (extensions or MIME types)
- **File Access** (iOS 14+): `copy` mode copies file to sandbox, `scoped` mode uses security-scoped access
- **Options**: `multiple`, `directory`, `recursive`, `defaultPath`, `title`

### Store (Persistent KV)
- **Location**: `~/.config/{app-name}/app-data/` (or platform equivalent)
- **API**: `Store.load()`, `set(key, value)`, `get(key)`, `save()`
- **LazyStore**: Loads on first access, auto-saves
- **Interop**: JS and Rust sides can share stores via same path; values must be `serde_json::Value`

### Autostart
- **API**: `enable()`, `disable()`, `isEnabled()`
- **Config**: Builder accepts `args(["--flag"])` and `app_name("Custom")`
- **No mobile** support

### OS Plugin
- **API**: `platform()`, `arch()`, `version()`, `type()`, `family()`
- **Returns**: Enums like `darwin` for macOS, `win32` for Windows

---

## Integration Pattern (Rust Side)

```rust
fn main() {
    tauri::Builder::default()
        // Simple plugins (no builder)
        .plugin(tauri_plugin_clipboard_manager::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_os::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        
        // setup() for config plugins
        .setup(|app| {
            #[cfg(desktop)]
            {
                use tauri::Emitter;
                use tauri_plugin_global_shortcut::{Code, Modifiers, ShortcutState};
                
                app.handle().plugin(
                    tauri_plugin_global_shortcut::Builder::new()
                        .with_shortcuts(["CommandOrControl+K"])?
                        .with_handler(|app, _shortcut, event| {
                            if event.state == ShortcutState::Pressed {
                                let _ = app.emit("cmd-k-pressed", ());
                            }
                        })
                        .build(),
                )?;
            }
            Ok(())
        })
        
        .setup(|app| {
            #[cfg(desktop)]
            app.handle().plugin(
                tauri_plugin_autostart::Builder::new()
                    .args(["--flag"])
                    .app_name("ZeroPass")
                    .build()
            )?;
            Ok(())
        })
        
        .run(tauri::generate_context!())
        .expect("fatal error");
}
```

---

## Permission Requirements

**Notification only requires permissions**:
```json
// src-tauri/capabilities/main.json
{
  "permissions": ["notification:default"]
}
```

All other plugins use implicit permissions or no permissions.

---

## Error Handling & Notes

- **Clipboard**: No errors for `clear()` if already empty
- **Global Shortcut**: Conflicts detected at registration time; returns error on duplicate
- **Store**: Auto-loads on first use; manual `await store.load()` if needed
- **Dialog**: Returns `null` if user cancels
- **Notification**: Permission denied silently fails; always check `isPermissionGranted()`

---

## Unresolved Questions

- Clipboard auto-clear behavior on security-sensitive platforms (macOS sandbox impact)?
- Does global-shortcut conflict detection prevent app startup or defer until use?
