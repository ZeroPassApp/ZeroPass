import { type CSSProperties, useState, useEffect } from "react";
import * as cmd from "../../lib/commands";
import { Icon } from "../common/icon";

interface HealthReport {
    score: number;
    total: number;
    weak: string[];
    reused: string[];
    old: string[];
    breached: string[];
}

export function HealthDashboard({ onClose }: { onClose: () => void }) {
    const [report, setReport] = useState<HealthReport | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        cmd.analyzeHealth()
            .then((json: string) => setReport(JSON.parse(json)))
            .catch(() =>
                setReport({
                    score: 0,
                    total: 0,
                    weak: [],
                    reused: [],
                    old: [],
                    breached: [],
                }),
            )
            .finally(() => setLoading(false));
    }, []);

    const score = report?.score ?? 0;
    const label = score >= 80 ? "Good" : score >= 60 ? "Fair" : "Poor";
    const labelColor = score >= 80 ? "#34C759" : score >= 60 ? "#FF9500" : "#FF3B30";
    const circumference = 2 * Math.PI * 70; // r=70
    const offset = circumference - (score / 100) * circumference;

    return (
        <div style={styles.overlay} onClick={onClose}>
            <div
                style={styles.container}
                onClick={(e) => e.stopPropagation()}
                className="animate-scale-in"
            >
                {/* Header */}
                <div style={styles.header}>
                    <div style={styles.headerLeft}>
                        <Icon
                            name="health_and_safety"
                            fill
                            size={24}
                            color="#818cf8"
                        />
                        <h1 style={styles.headerTitle}>Security Health</h1>
                    </div>
                    <div style={styles.headerRight}>
                        <button
                            style={styles.scanBtn}
                            onClick={() => {
                                setLoading(true);
                                cmd.analyzeHealth()
                                    .then((json: string) =>
                                        setReport(JSON.parse(json)),
                                    )
                                    .catch(() => {})
                                    .finally(() => setLoading(false));
                            }}
                        >
                            Scan Now
                        </button>
                        <button style={styles.closeBtn} onClick={onClose}>
                            <Icon name="close" size={18} color="#94a3b8" />
                        </button>
                    </div>
                </div>

                {loading ? (
                    <div style={styles.loadingState}>
                        <Icon name="sync" size={32} color="#64748b" />
                        <p style={{ color: "#64748b", fontSize: 13 }}>
                            Analyzing vault security…
                        </p>
                    </div>
                ) : (
                    <div style={styles.scrollArea}>
                        {/* Score + Summary Bento */}
                        <div style={styles.bentoGrid}>
                            {/* Score ring */}
                            <div style={styles.scoreCard}>
                                <div style={{ position: "relative", width: 160, height: 160 }}>
                                    <svg
                                        width="160"
                                        height="160"
                                        style={{ transform: "rotate(-90deg)" }}
                                    >
                                        <circle
                                            cx="80"
                                            cy="80"
                                            r="70"
                                            fill="transparent"
                                            stroke="#1e293b"
                                            strokeWidth="12"
                                        />
                                        <circle
                                            cx="80"
                                            cy="80"
                                            r="70"
                                            fill="transparent"
                                            stroke={labelColor}
                                            strokeWidth="12"
                                            strokeLinecap="round"
                                            strokeDasharray={circumference}
                                            strokeDashoffset={offset}
                                        />
                                    </svg>
                                    <div style={styles.scoreCenter}>
                                        <span style={styles.scoreNum}>
                                            {score}
                                            <span style={styles.scoreMax}>
                                                /100
                                            </span>
                                        </span>
                                        <span
                                            style={{
                                                ...styles.scoreLabel,
                                                color: labelColor,
                                            }}
                                        >
                                            {label}
                                        </span>
                                    </div>
                                </div>
                                <p style={styles.scoreDesc}>
                                    {score >= 80
                                        ? "Your vault security is looking strong. A few areas need attention."
                                        : "Your vault needs attention. Fix critical issues to improve your score."}
                                </p>
                            </div>

                            {/* Summary cards */}
                            <div style={styles.summaryGrid}>
                                <SummaryCard
                                    label="Weak Passwords"
                                    count={report?.weak.length ?? 0}
                                    sub="Requires urgent rotation"
                                    borderColor="#FF9500"
                                    icon="warning"
                                    iconColor="#FF9500"
                                />
                                <SummaryCard
                                    label="Reused Passwords"
                                    count={report?.reused.length ?? 0}
                                    sub="Compromised security risk"
                                    borderColor="#FF3B30"
                                    icon="content_copy"
                                    iconColor="#FF3B30"
                                />
                                <SummaryCard
                                    label="Old Passwords"
                                    count={report?.old.length ?? 0}
                                    sub="Expired over 90 days"
                                    borderColor="#eab308"
                                    icon="schedule"
                                    iconColor="#eab308"
                                />
                                <SummaryCard
                                    label="Breached"
                                    count={report?.breached.length ?? 0}
                                    sub="Found in public leaks"
                                    borderColor="#FF3B30"
                                    icon="skull"
                                    iconColor="#FF3B30"
                                    pulse={
                                        (report?.breached.length ?? 0) > 0
                                    }
                                />
                            </div>
                        </div>

                        {/* Findings */}
                        <div style={styles.findings}>
                            {/* Critical Issues */}
                            {(report?.breached.length ?? 0) > 0 && (
                                <FindingSection
                                    title="Critical Issues"
                                    color="#FF3B30"
                                    icon="error"
                                >
                                    {report!.breached.map((name) => (
                                        <FindingRow
                                            key={name}
                                            icon="security_update_warning"
                                            iconBg="rgba(255, 59, 48, 0.1)"
                                            iconColor="#FF3B30"
                                            title={`Breached: ${name}`}
                                            subtitle="Password appeared in known data breach"
                                            actionLabel="Fix Now"
                                            actionStyle={styles.fixBtn}
                                        />
                                    ))}
                                </FindingSection>
                            )}

                            {/* Warnings */}
                            {((report?.weak.length ?? 0) > 0 ||
                                (report?.reused.length ?? 0) > 0) && (
                                <FindingSection
                                    title="Warnings"
                                    color="#FF9500"
                                    icon="warning"
                                >
                                    {(report?.weak.length ?? 0) > 0 && (
                                        <div style={styles.findingCard}>
                                            <div style={styles.findingInfo}>
                                                <h5 style={styles.findingTitle}>
                                                    {report!.weak.length} weak passwords
                                                </h5>
                                                <p style={styles.findingSub}>
                                                    Short or easily guessable passwords found
                                                </p>
                                            </div>
                                            <div style={styles.findingItems}>
                                                {report!.weak.map((name) => (
                                                    <div key={name} style={styles.weakItem}>
                                                        <div style={styles.weakDot} />
                                                        <span style={styles.weakName}>{name}</span>
                                                        <button style={styles.genLink}>
                                                            Generate New
                                                        </button>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                    {(report?.reused.length ?? 0) > 0 && (
                                        <FindingRow
                                            icon="join_inner"
                                            iconBg="rgba(88, 86, 214, 0.1)"
                                            iconColor="#818cf8"
                                            title={`${report!.reused.length} reused passwords`}
                                            subtitle={`Same password used across: ${report!.reused.join(", ")}`}
                                            actionLabel="View All"
                                            actionStyle={styles.viewAllBtn}
                                        />
                                    )}
                                </FindingSection>
                            )}

                            {/* Suggestions */}
                            {(report?.old.length ?? 0) > 0 && (
                                <FindingSection
                                    title="Suggestions"
                                    color="#818cf8"
                                    icon="lightbulb"
                                >
                                    <FindingRow
                                        icon="update"
                                        iconBg="rgba(88, 86, 214, 0.1)"
                                        iconColor="#818cf8"
                                        title={`${report!.old.length} passwords older than 90 days`}
                                        subtitle="Consider rotating passwords for frequently used accounts"
                                        actionLabel="Review"
                                        actionStyle={styles.viewAllBtn}
                                    />
                                </FindingSection>
                            )}
                        </div>
                    </div>
                )}

                {/* Footer */}
                <div style={styles.footer}>
                    <div style={styles.footerLeft}>
                        <span style={styles.footerLabel}>STATUS</span>
                        <span style={styles.footerTime}>Last audit: just now</span>
                    </div>
                    <div style={styles.footerRight}>
                        <button style={styles.exportBtn}>Export Report</button>
                        <button
                            style={styles.auditBtn}
                            onClick={() => {
                                setLoading(true);
                                cmd.analyzeHealth()
                                    .then((json: string) =>
                                        setReport(JSON.parse(json)),
                                    )
                                    .catch(() => {})
                                    .finally(() => setLoading(false));
                            }}
                        >
                            <Icon name="bolt" size={14} color="#fff" />
                            Run Full Audit
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

function SummaryCard({
    label,
    count,
    sub,
    borderColor,
    icon,
    iconColor,
    pulse,
}: {
    label: string;
    count: number;
    sub: string;
    borderColor: string;
    icon: string;
    iconColor: string;
    pulse?: boolean;
}) {
    return (
        <div style={{ ...styles.summaryCard, borderLeft: `4px solid ${borderColor}` }}>
            <div>
                <p style={styles.cardLabel}>{label}</p>
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <h3 style={styles.cardCount}>{count}</h3>
                    {pulse && count > 0 && (
                        <span
                            style={{
                                width: 8,
                                height: 8,
                                borderRadius: "50%",
                                background: borderColor,
                                boxShadow: `0 0 8px ${borderColor}`,
                            }}
                        />
                    )}
                </div>
                <p style={styles.cardSub}>{sub}</p>
            </div>
            <Icon name={icon} size={20} color={iconColor} />
        </div>
    );
}

function FindingSection({
    title,
    color,
    icon,
    children,
}: {
    title: string;
    color: string;
    icon: string;
    children: React.ReactNode;
}) {
    return (
        <section style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            <div
                style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 8,
                    paddingBottom: 8,
                    borderBottom: `1px solid ${color}33`,
                }}
            >
                <Icon name={icon} size={14} color={color} />
                <h4
                    style={{
                        fontSize: 10,
                        fontWeight: 900,
                        color,
                        textTransform: "uppercase",
                        letterSpacing: "0.2em",
                    }}
                >
                    {title}
                </h4>
            </div>
            {children}
        </section>
    );
}

function FindingRow({
    icon,
    iconBg,
    iconColor,
    title,
    subtitle,
    actionLabel,
    actionStyle,
}: {
    icon: string;
    iconBg: string;
    iconColor: string;
    title: string;
    subtitle: string;
    actionLabel: string;
    actionStyle: CSSProperties;
}) {
    return (
        <div style={styles.findingCard}>
            <div style={{ display: "flex", alignItems: "center", gap: 16, flex: 1 }}>
                <div
                    style={{
                        background: iconBg,
                        padding: 10,
                        borderRadius: 8,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                    }}
                >
                    <Icon name={icon} size={20} color={iconColor} />
                </div>
                <div>
                    <h5 style={styles.findingTitle}>{title}</h5>
                    <p style={styles.findingSub}>{subtitle}</p>
                </div>
            </div>
            <button style={actionStyle}>{actionLabel}</button>
        </div>
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
    container: {
        background: "#0f172a",
        border: "1px solid #1e293b",
        borderRadius: 16,
        width: 860,
        maxWidth: "95vw",
        height: 620,
        maxHeight: "90vh",
        display: "flex",
        flexDirection: "column",
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
        overflow: "hidden",
    },
    header: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "12px 24px",
        borderBottom: "1px solid #1e293b",
        background: "rgba(15, 23, 42, 0.8)",
        backdropFilter: "blur(12px)",
    },
    headerLeft: {
        display: "flex",
        alignItems: "center",
        gap: 10,
    },
    headerTitle: {
        fontSize: 16,
        fontWeight: 900,
        color: "#fff",
        letterSpacing: "-0.01em",
    },
    headerRight: {
        display: "flex",
        alignItems: "center",
        gap: 8,
    },
    scanBtn: {
        fontSize: 11,
        fontWeight: 600,
        padding: "6px 16px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 20,
        cursor: "pointer",
    },
    closeBtn: {
        background: "none",
        border: "none",
        cursor: "pointer",
        padding: 4,
        display: "flex",
    },
    loadingState: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 12,
    },
    scrollArea: {
        flex: 1,
        overflow: "auto",
        padding: 24,
        display: "flex",
        flexDirection: "column",
        gap: 24,
    },
    bentoGrid: {
        display: "grid",
        gridTemplateColumns: "1fr 2fr",
        gap: 16,
    },
    scoreCard: {
        background: "rgba(24, 24, 27, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.05)",
        borderRadius: 16,
        padding: 24,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        textAlign: "center",
    },
    scoreCenter: {
        position: "absolute",
        inset: 0,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
    },
    scoreNum: {
        fontSize: 36,
        fontWeight: 900,
        color: "#fff",
    },
    scoreMax: {
        fontSize: 18,
        color: "#64748b",
    },
    scoreLabel: {
        fontSize: 12,
        fontWeight: 700,
        textTransform: "uppercase" as const,
        letterSpacing: "0.15em",
        marginTop: 4,
    },
    scoreDesc: {
        marginTop: 16,
        fontSize: 12,
        color: "#64748b",
        lineHeight: 1.7,
    },
    summaryGrid: {
        display: "grid",
        gridTemplateColumns: "1fr 1fr",
        gap: 12,
    },
    summaryCard: {
        background: "rgba(24, 24, 27, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.05)",
        borderRadius: 12,
        padding: 16,
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "space-between",
    },
    cardLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.08em",
        marginBottom: 4,
    },
    cardCount: {
        fontSize: 28,
        fontWeight: 700,
        color: "#fff",
    },
    cardSub: {
        fontSize: 10,
        color: "#64748b",
        marginTop: 6,
    },
    findings: {
        display: "flex",
        flexDirection: "column",
        gap: 20,
    },
    findingCard: {
        background: "rgba(24, 24, 27, 0.6)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.05)",
        borderRadius: 12,
        padding: 16,
        display: "flex",
        flexDirection: "column",
        gap: 12,
    },
    findingInfo: {
        display: "flex",
        flexDirection: "column",
        gap: 4,
    },
    findingTitle: {
        fontSize: 13,
        fontWeight: 700,
        color: "#fff",
    },
    findingSub: {
        fontSize: 11,
        color: "#64748b",
        marginTop: 2,
    },
    findingItems: {
        display: "flex",
        flexDirection: "column",
        gap: 8,
    },
    weakItem: {
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "8px 12px",
        background: "rgba(9, 9, 11, 0.5)",
        borderRadius: 8,
        border: "1px solid rgba(24, 24, 27, 0.5)",
    },
    weakDot: {
        width: 6,
        height: 6,
        borderRadius: "50%",
        background: "#ef4444",
    },
    weakName: {
        flex: 1,
        fontSize: 11,
        fontWeight: 500,
        color: "#e2e8f0",
    },
    genLink: {
        fontSize: 10,
        fontWeight: 700,
        color: "#818cf8",
        background: "none",
        border: "none",
        cursor: "pointer",
    },
    fixBtn: {
        padding: "6px 16px",
        background: "#FF3B30",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 11,
        fontWeight: 700,
        cursor: "pointer",
        flexShrink: 0,
        alignSelf: "flex-end",
    },
    viewAllBtn: {
        padding: "6px 16px",
        background: "#1e293b",
        color: "#fff",
        border: "1px solid #334155",
        borderRadius: 8,
        fontSize: 11,
        fontWeight: 700,
        cursor: "pointer",
        flexShrink: 0,
        alignSelf: "flex-end",
    },
    footer: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "12px 24px",
        borderTop: "1px solid #1e293b",
        background: "#0f172a",
    },
    footerLeft: {
        display: "flex",
        alignItems: "center",
        gap: 10,
    },
    footerLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.12em",
    },
    footerTime: {
        fontSize: 11,
        color: "#64748b",
        fontStyle: "italic",
    },
    footerRight: {
        display: "flex",
        alignItems: "center",
        gap: 8,
    },
    exportBtn: {
        padding: "6px 20px",
        background: "transparent",
        color: "#94a3b8",
        border: "1px solid #334155",
        borderRadius: 8,
        fontSize: 11,
        fontWeight: 700,
        cursor: "pointer",
    },
    auditBtn: {
        padding: "6px 20px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 11,
        fontWeight: 700,
        cursor: "pointer",
        display: "flex",
        alignItems: "center",
        gap: 6,
        boxShadow: "0 4px 12px rgba(88, 86, 214, 0.3)",
    },
};
