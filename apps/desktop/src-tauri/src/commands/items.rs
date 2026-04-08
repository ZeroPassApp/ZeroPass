use tauri::State;

use crate::bridge;
use crate::error::BridgeError;
use crate::state::VaultState;

#[tauri::command]
pub async fn list_items(
    state: State<'_, VaultState>,
    filter_json: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::list_items(handle, &filter_json))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn get_item(
    state: State<'_, VaultState>,
    item_id: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::get_item(handle, &item_id))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn create_item(
    state: State<'_, VaultState>,
    item_json: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::create_item(handle, &item_json))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn update_item(
    state: State<'_, VaultState>,
    item_id: String,
    item_json: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::update_item(handle, &item_id, &item_json))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn delete_item(
    state: State<'_, VaultState>,
    item_id: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::delete_item(handle, &item_id))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn search_items(
    state: State<'_, VaultState>,
    query: String,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::search(handle, &query))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}
