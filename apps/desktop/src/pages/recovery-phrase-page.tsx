import { type CSSProperties, useState } from "react";
import { useAuthStore } from "../stores/auth-store";
import { Icon } from "../components/common/icon";

export function RecoveryPhrasePage() {
    const { recoveryPhrase, confirmRecovery } = useAuthStore();
    const [confirmed, setConfirmed] = useState(false);
    const [copied, setCopied] = useState(false);

    const words = recoveryPhrase?.split(" ") ?? [];
    const isWordList = words.length >= 4;

    const handleCopy = async () => {
        if (recoveryPhrase) {
            await navigator.clipboard.writeText(recoveryPhrase);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    return (
        <div style={styles.container} className="animate-fade-in">
            {/* Background */}
            <div style={styles.bgLayer}>
                <div style={styles.orbTop} />
                <div style={styles.orbBottom} />
            </div>

            <main style={styles.main}>
                {/* Header */}
                <div style={styles.header}>
                    <div style={styles.iconBox}>
                        <Icon name="shield_lock" fill size={36} color="#818cf8" />
                    </div>
                    <h1 style={styles.title}>Recovery Phrase</h1>
                    <p style={styles.subtitle}>
                        Write down these words in order and store them securely.
                        <br />
                        This is your only backup to recover your vault.
                    </p>
                </div>

                {/* Word Grid */}
                {isWordList ? (
                    <div style={styles.wordGrid}>
                        {words.map((word, i) => (
                            <div key={i} style={styles.wordCell}>
                                <span style={styles.wordIndex}>
                                    {String(i + 1).padStart(2, "0")}
                                </span>
                                <span style={styles.word}>{word}</span>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div style={styles.keyBox}>
                        <code style={styles.keyText}>{recoveryPhrase}</code>
                    </div>
                )}

                {/* Warning Banner */}
                <div style={styles.warningBox}>
                    <Icon name="warning" fill size={20} color="#FF9500" style={{ flexShrink: 0 }} />
                    <p style={styles.warningText}>
                        <span style={{ fontWeight: 600, color: "#FF9500" }}>
                            This is the only time this phrase will be shown.
                        </span>{" "}
                        If you lose it, there is no way to recover your vault data.
                    </p>
                </div>

                {/* Action Buttons */}
                <div style={styles.actions}>
                    <button style={styles.copyBtn} onClick={handleCopy}>
                        <Icon
                            name={copied ? "check_circle" : "content_copy"}
                            size={16}
                            color={copied ? "#34C759" : "#5856D6"}
                        />
                        {copied ? "Copied!" : "Copy to Clipboard"}
                    </button>
                </div>

                {/* Confirm Checkbox */}
                <label style={styles.checkboxRow}>
                    <div
                        style={{
                            ...styles.checkBox,
                            ...(confirmed ? styles.checkBoxChecked : {}),
                        }}
                        onClick={() => setConfirmed(!confirmed)}
                    >
                        {confirmed && (
                            <Icon name="check" size={14} color="#fff" />
                        )}
                    </div>
                    <span style={styles.checkLabel}>
                        I have safely stored my recovery phrase
                    </span>
                </label>

                {/* Continue */}
                <button
                    style={{
                        ...styles.continueBtn,
                        opacity: confirmed ? 1 : 0.4,
                    }}
                    disabled={!confirmed}
                    onClick={confirmRecovery}
                >
                    Continue to Vault
                    <Icon name="arrow_forward" size={18} color="#fff" />
                </button>

                {/* Security Footer Badges */}
                <div style={styles.badges}>
                    {[
                        { icon: "verified_user", label: "AES-256 Encrypted" },
                        { icon: "visibility_off", label: "Zero-Knowledge" },
                        { icon: "lock", label: "BIP-39 Standard" },
                    ].map((b) => (
                        <div key={b.label} style={styles.badge}>
                            <Icon name={b.icon} size={12} color="#5856D6" />
                            <span style={styles.badgeText}>{b.label}</span>
                        </div>
                    ))}
                </div>
            </main>
        </div>
    );
}

/* ─── Styles ───────────────────────────────────────────────── */

const styles: Record<string, CSSProperties> = {
    container: {
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        background: "#020617",
        fontFamily: "'Inter', sans-serif",
        color: "#e2e8f0",
        position: "relative",
        overflow: "auto",
    },
    bgLayer: {
        position: "fixed",
        inset: 0,
        zIndex: 0,
        overflow: "hidden",
        pointerEvents: "none",
    },
    orbTop: {
        position: "absolute",
        top: "-15%",
        right: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(88, 86, 214, 0.15)",
        filter: "blur(120px)",
        borderRadius: "50%",
    },
    orbBottom: {
        position: "absolute",
        bottom: "-15%",
        left: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(88, 86, 214, 0.08)",
        filter: "blur(120px)",
        borderRadius: "50%",
    },
    main: {
        position: "relative",
        zIndex: 10,
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "48px 24px",
        maxWidth: 560,
        width: "100%",
        margin: "0 auto",
        gap: 24,
    },
    header: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        textAlign: "center",
        gap: 12,
        marginBottom: 8,
    },
    iconBox: {
        width: 64,
        height: 64,
        background: "rgba(79, 70, 229, 0.2)",
        border: "1px solid rgba(99, 102, 241, 0.3)",
        borderRadius: 16,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        marginBottom: 8,
    },
    title: {
        fontSize: 28,
        fontWeight: 800,
        color: "#fff",
        letterSpacing: "-0.025em",
    },
    subtitle: {
        fontSize: 14,
        color: "#94a3b8",
        lineHeight: 1.6,
    },
    wordGrid: {
        display: "grid",
        gridTemplateColumns: "repeat(3, 1fr)",
        gap: 8,
        width: "100%",
    },
    wordCell: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        background: "rgba(15, 23, 42, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 8,
        padding: "10px 12px",
    },
    wordIndex: {
        fontSize: 10,
        fontWeight: 700,
        color: "#475569",
        minWidth: 20,
    },
    word: {
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 14,
        fontWeight: 500,
        color: "#fff",
    },
    keyBox: {
        background: "rgba(15, 23, 42, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 8,
        padding: 16,
        width: "100%",
        wordBreak: "break-all",
    },
    keyText: {
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 13,
        color: "#fff",
    },
    warningBox: {
        display: "flex",
        gap: 12,
        padding: 16,
        background: "rgba(255, 149, 0, 0.06)",
        border: "1px solid rgba(255, 149, 0, 0.2)",
        borderRadius: 8,
        width: "100%",
    },
    warningText: {
        fontSize: 13,
        lineHeight: 1.6,
        color: "#94a3b8",
    },
    actions: {
        display: "flex",
        gap: 12,
    },
    copyBtn: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "10px 20px",
        background: "rgba(88, 86, 214, 0.1)",
        border: "1px solid rgba(88, 86, 214, 0.3)",
        borderRadius: 8,
        color: "#5856D6",
        fontSize: 13,
        fontWeight: 600,
        cursor: "pointer",
        transition: "all 0.15s",
    },
    checkboxRow: {
        display: "flex",
        alignItems: "center",
        gap: 12,
        cursor: "pointer",
        width: "100%",
        padding: "16px 0",
        borderTop: "1px solid #1e293b",
        borderBottom: "1px solid #1e293b",
    },
    checkBox: {
        width: 20,
        height: 20,
        borderRadius: 4,
        border: "2px solid #334155",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
        cursor: "pointer",
    },
    checkBoxChecked: {
        background: "#5856D6",
        borderColor: "#5856D6",
    },
    checkLabel: {
        fontSize: 13,
        color: "#94a3b8",
        fontWeight: 500,
    },
    continueBtn: {
        width: "100%",
        padding: "14px 24px",
        background: "#4f46e5",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 700,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 8,
        cursor: "pointer",
        boxShadow: "0 4px 14px rgba(79, 70, 229, 0.2)",
        transition: "all 0.15s",
    },
    badges: {
        display: "flex",
        gap: 16,
        marginTop: 8,
    },
    badge: {
        display: "flex",
        alignItems: "center",
        gap: 6,
    },
    badgeText: {
        fontSize: 10,
        fontWeight: 600,
        color: "#64748b",
    },
};
