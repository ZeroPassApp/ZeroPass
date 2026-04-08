import { create } from "zustand";
import type { VaultItem, ItemType } from "../lib/types";
import * as cmd from "../lib/commands";

type SortField = "name" | "updated_at" | "created_at" | "type";
type SortDir = "asc" | "desc";

interface VaultStore {
    items: VaultItem[];
    selectedItemId: string | null;
    filterType: ItemType | null;
    filterFavorites: boolean;
    searchQuery: string;
    sortField: SortField;
    sortDir: SortDir;
    loading: boolean;
    error: string | null;

    // Derived
    filteredItems: () => VaultItem[];
    selectedItem: () => VaultItem | null;

    // Actions
    loadItems: () => Promise<void>;
    selectItem: (id: string | null) => void;
    setFilterType: (type: ItemType | null) => void;
    setFilterFavorites: (fav: boolean) => void;
    setSearchQuery: (q: string) => void;
    setSort: (field: SortField, dir: SortDir) => void;
    createItem: (item: Partial<VaultItem>) => Promise<string | null>;
    updateItem: (id: string, item: Partial<VaultItem>) => Promise<void>;
    deleteItem: (id: string) => Promise<void>;
    toggleFavorite: (id: string) => Promise<void>;
}

export const useVaultStore = create<VaultStore>((set, get) => ({
    items: [],
    selectedItemId: null,
    filterType: null,
    filterFavorites: false,
    searchQuery: "",
    sortField: "updated_at",
    sortDir: "desc",
    loading: false,
    error: null,

    filteredItems: () => {
        const { items, filterType, filterFavorites, searchQuery, sortField, sortDir } = get();
        let result = [...items];

        if (filterType) {
            result = result.filter((i) => i.type === filterType);
        }
        if (filterFavorites) {
            result = result.filter((i) => i.favorite);
        }
        if (searchQuery) {
            const q = searchQuery.toLowerCase();
            result = result.filter(
                (i) =>
                    i.name.toLowerCase().includes(q) ||
                    Object.values(i.fields).some((v) => v.toLowerCase().includes(q)),
            );
        }

        result.sort((a, b) => {
            let cmp = 0;
            if (sortField === "name") {
                cmp = a.name.localeCompare(b.name);
            } else if (sortField === "type") {
                cmp = a.type.localeCompare(b.type);
            } else {
                cmp = (a[sortField] ?? "").localeCompare(b[sortField] ?? "");
            }
            return sortDir === "asc" ? cmp : -cmp;
        });

        return result;
    },

    selectedItem: () => {
        const { items, selectedItemId } = get();
        return items.find((i) => i.id === selectedItemId) ?? null;
    },

    loadItems: async () => {
        set({ loading: true, error: null });
        try {
            const json = await cmd.listItems("{}");
            const parsed: VaultItem[] = JSON.parse(json);
            set({ items: parsed, loading: false });
        } catch (e) {
            set({ error: String(e), loading: false });
        }
    },

    selectItem: (id) => set({ selectedItemId: id }),

    setFilterType: (type) => set({ filterType: type, selectedItemId: null }),

    setFilterFavorites: (fav) => set({ filterFavorites: fav, selectedItemId: null }),

    setSearchQuery: (q) => set({ searchQuery: q }),

    setSort: (field, dir) => set({ sortField: field, sortDir: dir }),

    createItem: async (item) => {
        try {
            set({ error: null });
            const json = JSON.stringify(item);
            const id = await cmd.createItem(json);
            await get().loadItems();
            set({ selectedItemId: id });
            return id;
        } catch (e) {
            set({ error: String(e) });
            return null;
        }
    },

    updateItem: async (id, item) => {
        try {
            set({ error: null });
            const json = JSON.stringify(item);
            await cmd.updateItem(id, json);
            await get().loadItems();
        } catch (e) {
            set({ error: String(e) });
        }
    },

    deleteItem: async (id) => {
        try {
            set({ error: null });
            await cmd.deleteItem(id);
            const { selectedItemId } = get();
            if (selectedItemId === id) {
                set({ selectedItemId: null });
            }
            await get().loadItems();
        } catch (e) {
            set({ error: String(e) });
        }
    },

    toggleFavorite: async (id) => {
        const item = get().items.find((i) => i.id === id);
        if (!item) return;
        await get().updateItem(id, { ...item, favorite: !item.favorite });
    },
}));
