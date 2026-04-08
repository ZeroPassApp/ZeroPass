use std::sync::Mutex;

use crate::error::BridgeError;

/// Shared application state managed by Tauri.
/// Holds the vault handle returned by the Go bridge.
pub struct VaultState {
    /// Go bridge handle (0 = no vault open)
    handle: Mutex<i64>,
}

impl VaultState {
    pub fn new() -> Self {
        Self {
            handle: Mutex::new(0),
        }
    }

    pub fn set_handle(&self, h: i64) {
        if let Ok(mut guard) = self.handle.lock() {
            *guard = h;
        }
    }

    pub fn get_handle(&self) -> Result<i64, BridgeError> {
        let guard = self.handle.lock().map_err(|e| {
            BridgeError::State(format!("lock poisoned: {e}"))
        })?;
        let h = *guard;
        if h == 0 {
            return Err(BridgeError::VaultNotOpen);
        }
        Ok(h)
    }

    pub fn clear_handle(&self) {
        if let Ok(mut guard) = self.handle.lock() {
            *guard = 0;
        }
    }
}
