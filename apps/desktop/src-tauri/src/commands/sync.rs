use tauri::State;

use crate::bridge;
use crate::error::BridgeError;
use crate::state::VaultState;

#[tauri::command]
pub async fn sync_setup(
    state: State<'_, VaultState>,
    config_json: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::sync_setup(handle, &config_json))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
        .map(|v| v.unwrap_or_default())
        .map_err(Into::into)
}

#[tauri::command]
pub async fn sync_register(
    state: State<'_, VaultState>,
    device_name: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::sync_register(handle, &device_name))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
        .map(|v| v.unwrap_or_default())
        .map_err(Into::into)
}

#[tauri::command]
pub async fn sync_pull(
    state: State<'_, VaultState>,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::sync_pull(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
        .map(|v| v.unwrap_or_default())
        .map_err(Into::into)
}

#[tauri::command]
pub async fn sync_push(
    state: State<'_, VaultState>,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::sync_push(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
        .map(|v| v.unwrap_or_default())
        .map_err(Into::into)
}

#[tauri::command]
pub async fn sync_full(
    state: State<'_, VaultState>,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::sync_full(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
        .map(|v| v.unwrap_or_default())
        .map_err(Into::into)
}
