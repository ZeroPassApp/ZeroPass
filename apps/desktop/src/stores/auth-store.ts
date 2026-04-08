import { create } from "zustand";
import type { AuthState } from "../lib/types";
import * as cmd from "../lib/commands";
import { saveLastVaultPath, getLastVaultPath, clearLastVaultPath } from "../lib/vault-persistence";

/** Extract a human-readable message from Tauri command errors.
 *  Tauri v2 serializes Rust enum errors as JSON objects, so
 *  `String(e)` yields "[object Object]". */
function extractError(e: unknown): string {
    if (typeof e === "string") return e;
    if (e instanceof Error) return e.message;
    if (e && typeof e === "object") {
        const obj = e as Record<string, unknown>;
        // BridgeError::Bridge { code, message }
        if ("Bridge" in obj && typeof obj.Bridge === "object" && obj.Bridge) {
            return (obj.Bridge as Record<string, unknown>).message as string;
        }
        // BridgeError::State(msg) / Json(msg) / InvalidArgument(msg)
        for (const key of ["State", "Json", "InvalidArgument"] as const) {
            if (key in obj && typeof obj[key] === "string") return obj[key] as string;
        }
        // Sentinel variants
        if ("VaultNotOpen" in obj) return "Vault is not open";
        if ("VaultLocked" in obj) return "Vault is locked";
        // Fallback: readable JSON
        return JSON.stringify(e);
    }
    return String(e);
}

interface AuthStore {
    state: AuthState;
    vaultPath: string | null;
    error: string | null;
    recoveryPhrase: string | null;

    // Actions
    restore: () => Promise<void>;
    setError: (error: string | null) => void;
    createVault: (path: string, password: string) => Promise<void>;
    openVault: (path: string) => Promise<void>;
    unlock: (password: string) => Promise<void>;
    unlockWithRecovery: (mnemonic: string) => Promise<void>;
    unlockWithKey: (key: string) => Promise<void>;
    lock: () => Promise<void>;
    confirmRecovery: () => void;
    reset: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
    state: "restoring",
    vaultPath: null,
    error: null,
    recoveryPhrase: null,

    restore: async () => {
        try {
            console.log("[auth-store] restore: starting...");
            const savedPath = await getLastVaultPath();
            console.log("[auth-store] restore: savedPath =", savedPath);
            if (!savedPath) {
                console.log("[auth-store] restore: no saved path, going to noVault");
                set({ state: "noVault" });
                return;
            }
            console.log("[auth-store] restore: opening vault at", savedPath);
            await cmd.openVault(savedPath);
            console.log("[auth-store] restore: vault opened, going to locked");
            set({ state: "locked", vaultPath: savedPath });
        } catch (e) {
            console.error("[auth-store] restore: error", e);
            // "vault is already open" means a prior call already opened it — treat as success
            const msg = extractError(e);
            if (msg.includes("already open")) {
                console.log("[auth-store] restore: vault already open, going to locked");
                const savedPath = await getLastVaultPath();
                set({ state: "locked", vaultPath: savedPath });
                return;
            }
            await clearLastVaultPath();
            set({ state: "noVault" });
        }
    },

    setError: (error) => set({ error }),

    createVault: async (path, password) => {
        try {
            set({ error: null });
            await cmd.createVault(path, password);
            // After creating, the vault is open but locked — unlock it
            await cmd.unlockVault(password);
            // Get recovery phrase
            const vaultKey = await cmd.getVaultKey();
            await saveLastVaultPath(path);
            set({
                state: "showingRecovery",
                vaultPath: path,
                recoveryPhrase: vaultKey,
            });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    openVault: async (path) => {
        try {
            set({ error: null });
            await cmd.openVault(path);
            await saveLastVaultPath(path);
            set({ state: "locked", vaultPath: path });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    unlock: async (password) => {
        try {
            set({ error: null });
            await cmd.unlockVault(password);
            set({ state: "unlocked" });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    unlockWithRecovery: async (mnemonic) => {
        try {
            set({ error: null });
            await cmd.unlockWithRecovery(mnemonic);
            set({ state: "unlocked" });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    unlockWithKey: async (key) => {
        try {
            set({ error: null });
            await cmd.unlockWithKey(key);
            set({ state: "unlocked" });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    lock: async () => {
        try {
            await cmd.lockVault();
            set({ state: "locked" });
        } catch (e) {
            set({ error: extractError(e) });
        }
    },

    confirmRecovery: () => {
        set({ state: "unlocked", recoveryPhrase: null });
    },

    reset: () => {
        set({
            state: "noVault",
            vaultPath: null,
            error: null,
            recoveryPhrase: null,
        });
    },
}));
