use tauri::State;

use crate::bridge;
use crate::error::BridgeError;
use crate::state::VaultState;

#[tauri::command]
pub async fn get_version_history(
    state: State<'_, VaultState>,
    item_id: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::get_version_history(handle, &item_id))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn restore_version(
    state: State<'_, VaultState>,
    item_id: String,
    version: i32,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::restore_version(handle, &item_id, version))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}
