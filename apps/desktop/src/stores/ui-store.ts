import { create } from "zustand";
import type { Toast } from "../lib/types";

interface UIStore {
    toasts: Toast[];
    isLoading: boolean;
    loadingMessage: string | null;

    addToast: (toast: Omit<Toast, "id">) => void;
    removeToast: (id: string) => void;
    setLoading: (loading: boolean, message?: string) => void;
}

let toastId = 0;

export const useUIStore = create<UIStore>((set, get) => ({
    toasts: [],
    isLoading: false,
    loadingMessage: null,

    addToast: (toast) => {
        const id = String(++toastId);
        const duration = toast.duration ?? 4000;
        set((s) => ({ toasts: [...s.toasts, { ...toast, id }] }));
        if (duration > 0) {
            setTimeout(() => get().removeToast(id), duration);
        }
    },

    removeToast: (id) => {
        set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }));
    },

    setLoading: (loading, message) => {
        set({ isLoading: loading, loadingMessage: message ?? null });
    },
}));
