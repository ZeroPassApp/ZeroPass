use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use crate::error::BridgeError;

// ─── FFI declarations ────────────────────────────────────────────────

#[repr(C)]
pub struct ZPResult {
    pub data: *mut c_char,
    pub error: *mut c_char,
    pub code: i32,
}

extern "C" {
    // Vault lifecycle
    pub fn ZPCreateVault(path: *mut c_char, master_password: *mut c_char) -> ZPResult;
    pub fn ZPOpenVault(path: *mut c_char) -> ZPResult;
    pub fn ZPUnlock(handle: i64, master_password: *mut c_char) -> ZPResult;
    pub fn ZPUnlockWithRecovery(handle: i64, mnemonic: *mut c_char) -> ZPResult;
    pub fn ZPUnlockWithKey(handle: i64, vault_key_base64: *mut c_char) -> ZPResult;
    pub fn ZPGetVaultKey(handle: i64) -> ZPResult;
    pub fn ZPLock(handle: i64) -> ZPResult;
    pub fn ZPCloseVault(handle: i64) -> ZPResult;
    pub fn ZPIsLocked(handle: i64) -> i32;
    pub fn ZPChangeMasterPassword(
        handle: i64,
        old_password: *mut c_char,
        new_password: *mut c_char,
    ) -> ZPResult;

    // Item CRUD
    pub fn ZPListItems(handle: i64, filter_json: *mut c_char) -> ZPResult;
    pub fn ZPGetItem(handle: i64, item_id: *mut c_char) -> ZPResult;
    pub fn ZPCreateItem(handle: i64, item_json: *mut c_char) -> ZPResult;
    pub fn ZPUpdateItem(handle: i64, item_id: *mut c_char, item_json: *mut c_char) -> ZPResult;
    pub fn ZPDeleteItem(handle: i64, item_id: *mut c_char) -> ZPResult;

    // Search
    pub fn ZPSearch(handle: i64, query: *mut c_char) -> ZPResult;
    pub fn ZPRebuildIndex(handle: i64) -> ZPResult;

    // Crypto
    pub fn ZPGeneratePassword(length: i32, options_json: *mut c_char) -> ZPResult;
    pub fn ZPGeneratePassphrase(words: i32, separator: *mut c_char) -> ZPResult;
    pub fn ZPScorePassword(password_str: *mut c_char) -> ZPResult;
    pub fn ZPCheckBreach(password_str: *mut c_char) -> ZPResult;

    // Health
    pub fn ZPAnalyzeHealth(handle: i64) -> ZPResult;

    // Import
    pub fn ZPImportCSV(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportChrome(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportFirefox(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImport1Password(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportBitwarden(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportSafari(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportLastPass(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImportKeePass(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPImport1PUX(handle: i64, path: *mut c_char) -> ZPResult;

    // Export
    pub fn ZPExportJSON(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPExportCSV(handle: i64, path: *mut c_char) -> ZPResult;
    pub fn ZPExportEncrypted(handle: i64, path: *mut c_char) -> ZPResult;

    // Recovery
    pub fn ZPRegenerateRecovery(handle: i64) -> ZPResult;

    // Sync
    pub fn ZPSyncSetup(handle: i64, config_json: *mut c_char) -> ZPResult;
    pub fn ZPSyncRegister(handle: i64, device_name: *mut c_char) -> ZPResult;
    pub fn ZPSyncPull(handle: i64) -> ZPResult;
    pub fn ZPSyncPush(handle: i64) -> ZPResult;
    pub fn ZPSyncFull(handle: i64) -> ZPResult;

    // Version history
    pub fn ZPGetVersionHistory(handle: i64, item_id: *mut c_char) -> ZPResult;
    pub fn ZPRestoreVersion(handle: i64, item_id: *mut c_char, version: i32) -> ZPResult;

    // Memory management
    pub fn ZPFree(ptr: *mut c_char);
    pub fn ZPFreeResult(r: ZPResult);
}

// ─── RAII guard for ZPResult ─────────────────────────────────────────

/// RAII wrapper that guarantees ZPResult memory is freed on drop.
pub struct ZPGuard {
    result: ZPResult,
}

impl ZPGuard {
    /// Wrap a raw ZPResult. Ownership transfers to the guard.
    pub fn new(result: ZPResult) -> Self {
        Self { result }
    }

    /// Check if the result represents an error.
    pub fn is_error(&self) -> bool {
        self.result.code != 0 || !self.result.error.is_null()
    }

    /// Extract the error as a BridgeError, if any.
    pub fn to_error(&self) -> Option<BridgeError> {
        if !self.is_error() {
            return None;
        }
        let message = if self.result.error.is_null() {
            "unknown bridge error".to_string()
        } else {
            unsafe { CStr::from_ptr(self.result.error) }
                .to_string_lossy()
                .into_owned()
        };
        Some(BridgeError::Bridge {
            code: self.result.code,
            message,
        })
    }

    /// Extract the data string, returning Ok(data) or Err(BridgeError).
    pub fn into_result(self) -> Result<Option<String>, BridgeError> {
        if let Some(err) = self.to_error() {
            return Err(err);
        }
        if self.result.data.is_null() {
            return Ok(None);
        }
        let data = unsafe { CStr::from_ptr(self.result.data) }
            .to_string_lossy()
            .into_owned();
        Ok(Some(data))
    }
}

impl Drop for ZPGuard {
    fn drop(&mut self) {
        // Free both data and error pointers
        if !self.result.data.is_null() {
            unsafe { ZPFree(self.result.data) };
            self.result.data = std::ptr::null_mut();
        }
        if !self.result.error.is_null() {
            unsafe { ZPFree(self.result.error) };
            self.result.error = std::ptr::null_mut();
        }
    }
}

// ─── Helper: Rust string → CString (mutable ptr for Go) ─────────────

fn to_c_str(s: &str) -> Result<CString, BridgeError> {
    CString::new(s).map_err(|_| BridgeError::InvalidArgument("string contains null byte".into()))
}

// ─── Safe wrappers ───────────────────────────────────────────────────

// -- Vault lifecycle --

pub fn create_vault(path: &str, master_password: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let c_pass = to_c_str(master_password)?;
    let guard = ZPGuard::new(unsafe {
        ZPCreateVault(c_path.into_raw(), c_pass.into_raw())
    });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no handle returned".into() })
}

pub fn open_vault(path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPOpenVault(c_path.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no handle returned".into() })
}

pub fn unlock(handle: i64, master_password: &str) -> Result<Option<String>, BridgeError> {
    let c_pass = to_c_str(master_password)?;
    let guard = ZPGuard::new(unsafe { ZPUnlock(handle, c_pass.into_raw()) });
    guard.into_result()
}

pub fn unlock_with_recovery(handle: i64, mnemonic: &str) -> Result<Option<String>, BridgeError> {
    let c_mnemonic = to_c_str(mnemonic)?;
    let guard = ZPGuard::new(unsafe { ZPUnlockWithRecovery(handle, c_mnemonic.into_raw()) });
    guard.into_result()
}

pub fn unlock_with_key(handle: i64, vault_key_base64: &str) -> Result<Option<String>, BridgeError> {
    let c_key = to_c_str(vault_key_base64)?;
    let guard = ZPGuard::new(unsafe { ZPUnlockWithKey(handle, c_key.into_raw()) });
    guard.into_result()
}

pub fn get_vault_key(handle: i64) -> Result<String, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPGetVaultKey(handle) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no vault key returned".into() })
}

pub fn lock(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPLock(handle) });
    guard.into_result()
}

pub fn close_vault(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPCloseVault(handle) });
    guard.into_result()
}

pub fn is_locked(handle: i64) -> bool {
    unsafe { ZPIsLocked(handle) != 0 }
}

pub fn change_master_password(
    handle: i64,
    old_password: &str,
    new_password: &str,
) -> Result<Option<String>, BridgeError> {
    let c_old = to_c_str(old_password)?;
    let c_new = to_c_str(new_password)?;
    let guard = ZPGuard::new(unsafe {
        ZPChangeMasterPassword(handle, c_old.into_raw(), c_new.into_raw())
    });
    guard.into_result()
}

// -- Item CRUD --

pub fn list_items(handle: i64, filter_json: &str) -> Result<String, BridgeError> {
    let c_filter = to_c_str(filter_json)?;
    let guard = ZPGuard::new(unsafe { ZPListItems(handle, c_filter.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no items returned".into() })
}

pub fn get_item(handle: i64, item_id: &str) -> Result<String, BridgeError> {
    let c_id = to_c_str(item_id)?;
    let guard = ZPGuard::new(unsafe { ZPGetItem(handle, c_id.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no item returned".into() })
}

pub fn create_item(handle: i64, item_json: &str) -> Result<String, BridgeError> {
    let c_json = to_c_str(item_json)?;
    let guard = ZPGuard::new(unsafe { ZPCreateItem(handle, c_json.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no item id returned".into() })
}

pub fn update_item(handle: i64, item_id: &str, item_json: &str) -> Result<Option<String>, BridgeError> {
    let c_id = to_c_str(item_id)?;
    let c_json = to_c_str(item_json)?;
    let guard = ZPGuard::new(unsafe { ZPUpdateItem(handle, c_id.into_raw(), c_json.into_raw()) });
    guard.into_result()
}

pub fn delete_item(handle: i64, item_id: &str) -> Result<Option<String>, BridgeError> {
    let c_id = to_c_str(item_id)?;
    let guard = ZPGuard::new(unsafe { ZPDeleteItem(handle, c_id.into_raw()) });
    guard.into_result()
}

// -- Search --

pub fn search(handle: i64, query: &str) -> Result<String, BridgeError> {
    let c_query = to_c_str(query)?;
    let guard = ZPGuard::new(unsafe { ZPSearch(handle, c_query.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no search results".into() })
}

pub fn rebuild_index(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPRebuildIndex(handle) });
    guard.into_result()
}

// -- Crypto --

pub fn generate_password(length: i32, options_json: &str) -> Result<String, BridgeError> {
    let c_opts = to_c_str(options_json)?;
    let guard = ZPGuard::new(unsafe { ZPGeneratePassword(length, c_opts.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no password generated".into() })
}

pub fn generate_passphrase(words: i32, separator: &str) -> Result<String, BridgeError> {
    let c_sep = to_c_str(separator)?;
    let guard = ZPGuard::new(unsafe { ZPGeneratePassphrase(words, c_sep.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no passphrase generated".into() })
}

pub fn score_password(password: &str) -> Result<String, BridgeError> {
    let c_pass = to_c_str(password)?;
    let guard = ZPGuard::new(unsafe { ZPScorePassword(c_pass.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no score returned".into() })
}

pub fn check_breach(password: &str) -> Result<String, BridgeError> {
    let c_pass = to_c_str(password)?;
    let guard = ZPGuard::new(unsafe { ZPCheckBreach(c_pass.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no breach result".into() })
}

// -- Health --

pub fn analyze_health(handle: i64) -> Result<String, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPAnalyzeHealth(handle) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no health result".into() })
}

// -- Import --

pub fn import_csv(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportCSV(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_chrome(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportChrome(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_firefox(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportFirefox(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_1password(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImport1Password(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_bitwarden(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportBitwarden(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_safari(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportSafari(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_lastpass(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportLastPass(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_keepass(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImportKeePass(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

pub fn import_1pux(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPImport1PUX(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "import failed".into() })
}

// -- Export --

pub fn export_json(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPExportJSON(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "export failed".into() })
}

pub fn export_csv(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPExportCSV(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "export failed".into() })
}

pub fn export_encrypted(handle: i64, path: &str) -> Result<String, BridgeError> {
    let c_path = to_c_str(path)?;
    let guard = ZPGuard::new(unsafe { ZPExportEncrypted(handle, c_path.into_raw()) });
    guard.into_result()?.ok_or_else(|| BridgeError::Bridge { code: -1, message: "export failed".into() })
}

// -- Recovery --

pub fn regenerate_recovery(handle: i64) -> Result<String, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPRegenerateRecovery(handle) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no recovery phrase returned".into() })
}

// -- Sync --

pub fn sync_setup(handle: i64, config_json: &str) -> Result<Option<String>, BridgeError> {
    let c_config = to_c_str(config_json)?;
    let guard = ZPGuard::new(unsafe { ZPSyncSetup(handle, c_config.into_raw()) });
    guard.into_result()
}

pub fn sync_register(handle: i64, device_name: &str) -> Result<Option<String>, BridgeError> {
    let c_name = to_c_str(device_name)?;
    let guard = ZPGuard::new(unsafe { ZPSyncRegister(handle, c_name.into_raw()) });
    guard.into_result()
}

pub fn sync_pull(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPSyncPull(handle) });
    guard.into_result()
}

pub fn sync_push(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPSyncPush(handle) });
    guard.into_result()
}

pub fn sync_full(handle: i64) -> Result<Option<String>, BridgeError> {
    let guard = ZPGuard::new(unsafe { ZPSyncFull(handle) });
    guard.into_result()
}

// -- Version history --

pub fn get_version_history(handle: i64, item_id: &str) -> Result<String, BridgeError> {
    let c_id = to_c_str(item_id)?;
    let guard = ZPGuard::new(unsafe { ZPGetVersionHistory(handle, c_id.into_raw()) });
    guard
        .into_result()?
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "no version history".into() })
}

pub fn restore_version(handle: i64, item_id: &str, version: i32) -> Result<Option<String>, BridgeError> {
    let c_id = to_c_str(item_id)?;
    let guard = ZPGuard::new(unsafe { ZPRestoreVersion(handle, c_id.into_raw(), version) });
    guard.into_result()
}
