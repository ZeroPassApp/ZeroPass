use tauri::State;

use crate::bridge;
use crate::error::BridgeError;
use crate::state::VaultState;

/// Close any previously-open vault session so the file lock is released.
fn close_existing(state: &VaultState) {
    if let Ok(old) = state.get_handle() {
        let _ = bridge::close_vault(old);
        state.clear_handle();
    }
}

/// Extract the "handle" field from JSON returned by Go bridge.
fn parse_handle(json_str: &str) -> Result<i64, BridgeError> {
    let v: serde_json::Value = serde_json::from_str(json_str)?;
    v.get("handle")
        .and_then(|h| h.as_i64())
        .ok_or_else(|| BridgeError::Bridge { code: -1, message: "missing handle in response".into() })
}

#[tauri::command]
pub async fn create_vault(
    state: State<'_, VaultState>,
    path: String,
    master_password: String,
) -> Result<String, BridgeError> {
    // Release any stale handle first
    close_existing(&state);
    let result_json =
        tokio::task::spawn_blocking(move || bridge::create_vault(&path, &master_password))
            .await
            .map_err(|e| BridgeError::State(e.to_string()))??;
    let handle = parse_handle(&result_json)?;
    state.set_handle(handle);
    Ok(result_json)
}

#[tauri::command]
pub async fn open_vault(
    state: State<'_, VaultState>,
    path: String,
) -> Result<String, BridgeError> {
    // Release any stale handle first
    close_existing(&state);
    let result_json =
        tokio::task::spawn_blocking(move || bridge::open_vault(&path))
            .await
            .map_err(|e| BridgeError::State(e.to_string()))??;
    let handle = parse_handle(&result_json)?;
    state.set_handle(handle);
    Ok(result_json)
}

#[tauri::command]
pub async fn unlock_vault(
    state: State<'_, VaultState>,
    master_password: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::unlock(handle, &master_password))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn unlock_with_recovery(
    state: State<'_, VaultState>,
    mnemonic: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::unlock_with_recovery(handle, &mnemonic))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn unlock_with_key(
    state: State<'_, VaultState>,
    vault_key_base64: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::unlock_with_key(handle, &vault_key_base64))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn get_vault_key(state: State<'_, VaultState>) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::get_vault_key(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn lock_vault(state: State<'_, VaultState>) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::lock(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn close_vault(state: State<'_, VaultState>) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::close_vault(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))??;
    state.clear_handle();
    Ok(())
}

#[tauri::command]
pub async fn is_vault_locked(state: State<'_, VaultState>) -> Result<bool, BridgeError> {
    let handle = state.get_handle()?;
    Ok(tokio::task::spawn_blocking(move || bridge::is_locked(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?)
}

#[tauri::command]
pub async fn change_master_password(
    state: State<'_, VaultState>,
    old_password: String,
    new_password: String,
) -> Result<(), BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || {
        bridge::change_master_password(handle, &old_password, &new_password)
    })
    .await
    .map_err(|e| BridgeError::State(e.to_string()))??;
    Ok(())
}

#[tauri::command]
pub async fn generate_password(
    length: i32,
    options_json: String,
) -> Result<String, BridgeError> {
    tokio::task::spawn_blocking(move || bridge::generate_password(length, &options_json))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn generate_passphrase(
    words: i32,
    separator: String,
) -> Result<String, BridgeError> {
    tokio::task::spawn_blocking(move || bridge::generate_passphrase(words, &separator))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}

#[tauri::command]
pub async fn score_password(password: String) -> Result<String, BridgeError> {
    tokio::task::spawn_blocking(move || bridge::score_password(&password))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}
