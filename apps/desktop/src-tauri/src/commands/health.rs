use tauri::State;

use crate::bridge;
use crate::error::BridgeError;
use crate::state::VaultState;

#[tauri::command]
pub async fn analyze_health(
    state: State<'_, VaultState>,
) -> Result<String, BridgeError> {
    let handle = state.get_handle()?;
    tokio::task::spawn_blocking(move || bridge::analyze_health(handle))
        .await
        .map_err(|e| BridgeError::State(e.to_string()))?
}
