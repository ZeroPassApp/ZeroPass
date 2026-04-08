import { invoke } from "@tauri-apps/api/core";
import type { PasswordOptions } from "./types";

// -- Vault lifecycle --

export async function createVault(
    path: string,
    masterPassword: string,
): Promise<string> {
    return invoke("create_vault", { path, masterPassword });
}

export async function openVault(path: string): Promise<string> {
    return invoke("open_vault", { path });
}

export async function unlockVault(masterPassword: string): Promise<void> {
    return invoke("unlock_vault", { masterPassword });
}

export async function unlockWithRecovery(mnemonic: string): Promise<void> {
    return invoke("unlock_with_recovery", { mnemonic });
}

export async function unlockWithKey(vaultKeyBase64: string): Promise<void> {
    return invoke("unlock_with_key", { vaultKeyBase64 });
}

export async function getVaultKey(): Promise<string> {
    return invoke("get_vault_key");
}

export async function lockVault(): Promise<void> {
    return invoke("lock_vault");
}

export async function closeVault(): Promise<void> {
    return invoke("close_vault");
}

export async function isVaultLocked(): Promise<boolean> {
    return invoke("is_vault_locked");
}

export async function changeMasterPassword(
    oldPassword: string,
    newPassword: string,
): Promise<void> {
    return invoke("change_master_password", { oldPassword, newPassword });
}

// -- Crypto --

export async function generatePassword(
    length: number,
    options: PasswordOptions,
): Promise<string> {
    const optionsJson = JSON.stringify(options);
    return invoke("generate_password", { length, optionsJson });
}

export async function generatePassphrase(
    words: number,
    separator: string,
): Promise<string> {
    return invoke("generate_passphrase", { words, separator });
}

export async function scorePassword(
    password: string,
): Promise<string> {
    return invoke("score_password", { password });
}

// -- Items CRUD --

export async function listItems(filterJson: string): Promise<string> {
    return invoke("list_items", { filterJson });
}

export async function getItem(itemId: string): Promise<string> {
    return invoke("get_item", { itemId });
}

export async function createItem(itemJson: string): Promise<string> {
    return invoke("create_item", { itemJson });
}

export async function updateItem(
    itemId: string,
    itemJson: string,
): Promise<void> {
    return invoke("update_item", { itemId, itemJson });
}

export async function deleteItem(itemId: string): Promise<void> {
    return invoke("delete_item", { itemId });
}

// -- Search --

export async function searchItems(query: string): Promise<string> {
    return invoke("search_items", { query });
}

// -- Health --

export async function analyzeHealth(): Promise<string> {
    return invoke("analyze_health");
}

// -- Version history --

export async function getVersionHistory(itemId: string): Promise<string> {
    return invoke("get_version_history", { itemId });
}

export async function restoreVersion(
    itemId: string,
    version: number,
): Promise<void> {
    return invoke("restore_version", { itemId, version });
}

// -- Sync --

export async function syncSetup(configJson: string): Promise<string> {
    return invoke("sync_setup", { configJson });
}

export async function syncRegister(deviceName: string): Promise<string> {
    return invoke("sync_register", { deviceName });
}

export async function syncPull(): Promise<string> {
    return invoke("sync_pull");
}

export async function syncPush(): Promise<string> {
    return invoke("sync_push");
}

export async function syncFull(): Promise<string> {
    return invoke("sync_full");
}
