import { type CSSProperties } from "react";
import type { VaultItem } from "../../lib/types";
import { useVaultStore } from "../../stores/vault-store";
import { Icon } from "../common/icon";

const typeIcons: Record<string, string> = {
    login: "key",
    apikey: "api",
    sshkey: "terminal",
    note: "sticky_note_2",
    creditcard: "credit_card",
    identity: "person",
    passkey: "passkey",
    custom: "inventory_2",
};

const typeColors: Record<string, string> = {
    login: "#5856D6",
    apikey: "#34C759",
    sshkey: "#FF9500",
    note: "#818cf8",
    creditcard: "#f472b6",
    identity: "#38bdf8",
    passkey: "#a78bfa",
    custom: "#94a3b8",
};

export function ItemList({ onNewItem }: { onNewItem: () => void }) {
    const filteredItems = useVaultStore((s) => s.filteredItems);
    const selectedItemId = useVaultStore((s) => s.selectedItemId);
    const selectItem = useVaultStore((s) => s.selectItem);
    const searchQuery = useVaultStore((s) => s.searchQuery);
    const setSearchQuery = useVaultStore((s) => s.setSearchQuery);

    const items = filteredItems();

    return (
        <div style={styles.container}>
            {/* Header with search and add */}
            <div style={styles.header}>
                <div style={styles.searchRow}>
                    <Icon name="search" size={16} color="#64748b" />
                    <input
                        style={styles.searchInput}
                        placeholder="Search items…"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
                <div style={styles.headerActions}>
                    <button style={styles.sortBtn} title="Sort">
                        <Icon name="swap_vert" size={18} color="#64748b" />
                    </button>
                    <button style={styles.addBtn} onClick={onNewItem} title="New Item">
                        <Icon name="add" size={18} color="#fff" />
                    </button>
                </div>
            </div>

            {/* Item count */}
            <div style={styles.counter}>
                {items.length} item{items.length !== 1 ? "s" : ""}
            </div>

            {/* List */}
            <div style={styles.list}>
                {items.length === 0 && (
                    <div style={styles.empty}>
                        <Icon name="folder_off" size={40} color="#334155" />
                        <p style={styles.emptyText}>No items found</p>
                        <button style={styles.emptyBtn} onClick={onNewItem}>
                            <Icon name="add" size={16} color="#fff" />
                            Create your first item
                        </button>
                    </div>
                )}
                {items.map((item) => (
                    <ItemRow
                        key={item.id}
                        item={item}
                        selected={item.id === selectedItemId}
                        onClick={() => selectItem(item.id)}
                    />
                ))}
            </div>
        </div>
    );
}

function ItemRow({
    item,
    selected,
    onClick,
}: {
    item: VaultItem;
    selected: boolean;
    onClick: () => void;
}) {
    const subtitle =
        item.fields.username || item.fields.email || item.fields.url || "";
    const iconName = typeIcons[item.type] ?? "inventory_2";
    const iconColor = typeColors[item.type] ?? "#94a3b8";

    return (
        <button
            style={{
                ...styles.row,
                ...(selected ? styles.rowSelected : {}),
            }}
            onClick={onClick}
        >
            {selected && <div style={styles.activeBar} />}
            <div style={{ ...styles.rowIconBox, background: `${iconColor}15` }}>
                <Icon name={iconName} fill size={18} color={iconColor} />
            </div>
            <div style={styles.rowContent}>
                <span style={styles.rowName}>{item.name}</span>
                {subtitle && <span style={styles.rowSub}>{subtitle}</span>}
            </div>
            {item.favorite && (
                <Icon name="star" fill size={14} color="#f59e0b" />
            )}
        </button>
    );
}

const styles: Record<string, CSSProperties> = {
    container: {
        width: 300,
        minWidth: 280,
        borderRight: "1px solid #1e293b",
        display: "flex",
        flexDirection: "column",
        background: "#1e293b",
        fontFamily: "'Inter', sans-serif",
    },
    header: {
        display: "flex",
        gap: 8,
        padding: "12px 12px 8px",
        alignItems: "center",
    },
    searchRow: {
        flex: 1,
        display: "flex",
        alignItems: "center",
        gap: 8,
        background: "#0f172a",
        border: "1px solid #334155",
        borderRadius: 8,
        padding: "6px 10px",
    },
    searchInput: {
        flex: 1,
        background: "transparent",
        border: "none",
        color: "#e2e8f0",
        fontSize: 13,
        outline: "none",
    },
    headerActions: {
        display: "flex",
        gap: 4,
    },
    sortBtn: {
        width: 32,
        height: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "none",
        border: "none",
        cursor: "pointer",
        borderRadius: 6,
    },
    addBtn: {
        width: 32,
        height: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "#5856D6",
        border: "none",
        borderRadius: 6,
        cursor: "pointer",
    },
    counter: {
        fontSize: 10,
        fontWeight: 600,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.05em",
        padding: "2px 16px 8px",
    },
    list: {
        flex: 1,
        overflow: "auto",
        padding: "0 6px 6px",
    },
    empty: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 12,
        padding: "48px 16px",
    },
    emptyText: {
        color: "#64748b",
        fontSize: 13,
    },
    emptyBtn: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        padding: "8px 16px",
        fontSize: 13,
        fontWeight: 600,
        cursor: "pointer",
    },
    row: {
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "10px 10px",
        border: "none",
        background: "transparent",
        borderRadius: 8,
        cursor: "pointer",
        width: "100%",
        textAlign: "left",
        transition: "all 0.12s",
        position: "relative",
    },
    rowSelected: {
        background: "rgba(88, 86, 214, 0.1)",
        borderLeft: "4px solid #5856D6",
        paddingLeft: 6,
    },
    activeBar: {
        position: "absolute",
        left: 0,
        top: "50%",
        transform: "translateY(-50%)",
        width: 3,
        height: "60%",
        background: "#5856D6",
        borderRadius: 2,
    },
    rowIconBox: {
        width: 40,
        height: 40,
        borderRadius: 8,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
    },
    rowContent: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        minWidth: 0,
    },
    rowName: {
        fontSize: 13,
        fontWeight: 600,
        color: "#e2e8f0",
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
    },
    rowSub: {
        fontSize: 11,
        color: "#64748b",
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
        marginTop: 1,
    },
};
