# Keyring Research Report

keyring v3.6.3 is production-ready for cross-platform OS keychain access in ZeroPass Tauri desktop app.

## Key Findings

### 1. Platform Support
- macOS/iOS: Native Keychain  
- Windows: Credential Manager
- Linux: keyutils + DBus Secret Service
- FreeBSD/OpenBSD: DBus Secret Service

### 2. Biometric Integration (CRITICAL)
**Biometric IS NOT controlled by keyring crate.** Platform OS handles it automatically:
- When calling entry.get_password() on macOS, OS checks if item is biometric-protected
- If yes: OS prompts for Touch ID/Face ID automatically before returning plaintext
- keyring unaware; receives plaintext post-auth
- No explicit biometric code needed in Rust

### 3. Thread Safety
| Platform | Concurrent Access |
|----------|-------------------|
| macOS    | ✅ Thread-safe    |
| Windows  | ❌ Single thread  |
| Linux    | ⚠️ Dedicated thread |

### 4. Tauri Integration
Directly embed keyring in Tauri commands:
```rust
#[tauri::command]
fn unlock_vault() -> Result<String> {
    Entry::new(\"zeropass\", \"vault_key\")?.get_password()
}
```

### 5. Alternatives
- security-framework: macOS-only, overkill for password storage
- tauri-plugin-stronghold: Encrypted DB, NOT for OS keychain
- Direct platform APIs: Unnecessary complexity

## Recommendation
**✅ Use keyring v3.6.3** - Production-ready, cross-platform, biometric auto-prompt on macOS, trivial Tauri integration. Covers 1000f vault key storage needs.

## Implementation
```toml
[dependencies]
keyring = { version = \"3\", features = [\"apple-native\", \"windows-native\", \"sync-secret-service\"] }
```

Store: `Entry::new(\"zeropass-vault\", \"default\")?.set_password(&key)?`
Retrieve: `Entry::new(\"zeropass-vault\", \"default\")?.get_password()?`
