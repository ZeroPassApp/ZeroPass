# Research Report: Rust FFI with Go C-Archive Libraries (libzeropass)

## Executive Summary

Rust FFI with Go C-archive requires: (1) **build.rs** linking arch-specific `.a` files, (2) **safe wrappers** managing CString/ZPResult memory, (3) **error conversion** from C to Rust Result<T>. Go runtime auto-initializes; system libraries (macOS: CoreFoundation, Security; Linux: pthread, dl) must be linked. Go goroutines are thread-safe from C—no extra sync needed.

## Key Findings

### 1. Build Configuration

```rust
// build.rs
use std::env;

fn main() {
    let target_arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap();
    let lib_suffix = match target_arch.as_str() {
        "aarch64" => "arm64",
        "x86_64" => "amd64",
        _ => panic!("unsupported arch: {}", target_arch),
    };
    
    let lib_dir = format!("{}/../../bridge",
        env::var("CARGO_MANIFEST_DIR").unwrap());
    
    println!("cargo:rustc-link-search=native={}", lib_dir);
    println!("cargo:rustc-link-lib=static=zeropass_{}", lib_suffix);
    
    if cfg!(target_os = "macos") {
        println!("cargo:rustc-link-lib=framework=CoreFoundation");
        println!("cargo:rustc-link-lib=framework=Security");
    }
    if cfg!(target_os = "linux") {
        println!("cargo:rustc-link-lib=dylib=pthread");
        println!("cargo:rustc-link-lib=dylib=dl");
    }
    
    println!("cargo:rerun-if-changed=../../bridge/libzeropass.h");
}
```

**Critical:** Link name includes arch suffix (zeropass_arm64, zeropass_amd64). File is libzeropass_arm64.a, but link as `static=zeropass_arm64`.

### 2. Safe FFI Wrapper Pattern

```rust
use std::ffi::CStr;

#[repr(C)]
pub struct ZPResult {
    data: *mut u8,
    error: *mut u8,
    code: i32,
}

pub fn unlock_vault(handle: i64, password: &str) -> Result<(), String> {
    let c_password = std::ffi::CString::new(password)
        .map_err(|_| "null byte in password")?;
    
    unsafe {
        let result = ZPUnlock(handle, c_password.as_ptr() as *mut _);
        let outcome = if result.code == 0 {
            Ok(())
        } else {
            let msg = if !result.error.is_null() {
                CStr::from_ptr(result.error as *const _)
                    .to_string_lossy()
                    .into_owned()
            } else {
                format!("error code {}", result.code)
            };
            Err(msg)
        };
        ZPFreeResult(result); // ALWAYS free, even on error
        outcome
    }
}

extern "C" {
    fn ZPUnlock(handle: i64, password: *mut u8) -> ZPResult;
    fn ZPFreeResult(r: ZPResult);
}
```

### 3. Memory Safety Rules

- **CString ownership:** Rust owns; C borrows. Use `as_ptr() as *mut _`.
- **ZPResult cleanup:** Use RAII guard to prevent use-after-free:

```rust
pub struct ZPGuard(ZPResult);
impl Drop for ZPGuard { fn drop(&mut self) { unsafe { ZPFreeResult(self.0) }; } }

pub fn safe_op(handle: i64) -> Result<String, String> {
    let guard = ZPGuard(unsafe { ZPSomeOp(handle) });
    if guard.0.code == 0 {
        let data = unsafe { 
            CStr::from_ptr(guard.0.data as *const _).to_string_lossy().into_owned()
        };
        Ok(data)
    } else {
        Err("failed".into())
    }
}
```

### 4. Thread Safety & Async

- Go is thread-safe from C; handles can be shared across threads.
- Tauri async: `#[tauri::command] async fn cmd() -> Result<T, String>` → `task::block_in_place(|| unsafe { ffi_call() })`.
- No mutex needed around individual calls unless vault teardown races with operations.

### 5. Error Handling

```rust
pub enum ZPError {
    FFI(String),
    JSON(serde_json::Error),
    NullByte,
}

pub fn list_items(handle: i64) -> Result<Vec<Item>, ZPError> {
    unsafe {
        let result = ZPListItems(handle, std::ptr::null_mut());
        let data = CStr::from_ptr(result.data as *const _).to_string_lossy();
        let items: Vec<Item> = serde_json::from_str(&data)?;
        ZPFreeResult(result);
        Ok(items)
    }
}
```

## System Dependencies

| OS | Libraries | Flags |
|---|---|---|
| macOS 14+ | CoreFoundation, Security, libc | `-framework CoreFoundation` |
| Linux | libc, pthread, dl | `-lpthread -ldl` |

## Recommendations

1. Wrap each ZP function in a type-safe Rust function returning Result<T>
2. Create helper: `fn zp_cstring(s: &str) -> Result<CString>` to handle null bytes
3. Keep build.rs minimal; per-arch logic uses env vars
4. Test integration with real libzeropass.a; mock Go stubs for unit tests
5. Document FFI boundaries: who owns memory

## Unresolved Questions

- Signal handling in Tauri when Rust calls Go (test kill -9 vs graceful shutdown)
- Performance: measure CString alloc overhead in hot paths
- Windows platform: does libzeropass build on MSVC? (currently macOS/Linux/amd64)

