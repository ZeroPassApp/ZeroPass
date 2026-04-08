use serde::Serialize;
use thiserror::Error;

/// Structured error type for all bridge operations.
/// Serializable for Tauri IPC transport.
#[derive(Debug, Error, Serialize)]
pub enum BridgeError {
    #[error("vault not open")]
    VaultNotOpen,

    #[error("vault is locked")]
    VaultLocked,

    #[error("invalid argument: {0}")]
    InvalidArgument(String),

    #[error("bridge error ({code}): {message}")]
    Bridge { code: i32, message: String },

    #[error("json error: {0}")]
    Json(String),

    #[error("state error: {0}")]
    State(String),
}

impl From<serde_json::Error> for BridgeError {
    fn from(e: serde_json::Error) -> Self {
        BridgeError::Json(e.to_string())
    }
}
