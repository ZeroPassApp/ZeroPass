mod bridge;
mod commands;
mod error;
mod state;

use state::VaultState;
use tauri_plugin_autostart::MacosLauncher;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_clipboard_manager::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_global_shortcut::Builder::new().build())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_os::init())
        .plugin(tauri_plugin_autostart::init(
            MacosLauncher::LaunchAgent,
            None,
        ))
        .plugin(tauri_plugin_store::Builder::new().build())
        .manage(VaultState::new())
        .invoke_handler(tauri::generate_handler![
            commands::vault::create_vault,
            commands::vault::open_vault,
            commands::vault::unlock_vault,
            commands::vault::unlock_with_recovery,
            commands::vault::unlock_with_key,
            commands::vault::get_vault_key,
            commands::vault::lock_vault,
            commands::vault::close_vault,
            commands::vault::is_vault_locked,
            commands::vault::change_master_password,
            commands::vault::generate_password,
            commands::vault::generate_passphrase,
            commands::vault::score_password,
            commands::items::list_items,
            commands::items::get_item,
            commands::items::create_item,
            commands::items::update_item,
            commands::items::delete_item,
            commands::items::search_items,
            commands::health::analyze_health,
            commands::version::get_version_history,
            commands::version::restore_version,
            commands::sync::sync_setup,
            commands::sync::sync_register,
            commands::sync::sync_pull,
            commands::sync::sync_push,
            commands::sync::sync_full,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
