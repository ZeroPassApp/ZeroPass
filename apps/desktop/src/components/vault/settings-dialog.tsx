import { type CSSProperties, useState } from "react";
import * as cmd from "../../lib/commands";
import { Icon } from "../common/icon";

type SettingsTab = "general" | "security" | "sync" | "about";

export function SettingsDialog({ onClose }: { onClose: () => void }) {
    const [tab, setTab] = useState<SettingsTab>("general");

    const tabs: { id: SettingsTab; label: string; icon: string }[] = [
        { id: "general", label: "General", icon: "tune" },
        { id: "security", label: "Security", icon: "shield" },
        { id: "sync", label: "Sync", icon: "sync" },
        { id: "about", label: "About", icon: "info" },
    ];

    return (
        <div style={styles.overlay} onClick={onClose}>
            <div
                style={styles.dialog}
                onClick={(e) => e.stopPropagation()}
                className="animate-scale-in"
            >
                {/* Sidebar */}
                <div style={styles.sidebar}>
                    <div style={styles.sidebarBrand}>
                        <div style={styles.brandDot}>
                            <Icon name="lock" fill size={14} color="#fff" />
                        </div>
                        <div>
                            <div style={styles.brandName}>ZeroPass</div>
                            <div style={styles.brandVer}>v0.1.0</div>
                        </div>
                    </div>

                    <div style={styles.tabList}>
                        {tabs.map((t) => (
                            <button
                                key={t.id}
                                style={tab === t.id ? styles.tabActive : styles.tab}
                                onClick={() => setTab(t.id)}
                            >
                                <Icon
                                    name={t.icon}
                                    size={16}
                                    color={tab === t.id ? "#818cf8" : "#64748b"}
                                />
                                <span>{t.label}</span>
                            </button>
                        ))}
                    </div>

                    <div style={styles.sidebarSpacer} />

                    <div style={styles.sidebarUser}>
                        <div style={styles.userAvatar}>
                            <Icon name="person" size={14} color="#94a3b8" />
                        </div>
                        <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={styles.sidebarUserName}>Local User</div>
                            <div style={styles.sidebarUserPlan}>Pro Plan</div>
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div style={styles.content}>
                    <div style={styles.contentHeader}>
                        <h2 style={styles.contentTitle}>
                            {tabs.find((t) => t.id === tab)?.label}
                        </h2>
                    </div>

                    {tab === "general" && <GeneralTab />}
                    {tab === "security" && <SecurityTab />}
                    {tab === "sync" && <SyncTab />}
                    {tab === "about" && <AboutTab />}

                    {/* Footer */}
                    <div style={styles.footer}>
                        <button style={styles.doneBtn} onClick={onClose}>
                            Done
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

function GeneralTab() {
    const [appearance, setAppearance] = useState<"system" | "light" | "dark">("system");

    return (
        <div style={styles.section}>
            {/* Appearance Toggle */}
            <div style={styles.settingBlock}>
                <span style={styles.settingLabel}>Appearance</span>
                <div style={styles.segmentedControl}>
                    {(["system", "light", "dark"] as const).map((mode) => (
                        <button
                            key={mode}
                            style={
                                appearance === mode
                                    ? styles.segActive
                                    : styles.segBtn
                            }
                            onClick={() => setAppearance(mode)}
                        >
                            <Icon
                                name={
                                    mode === "system"
                                        ? "contrast"
                                        : mode === "light"
                                          ? "light_mode"
                                          : "dark_mode"
                                }
                                size={14}
                                color={appearance === mode ? "#fff" : "#64748b"}
                            />
                            <span>{mode.charAt(0).toUpperCase() + mode.slice(1)}</span>
                        </button>
                    ))}
                </div>
            </div>

            <div style={styles.separator} />

            <span style={styles.groupLabel}>Startup</span>
            <SettingRow
                label="Launch at Login"
                description="Start ZeroPass when you log in"
            >
                <ToggleSwitch defaultChecked={false} />
            </SettingRow>
            <SettingRow
                label="Show in Menu Bar"
                description="Add ZeroPass to the system tray"
            >
                <ToggleSwitch defaultChecked={true} />
            </SettingRow>

            <div style={styles.separator} />

            <span style={styles.groupLabel}>Keyboard</span>
            <div style={styles.hotkeyRow}>
                <span style={styles.hotkeyLabel}>Global Hotkey</span>
                <div style={styles.hotkeyDisplay}>
                    <kbd style={styles.hotkeyKbd}>⌘</kbd>
                    <kbd style={styles.hotkeyKbd}>⇧</kbd>
                    <kbd style={styles.hotkeyKbd}>P</kbd>
                </div>
            </div>

            <div style={styles.proTip}>
                <Icon name="lightbulb" size={16} color="#FF9500" />
                <div>
                    <span style={styles.proTipTitle}>Pro Tip</span>
                    <span style={styles.proTipText}>
                        Press <strong>⌘K</strong> anywhere to instantly search your vault
                    </span>
                </div>
            </div>
        </div>
    );
}

function SecurityTab() {
    const [autoLock, setAutoLock] = useState("5");
    const [clipClear, setClipClear] = useState("30");
    const [showChangePw, setShowChangePw] = useState(false);

    return (
        <div style={styles.section}>
            <SettingRow label="Auto-Lock" description="Lock vault after inactivity">
                <select
                    style={styles.select}
                    value={autoLock}
                    onChange={(e) => setAutoLock(e.target.value)}
                >
                    <option value="1">1 minute</option>
                    <option value="5">5 minutes</option>
                    <option value="15">15 minutes</option>
                    <option value="30">30 minutes</option>
                    <option value="0">Never</option>
                </select>
            </SettingRow>
            <SettingRow label="Clipboard Clear" description="Auto-clear clipboard after copy">
                <select
                    style={styles.select}
                    value={clipClear}
                    onChange={(e) => setClipClear(e.target.value)}
                >
                    <option value="10">10 seconds</option>
                    <option value="30">30 seconds</option>
                    <option value="60">1 minute</option>
                    <option value="0">Never</option>
                </select>
            </SettingRow>
            <SettingRow label="Biometric Unlock" description="Use Touch ID or system biometric">
                <ToggleSwitch defaultChecked={false} />
            </SettingRow>
            <div style={styles.separator} />
            <button
                style={styles.actionButton}
                onClick={() => setShowChangePw(true)}
            >
                <Icon name="lock_reset" size={16} color="#5856D6" />
                Change Master Password
            </button>
            {showChangePw && (
                <ChangePasswordForm onClose={() => setShowChangePw(false)} />
            )}
        </div>
    );
}

function SyncTab() {
    return (
        <div style={styles.section}>
            <SettingRow label="Sync Server" description="Self-hosted sync server URL">
                <input style={styles.inputSmall} placeholder="https://sync.example.com" />
            </SettingRow>
            <SettingRow label="Device Name" description="Name for this device">
                <input style={styles.inputSmall} placeholder="My Mac" />
            </SettingRow>
            <div style={styles.infoBox}>
                <Icon name="info" size={16} color="#5856D6" />
                <p style={styles.infoText}>
                    Sync is an optional preview feature. Your data is end-to-end encrypted.
                </p>
            </div>
        </div>
    );
}

function AboutTab() {
    return (
        <div style={styles.section}>
            <div style={styles.aboutContent}>
                <div style={styles.aboutLogoBox}>
                    <Icon name="lock" fill size={28} color="#fff" />
                </div>
                <h3 style={styles.aboutName}>ZeroPass Desktop</h3>
                <p style={styles.aboutVersion}>Version 0.1.0 (Build 1)</p>
                <p style={styles.aboutDesc}>
                    Zero-knowledge secrets manager.
                    <br />
                    Built with Tauri, React, and a Go cryptographic core.
                </p>
                <div style={styles.aboutBadges}>
                    <span style={styles.aboutBadge}>AES-256</span>
                    <span style={styles.aboutBadge}>Argon2id</span>
                    <span style={styles.aboutBadge}>BIP-39</span>
                </div>
            </div>
        </div>
    );
}

function ChangePasswordForm({ onClose }: { onClose: () => void }) {
    const [oldPw, setOldPw] = useState("");
    const [newPw, setNewPw] = useState("");
    const [confirm, setConfirm] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const canSubmit = oldPw && newPw.length >= 8 && newPw === confirm && !submitting;

    const handleSubmit = async () => {
        if (!canSubmit) return;
        setSubmitting(true);
        setError(null);
        try {
            await cmd.changeMasterPassword(oldPw, newPw);
            onClose();
        } catch (e) {
            setError(String(e));
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div style={styles.inlineForm}>
            <input type="password" style={styles.formInput} placeholder="Current password" value={oldPw} onChange={(e) => setOldPw(e.target.value)} />
            <input type="password" style={styles.formInput} placeholder="New password (min 8 chars)" value={newPw} onChange={(e) => setNewPw(e.target.value)} />
            <input type="password" style={styles.formInput} placeholder="Confirm new password" value={confirm} onChange={(e) => setConfirm(e.target.value)} />
            {error && <p style={{ color: "#FF3B30", fontSize: 12 }}>{error}</p>}
            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 4 }}>
                <button style={styles.cancelBtn} onClick={onClose}>Cancel</button>
                <button style={{ ...styles.saveBtn, opacity: canSubmit ? 1 : 0.5 }} onClick={handleSubmit} disabled={!canSubmit}>
                    {submitting ? "Changing…" : "Change"}
                </button>
            </div>
        </div>
    );
}

function SettingRow({ label, description, children }: { label: string; description: string; children: React.ReactNode }) {
    return (
        <div style={styles.settingRow}>
            <div style={styles.settingLeft}>
                <span style={styles.settingLabel}>{label}</span>
                <span style={styles.settingDesc}>{description}</span>
            </div>
            {children}
        </div>
    );
}

function ToggleSwitch({ defaultChecked }: { defaultChecked: boolean }) {
    const [on, setOn] = useState(defaultChecked);
    return (
        <button
            style={{
                ...styles.toggle,
                background: on ? "#5856D6" : "#334155",
            }}
            onClick={() => setOn(!on)}
        >
            <span
                style={{
                    ...styles.toggleThumb,
                    transform: on ? "translateX(18px)" : "translateX(2px)",
                }}
            />
        </button>
    );
}

const styles: Record<string, CSSProperties> = {
    overlay: {
        position: "fixed",
        inset: 0,
        background: "rgba(2, 6, 23, 0.7)",
        backdropFilter: "blur(8px)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1500,
        fontFamily: "'Inter', sans-serif",
    },
    dialog: {
        background: "rgba(15, 23, 42, 0.95)",
        backdropFilter: "blur(12px)",
        border: "1px solid #1e293b",
        borderRadius: 16,
        display: "flex",
        width: 700,
        maxWidth: "90vw",
        height: 500,
        maxHeight: "80vh",
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
        overflow: "hidden",
    },
    sidebar: {
        width: 200,
        background: "#0f172a",
        borderRight: "1px solid #1e293b",
        padding: 16,
        display: "flex",
        flexDirection: "column",
    },
    sidebarBrand: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "4px 0 16px",
    },
    brandDot: {
        width: 24,
        height: 24,
        borderRadius: 6,
        background: "#5856D6",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    brandName: {
        fontSize: 13,
        fontWeight: 700,
        color: "#fff",
    },
    brandVer: {
        fontSize: 10,
        color: "#64748b",
    },
    tabList: {
        display: "flex",
        flexDirection: "column",
        gap: 2,
    },
    tab: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "8px 10px",
        border: "none",
        background: "transparent",
        borderRadius: 6,
        color: "#94a3b8",
        fontSize: 13,
        fontWeight: 500,
        cursor: "pointer",
        textAlign: "left",
        width: "100%",
    },
    tabActive: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "8px 10px",
        border: "none",
        background: "rgba(88, 86, 214, 0.15)",
        color: "#e2e8f0",
        borderRadius: 6,
        fontSize: 13,
        fontWeight: 500,
        cursor: "pointer",
        textAlign: "left",
        width: "100%",
    },
    sidebarSpacer: { flex: 1 },
    sidebarUser: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        paddingTop: 12,
        borderTop: "1px solid #1e293b",
    },
    userAvatar: {
        width: 28,
        height: 28,
        borderRadius: "50%",
        background: "#1e293b",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
    },
    sidebarUserName: { fontSize: 12, fontWeight: 600, color: "#e2e8f0" },
    sidebarUserPlan: { fontSize: 10, color: "#64748b" },
    content: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        overflow: "auto",
    },
    contentHeader: {
        padding: "20px 24px 12px",
    },
    contentTitle: {
        fontSize: 16,
        fontWeight: 700,
        color: "#fff",
        letterSpacing: "-0.01em",
    },
    section: {
        display: "flex",
        flexDirection: "column",
        gap: 12,
        padding: "0 24px",
        flex: 1,
    },
    settingBlock: {
        display: "flex",
        flexDirection: "column",
        gap: 8,
    },
    groupLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#475569",
        textTransform: "uppercase",
        letterSpacing: "0.08em",
    },
    segmentedControl: {
        display: "flex",
        padding: 3,
        background: "#020617",
        borderRadius: 8,
        border: "1px solid #1e293b",
    },
    segBtn: {
        flex: 1,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 6,
        padding: "6px 0",
        background: "none",
        border: "none",
        color: "#64748b",
        fontSize: 12,
        fontWeight: 500,
        cursor: "pointer",
        borderRadius: 5,
    },
    segActive: {
        flex: 1,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 6,
        padding: "6px 0",
        background: "#5856D6",
        border: "none",
        color: "#fff",
        fontSize: 12,
        fontWeight: 600,
        cursor: "pointer",
        borderRadius: 5,
    },
    hotkeyRow: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
    },
    hotkeyLabel: {
        fontSize: 13,
        fontWeight: 500,
        color: "#e2e8f0",
    },
    hotkeyDisplay: {
        display: "flex",
        gap: 4,
    },
    hotkeyKbd: {
        padding: "4px 8px",
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 4,
        fontSize: 12,
        fontWeight: 600,
        color: "#e2e8f0",
        fontFamily: "'Inter', sans-serif",
    },
    proTip: {
        display: "flex",
        gap: 10,
        padding: 12,
        background: "rgba(255, 149, 0, 0.06)",
        border: "1px solid rgba(255, 149, 0, 0.15)",
        borderRadius: 8,
        marginTop: 4,
    },
    proTipTitle: {
        display: "block",
        fontSize: 11,
        fontWeight: 700,
        color: "#FF9500",
        marginBottom: 2,
    },
    proTipText: {
        display: "block",
        fontSize: 11,
        color: "#94a3b8",
        lineHeight: 1.5,
    },
    separator: {
        height: 1,
        background: "#1e293b",
        margin: "4px 0",
    },
    settingRow: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        gap: 16,
        padding: "4px 0",
    },
    settingLeft: {
        display: "flex",
        flexDirection: "column",
    },
    settingLabel: {
        fontSize: 13,
        fontWeight: 500,
        color: "#e2e8f0",
    },
    settingDesc: {
        fontSize: 11,
        color: "#64748b",
    },
    select: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 6,
        padding: "4px 8px",
        color: "#e2e8f0",
        fontSize: 12,
        outline: "none",
    },
    inputSmall: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 6,
        padding: "6px 10px",
        color: "#e2e8f0",
        fontSize: 12,
        outline: "none",
        width: 200,
    },
    toggle: {
        width: 44,
        height: 24,
        borderRadius: 12,
        border: "none",
        cursor: "pointer",
        position: "relative",
        transition: "background 0.15s",
        flexShrink: 0,
    },
    toggleThumb: {
        display: "block",
        width: 20,
        height: 20,
        borderRadius: "50%",
        background: "#fff",
        position: "absolute",
        top: 2,
        transition: "transform 0.15s",
    },
    actionButton: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        background: "transparent",
        color: "#5856D6",
        border: "1px solid rgba(88, 86, 214, 0.3)",
        borderRadius: 8,
        padding: "8px 16px",
        fontSize: 13,
        fontWeight: 500,
        cursor: "pointer",
        alignSelf: "flex-start",
    },
    infoBox: {
        display: "flex",
        gap: 10,
        padding: 12,
        background: "rgba(88, 86, 214, 0.05)",
        border: "1px solid rgba(88, 86, 214, 0.2)",
        borderRadius: 8,
    },
    infoText: {
        fontSize: 12,
        color: "#94a3b8",
        lineHeight: 1.6,
    },
    aboutContent: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 8,
        textAlign: "center",
        paddingTop: 24,
    },
    aboutLogoBox: {
        width: 56,
        height: 56,
        borderRadius: 12,
        background: "#5856D6",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        marginBottom: 8,
    },
    aboutName: { fontSize: 18, fontWeight: 700, color: "#fff" },
    aboutVersion: { fontSize: 12, color: "#64748b" },
    aboutDesc: { fontSize: 13, color: "#94a3b8", lineHeight: 1.6 },
    aboutBadges: { display: "flex", gap: 8, marginTop: 8 },
    aboutBadge: {
        padding: "4px 10px",
        background: "rgba(88, 86, 214, 0.1)",
        border: "1px solid rgba(88, 86, 214, 0.2)",
        borderRadius: 4,
        fontSize: 10,
        fontWeight: 700,
        color: "#818cf8",
    },
    inlineForm: {
        display: "flex",
        flexDirection: "column",
        gap: 8,
        padding: 12,
        background: "rgba(2, 6, 23, 0.5)",
        borderRadius: 8,
        border: "1px solid #1e293b",
    },
    formInput: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 6,
        padding: "8px 12px",
        color: "#e2e8f0",
        fontSize: 13,
        outline: "none",
    },
    cancelBtn: {
        background: "transparent",
        color: "#94a3b8",
        border: "1px solid #334155",
        borderRadius: 6,
        padding: "6px 14px",
        fontSize: 12,
        cursor: "pointer",
    },
    saveBtn: {
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 6,
        padding: "6px 14px",
        fontSize: 12,
        fontWeight: 600,
        cursor: "pointer",
    },
    footer: {
        padding: "12px 24px",
        borderTop: "1px solid #1e293b",
        display: "flex",
        justifyContent: "flex-end",
    },
    doneBtn: {
        padding: "8px 24px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 13,
        fontWeight: 600,
        cursor: "pointer",
    },
};
