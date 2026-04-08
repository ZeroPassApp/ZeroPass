import { type CSSProperties } from "react";
import type { ItemType } from "../../lib/types";
import { useVaultStore } from "../../stores/vault-store";
import { useAuthStore } from "../../stores/auth-store";
import { Icon } from "../common/icon";

const categories: { type: ItemType; label: string; icon: string }[] = [
    { type: "login", label: "Logins", icon: "key" },
    { type: "apikey", label: "API Keys", icon: "api" },
    { type: "sshkey", label: "SSH Keys", icon: "terminal" },
    { type: "note", label: "Secure Notes", icon: "sticky_note_2" },
    { type: "creditcard", label: "Cards", icon: "credit_card" },
    { type: "identity", label: "Identities", icon: "person" },
    { type: "passkey", label: "Passkeys", icon: "passkey" },
    { type: "custom", label: "Custom", icon: "inventory_2" },
];

export function Sidebar({ onOpenSettings, onOpenHealth }: { onOpenSettings?: () => void; onOpenHealth?: () => void }) {
    const filterType = useVaultStore((s) => s.filterType);
    const filterFavorites = useVaultStore((s) => s.filterFavorites);
    const setFilterType = useVaultStore((s) => s.setFilterType);
    const setFilterFavorites = useVaultStore((s) => s.setFilterFavorites);
    const items = useVaultStore((s) => s.items);
    const lock = useAuthStore((s) => s.lock);

    const countByType = (t: ItemType) => items.filter((i) => i.type === t).length;
    const favCount = items.filter((i) => i.favorite).length;

    return (
        <aside style={styles.sidebar}>
            {/* Brand */}
            <div style={styles.brand}>
                <div style={styles.brandIcon}>
                    <Icon name="lock" fill size={16} color="#fff" />
                </div>
                <div>
                    <div style={styles.brandName}>ZeroPass</div>
                    <div style={styles.brandSub}>Personal Vault</div>
                </div>
            </div>

            {/* Library */}
            <div style={styles.section}>
                <div style={styles.sectionLabel}>Library</div>
                <NavItem
                    icon="folder"
                    label="All Items"
                    count={items.length}
                    active={!filterType && !filterFavorites}
                    onClick={() => { setFilterType(null); setFilterFavorites(false); }}
                />
                <NavItem
                    icon="star"
                    label="Favorites"
                    count={favCount}
                    active={filterFavorites}
                    onClick={() => { setFilterType(null); setFilterFavorites(true); }}
                />
                <NavItem
                    icon="schedule"
                    label="Recent"
                    count={0}
                    active={false}
                    onClick={() => { setFilterType(null); setFilterFavorites(false); }}
                />
            </div>

            {/* Categories */}
            <div style={styles.section}>
                <div style={styles.sectionLabel}>Categories</div>
                {categories.map((c) => (
                    <NavItem
                        key={c.type}
                        icon={c.icon}
                        label={c.label}
                        count={countByType(c.type)}
                        active={filterType === c.type}
                        onClick={() => {
                            setFilterFavorites(false);
                            setFilterType(filterType === c.type ? null : c.type);
                        }}
                    />
                ))}
            </div>

            <div style={styles.spacer} />

            {/* User profile + actions */}
            <div style={styles.bottomSection}>
                <div style={styles.userRow}>
                    <div style={styles.avatar}>
                        <Icon name="person" size={16} color="#94a3b8" />
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={styles.userName}>Local User</div>
                        <div style={styles.userPlan}>Pro Plan</div>
                    </div>
                    {onOpenSettings && (
                        <button style={styles.gearBtn} onClick={onOpenSettings}>
                            <Icon name="settings" size={18} color="#64748b" />
                        </button>
                    )}
                </div>
                <div style={styles.bottomActions}>
                    {onOpenHealth && (
                        <button style={styles.actionBtn} onClick={onOpenHealth}>
                            <Icon name="shield" size={16} color="#64748b" />
                            <span>Health</span>
                        </button>
                    )}
                    <button style={styles.lockBtn} onClick={lock}>
                        <Icon name="lock" fill size={14} color="#64748b" />
                        Lock Vault
                    </button>
                </div>
            </div>
        </aside>
    );
}

function NavItem({
    icon,
    label,
    count,
    active,
    onClick,
}: {
    icon: string;
    label: string;
    count: number;
    active: boolean;
    onClick: () => void;
}) {
    return (
        <button
            style={{
                ...styles.navItem,
                ...(active ? styles.navItemActive : {}),
            }}
            onClick={onClick}
        >
            <Icon name={icon} size={18} color={active ? "#818cf8" : "#64748b"} />
            <span style={styles.navLabel}>{label}</span>
            {count > 0 && <span style={styles.badge}>{count}</span>}
        </button>
    );
}

const styles: Record<string, CSSProperties> = {
    sidebar: {
        width: 220,
        minWidth: 220,
        background: "#0f172a",
        borderRight: "1px solid #1e293b",
        display: "flex",
        flexDirection: "column",
        fontFamily: "'Inter', sans-serif",
    },
    brand: {
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "20px 16px 16px",
    },
    brandIcon: {
        width: 28,
        height: 28,
        background: "#5856D6",
        borderRadius: 6,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    brandName: {
        fontSize: 14,
        fontWeight: 700,
        color: "#fff",
        letterSpacing: "-0.01em",
    },
    brandSub: {
        fontSize: 10,
        color: "#64748b",
        fontWeight: 500,
    },
    section: {
        display: "flex",
        flexDirection: "column",
        gap: 1,
        padding: "0 8px",
        marginBottom: 16,
    },
    sectionLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#475569",
        textTransform: "uppercase",
        letterSpacing: "0.08em",
        padding: "12px 8px 6px",
    },
    navItem: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "7px 8px",
        border: "none",
        background: "transparent",
        borderRadius: 6,
        color: "#94a3b8",
        fontSize: 13,
        cursor: "pointer",
        textAlign: "left",
        width: "100%",
        transition: "all 0.12s",
    },
    navItemActive: {
        background: "rgba(88, 86, 214, 0.15)",
        color: "#e2e8f0",
    },
    navLabel: {
        flex: 1,
        fontWeight: 500,
    },
    badge: {
        fontSize: 11,
        color: "#64748b",
        fontWeight: 500,
        fontVariantNumeric: "tabular-nums",
    },
    spacer: { flex: 1 },
    bottomSection: {
        padding: 12,
        borderTop: "1px solid #1e293b",
    },
    userRow: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "4px 0 8px",
    },
    avatar: {
        width: 28,
        height: 28,
        borderRadius: "50%",
        background: "#1e293b",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    userName: {
        fontSize: 12,
        fontWeight: 600,
        color: "#e2e8f0",
    },
    userPlan: {
        fontSize: 10,
        color: "#64748b",
    },
    gearBtn: {
        background: "none",
        border: "none",
        cursor: "pointer",
        padding: 2,
        display: "flex",
    },
    bottomActions: {
        display: "flex",
        gap: 6,
        marginTop: 4,
    },
    actionBtn: {
        display: "flex",
        alignItems: "center",
        gap: 4,
        padding: "6px 8px",
        background: "none",
        border: "none",
        color: "#64748b",
        fontSize: 11,
        fontWeight: 500,
        cursor: "pointer",
        borderRadius: 4,
    },
    lockBtn: {
        flex: 1,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 6,
        padding: "8px 0",
        background: "rgba(30, 41, 59, 0.5)",
        border: "1px solid #334155",
        borderRadius: 6,
        color: "#94a3b8",
        fontSize: 12,
        fontWeight: 600,
        cursor: "pointer",
        transition: "all 0.15s",
    },
};
