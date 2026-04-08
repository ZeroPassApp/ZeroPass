export type AuthState =
    | "restoring"
    | "noVault"
    | "locked"
    | "showingRecovery"
    | "unlocked";

export type ItemType =
    | "login"
    | "apikey"
    | "sshkey"
    | "note"
    | "creditcard"
    | "identity"
    | "passkey"
    | "custom";

export interface VaultItem {
    id: string;
    type: ItemType;
    name: string;
    fields: Record<string, string>;
    tags: string[];
    favorite: boolean;
    created_at: string;
    updated_at: string;
}

export interface PasswordScore {
    score: number;
    feedback: string;
    crack_time: string;
}

export interface PasswordOptions {
    uppercase: boolean;
    lowercase: boolean;
    numbers: boolean;
    symbols: boolean;
}

export interface Toast {
    id: string;
    type: "success" | "error" | "info" | "warning";
    message: string;
    duration?: number;
}
