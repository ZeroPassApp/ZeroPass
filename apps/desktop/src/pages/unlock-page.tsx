import { type CSSProperties, useState } from "react";
import { useAuthStore } from "../stores/auth-store";
import { Icon } from "../components/common/icon";

export function UnlockPage() {
    const { unlock, unlockWithRecovery, error, setError, vaultPath } =
        useAuthStore();
    const [mode, setMode] = useState<"password" | "recovery" | "biometric">(
        "password",
    );

    const vaultName = vaultPath?.split("/").pop() ?? "Default Vault";
    const displayPath = vaultPath ?? "~/.zeropass/vaults/default";

    return (
        <div style={styles.container} className="animate-fade-in">
            {/* Background mesh + decorative orbs */}
            <div style={styles.bgDecor}>
                <div style={styles.orbTopRight} />
                <div style={styles.orbBottomLeft} />
            </div>

            <main style={styles.main}>
                {/* Brand & Vault Header */}
                <div style={styles.header}>
                    <div style={styles.iconBox}>
                        <Icon
                            name="enhanced_encryption"
                            fill
                            size={36}
                            color="#818cf8"
                        />
                    </div>
                    <h1 style={styles.title}>{vaultName}</h1>
                    <p style={styles.path}>{displayPath}</p>
                </div>

                {/* Glassmorphic Interaction Card */}
                <div style={styles.card}>
                    {/* Tab Switcher */}
                    <div style={styles.tabBar}>
                        {(["password", "recovery", "biometric"] as const).map(
                            (m) => (
                                <button
                                    key={m}
                                    style={
                                        mode === m
                                            ? styles.tabActive
                                            : styles.tab
                                    }
                                    onClick={() => {
                                        setMode(m);
                                        setError(null);
                                    }}
                                >
                                    {m.charAt(0).toUpperCase() + m.slice(1)}
                                </button>
                            ),
                        )}
                    </div>

                    {mode === "password" && (
                        <PasswordForm onSubmit={unlock} error={error} />
                    )}
                    {mode === "recovery" && (
                        <RecoveryForm
                            onSubmit={unlockWithRecovery}
                            error={error}
                        />
                    )}
                    {mode === "biometric" && (
                        <div style={styles.bioSection}>
                            <div style={styles.bioGlow} />
                            <Icon
                                name="fingerprint"
                                size={72}
                                color="#5856D6"
                                style={{
                                    position: "relative",
                                    fontVariationSettings:
                                        "'FILL' 0, 'wght' 200",
                                }}
                            />
                            <button style={styles.bioBtn}>
                                Authenticate with Touch ID
                            </button>
                        </div>
                    )}
                </div>

                {/* Footer Links */}
                <div style={styles.links}>
                    {mode !== "recovery" && (
                        <button
                            style={styles.linkBtn}
                            onClick={() => {
                                setMode("recovery");
                                setError(null);
                            }}
                        >
                            Forgot password? Use Recovery Phrase
                        </button>
                    )}
                    <div style={styles.linkSep} />
                    <button
                        style={styles.linkBtnUpper}
                        onClick={() => useAuthStore.getState().reset()}
                    >
                        Open Different Vault
                    </button>
                </div>
            </main>

            {/* Footer */}
            <footer style={styles.footer}>
                <span style={styles.footerLeft}>
                    © 2024 ZeroPass Security. AES-256 Encrypted.
                </span>
                <div style={styles.footerLinks}>
                    <span style={styles.footerLink}>Privacy Policy</span>
                    <span style={styles.footerLink}>Security Whitepaper</span>
                    <span style={styles.footerLink}>Support</span>
                </div>
            </footer>
        </div>
    );
}

/* ─── Password Form ────────────────────────────────────────── */

function PasswordForm({
    onSubmit,
    error,
}: {
    onSubmit: (pw: string) => Promise<void>;
    error: string | null;
}) {
    const [password, setPassword] = useState("");
    const [showPw, setShowPw] = useState(false);
    const [loading, setLoading] = useState(false);

    const handleSubmit = async () => {
        if (!password || loading) return;
        setLoading(true);
        await onSubmit(password);
        setLoading(false);
    };

    return (
        <div style={styles.formSection}>
            <div>
                <label style={styles.inputLabel}>Master Password</label>
                <div style={styles.inputRow}>
                    <input
                        style={styles.input}
                        type={showPw ? "text" : "password"}
                        placeholder="Enter your password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
                        autoFocus
                    />
                    <button
                        style={styles.visBtn}
                        onClick={() => setShowPw(!showPw)}
                        type="button"
                    >
                        <Icon
                            name={showPw ? "visibility_off" : "visibility"}
                            size={20}
                            color="#64748b"
                        />
                    </button>
                </div>
            </div>

            {error && (
                <div style={styles.errorBox}>
                    <Icon name="error" size={16} color="#f87171" />
                    <span>{error}</span>
                </div>
            )}

            <button
                style={{
                    ...styles.unlockBtn,
                    opacity: loading ? 0.7 : 1,
                }}
                onClick={handleSubmit}
                disabled={loading}
            >
                <span>{loading ? "Unlocking…" : "Unlock"}</span>
                <Icon name="lock_open" size={18} color="#fff" />
            </button>
        </div>
    );
}

/* ─── Recovery Form ────────────────────────────────────────── */

function RecoveryForm({
    onSubmit,
    error,
}: {
    onSubmit: (phrase: string) => Promise<void>;
    error: string | null;
}) {
    const [words, setWords] = useState<string[]>(Array(12).fill(""));
    const [loading, setLoading] = useState(false);

    const handleChange = (i: number, v: string) => {
        const next = [...words];
        next[i] = v;
        setWords(next);
    };

    const handleSubmit = async () => {
        const phrase = words.join(" ").trim();
        if (!phrase || loading) return;
        setLoading(true);
        await onSubmit(phrase);
        setLoading(false);
    };

    return (
        <div style={styles.formSection}>
            <div style={styles.wordGrid}>
                {words.map((w, i) => (
                    <div key={i} style={styles.wordSlot}>
                        <span style={styles.wordNum}>
                            {String(i + 1).padStart(2, "0")}
                        </span>
                        <input
                            style={styles.wordInput}
                            type="text"
                            value={w}
                            onChange={(e) => handleChange(i, e.target.value)}
                        />
                    </div>
                ))}
            </div>

            {error && (
                <div style={styles.errorBox}>
                    <Icon name="error" size={16} color="#f87171" />
                    <span>{error}</span>
                </div>
            )}

            <button
                style={{
                    ...styles.recoveryBtn,
                    opacity: loading ? 0.7 : 1,
                }}
                onClick={handleSubmit}
                disabled={loading}
            >
                {loading ? "Unlocking…" : "Unlock with Recovery"}
            </button>
        </div>
    );
}

/* ─── Styles ───────────────────────────────────────────────── */

const styles: Record<string, CSSProperties> = {
    container: {
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        background: "radial-gradient(circle at 50% 50%, #1e1b4b 0%, #020617 100%)",
        fontFamily: "'Inter', sans-serif",
        color: "#e2e8f0",
        position: "relative",
    },
    bgDecor: {
        position: "fixed",
        inset: 0,
        zIndex: 0,
        overflow: "hidden",
        pointerEvents: "none",
    },
    orbTopRight: {
        position: "absolute",
        top: "-10%",
        right: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(30, 27, 75, 0.4)",
        filter: "blur(120px)",
        borderRadius: "50%",
    },
    orbBottomLeft: {
        position: "absolute",
        bottom: "-10%",
        left: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(88, 86, 214, 0.05)",
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
        padding: 24,
        maxWidth: 440,
        width: "100%",
        margin: "0 auto",
        gap: 32,
    },
    header: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        textAlign: "center",
        gap: 16,
    },
    iconBox: {
        width: 64,
        height: 64,
        background: "rgba(79, 70, 229, 0.2)",
        borderRadius: 16,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        border: "1px solid rgba(99, 102, 241, 0.3)",
        boxShadow: "0 0 20px rgba(88, 86, 214, 0.3)",
    },
    title: {
        fontSize: 24,
        fontWeight: 800,
        letterSpacing: "-0.025em",
        color: "#fff",
    },
    path: {
        fontSize: 12,
        fontFamily: "monospace",
        color: "#64748b",
        opacity: 0.8,
    },
    card: {
        width: "100%",
        background: "rgba(15, 23, 42, 0.8)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.1)",
        borderRadius: 12,
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
        padding: 32,
        display: "flex",
        flexDirection: "column",
        gap: 24,
    },
    tabBar: {
        display: "flex",
        padding: 4,
        background: "rgba(15, 23, 42, 0.5)",
        borderRadius: 8,
        border: "1px solid #1e293b",
    },
    tab: {
        flex: 1,
        padding: "8px 0",
        fontSize: 12,
        fontWeight: 600,
        color: "#94a3b8",
        background: "none",
        border: "none",
        borderRadius: 6,
        cursor: "pointer",
        transition: "all 0.15s",
    },
    tabActive: {
        flex: 1,
        padding: "8px 0",
        fontSize: 12,
        fontWeight: 600,
        color: "#fff",
        background: "#4f46e5",
        border: "none",
        borderRadius: 6,
        cursor: "pointer",
        boxShadow: "0 1px 3px rgba(0,0,0,0.2)",
    },
    formSection: {
        display: "flex",
        flexDirection: "column",
        gap: 16,
    },
    inputLabel: {
        display: "block",
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.1em",
        marginBottom: 6,
        marginLeft: 4,
    },
    inputRow: {
        position: "relative",
    },
    input: {
        width: "100%",
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: "12px 48px 12px 16px",
        color: "#fff",
        fontSize: 14,
        outline: "none",
        transition: "all 0.15s",
    },
    visBtn: {
        position: "absolute",
        right: 12,
        top: "50%",
        transform: "translateY(-50%)",
        background: "none",
        border: "none",
        cursor: "pointer",
        padding: 0,
        display: "flex",
    },
    errorBox: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: 12,
        background: "rgba(239, 68, 68, 0.1)",
        border: "1px solid rgba(239, 68, 68, 0.2)",
        borderRadius: 8,
        color: "#f87171",
        fontSize: 14,
    },
    unlockBtn: {
        width: "100%",
        padding: "14px 0",
        background: "#4f46e5",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 700,
        cursor: "pointer",
        boxShadow: "0 4px 14px rgba(79, 70, 229, 0.2)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 8,
        transition: "all 0.15s",
    },
    wordGrid: {
        display: "grid",
        gridTemplateColumns: "repeat(3, 1fr)",
        gap: 8,
    },
    wordSlot: {
        display: "flex",
        flexDirection: "column",
        gap: 2,
    },
    wordNum: {
        fontSize: 9,
        fontWeight: 700,
        color: "#475569",
        marginLeft: 4,
    },
    wordInput: {
        width: "100%",
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 4,
        padding: "6px 8px",
        fontSize: 12,
        color: "#fff",
        outline: "none",
    },
    recoveryBtn: {
        width: "100%",
        padding: "14px 0",
        background: "#1e293b",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 700,
        cursor: "pointer",
        transition: "all 0.15s",
    },
    bioSection: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        padding: "16px 0",
        gap: 24,
        position: "relative",
    },
    bioGlow: {
        position: "absolute",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -70%)",
        width: 120,
        height: 120,
        background: "rgba(88, 86, 214, 0.2)",
        filter: "blur(40px)",
        borderRadius: "50%",
    },
    bioBtn: {
        width: "100%",
        padding: "14px 0",
        background: "#4f46e5",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 700,
        cursor: "pointer",
        transition: "all 0.15s",
    },
    links: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 16,
    },
    linkBtn: {
        fontSize: 12,
        color: "#94a3b8",
        background: "none",
        border: "none",
        cursor: "pointer",
        fontWeight: 500,
        transition: "color 0.15s",
    },
    linkSep: {
        height: 1,
        width: 32,
        background: "#1e293b",
    },
    linkBtnUpper: {
        fontSize: 11,
        color: "#64748b",
        background: "none",
        border: "none",
        cursor: "pointer",
        fontWeight: 700,
        textTransform: "uppercase",
        letterSpacing: "0.05em",
    },
    footer: {
        position: "relative",
        zIndex: 20,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        padding: "16px 32px",
        background: "#020617",
        borderTop: "1px solid #1e293b",
    },
    footerLeft: {
        fontSize: 12,
        color: "#64748b",
        fontWeight: 500,
    },
    footerLinks: {
        display: "flex",
        gap: 24,
    },
    footerLink: {
        fontSize: 12,
        color: "#64748b",
        fontWeight: 500,
        cursor: "pointer",
    },
};
