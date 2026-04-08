import { type CSSProperties, useState, useEffect } from "react";
import type { VaultItem, ItemType } from "../../lib/types";
import { useVaultStore } from "../../stores/vault-store";
import * as cmd from "../../lib/commands";
import { Icon } from "../common/icon";

const itemTypes: { type: ItemType; label: string; icon: string }[] = [
    { type: "login", label: "Login", icon: "key" },
    { type: "apikey", label: "API Key", icon: "api" },
    { type: "sshkey", label: "SSH Key", icon: "terminal" },
    { type: "note", label: "Secure Note", icon: "sticky_note_2" },
    { type: "creditcard", label: "Credit Card", icon: "credit_card" },
    { type: "identity", label: "Identity", icon: "person" },
    { type: "passkey", label: "Passkey", icon: "passkey" },
    { type: "custom", label: "Custom", icon: "inventory_2" },
];

const defaultFieldsByType: Record<ItemType, string[]> = {
    login: ["username", "password", "url", "notes"],
    apikey: ["api_key", "secret", "url", "notes"],
    sshkey: ["private_key", "public_key", "passphrase", "notes"],
    note: ["content"],
    creditcard: ["cardholder", "number", "expiry", "cvv", "pin", "notes"],
    identity: ["first_name", "last_name", "email", "phone", "address", "notes"],
    passkey: ["relying_party", "username", "credential_id", "notes"],
    custom: ["notes"],
};

const fieldIcons: Record<string, string> = {
    username: "person",
    password: "lock",
    url: "link",
    notes: "sticky_note_2",
    api_key: "key",
    secret: "lock",
    private_key: "vpn_key",
    public_key: "vpn_key",
    passphrase: "lock",
    content: "description",
    cardholder: "badge",
    number: "credit_card",
    expiry: "schedule",
    cvv: "pin",
    pin: "pin",
    first_name: "person",
    last_name: "person",
    email: "email",
    phone: "phone",
    address: "home",
    relying_party: "dns",
    credential_id: "fingerprint",
};

const sensitiveFields = new Set([
    "password", "secret", "pin", "cvv", "passphrase", "private_key",
]);

interface ItemEditorProps {
    item: VaultItem | null;
    onClose: () => void;
}

export function ItemEditor({ item, onClose }: ItemEditorProps) {
    const createItem = useVaultStore((s) => s.createItem);
    const updateItem = useVaultStore((s) => s.updateItem);

    const [itemType, setItemType] = useState<ItemType>(item?.type ?? "login");
    const [name, setName] = useState(item?.name ?? "");
    const [fields, setFields] = useState<Record<string, string>>(item?.fields ?? {});
    const [tagList, setTagList] = useState<string[]>(item?.tags ?? []);
    const [tagInput, setTagInput] = useState("");
    const [favorite, setFavorite] = useState(item?.favorite ?? false);
    const [submitting, setSubmitting] = useState(false);
    const [showGen, setShowGen] = useState(false);
    const [genLength, setGenLength] = useState(24);
    const [genOpts, setGenOpts] = useState({ upper: true, lower: true, numbers: true, symbols: true });
    const [genPreview, setGenPreview] = useState("");
    const [showTypeDropdown, setShowTypeDropdown] = useState(false);
    const [visibleFields, setVisibleFields] = useState<Set<string>>(new Set());

    useEffect(() => {
        if (!item) {
            const defaults: Record<string, string> = {};
            for (const f of defaultFieldsByType[itemType]) {
                defaults[f] = fields[f] ?? "";
            }
            setFields(defaults);
        }
    }, [itemType]); // eslint-disable-line react-hooks/exhaustive-deps

    const fieldKeys = item ? Object.keys(fields) : defaultFieldsByType[itemType];

    const setField = (key: string, value: string) => {
        setFields((prev) => ({ ...prev, [key]: value }));
    };

    const toggleFieldVisibility = (key: string) => {
        setVisibleFields((prev) => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key);
            else next.add(key);
            return next;
        });
    };

    const handleGenPassword = async () => {
        try {
            const pw = await cmd.generatePassword(genLength, {
                uppercase: genOpts.upper,
                lowercase: genOpts.lower,
                numbers: genOpts.numbers,
                symbols: genOpts.symbols,
            });
            setGenPreview(pw);
        } catch { /* ignore */ }
    };

    const applyGenerated = () => {
        if (genPreview) {
            setField("password", genPreview);
            setShowGen(false);
        }
    };

    const handleAddTag = () => {
        const t = tagInput.trim();
        if (t && !tagList.includes(t)) {
            setTagList([...tagList, t]);
        }
        setTagInput("");
    };

    const removeTag = (t: string) => {
        setTagList(tagList.filter((x) => x !== t));
    };

    const handleSave = async () => {
        if (!name.trim()) return;
        setSubmitting(true);
        const payload: Partial<VaultItem> = {
            type: itemType,
            name: name.trim(),
            fields,
            tags: tagList,
            favorite,
        };
        if (item) {
            await updateItem(item.id, payload);
        } else {
            await createItem(payload);
        }
        setSubmitting(false);
        onClose();
    };

    const isEdit = !!item;
    const currentTypeInfo = itemTypes.find((t) => t.type === itemType) ?? itemTypes[0];

    // Password strength
    const pw = fields.password ?? "";
    const pwStrength = pw.length >= 16 ? 4 : pw.length >= 12 ? 3 : pw.length >= 8 ? 2 : pw.length >= 4 ? 1 : 0;
    const strengthLabels = ["", "Weak", "Fair", "Good", "Strong"];
    const strengthColors = ["#334155", "#FF3B30", "#FF9500", "#eab308", "#34C759"];

    return (
        <div style={styles.overlay} onClick={onClose}>
            <div style={styles.modal} onClick={(e) => e.stopPropagation()} className="animate-scale-in">
                {/* Header */}
                <div style={styles.header}>
                    <div style={styles.headerLeft}>
                        <h2 style={styles.headerTitle}>
                            {isEdit ? `Edit ${currentTypeInfo.label}` : `New ${currentTypeInfo.label}`}
                        </h2>

                        {/* Type selector pill */}
                        {!isEdit && (
                            <div style={{ position: "relative" }}>
                                <button
                                    style={styles.typePill}
                                    onClick={() => setShowTypeDropdown(!showTypeDropdown)}
                                >
                                    <Icon name={currentTypeInfo.icon} fill size={14} color="#818cf8" />
                                    <span>{currentTypeInfo.label}</span>
                                    <Icon name="expand_more" size={14} color="#64748b" />
                                </button>
                                {showTypeDropdown && (
                                    <div style={styles.typeDropdown}>
                                        {itemTypes.map((t) => (
                                            <button
                                                key={t.type}
                                                style={{
                                                    ...styles.typeOption,
                                                    background: t.type === itemType ? "rgba(88,86,214,0.15)" : "transparent",
                                                }}
                                                onClick={() => {
                                                    setItemType(t.type);
                                                    setShowTypeDropdown(false);
                                                }}
                                            >
                                                <Icon name={t.icon} size={16} color={t.type === itemType ? "#818cf8" : "#64748b"} />
                                                <span>{t.label}</span>
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                    <button style={styles.closeBtn} onClick={onClose}>
                        <Icon name="close" size={20} color="#94a3b8" />
                    </button>
                </div>

                {/* Scrollable Content */}
                <div style={styles.scrollArea}>
                    {/* Name field */}
                    <div style={styles.fieldGroup}>
                        <label style={styles.label}>NAME</label>
                        <input
                            style={styles.input}
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder={`e.g. ${currentTypeInfo.label === "Login" ? "GitHub" : currentTypeInfo.label}`}
                            autoFocus
                        />
                    </div>

                    {/* Fields */}
                    {fieldKeys.map((key) => {
                        const isSensitive = sensitiveFields.has(key);
                        const isVisible = visibleFields.has(key);
                        const isTextarea = key === "content" || key === "notes" || key === "private_key";
                        const icon = fieldIcons[key] ?? "text_fields";

                        return (
                            <div key={key} style={styles.fieldGroup}>
                                <div style={styles.labelRow}>
                                    <label style={styles.label}>
                                        {key.replace(/_/g, " ").toUpperCase()}
                                    </label>
                                    {key === "password" && (
                                        <button style={styles.genLink} onClick={() => { setShowGen(!showGen); handleGenPassword(); }}>
                                            Generate
                                        </button>
                                    )}
                                </div>

                                {isTextarea ? (
                                    <textarea
                                        style={styles.textarea}
                                        value={fields[key] ?? ""}
                                        onChange={(e) => setField(key, e.target.value)}
                                        rows={3}
                                        placeholder="Additional details..."
                                    />
                                ) : (
                                    <div style={styles.inputWithIcon}>
                                        <Icon name={icon} size={18} color="#64748b" style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)" }} />
                                        <input
                                            style={styles.iconInput}
                                            type={isSensitive && !isVisible ? "password" : "text"}
                                            value={fields[key] ?? ""}
                                            onChange={(e) => setField(key, e.target.value)}
                                            placeholder={key === "url" ? "https://example.com" : `Enter ${key.replace(/_/g, " ")}`}
                                        />
                                        {isSensitive && (
                                            <button
                                                style={styles.visBtn}
                                                onClick={() => toggleFieldVisibility(key)}
                                            >
                                                <Icon name={isVisible ? "visibility_off" : "visibility"} size={18} color="#64748b" />
                                            </button>
                                        )}
                                    </div>
                                )}

                                {/* Password strength bar */}
                                {key === "password" && pw && (
                                    <div style={styles.strengthRow}>
                                        {[1, 2, 3, 4].map((i) => (
                                            <div
                                                key={i}
                                                style={{
                                                    ...styles.strengthSeg,
                                                    background: i <= pwStrength ? strengthColors[pwStrength] : "#1e293b",
                                                }}
                                            />
                                        ))}
                                        <span style={{ fontSize: 10, fontWeight: 700, color: strengthColors[pwStrength], textTransform: "uppercase", marginLeft: 8 }}>
                                            {strengthLabels[pwStrength]}
                                        </span>
                                    </div>
                                )}
                            </div>
                        );
                    })}

                    {/* Password Generator Panel */}
                    {showGen && fieldKeys.includes("password") && (
                        <div style={styles.genPanel}>
                            <div style={styles.genHeader}>
                                <span style={styles.genPreviewLabel}>PREVIEW</span>
                                <p style={styles.genPreviewText}>{genPreview || "…"}</p>
                            </div>
                            <div style={styles.genGrid}>
                                <div>
                                    <div style={styles.genLenRow}>
                                        <span style={styles.genLenLabel}>Length</span>
                                        <span style={styles.genLenVal}>{genLength}</span>
                                    </div>
                                    <input
                                        type="range"
                                        min={8}
                                        max={64}
                                        value={genLength}
                                        onChange={(e) => { setGenLength(+e.target.value); }}
                                        onMouseUp={handleGenPassword}
                                        style={styles.rangeInput}
                                    />
                                </div>
                                <div style={styles.checkGrid}>
                                    {([
                                        ["upper", "A-Z"],
                                        ["lower", "a-z"],
                                        ["numbers", "0-9"],
                                        ["symbols", "!@#"],
                                    ] as const).map(([key, label]) => (
                                        <label key={key} style={styles.checkLabel}>
                                            <div
                                                style={{
                                                    ...styles.checkbox,
                                                    background: genOpts[key] ? "#5856D6" : "transparent",
                                                    borderColor: genOpts[key] ? "#5856D6" : "#334155",
                                                }}
                                                onClick={() => {
                                                    setGenOpts((p) => ({ ...p, [key]: !p[key] }));
                                                    setTimeout(handleGenPassword, 50);
                                                }}
                                            >
                                                {genOpts[key] && <Icon name="check" size={12} color="#fff" />}
                                            </div>
                                            <span>{label}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                            <button style={styles.usePassBtn} onClick={applyGenerated}>
                                Use Password
                            </button>
                        </div>
                    )}

                    {/* Notes if not already in fields */}
                    {/* Tags */}
                    <div style={styles.fieldGroup}>
                        <label style={styles.label}>TAGS</label>
                        <div style={styles.tagsContainer}>
                            {tagList.map((t) => (
                                <span key={t} style={styles.tag}>
                                    {t}
                                    <button style={styles.tagX} onClick={() => removeTag(t)}>
                                        <Icon name="close" size={14} color="currentColor" />
                                    </button>
                                </span>
                            ))}
                            <input
                                style={styles.tagInput}
                                value={tagInput}
                                onChange={(e) => setTagInput(e.target.value)}
                                onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); handleAddTag(); } }}
                                placeholder="Add tag..."
                            />
                        </div>
                    </div>

                    {/* Custom field + Favorite */}
                    <div style={styles.bottomRow}>
                        <button style={styles.addFieldBtn}>
                            <Icon name="add_circle" size={18} color="#818cf8" />
                            Add Custom Field
                        </button>
                        <div style={styles.favToggle}>
                            <span style={styles.favLabel}>Favorite</span>
                            <button
                                style={{
                                    ...styles.toggle,
                                    background: favorite ? "#5856D6" : "#334155",
                                }}
                                onClick={() => setFavorite(!favorite)}
                            >
                                <span style={{
                                    ...styles.toggleThumb,
                                    transform: favorite ? "translateX(18px)" : "translateX(2px)",
                                }} />
                            </button>
                        </div>
                    </div>
                </div>

                {/* Footer */}
                <div style={styles.footer}>
                    <button style={styles.cancelBtn} onClick={onClose}>
                        Cancel
                    </button>
                    <button
                        style={{ ...styles.saveBtn, opacity: name.trim() && !submitting ? 1 : 0.5 }}
                        onClick={handleSave}
                        disabled={!name.trim() || submitting}
                    >
                        {submitting ? "Saving…" : isEdit ? "Save" : "Save"}
                    </button>
                </div>
            </div>
        </div>
    );
}

const styles: Record<string, CSSProperties> = {
    overlay: {
        position: "fixed",
        inset: 0,
        background: "rgba(2, 6, 23, 0.6)",
        backdropFilter: "blur(4px)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
        fontFamily: "'Inter', sans-serif",
    },
    modal: {
        background: "#0f172a",
        border: "1px solid #1e293b",
        borderRadius: 12,
        width: 640,
        maxWidth: "95vw",
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
        padding: "16px 24px",
        borderBottom: "1px solid #1e293b",
    },
    headerLeft: {
        display: "flex",
        alignItems: "center",
        gap: 12,
    },
    headerTitle: {
        fontSize: 16,
        fontWeight: 700,
        color: "#fff",
    },
    typePill: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        padding: "4px 12px",
        background: "#1e293b",
        border: "none",
        borderRadius: 20,
        color: "#cbd5e1",
        fontSize: 12,
        fontWeight: 600,
        cursor: "pointer",
    },
    typeDropdown: {
        position: "absolute",
        top: "100%",
        left: 0,
        marginTop: 4,
        background: "#1e293b",
        border: "1px solid #334155",
        borderRadius: 8,
        padding: 4,
        minWidth: 160,
        zIndex: 10,
        display: "flex",
        flexDirection: "column",
        gap: 2,
    },
    typeOption: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "6px 10px",
        border: "none",
        borderRadius: 6,
        color: "#e2e8f0",
        fontSize: 12,
        fontWeight: 500,
        cursor: "pointer",
        textAlign: "left",
        width: "100%",
    },
    closeBtn: {
        width: 32,
        height: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        borderRadius: "50%",
        background: "transparent",
        border: "none",
        cursor: "pointer",
    },
    scrollArea: {
        flex: 1,
        overflow: "auto",
        padding: 24,
        display: "flex",
        flexDirection: "column",
        gap: 20,
    },
    fieldGroup: {
        display: "flex",
        flexDirection: "column",
        gap: 6,
    },
    labelRow: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        paddingLeft: 4,
        paddingRight: 4,
    },
    label: {
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.12em",
    },
    genLink: {
        fontSize: 11,
        fontWeight: 700,
        color: "#818cf8",
        background: "none",
        border: "none",
        cursor: "pointer",
    },
    input: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: "10px 16px",
        color: "#fff",
        fontSize: 13,
        outline: "none",
        width: "100%",
    },
    inputWithIcon: {
        position: "relative",
    },
    iconInput: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: "10px 16px 10px 40px",
        color: "#fff",
        fontSize: 13,
        outline: "none",
        width: "100%",
        fontFamily: "'Inter', sans-serif",
    },
    visBtn: {
        position: "absolute",
        right: 12,
        top: "50%",
        transform: "translateY(-50%)",
        background: "none",
        border: "none",
        cursor: "pointer",
        display: "flex",
    },
    textarea: {
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: "10px 16px",
        color: "#fff",
        fontSize: 13,
        outline: "none",
        resize: "none",
        width: "100%",
    },
    strengthRow: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        paddingLeft: 4,
    },
    strengthSeg: {
        height: 4,
        borderRadius: 2,
        flex: 1,
    },
    genPanel: {
        background: "rgba(2, 6, 23, 0.5)",
        border: "1px solid #1e293b",
        borderRadius: 12,
        padding: 20,
        display: "flex",
        flexDirection: "column",
        gap: 16,
    },
    genHeader: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
    },
    genPreviewLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
    },
    genPreviewText: {
        fontSize: 13,
        fontFamily: "'JetBrains Mono', monospace",
        color: "#818cf8",
        letterSpacing: "0.08em",
    },
    genGrid: {
        display: "grid",
        gridTemplateColumns: "1fr 1fr",
        gap: 24,
    },
    genLenRow: {
        display: "flex",
        justifyContent: "space-between",
        marginBottom: 8,
    },
    genLenLabel: {
        fontSize: 11,
        fontWeight: 700,
        color: "#64748b",
    },
    genLenVal: {
        fontSize: 11,
        fontWeight: 700,
        color: "#818cf8",
    },
    rangeInput: {
        width: "100%",
        accentColor: "#5856D6",
    },
    checkGrid: {
        display: "grid",
        gridTemplateColumns: "1fr 1fr",
        gap: 10,
    },
    checkLabel: {
        display: "flex",
        alignItems: "center",
        gap: 8,
        fontSize: 12,
        fontWeight: 500,
        color: "#cbd5e1",
        cursor: "pointer",
    },
    checkbox: {
        width: 16,
        height: 16,
        borderRadius: 3,
        border: "1px solid #334155",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        cursor: "pointer",
        flexShrink: 0,
    },
    usePassBtn: {
        width: "100%",
        padding: "8px 0",
        background: "#1e293b",
        border: "1px solid #334155",
        borderRadius: 8,
        color: "#e2e8f0",
        fontSize: 12,
        fontWeight: 700,
        cursor: "pointer",
    },
    tagsContainer: {
        display: "flex",
        flexWrap: "wrap",
        gap: 8,
        alignItems: "center",
        background: "#020617",
        border: "1px solid #1e293b",
        borderRadius: 8,
        padding: 8,
        minHeight: 46,
    },
    tag: {
        display: "flex",
        alignItems: "center",
        gap: 4,
        padding: "4px 10px",
        background: "rgba(88, 86, 214, 0.1)",
        border: "1px solid rgba(88, 86, 214, 0.3)",
        color: "#818cf8",
        fontSize: 12,
        fontWeight: 700,
        borderRadius: 20,
    },
    tagX: {
        background: "none",
        border: "none",
        cursor: "pointer",
        display: "flex",
        color: "inherit",
        padding: 0,
    },
    tagInput: {
        background: "transparent",
        border: "none",
        color: "#cbd5e1",
        fontSize: 12,
        fontWeight: 500,
        outline: "none",
        width: 80,
        marginLeft: 4,
    },
    bottomRow: {
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        paddingTop: 8,
    },
    addFieldBtn: {
        display: "flex",
        alignItems: "center",
        gap: 6,
        background: "none",
        border: "none",
        color: "#818cf8",
        fontSize: 12,
        fontWeight: 700,
        cursor: "pointer",
    },
    favToggle: {
        display: "flex",
        alignItems: "center",
        gap: 10,
    },
    favLabel: {
        fontSize: 10,
        fontWeight: 700,
        color: "#64748b",
        textTransform: "uppercase",
        letterSpacing: "0.12em",
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
    footer: {
        display: "flex",
        justifyContent: "flex-end",
        gap: 12,
        padding: "16px 24px",
        borderTop: "1px solid #1e293b",
        background: "rgba(15, 23, 42, 0.5)",
    },
    cancelBtn: {
        padding: "8px 24px",
        background: "transparent",
        color: "#94a3b8",
        border: "none",
        fontSize: 13,
        fontWeight: 700,
        cursor: "pointer",
    },
    saveBtn: {
        padding: "8px 32px",
        background: "#5856D6",
        color: "#fff",
        border: "none",
        borderRadius: 8,
        fontSize: 13,
        fontWeight: 700,
        cursor: "pointer",
        boxShadow: "0 4px 12px rgba(88, 86, 214, 0.2)",
    },
};
