import { type CSSProperties } from "react";
import { useUIStore } from "../../stores/ui-store";

export function ToastOverlay() {
    const toasts = useUIStore((s) => s.toasts);
    const removeToast = useUIStore((s) => s.removeToast);

    if (toasts.length === 0) return null;

    return (
        <div style={styles.container}>
            {toasts.map((toast) => (
                <div
                    key={toast.id}
                    style={{ ...styles.toast, ...typeStyles[toast.type] }}
                    onClick={() => removeToast(toast.id)}
                    className="animate-slide-up"
                >
                    <span style={styles.icon}>{icons[toast.type]}</span>
                    <span style={styles.message}>{toast.message}</span>
                </div>
            ))}
        </div>
    );
}

const icons: Record<string, string> = {
    success: "✓",
    error: "✕",
    warning: "⚠",
    info: "ℹ",
};

const typeStyles: Record<string, CSSProperties> = {
    success: { borderLeft: "3px solid var(--color-success)" },
    error: { borderLeft: "3px solid var(--color-destructive)" },
    warning: { borderLeft: "3px solid var(--color-warning)" },
    info: { borderLeft: "3px solid var(--color-info)" },
};

const styles: Record<string, CSSProperties> = {
    container: {
        position: "fixed",
        bottom: "var(--space-6)",
        right: "var(--space-6)",
        display: "flex",
        flexDirection: "column",
        gap: "var(--space-2)",
        zIndex: 9999,
        pointerEvents: "none",
    },
    toast: {
        background: "var(--color-bg-elevated)",
        borderRadius: "var(--radius-md)",
        padding: "var(--space-3) var(--space-4)",
        display: "flex",
        alignItems: "center",
        gap: "var(--space-2)",
        boxShadow: "var(--shadow-lg)",
        pointerEvents: "auto",
        cursor: "pointer",
        maxWidth: 360,
    },
    icon: {
        fontSize: "var(--text-sm)",
        fontWeight: 700,
    },
    message: {
        fontSize: "var(--text-sm)",
        color: "var(--color-text)",
    },
};
