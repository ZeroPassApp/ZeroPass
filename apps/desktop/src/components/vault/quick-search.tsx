import { type CSSProperties, useState, useEffect, useRef, useCallback } from "react";
import type { VaultItem } from "../../lib/types";
import * as cmd from "../../lib/commands";
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

export function QuickSearch({ onClose }: { onClose: () => void }) {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<VaultItem[]>([]);
    const [selectedIdx, setSelectedIdx] = useState(0);
    const [loading, setLoading] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);
    const selectItem = useVaultStore((s) => s.selectItem);

    useEffect(() => {
        inputRef.current?.focus();
    }, []);

    const doSearch = useCallback(async (q: string) => {
        if (!q.trim()) {
            setResults([]);
            return;
        }
        setLoading(true);
        try {
            const json = await cmd.searchItems(q);
            const parsed: VaultItem[] = JSON.parse(json);
            setResults(parsed);
            setSelectedIdx(0);
        } catch {
            setResults([]);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        const timer = setTimeout(() => doSearch(query), 150);
        return () => clearTimeout(timer);
    }, [query, doSearch]);

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Escape") {
            onClose();
        } else if (e.key === "ArrowDown") {
            e.preventDefault();
            setSelectedIdx((i) => Math.min(i + 1, results.length - 1));
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            setSelectedIdx((i) => Math.max(i - 1, 0));
        } else if (e.key === "Enter" && results[selectedIdx]) {
            selectItem(results[selectedIdx].id);
            onClose();
        }
    };

    // Group results by type
    const grouped = results.reduce<Record<string, VaultItem[]>>((acc, item) => {
        const t = item.type;
        if (!acc[t]) acc[t] = [];
        acc[t].push(item);
        return acc;
    }, {});

    let flatIdx = 0;

    return (
        <div style={styles.overlay} onClick={onClose}>
            <div
                style={styles.panel}
                onClick={(e) => e.stopPropagation()}
                className="animate-scale-in"
            >
                {/* Search Input */}
                <div style={styles.inputRow}>
                    <Icon name="search" size={20} color="#64748b" />
                    <input
                        ref={inputRef}
                        style={styles.input}
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Search vault items…"
                    />
                    <div style={styles.kbdBadge}>
                        <span style={styles.kbdText}>⌘K</span>
                    </div>
                </div>

                {loading && (
                    <div style={styles.status}>
                        <Icon name="sync" size={16} color="#64748b" />
                        <span>Searching…</span>
                    </div>
                )}

                {/* Grouped Results */}
                {Object.keys(grouped).length > 0 && (
                    <div style={styles.results}>
                        {Object.entries(grouped).map(([type, items]) => (
                            <div key={type}>
                                <div style={styles.groupLabel}>
                                    {type.charAt(0).toUpperCase() + type.slice(1)}s
                                </div>
                                {items.map((item) => {
                                    const idx = flatIdx++;
                                    return (
                                        <button
                                            key={item.id}
                                            style={{
                                                ...styles.resultRow,
                                                ...(idx === selectedIdx
                                                    ? styles.resultSelected
                                                    : {}),
                                            }}
                                            onClick={() => {
                                                selectItem(item.id);
                                                onClose();
                                            }}
                                            onMouseEnter={() =>
                                                setSelectedIdx(idx)
                                            }
                                        >
                                            <div style={styles.resultIcon}>
                                                <Icon
                                                    name={typeIcons[item.type] ?? "inventory_2"}
                                                    fill
                                                    size={16}
                                                    color="#818cf8"
                                                />
                                            </div>
                                            <div style={styles.resultContent}>
                                                <span
                                                    style={styles.resultName}
                                                >
                                                    {item.name}
                                                </span>
                                                <span
                                                    style={styles.resultSub}
                                                >
                                                    {item.fields.username ||
                                                        item.fields.url ||
                                                        item.type}
                                                </span>
                                            </div>
                                            {idx === selectedIdx && (
                                                <Icon
                                                    name="subdirectory_arrow_left"
                                                    size={14}
                                                    color="#64748b"
                                                />
                                            )}
                                        </button>
                                    );
                                })}
                            </div>
                        ))}
                    </div>
                )}

                {query && !loading && results.length === 0 && (
                    <div style={styles.emptyState}>
                        <Icon name="search_off" size={32} color="#334155" />
                        <p style={styles.emptyText}>No results for "{query}"</p>
                    </div>
                )}

                {/* Footer Hints */}
                <div style={styles.footer}>
                    <div style={styles.footerHint}>
                        <kbd style={styles.kbd}>↑↓</kbd>
                        <span>Navigate</span>
                    </div>
                    <div style={styles.footerHint}>
                        <kbd style={styles.kbd}>↵</kbd>
                        <span>Open</span>
                    </div>
                    <div style={styles.footerHint}>
                        <kbd style={styles.kbd}>⌘C</kbd>
                        <span>Copy Password</span>
                    </div>
                    <div style={styles.footerHint}>
                        <kbd style={styles.kbd}>ESC</kbd>
                        <span>Close</span>
                    </div>
                </div>
            </div>
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    overlay: {
        position: "fixed",
        inset: 0,
        background: "rgba(2, 6, 23, 0.8)",
        backdropFilter: "blur(8px)",
        display: "flex",
        justifyContent: "center",
        paddingTop: 153,
        zIndex: 2000,
        fontFamily: "'Inter', sans-serif",
    },
    panel: {
        background: "rgba(15, 23, 42, 0.95)",
        border: "1px solid #1e293b",
        borderRadius: 16,
        width: 560,
        maxWidth: "90vw",
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
        overflow: "hidden",
        maxHeight: "60vh",
        display: "flex",
        flexDirection: "column",
    },
    inputRow: {
        display: "flex",
        alignItems: "center",
        gap: 12,
        padding: "16px 20px",
        borderBottom: "1px solid #1e293b",
    },
    input: {
        flex: 1,
        background: "transparent",
        border: "none",
        color: "#fff",
        fontSize: 16,
        fontWeight: 500,
        outline: "none",
    },
    kbdBadge: {
        background: "rgba(255, 255, 255, 0.06)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 6,
        padding: "4px 8px",
    },
    kbdText: {
        fontSize: 11,
        fontWeight: 600,
        color: "#64748b",
        fontFamily: "'Inter', sans-serif",
    },
    status: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 8,
        padding: 16,
        color: "#64748b",
        fontSize: 13,
    },
    results: {
        flex: 1,
        overflow: "auto",
        padding: "8px 8px",
    },
    groupLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#475569",
        textTransform: "uppercase",
        letterSpacing: "0.08em",
        padding: "12px 12px 6px",
    },
    resultRow: {
        display: "flex",
        alignItems: "center",
        gap: 12,
        padding: "8px 12px",
        border: "none",
        background: "transparent",
        borderRadius: 8,
        cursor: "pointer",
        width: "100%",
        textAlign: "left",
        transition: "background 0.1s",
    },
    resultSelected: {
        background: "rgba(88, 86, 214, 0.15)",
    },
    resultIcon: {
        width: 32,
        height: 32,
        borderRadius: 6,
        background: "rgba(88, 86, 214, 0.1)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
    },
    resultContent: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        minWidth: 0,
    },
    resultName: {
        fontSize: 13,
        fontWeight: 600,
        color: "#e2e8f0",
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
    },
    resultSub: {
        fontSize: 11,
        color: "#64748b",
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
    },
    emptyState: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 8,
        padding: "32px 16px",
    },
    emptyText: {
        fontSize: 13,
        color: "#64748b",
    },
    footer: {
        display: "flex",
        alignItems: "center",
        gap: 16,
        padding: "10px 20px",
        borderTop: "1px solid #1e293b",
        background: "rgba(2, 6, 23, 0.5)",
    },
    footerHint: {
        display: "flex",
        alignItems: "center",
        gap: 4,
        fontSize: 11,
        color: "#475569",
    },
    kbd: {
        background: "rgba(255, 255, 255, 0.06)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 3,
        padding: "1px 5px",
        fontSize: 10,
        fontFamily: "'JetBrains Mono', monospace",
        color: "#64748b",
    },
};
