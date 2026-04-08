import { type CSSProperties, useEffect, useState, useCallback } from "react";
import type { VaultItem } from "../lib/types";
import { useVaultStore } from "../stores/vault-store";
import { Sidebar } from "../components/vault/sidebar";
import { ItemList } from "../components/vault/item-list";
import { ItemDetail } from "../components/vault/item-detail";
import { ItemEditor } from "../components/vault/item-editor";
import { QuickSearch } from "../components/vault/quick-search";
import { SettingsDialog } from "../components/vault/settings-dialog";
import { HealthDashboard } from "../components/vault/health-dashboard";
import { VersionHistory } from "../components/vault/version-history";

export function VaultPage() {
    const loadItems = useVaultStore((s) => s.loadItems);
    const [editorItem, setEditorItem] = useState<VaultItem | null | undefined>(
        undefined,
    );
    const [showSearch, setShowSearch] = useState(false);
    const [showSettings, setShowSettings] = useState(false);
    const [showHealth, setShowHealth] = useState(false);
    const [versionItem, setVersionItem] = useState<VaultItem | null>(null);

    useEffect(() => {
        loadItems();
    }, [loadItems]);

    // Cmd+K for quick search, Cmd+, for settings
    const handleKeyDown = useCallback((e: KeyboardEvent) => {
        if ((e.metaKey || e.ctrlKey) && e.key === "k") {
            e.preventDefault();
            setShowSearch((s) => !s);
        }
        if ((e.metaKey || e.ctrlKey) && e.key === ",") {
            e.preventDefault();
            setShowSettings((s) => !s);
        }
    }, []);

    useEffect(() => {
        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [handleKeyDown]);

    const openNewItem = () => setEditorItem(null);
    const openEditItem = (item: VaultItem) => setEditorItem(item);
    const closeEditor = () => setEditorItem(undefined);

    return (
        <div style={styles.container}>
            <Sidebar
                onOpenSettings={() => setShowSettings(true)}
                onOpenHealth={() => setShowHealth(true)}
            />
            <ItemList onNewItem={openNewItem} />
            <ItemDetail
                onEdit={openEditItem}
                onViewVersions={(item) => setVersionItem(item)}
            />

            {editorItem !== undefined && (
                <ItemEditor item={editorItem} onClose={closeEditor} />
            )}
            {showSearch && <QuickSearch onClose={() => setShowSearch(false)} />}
            {showSettings && <SettingsDialog onClose={() => setShowSettings(false)} />}
            {showHealth && <HealthDashboard onClose={() => setShowHealth(false)} />}
            {versionItem && (
                <VersionHistory
                    itemId={versionItem.id}
                    itemName={versionItem.name}
                    onClose={() => { setVersionItem(null); loadItems(); }}
                />
            )}
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    container: {
        display: "flex",
        height: "100%",
        background: "#0f172a",
        fontFamily: "'Inter', sans-serif",
    },
};
