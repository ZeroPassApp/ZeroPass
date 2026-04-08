import type { CSSProperties } from "react";

/** Thin wrapper around Google Material Symbols Outlined. */
export function Icon({
    name,
    fill,
    size,
    color,
    style,
}: {
    name: string;
    fill?: boolean;
    size?: number;
    color?: string;
    style?: CSSProperties;
}) {
    return (
        <span
            className="material-symbols-outlined"
            style={{
                fontSize: size ?? 24,
                fontVariationSettings: fill
                    ? "'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24"
                    : "'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24",
                color,
                lineHeight: 1,
                ...style,
            }}
        >
            {name}
        </span>
    );
}
