import { type CSSProperties, useEffect, useState } from "react";
import * as cmd from "../../lib/commands";

interface ItemVersion {
    version: number;
    timestamp: string;
    fields_changed: string[];
}

export function VersionHistory({
    itemId,
    itemName,
    onClose,
}: {
    itemId: string;
    itemName: string;
    onClose: () => void;
}) {
    const [versions, setVersions] = useState<ItemVersion[]>([]);
    const [loading, setLoading] = useState(true);
    const [restoring, setRestoring] = useState<number | null>(null);

    useEffect(() => {
        let active = true;
        (async () => {
            try {
                const json = await cmd.getVersionHistory(itemId);
                if (active) setVersions(JSON.parse(json));
            } catch {
                // no versions available
            } finally {
                if (active) setLoading(false);
            }
        })();
        return () => { active = false; };
    }, [itemId]);

    const handleRestore = async (version: number) => {
        setRestoring(version);
        try {
            await cmd.restoreVersion(itemId, version);
            onClose();
        } catch {
            // noop
        } finally {
            setRestoring(null);
        }
    };

    return (
        <div style={styles.overlay} onClick={onClose}>
            <div style={styles.dialog} onClick={(e) => e.stopPropagation()} className="animate-scale-in">
                <div style={styles.header}>
                    <div>
                        <h2 style={styles.title}>Version History</h2>
                        <p style={styles.subtitle}>{itemName}</p>
                    </div>
                    <button style={styles.closeBtn} onClick={onClose}>✕</button>
                </div>

                {loading && <p style={styles.hint}>Loading versions…</p>}

                {!loading && versions.length === 0 && (
                    <p style={styles.hint}>No version history available for this item.</p>
                )}

                {versions.length > 0 && (
                    <div style={styles.list}>
                        {versions.map((v) => (
                            <div key={v.version} style={styles.row}>
                                <div style={styles.rowLeft}>
                                    <span style={styles.versionNum}>v{v.version}</span>
                                    <span style={styles.timestamp}>
                                        {new Date(v.timestamp).toLocaleString()}
                                    </span>
                                    {v.fields_changed.length > 0 && (
                                        <span style={styles.fields}>
                                            Changed: {v.fields_changed.join(", ")}
                                        </span>
                                    )}
                                </div>
                                <button
                                    style={styles.restoreBtn}
                                    onClick={() => handleRestore(v.version)}
                                    disabled={restoring !== null}
                                >
                                    {restoring === v.version ? "Restoring…" : "Restore"}
                                </button>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    overlay: {
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.6)",
        backdropFilter: "blur(8px)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1500,
    },
    dialog: {
        background: "#1a1a2e",
        borderRadius: 16,
        width: 480,
        maxWidth: "90vw",
        maxHeight: "70vh",
        overflow: "auto",
        boxShadow: "0 24px 64px rgba(0,0,0,0.5)",
        padding: 24,
        border: "1px solid rgba(255,255,255,0.06)",
    },
    header: {
        display: "flex",
        justifyContent: "space-between",
        marginBottom: 16,
    },
    title: { fontSize: 18, fontWeight: 700, color: "#fff" },
    subtitle: { fontSize: 13, color: "rgba(255,255,255,0.5)" },
    closeBtn: {
        background: "transparent",
        border: "none",
        fontSize: 18,
        color: "rgba(255,255,255,0.35)",
        cursor: "pointer",
    },
    hint: { fontSize: 13, color: "rgba(255,255,255,0.35)", textAlign: "center" },
    list: {
        display: "flex",
        flexDirection: "column",
        gap: 8,
    },
    row: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: 12,
        border: "1px solid rgba(255,255,255,0.06)",
        borderRadius: 10,
        background: "rgba(255,255,255,0.02)",
    },
    rowLeft: {
        display: "flex",
        flexDirection: "column",
        gap: 2,
    },
    versionNum: { fontSize: 13, fontWeight: 600, color: "#fff" },
    timestamp: { fontSize: 11, color: "rgba(255,255,255,0.35)" },
    fields: { fontSize: 11, color: "rgba(255,255,255,0.5)" },
    restoreBtn: {
        background: "transparent",
        color: "#5856D6",
        border: "1px solid rgba(88,86,214,0.3)",
        borderRadius: 6,
        padding: "4px 12px",
        fontSize: 11,
        cursor: "pointer",
        flexShrink: 0,
    },
};
