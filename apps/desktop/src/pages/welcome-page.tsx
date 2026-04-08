import { type CSSProperties, useEffect, useState } from "react";
import { open } from "@tauri-apps/plugin-dialog";
import { homeDir, join } from "@tauri-apps/api/path";
import { useAuthStore } from "../stores/auth-store";
import { Icon } from "../components/common/icon";

const features = [
    { icon: "lock", label: "AES-256" },
    { icon: "sync", label: "Cloud Sync" },
    { icon: "devices", label: "Cross-Platform" },
    { icon: "fingerprint", label: "Biometric" },
];

export function WelcomePage() {
    const { openVault, error } = useAuthStore();
    const [showCreate, setShowCreate] = useState(false);

    const handleOpenVault = async () => {
        const selected = await open({
            directory: true,
            title: "Select vault folder",
        });
        if (selected) {
            await openVault(selected);
        }
    };

    return (
        <div style={styles.container} className="animate-fade-in">
            {/* Background Layer: Ambient Gradients and Security Shapes */}
            <div style={styles.bgLayer}>
                <div style={styles.glowTopLeft} />
                <div style={styles.glowBottomRight} />
                <div style={styles.securityGrid} />
                {/* Abstract Security Circles */}
                <div style={styles.circle1Outer}>
                    <div style={styles.circle1Inner} />
                </div>
                <div style={styles.circle2Outer}>
                    <div style={styles.circle2Inner} />
                </div>
            </div>

            {/* Main Content */}
            <main style={styles.main}>
                {/* Logo Branding */}
                <div style={styles.logoGroup}>
                    <div style={styles.logoContainer}>
                        <div style={styles.logoBgTilt} />
                        <div style={styles.logoBox}>
                            <Icon name="shield" fill size={40} color="#fff" />
                            <span style={styles.logoZP}>ZP</span>
                        </div>
                    </div>
                    <h2 style={styles.brandName}>ZeroPass</h2>
                </div>

                {/* Hero Text */}
                <div style={styles.heroSection}>
                    <h1 style={styles.heading}>
                        Your Secrets, <br />
                        <span style={styles.headingGradient}>Zero Knowledge</span>
                    </h1>
                    <p style={styles.subtitle}>
                        A cross-platform password manager that never sees your data.
                        Built with end-to-end encryption for the modern web.
                    </p>
                </div>

                {/* CTA Action Cluster */}
                <div style={styles.actions}>
                    <button style={styles.primaryBtn} onClick={() => setShowCreate(true)}>
                        <Icon name="add_circle" fill size={20} color="#fff" />
                        Create New Vault
                    </button>
                    <button style={styles.secondaryBtn} onClick={handleOpenVault}>
                        <Icon name="key" size={20} color="#e2e8f0" />
                        Open Existing Vault
                    </button>
                </div>

                {/* Feature Pills */}
                <div style={styles.pillGrid}>
                    {features.map((f) => (
                        <div key={f.label} style={styles.pill}>
                            <Icon name={f.icon} size={18} color="#5856D6" />
                            <span style={styles.pillLabel}>{f.label}</span>
                        </div>
                    ))}
                </div>

                {error && <p style={styles.error}>{error}</p>}
            </main>

            {/* Footer */}
            <footer style={styles.footer}>
                <span style={styles.footerLeft}>
                    © 2024 ZeroPass Security. Version 1.0.0
                </span>
                <div style={styles.footerLinks}>
                    <span style={styles.footerLink}>Security Audit</span>
                    <span style={styles.footerLink}>Privacy Policy</span>
                    <span style={styles.footerLink}>Help Center</span>
                </div>
            </footer>

            {showCreate && (
                <CreateVaultDialog onClose={() => setShowCreate(false)} />
            )}
        </div>
    );
}

/* ─── Create Vault Dialog ──────────────────────────────────── */

function CreateVaultDialog({ onClose }: { onClose: () => void }) {
    const { createVault, error } = useAuthStore();
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");
    const defaultPath = "~/.zeropass/vaults/default";
    const [path, setPath] = useState(defaultPath);
    const [showPw, setShowPw] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    // Resolve ~ to actual home directory on mount
    useEffect(() => {
        homeDir().then((home) =>
            join(home, ".zeropass", "vaults", "default")
        ).then(setPath).catch(() => { });
    }, []);

    const handlePickFolder = async () => {
        const selected = await open({
            directory: true,
            title: "Choose vault location",
        });
        if (selected) setPath(selected);
    };

    const canSubmit =
        password.length >= 8 && password === confirm && path.length > 0;

    const handleCreate = async () => {
        if (!canSubmit) return;
        setSubmitting(true);
        await createVault(path, password);
        setSubmitting(false);
    };

    return (
        <div style={dlgStyles.overlay} onClick={onClose}>
            <div
                style={dlgStyles.card}
                onClick={(e) => e.stopPropagation()}
                className="animate-scale-in"
            >
                {/* Header */}
                <div style={dlgStyles.header}>
                    <div style={dlgStyles.headerIcon}>
                        <Icon name="encrypted" size={28} color="#5856D6" />
                    </div>
                    <div>
                        <h1 style={dlgStyles.title}>Create New Vault</h1>
                        <p style={dlgStyles.subtitle}>Initialize a new secure container</p>
                    </div>
                </div>

                {/* Body */}
                <div style={dlgStyles.body}>
                    {/* Vault Location */}
                    <div style={dlgStyles.field}>
                        <label style={dlgStyles.label}>Vault Location</label>
                        <div style={{ display: "flex", gap: 8 }}>
                            <div style={{ position: "relative", flex: 1 }}>
                                <Icon
                                    name="folder_open"
                                    size={14}
                                    color="#64748b"
                                    style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)" }}
                                />
                                <input
                                    style={dlgStyles.inputIcon}
                                    type="text"
                                    readOnly
                                    value={path}
                                />
                            </div>
                            <button style={dlgStyles.browseBtn} onClick={handlePickFolder}>
                                Browse…
                            </button>
                        </div>
                    </div>

                    {/* Master Password */}
                    <div style={dlgStyles.field}>
                        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
                            <label style={dlgStyles.label}>Master Password</label>
                            {password.length >= 8 && (
                                <span style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.05em", fontWeight: 700, color: "#34C759" }}>
                                    Strong
                                </span>
                            )}
                        </div>
                        <div style={{ position: "relative" }}>
                            <Icon
                                name="key"
                                size={14}
                                color="#64748b"
                                style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)", zIndex: 1 }}
                            />
                            <input
                                style={dlgStyles.inputIcon}
                                type={showPw ? "text" : "password"}
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="Minimum 8 characters"
                            />
                            <button
                                style={dlgStyles.visBtn}
                                onClick={() => setShowPw(!showPw)}
                                type="button"
                            >
                                <Icon name={showPw ? "visibility_off" : "visibility"} size={14} color="#64748b" />
                            </button>
                        </div>
                        {password.length > 0 && (
                            <div style={{ display: "flex", gap: 4, paddingTop: 4 }}>
                                {[0, 1, 2, 3].map((i) => (
                                    <div
                                        key={i}
                                        style={{
                                            height: 6,
                                            flex: 1,
                                            borderRadius: 3,
                                            background:
                                                password.length >= 8
                                                    ? "#34C759"
                                                    : i < 2
                                                        ? "#FF9500"
                                                        : "rgba(255,255,255,0.08)",
                                        }}
                                    />
                                ))}
                            </div>
                        )}
                    </div>

                    {/* Confirm Password */}
                    <div style={dlgStyles.field}>
                        <label style={dlgStyles.label}>Confirm Password</label>
                        <div style={{ position: "relative" }}>
                            <Icon
                                name="lock_reset"
                                size={14}
                                color="#64748b"
                                style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)", zIndex: 1 }}
                            />
                            <input
                                style={dlgStyles.inputIcon}
                                type="password"
                                value={confirm}
                                onChange={(e) => setConfirm(e.target.value)}
                                placeholder="Re-enter master password"
                            />
                        </div>
                        {confirm.length > 0 && password !== confirm && (
                            <span style={dlgStyles.fieldError}>Passwords do not match</span>
                        )}
                    </div>

                    {/* Info Box */}
                    <div style={dlgStyles.infoBox}>
                        <Icon name="info" size={20} color="#5856D6" style={{ flexShrink: 0 }} />
                        <p style={dlgStyles.infoText}>
                            Your master password{" "}
                            <strong style={{ color: "#e2e8f0", fontWeight: 500 }}>
                                cannot be recovered
                            </strong>
                            . Make sure to save your recovery phrase after creation.
                        </p>
                    </div>
                </div>

                {error && <p style={{ color: "#FF3B30", fontSize: 13, padding: "0 24px 12px" }}>{error}</p>}

                {/* Footer */}
                <div style={dlgStyles.footer}>
                    <button style={dlgStyles.cancelBtn} onClick={onClose}>
                        Cancel
                    </button>
                    <button
                        style={{
                            ...dlgStyles.createBtn,
                            opacity: canSubmit && !submitting ? 1 : 0.5,
                        }}
                        onClick={handleCreate}
                        disabled={!canSubmit || submitting}
                    >
                        {submitting ? "Creating…" : "Create Vault"}
                    </button>
                </div>
            </div>

            {/* Footnote */}
            <div style={{ marginTop: 32, textAlign: "center" as const }}>
                <div style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 8, marginBottom: 8 }}>
                    <div style={{ height: 1, width: 32, background: "#1e293b" }} />
                    <span style={{ fontSize: 10, fontWeight: 900, letterSpacing: "0.12em", textTransform: "uppercase" as const, color: "#475569" }}>
                        ZeroPass Security
                    </span>
                    <div style={{ height: 1, width: 32, background: "#1e293b" }} />
                </div>
                <p style={{ fontSize: 10, color: "#64748b" }}>
                    AES-256 Bit Encryption • Zero-Knowledge Architecture
                </p>
            </div>
        </div>
    );
}

/* ─── Styles: Welcome Page ─────────────────────────────────── */

const styles: Record<string, CSSProperties> = {
    container: {
        display: "flex",
        flexDirection: "column",
        height: "100%",
        background: "#020617",
        position: "relative",
        overflow: "hidden",
        color: "#e2e8f0",
        fontFamily: "'Inter', sans-serif",
    },
    bgLayer: {
        position: "fixed",
        inset: 0,
        zIndex: 0,
        overflow: "hidden",
    },
    glowTopLeft: {
        position: "absolute",
        top: "-10%",
        left: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(88, 86, 214, 0.2)",
        filter: "blur(120px)",
        borderRadius: "50%",
    },
    glowBottomRight: {
        position: "absolute",
        bottom: "-10%",
        right: "-10%",
        width: "40%",
        height: "40%",
        background: "rgba(88, 86, 214, 0.1)",
        filter: "blur(120px)",
        borderRadius: "50%",
    },
    securityGrid: {
        position: "absolute",
        inset: 0,
        backgroundImage:
            "radial-gradient(circle at 2px 2px, rgba(88, 86, 214, 0.15) 1px, transparent 0)",
        backgroundSize: "40px 40px",
        opacity: 0.3,
    },
    circle1Outer: {
        position: "absolute",
        top: "25%",
        left: "25%",
        width: 256,
        height: 256,
        border: "1px solid rgba(88, 86, 214, 0.1)",
        borderRadius: "50%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    circle1Inner: {
        width: 192,
        height: 192,
        border: "1px solid rgba(88, 86, 214, 0.05)",
        borderRadius: "50%",
    },
    circle2Outer: {
        position: "absolute",
        bottom: "25%",
        right: "25%",
        width: 384,
        height: 384,
        border: "1px solid rgba(88, 86, 214, 0.1)",
        borderRadius: "50%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        transform: "rotate(45deg)",
    },
    circle2Inner: {
        width: 256,
        height: 256,
        border: "1px solid rgba(88, 86, 214, 0.05)",
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
        padding: "0 24px",
        textAlign: "center",
    },
    logoGroup: {
        marginBottom: 48,
        cursor: "default",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
    },
    logoContainer: {
        position: "relative",
        width: 96,
        height: 96,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        marginBottom: 24,
    },
    logoBgTilt: {
        position: "absolute",
        inset: 0,
        background: "rgba(88, 86, 214, 0.2)",
        borderRadius: 16,
        transform: "rotate(6deg)",
        filter: "blur(2px)",
    },
    logoBox: {
        position: "relative",
        width: 80,
        height: 80,
        background: "#5856D6",
        borderRadius: 16,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        boxShadow: "0 0 40px rgba(88, 86, 214, 0.4)",
    },
    logoZP: {
        position: "absolute",
        color: "#fff",
        fontWeight: 800,
        fontSize: 10,
        letterSpacing: -0.5,
        marginTop: 4,
    },
    brandName: {
        fontSize: 24,
        fontWeight: 700,
        letterSpacing: "-0.025em",
        color: "#fff",
    },
    heroSection: {
        maxWidth: 640,
        marginBottom: 48,
    },
    heading: {
        fontSize: 56,
        fontWeight: 800,
        color: "#fff",
        letterSpacing: "-0.025em",
        lineHeight: 1.1,
        marginBottom: 24,
    },
    headingGradient: {
        background: "linear-gradient(to right, #5856D6, #818cf8)",
        WebkitBackgroundClip: "text",
        WebkitTextFillColor: "transparent",
    },
    subtitle: {
        fontSize: 18,
        color: "#94a3b8",
        fontWeight: 500,
        lineHeight: 1.6,
        maxWidth: 480,
        margin: "0 auto",
    },
    actions: {
        display: "flex",
        gap: 16,
        alignItems: "center",
        justifyContent: "center",
        marginBottom: 80,
    },
    primaryBtn: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "16px 32px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 600,
        cursor: "pointer",
        boxShadow: "0 0 20px rgba(88, 86, 214, 0.3)",
        transition: "all 0.2s",
    },
    secondaryBtn: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "16px 32px",
        background: "rgba(15, 23, 42, 0.5)",
        color: "#e2e8f0",
        border: "1px solid #334155",
        borderRadius: 8,
        fontSize: 15,
        fontWeight: 600,
        cursor: "pointer",
        backdropFilter: "blur(4px)",
        transition: "all 0.2s",
    },
    pillGrid: {
        display: "grid",
        gridTemplateColumns: "repeat(4, auto)",
        gap: 16,
        opacity: 0.6,
        transition: "opacity 0.5s",
    },
    pill: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "8px 16px",
        background: "rgba(15, 23, 42, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        borderRadius: 9999,
    },
    pillLabel: {
        fontSize: 11,
        fontWeight: 500,
        textTransform: "uppercase",
        letterSpacing: "0.1em",
    },
    error: {
        color: "#FF3B30",
        fontSize: 14,
        marginTop: 16,
    },
    footer: {
        position: "relative",
        zIndex: 20,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        padding: "32px",
        borderTop: "1px solid rgba(30, 41, 59, 0.5)",
    },
    footerLeft: {
        fontSize: 13,
        color: "#64748b",
        fontWeight: 500,
    },
    footerLinks: {
        display: "flex",
        gap: 24,
    },
    footerLink: {
        fontSize: 13,
        color: "#64748b",
        fontWeight: 500,
        cursor: "pointer",
        transition: "color 0.2s",
    },
};

/* ─── Styles: Create Vault Dialog ──────────────────────────── */

const dlgStyles: Record<string, CSSProperties> = {
    overlay: {
        position: "fixed",
        inset: 0,
        zIndex: 100,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: 16,
    },
    card: {
        background: "#0f172a",
        border: "1px solid #1e293b",
        borderRadius: 12,
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
        width: 480,
        maxWidth: "100%",
        overflow: "hidden",
    },
    header: {
        padding: 24,
        borderBottom: "1px solid #1e293b",
        display: "flex",
        alignItems: "center",
        gap: 16,
    },
    headerIcon: {
        width: 48,
        height: 48,
        borderRadius: 8,
        background: "rgba(88, 86, 214, 0.1)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    title: {
        fontSize: 20,
        fontWeight: 700,
        color: "#fff",
        letterSpacing: "-0.025em",
        lineHeight: 1.2,
    },
    subtitle: {
        fontSize: 14,
        color: "#94a3b8",
    },
    body: {
        padding: 24,
        display: "flex",
        flexDirection: "column",
        gap: 24,
    },
    field: {
        display: "flex",
        flexDirection: "column",
        gap: 8,
    },
    label: {
        fontSize: 14,
        fontWeight: 500,
        color: "#cbd5e1",
    },
    inputIcon: {
        width: "100%",
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: "8px 12px 8px 36px",
        fontSize: 14,
        color: "#fff",
        outline: "none",
    },
    browseBtn: {
        padding: "8px 16px",
        background: "#1e293b",
        border: "1px solid #334155",
        borderRadius: 8,
        color: "#fff",
        fontSize: 14,
        fontWeight: 500,
        cursor: "pointer",
        whiteSpace: "nowrap",
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
    infoBox: {
        display: "flex",
        gap: 12,
        padding: 16,
        background: "rgba(88, 86, 214, 0.05)",
        border: "1px solid rgba(88, 86, 214, 0.2)",
        borderRadius: 8,
    },
    infoText: {
        fontSize: 12,
        lineHeight: 1.6,
        color: "#94a3b8",
    },
    fieldError: {
        color: "#FF3B30",
        fontSize: 11,
    },
    footer: {
        padding: 24,
        background: "rgba(2, 6, 23, 0.5)",
        borderTop: "1px solid #1e293b",
        display: "flex",
        justifyContent: "flex-end",
        alignItems: "center",
        gap: 12,
    },
    cancelBtn: {
        padding: "10px 24px",
        background: "none",
        border: "none",
        color: "#94a3b8",
        fontSize: 14,
        fontWeight: 500,
        cursor: "pointer",
    },
    createBtn: {
        padding: "10px 24px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 14,
        fontWeight: 700,
        cursor: "pointer",
        boxShadow: "0 4px 14px rgba(88, 86, 214, 0.2)",
        transition: "all 0.2s",
    },
};
