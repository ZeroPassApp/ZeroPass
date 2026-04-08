import { type CSSProperties, useState } from "react";
import type { VaultItem } from "../../lib/types";
import { useVaultStore } from "../../stores/vault-store";
import { Icon } from "../common/icon";

const sensitiveFields = new Set([
    "password", "secret", "private_key", "api_key",
    "recovery_codes", "pin", "cvv", "security_code",
]);

const fieldIcons: Record<string, string> = {
    username: "person",
    email: "mail",
    password: "lock",
    url: "link",
    notes: "sticky_note_2",
    api_key: "key",
    secret: "key",
    private_key: "vpn_key",
    public_key: "vpn_key",
};

export function ItemDetail({
    onEdit,
    onViewVersions,
}: {
    onEdit: (item: VaultItem) => void;
    onViewVersions?: (item: VaultItem) => void;
}) {
    const selectedItem = useVaultStore((s) => s.selectedItem);
    const toggleFavorite = useVaultStore((s) => s.toggleFavorite);
    const deleteItem = useVaultStore((s) => s.deleteItem);

    const item = selectedItem();

    if (!item) {
        return (
            <div style={styles.empty}>
                <Icon name="lock" size={48} color="#334155" />
                <p style={styles.emptyTitle}>Select an Item</p>
                <p style={styles.emptyText}>Choose an item from the list to view its details</p>
            </div>
        );
    }

    const handleCopy = (value: string) => {
        navigator.clipboard.writeText(value);
    };

    const handleDelete = async () => {
        if (confirm(`Delete "${item.name}"? This cannot be undone.`)) {
            await deleteItem(item.id);
        }
    };

    const fieldEntries = Object.entries(item.fields).filter(
        ([, v]) => v.length > 0,
    );

    return (
        <div style={styles.container} className="animate-fade-in" key={item.id}>
            {/* Header */}
            <div style={styles.header}>
                <div style={styles.headerLeft}>
                    <div style={styles.titleIcon}>
                        <Icon name="key" fill size={20} color="#5856D6" />
                    </div>
                    <div>
                        <h2 style={styles.title}>{item.name}</h2>
                        <div style={styles.metaRow}>
                            <span style={styles.typeBadge}>{item.type}</span>
                            {item.tags.map((t) => (
                                <span key={t} style={styles.tag}>{t}</span>
                            ))}
                        </div>
                    </div>
                </div>
                <div style={styles.headerActions}>
                    <button
                        style={styles.iconBtn}
                        onClick={() => toggleFavorite(item.id)}
                        title="Favorite"
                    >
                        <Icon
                            name="star"
                            fill={item.favorite}
                            size={18}
                            color={item.favorite ? "#f59e0b" : "#64748b"}
                        />
                    </button>
                    <button style={styles.actionBtnPrimary} onClick={() => onEdit(item)}>
                        <Icon name="edit" size={16} color="#fff" />
                        Edit
                    </button>
                    <button style={styles.iconBtn} onClick={handleDelete} title="Delete">
                        <Icon name="delete" size={18} color="#ef4444" />
                    </button>
                    {onViewVersions && (
                        <button style={styles.iconBtn} onClick={() => onViewVersions(item)} title="History">
                            <Icon name="history" size={18} color="#64748b" />
                        </button>
                    )}
                </div>
            </div>

            {/* Credential Card */}
            <div style={styles.credCard}>
                <div style={styles.credHeader}>
                    <Icon name="badge" size={16} color="#64748b" />
                    <span style={styles.credTitle}>Credentials</span>
                </div>

                {fieldEntries.map(([key, value]) => (
                    <FieldRow
                        key={key}
                        label={key}
                        value={value}
                        icon={fieldIcons[key] ?? "text_fields"}
                        sensitive={sensitiveFields.has(key)}
                        onCopy={() => handleCopy(value)}
                    />
                ))}
            </div>

            {/* Footer */}
            <div style={styles.footer}>
                <span style={styles.footerText}>
                    Created {new Date(item.created_at).toLocaleDateString()}
                </span>
                <span style={styles.footerDot}>·</span>
                <span style={styles.footerText}>
                    Modified {new Date(item.updated_at).toLocaleDateString()}
                </span>
            </div>
        </div>
    );
}

function FieldRow({
    label,
    value,
    icon,
    sensitive,
    onCopy,
}: {
    label: string;
    value: string;
    icon: string;
    sensitive: boolean;
    onCopy: () => void;
}) {
    const [revealed, setRevealed] = useState(false);

    return (
        <div style={styles.fieldRow}>
            <div style={styles.fieldLeft}>
                <Icon name={icon} size={14} color="#64748b" />
                <span style={styles.fieldLabel}>
                    {label.replace(/_/g, " ")}
                </span>
            </div>
            <div style={styles.fieldRight}>
                <span style={styles.fieldValue}>
                    {sensitive && !revealed ? "••••••••••••" : value}
                </span>
                <div style={styles.fieldActions}>
                    {sensitive && (
                        <button
                            style={styles.fieldBtn}
                            onClick={() => setRevealed(!revealed)}
                        >
                            <Icon
                                name={revealed ? "visibility_off" : "visibility"}
                                size={14}
                                color="#64748b"
                            />
                        </button>
                    )}
                    <button style={styles.fieldBtn} onClick={onCopy}>
                        <Icon name="content_copy" size={14} color="#64748b" />
                    </button>
                </div>
            </div>
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    container: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        padding: 24,
        overflow: "auto",
        background: "#334155",
        fontFamily: "'Inter', sans-serif",
        color: "#e2e8f0",
    },
    empty: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 12,
        background: "#334155",
    },
    emptyTitle: {
        fontSize: 16,
        fontWeight: 600,
        color: "#94a3b8",
    },
    emptyText: {
        fontSize: 13,
        color: "#64748b",
    },
    header: {
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "space-between",
        gap: 16,
        marginBottom: 24,
    },
    headerLeft: {
        display: "flex",
        alignItems: "center",
        gap: 12,
    },
    titleIcon: {
        width: 40,
        height: 40,
        background: "rgba(88, 86, 214, 0.15)",
        borderRadius: 8,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    title: {
        fontSize: 20,
        fontWeight: 700,
        color: "#fff",
        letterSpacing: "-0.01em",
    },
    metaRow: {
        display: "flex",
        gap: 6,
        marginTop: 4,
    },
    typeBadge: {
        background: "rgba(88, 86, 214, 0.2)",
        color: "#818cf8",
        padding: "2px 8px",
        borderRadius: 4,
        fontSize: 10,
        fontWeight: 700,
        textTransform: "capitalize",
    },
    tag: {
        background: "rgba(255, 255, 255, 0.06)",
        color: "#94a3b8",
        padding: "2px 8px",
        borderRadius: 4,
        fontSize: 10,
        fontWeight: 500,
    },
    headerActions: {
        display: "flex",
        gap: 6,
        alignItems: "center",
    },
    iconBtn: {
        width: 32,
        height: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "rgba(255, 255, 255, 0.06)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 6,
        cursor: "pointer",
    },
    actionBtnPrimary: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        padding: "6px 14px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 6,
        fontSize: 12,
        fontWeight: 600,
        cursor: "pointer",
    },
    credCard: {
        background: "rgba(15, 23, 42, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 12,
        overflow: "hidden",
        flex: 1,
    },
    credHeader: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "12px 16px",
        borderBottom: "1px solid rgba(255, 255, 255, 0.06)",
    },
    credTitle: {
        fontSize: 12,
        fontWeight: 600,
        color: "#94a3b8",
    },
    fieldRow: {
        padding: "12px 16px",
        borderBottom: "1px solid rgba(255, 255, 255, 0.04)",
    },
    fieldLeft: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        marginBottom: 4,
    },
    fieldLabel: {
        fontSize: 10,
        fontWeight: 600,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.05em",
    },
    fieldRight: {
        display: "flex",
        alignItems: "center",
        gap: 8,
    },
    fieldValue: {
        flex: 1,
        fontSize: 14,
        color: "#fff",
        fontFamily: "'JetBrains Mono', monospace",
        wordBreak: "break-all",
    },
    fieldActions: {
        display: "flex",
        gap: 2,
    },
    fieldBtn: {
        width: 28,
        height: 28,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "rgba(255, 255, 255, 0.04)",
        border: "1px solid rgba(255, 255, 255, 0.06)",
        borderRadius: 4,
        cursor: "pointer",
    },
    footer: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        paddingTop: 16,
        marginTop: 16,
        borderTop: "1px solid rgba(255, 255, 255, 0.06)",
    },
    footerText: {
        fontSize: 11,
        color: "#64748b",
    },
    footerDot: {
        fontSize: 11,
        color: "#475569",
    },
};
