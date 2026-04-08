import { type CSSProperties } from "react";

interface PasswordStrengthBarProps {
    score: number; // 0-4
}

const labels = ["Very Weak", "Weak", "Fair", "Good", "Strong"];
const colors = ["#FF3B30", "#FF9500", "#FFD60A", "#34C759", "#34C759"];

export function PasswordStrengthBar({ score }: PasswordStrengthBarProps) {
    const clampedScore = Math.max(0, Math.min(4, score));

    return (
        <div style={styles.container}>
            <div style={styles.barTrack}>
                {[0, 1, 2, 3].map((i) => (
                    <div
                        key={i}
                        style={{
                            ...styles.barSegment,
                            backgroundColor:
                                i <= clampedScore - 1
                                    ? colors[clampedScore]
                                    : "var(--color-border)",
                        }}
                    />
                ))}
            </div>
            {score > 0 && (
                <span
                    style={{ ...styles.label, color: colors[clampedScore] }}
                >
                    {labels[clampedScore]}
                </span>
            )}
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    container: {
        display: "flex",
        alignItems: "center",
        gap: "var(--space-2)",
        marginTop: "var(--space-1)",
    },
    barTrack: {
        display: "flex",
        gap: 3,
        flex: 1,
    },
    barSegment: {
        height: 4,
        flex: 1,
        borderRadius: 2,
        transition: "background-color var(--transition-fast)",
    },
    label: {
        fontSize: "var(--text-xs)",
        fontWeight: 500,
        minWidth: 60,
        textAlign: "right" as const,
    },
};
