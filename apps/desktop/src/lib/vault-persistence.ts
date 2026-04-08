import { load } from "@tauri-apps/plugin-store";

const STORE_NAME = "settings.json";
const KEY_VAULT_PATH = "lastVaultPath";

let storePromise: ReturnType<typeof load> | null = null;

function getStore() {
    if (!storePromise) {
        storePromise = load(STORE_NAME, { defaults: {}, autoSave: true });
    }
    return storePromise;
}

export async function saveLastVaultPath(path: string): Promise<void> {
    console.log("[vault-persistence] saving vault path:", path);
    const store = await getStore();
    await store.set(KEY_VAULT_PATH, path);
    await store.save();
    console.log("[vault-persistence] saved successfully");
}

export async function getLastVaultPath(): Promise<string | null> {
    console.log("[vault-persistence] loading last vault path...");
    const store = await getStore();
    const value = await store.get<string>(KEY_VAULT_PATH);
    console.log("[vault-persistence] loaded value:", value);
    return value ?? null;
}

export async function clearLastVaultPath(): Promise<void> {
    console.log("[vault-persistence] clearing vault path");
    const store = await getStore();
    await store.delete(KEY_VAULT_PATH);
    await store.save();
}
