import SwiftUI

struct ItemEditorView: View, Identifiable {
    let id: String

    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State var item: VaultItem

    init(item: VaultItem) {
        self.id = item.id.isEmpty ? UUID().uuidString : item.id
        self._item = State(initialValue: item)
    }
    @State private var isBusy = false
    @State private var newFieldKey: String = ""
    @State private var newFieldValue: String = ""
    @State private var newTag: String = ""
    @State private var showPasswordGenerator = false
    @State private var passwordStrength: String = ""
    @State private var generatorTargetKey: String?
    @State private var revealedSecretKeys: Set<String> = []

    @FocusState private var nameFieldFocused: Bool

    // MARK: - Computed Helpers

    /// Orders field keys with well-known fields first, then alphabetical custom keys.
    private var orderedFieldKeys: [String] {
        let priority = [
            "username", "email", "user", "login",
            "password", "api_key", "api_secret", "secret",
            "private_key", "public_key", "passphrase",
            "url", "endpoint",
            "card_number", "cardholder", "expiry", "cvv",
            "full_name", "phone", "address",
            "credential_id", "relying_party", "user_handle",
        ]
        let allKeys = item.fields.keys.sorted()
        var ordered: [String] = []
        for key in priority where allKeys.contains(key) {
            ordered.append(key)
        }
        for key in allKeys where !ordered.contains(key) {
            ordered.append(key)
        }
        return ordered
    }

    private func isSecretField(_ key: String) -> Bool {
        item.type.sensitiveFieldKeys.contains(key)
    }

    private func fieldLabel(_ key: String) -> String {
        key.replacingOccurrences(of: "_", with: " ").capitalized
    }

    private var metadataKeys: Set<String> {
        ["url", "endpoint", "relying_party"]
    }

    private var credentialFieldKeys: [String] {
        orderedFieldKeys.filter { !metadataKeys.contains($0) }
    }

    private var metadataFieldKeys: [String] {
        orderedFieldKeys.filter { metadataKeys.contains($0) }
    }

    private var editorTitle: String {
        let trimmedName = item.name.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmedName.isEmpty {
            return trimmedName
        }

        return item.id.isEmpty ? "New \(item.type.displayName)" : item.type.displayName
    }

    private var saveButtonTitle: String {
        item.id.isEmpty ? "Save Item" : "Save Changes"
    }

    // MARK: - Body

    var body: some View {
        ZStack {
            ZPTheme.workspaceBackground
                .ignoresSafeArea()

            ScrollView {
                VStack(spacing: 0) {
                    editorCard
                        .frame(maxWidth: 680)
                        .frame(maxWidth: .infinity)
                }
                .padding(.horizontal, ZPTheme.spacing24)
                .padding(.vertical, ZPTheme.spacing24)
            }
        }
        .frame(minWidth: 620, idealWidth: 720, minHeight: 620, idealHeight: 760)
        .popover(isPresented: $showPasswordGenerator) {
            PasswordGeneratorPopover(
                vault: vault,
                onSelect: { generated in
                    if let key = generatorTargetKey {
                        item.fields[key] = generated
                    }
                    showPasswordGenerator = false
                    updatePasswordStrength(generated)
                }
            )
        }
        .onAppear {
            if item.id.isEmpty && item.fields.isEmpty {
                item.fields = item.type.defaultFields
            }

            DispatchQueue.main.async {
                nameFieldFocused = true
            }
        }
        .onChange(of: item.type) { oldType, newType in
            guard item.id.isEmpty else { return }
            // Apply template when type changes on new items with default/empty fields
            let oldDefaults = oldType.defaultFields
            if item.fields == oldDefaults || item.fields.isEmpty {
                item.fields = newType.defaultFields
            }
        }
    }

    private var editorCard: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing24) {
            headerView
            credentialsSection

            if !metadataFieldKeys.isEmpty || !newFieldKey.isEmpty || !newFieldValue.isEmpty || !item.fields.keys.isEmpty {
                metadataSection
            }

            notesAndTagsSection
            footerView
        }
        .padding(ZPTheme.spacing24)
        .zpSurface(.elevated, radius: 30)
    }

    private var headerView: some View {
        HStack(alignment: .top, spacing: ZPTheme.spacing16) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: ZPTheme.spacing8) {
                        ForEach(VaultItemType.allCases) { type in
                            Button {
                                item.type = type
                            } label: {
                                Label(type.displayName, systemImage: type.symbolName)
                                    .font(.caption.weight(.semibold))
                                    .foregroundStyle(item.type == type ? ZPTheme.accent : ZPTheme.textSecondary)
                                    .padding(.horizontal, ZPTheme.spacing10)
                                    .padding(.vertical, ZPTheme.spacing6)
                                    .background(item.type == type ? ZPTheme.pillBackground : ZPTheme.chipBackground, in: Capsule())
                                    .overlay(
                                        Capsule()
                                            .stroke(item.type == type ? ZPTheme.pillBorder : ZPTheme.panelBorder, lineWidth: 1)
                                    )
                            }
                            .buttonStyle(.plain)
                            .accessibilityLabel("Set item type to \(type.displayName)")
                        }
                    }
                }

                TextField("Item name", text: $item.name)
                    .textFieldStyle(.plain)
                    .font(.system(size: 30, weight: .bold, design: .rounded))
                    .foregroundStyle(ZPTheme.textPrimary)
                    .focused($nameFieldFocused)

                Text("Securely encrypted locally")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.textSecondary)
            }

            Spacer(minLength: ZPTheme.spacing12)

            Button {
                item.favorite.toggle()
            } label: {
                Image(systemName: item.favorite ? "star.fill" : "star")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(item.favorite ? .yellow : ZPTheme.textSecondary)
                    .frame(width: 30, height: 30)
            }
            .buttonStyle(.plain)
            .accessibilityLabel(item.favorite ? "Remove favorite" : "Mark as favorite")
        }
    }

    private var credentialsSection: some View {
        editorSection(title: "Credentials", accessory: AnyView(securityNoteView)) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
                ForEach(credentialFieldKeys, id: \.self) { key in
                    fieldEditorRow(for: key)
                }

                if !passwordStrength.isEmpty {
                    Text(passwordStrength)
                        .font(.caption)
                        .foregroundStyle(ZPTheme.textSecondary)
                        .padding(.top, 2)
                }
            }
        }
    }

    private var metadataSection: some View {
        editorSection(title: "Metadata") {
            VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
                ForEach(metadataFieldKeys, id: \.self) { key in
                    fieldEditorRow(for: key)
                }

                addFieldComposer
            }
        }
    }

    private var notesAndTagsSection: some View {
        editorSection(title: "Notes & Tags") {
            VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
                VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                    Text("Secure Notes")
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(ZPTheme.textSecondary)

                    TextEditor(text: $item.notes)
                        .font(.body)
                        .scrollContentBackground(.hidden)
                        .padding(ZPTheme.spacing12)
                        .frame(minHeight: 92, maxHeight: 148)
                        .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
                }

                VStack(alignment: .leading, spacing: ZPTheme.spacing10) {
                    Text("Tags")
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(ZPTheme.textSecondary)

                    FlowLayout(spacing: 8) {
                        ForEach(item.tags, id: \.self) { tag in
                            HStack(spacing: ZPTheme.spacing6) {
                                Text(tag)
                                    .font(.caption.weight(.semibold))
                                    .foregroundStyle(ZPTheme.textPrimary)

                                Button {
                                    removeTag(tag)
                                } label: {
                                    Image(systemName: "xmark")
                                        .font(.system(size: 9, weight: .bold))
                                }
                                .buttonStyle(.plain)
                                .foregroundStyle(ZPTheme.textSecondary)
                                .accessibilityLabel("Remove tag \(tag)")
                            }
                            .padding(.horizontal, ZPTheme.spacing10)
                            .padding(.vertical, ZPTheme.spacing6)
                            .background(ZPTheme.chipBackground, in: Capsule())
                            .overlay(
                                Capsule()
                                    .stroke(ZPTheme.panelBorder, lineWidth: 1)
                            )
                        }
                    }
                    .frame(maxWidth: CGFloat.infinity, alignment: Alignment.leading)

                    HStack(spacing: ZPTheme.spacing10) {
                        TextField("Add tag…", text: $newTag)
                            .textFieldStyle(.plain)
                            .onSubmit(addTag)

                        Button {
                            addTag()
                        } label: {
                            Image(systemName: "plus")
                                .font(.system(size: 11, weight: .bold))
                        }
                        .buttonStyle(.plain)
                        .disabled(newTag.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                        .accessibilityLabel("Add tag")
                    }
                    .padding(.horizontal, ZPTheme.spacing12)
                    .padding(.vertical, ZPTheme.spacing10)
                    .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
                }
            }
        }
    }

    private var footerView: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing12) {
            Divider()
                .overlay(ZPTheme.separator)

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(ZPTheme.destructive)
                    .font(.callout)
                    .textSelection(.enabled)
            }

            HStack(spacing: ZPTheme.spacing12) {
                Label("Securely encrypted locally", systemImage: "lock.shield")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.textSecondary)

                Spacer()

                Button("Cancel") {
                    dismiss()
                }
                .keyboardShortcut(.cancelAction)
                .disabled(isBusy)

                Button(saveButtonTitle) {
                    save()
                }
                .buttonStyle(.borderedProminent)
                .disabled(isBusy || item.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .keyboardShortcut(.defaultAction)
            }
        }
    }

    private var securityNoteView: some View {
        Label("End-to-end encrypted", systemImage: "checkmark.shield")
            .font(.caption)
            .foregroundStyle(ZPTheme.textSecondary)
    }

    @ViewBuilder
    private func editorSection(title: String, accessory: AnyView? = nil, @ViewBuilder content: () -> some View) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
            HStack(alignment: .firstTextBaseline) {
                Text(title.uppercased())
                    .font(.caption.weight(.semibold))
                    .tracking(1.4)
                    .foregroundStyle(ZPTheme.textTertiary)

                Spacer()

                if let accessory {
                    accessory
                }
            }

            content()
        }
    }

    @ViewBuilder
    private func fieldEditorRow(for key: String) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
            Text(fieldLabel(key))
                .font(.caption.weight(.semibold))
                .foregroundStyle(ZPTheme.textSecondary)

            HStack(spacing: ZPTheme.spacing10) {
                Image(systemName: iconName(for: key))
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(ZPTheme.textSecondary)
                    .frame(width: 18)

                Group {
                    if isSecretField(key) && !revealedSecretKeys.contains(key) {
                        SecureField(fieldLabel(key), text: fieldBinding(for: key))
                    } else {
                        TextField(fieldLabel(key), text: fieldBinding(for: key))
                    }
                }
                .textFieldStyle(.plain)
                .font(.body.weight(.medium))
                .autocorrectionDisabled()
                .textContentType(isSecretField(key) ? .password : .none)
                .privacySensitive(isSecretField(key))
                .onChange(of: item.fields[key, default: ""]) { _, newValue in
                    if key == "password" || key == "api_key" || key == "api_secret" || key == "secret" || key == "passphrase" {
                        updatePasswordStrength(newValue)
                    }
                }

                if isSecretField(key) {
                    Button {
                        toggleSecretVisibility(for: key)
                    } label: {
                        Image(systemName: revealedSecretKeys.contains(key) ? "eye" : "eye.slash")
                            .font(.system(size: 13, weight: .semibold))
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(ZPTheme.accent)
                    .accessibilityLabel(revealedSecretKeys.contains(key) ? "Hide \(fieldLabel(key))" : "Show \(fieldLabel(key))")

                    Button {
                        generatorTargetKey = key
                        showPasswordGenerator = true
                    } label: {
                        Image(systemName: "wand.and.stars")
                            .font(.system(size: 13, weight: .semibold))
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(ZPTheme.accent)
                    .help("Generate value")
                    .accessibilityLabel("Generate value for \(fieldLabel(key))")
                }

                if isURLKey(key), let value = item.fields[key], !value.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    Button {
                        openURL(value)
                    } label: {
                        Image(systemName: "arrow.up.right.square")
                            .font(.system(size: 13, weight: .semibold))
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(ZPTheme.accent)
                    .accessibilityLabel("Open \(fieldLabel(key))")
                }

                Button(role: .destructive) {
                    item.fields.removeValue(forKey: key)
                    revealedSecretKeys.remove(key)
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 13, weight: .semibold))
                }
                .buttonStyle(.plain)
                .foregroundStyle(ZPTheme.textMuted)
                .help("Remove field")
                .accessibilityLabel("Remove \(fieldLabel(key)) field")
            }
            .padding(.horizontal, ZPTheme.spacing12)
            .padding(.vertical, ZPTheme.spacing10)
            .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
        }
    }

    private var addFieldComposer: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
            Text("Add Field")
                .font(.caption.weight(.semibold))
                .foregroundStyle(ZPTheme.textSecondary)

            HStack(spacing: ZPTheme.spacing10) {
                TextField("Field key", text: $newFieldKey)
                    .textFieldStyle(.plain)

                Divider()
                    .frame(height: 18)

                TextField("Value", text: $newFieldValue)
                    .textFieldStyle(.plain)

                Button {
                    addField()
                } label: {
                    Image(systemName: "plus")
                        .font(.system(size: 11, weight: .bold))
                }
                .buttonStyle(.plain)
                .disabled(newFieldKey.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                .accessibilityLabel("Add new field")
            }
            .padding(.horizontal, ZPTheme.spacing12)
            .padding(.vertical, ZPTheme.spacing10)
            .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
        }
    }

    // MARK: - Actions

    private func updatePasswordStrength(_ password: String) {
        Task {
            let summary = await vault.scorePasswordSummary(password)
            passwordStrength = summary
        }
    }

    private func fieldBinding(for key: String) -> Binding<String> {
        Binding(
            get: { item.fields[key] ?? "" },
            set: { item.fields[key] = $0 }
        )
    }

    private func iconName(for key: String) -> String {
        switch key {
        case "username", "email", "user", "login", "full_name":
            return "person"
        case "password", "api_key", "api_secret", "secret", "private_key", "public_key", "passphrase", "credential_id", "cvv", "card_number":
            return "key"
        case "url", "endpoint":
            return "globe"
        case "relying_party":
            return "network"
        case "phone":
            return "phone"
        case "address":
            return "location"
        case "cardholder", "expiry":
            return "creditcard"
        default:
            return "square.grid.2x2"
        }
    }

    private func isURLKey(_ key: String) -> Bool {
        key == "url" || key == "endpoint"
    }

    private func openURL(_ value: String) {
        let normalized = value.hasPrefix("http://") || value.hasPrefix("https://") ? value : "https://\(value)"
        guard let url = URL(string: normalized) else { return }
        NSWorkspace.shared.open(url)
    }

    private func toggleSecretVisibility(for key: String) {
        if revealedSecretKeys.contains(key) {
            revealedSecretKeys.remove(key)
        } else {
            revealedSecretKeys.insert(key)
        }
    }

    private func addField() {
        let trimmedKey = newFieldKey.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedKey.isEmpty else { return }

        item.fields[trimmedKey] = newFieldValue
        newFieldKey = ""
        newFieldValue = ""
    }

    private func addTag() {
        let trimmedTag = newTag.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTag.isEmpty else { return }
        guard !item.tags.contains(trimmedTag) else {
            newTag = ""
            return
        }

        item.tags.append(trimmedTag)
        item.tags.sort()
        newTag = ""
    }

    private func removeTag(_ tag: String) {
        item.tags.removeAll { $0 == tag }
    }

    private func save() {
        isBusy = true
        vault.lastError = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.saveItem(item)
                dismiss()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}

// MARK: - Password Generator Popover

private struct PasswordGeneratorPopover: View {
    let vault: VaultClient
    let onSelect: (String) -> Void

    @State private var mode: GeneratorMode = .random
    @State private var length: Double = 20
    @State private var words: Double = 5
    @State private var separator: String = "-"
    @State private var uppercase = true
    @State private var lowercase = true
    @State private var digits = true
    @State private var symbols = true
    @State private var excludeAmbiguous = false
    @State private var preview: String = ""
    @State private var strength: String = ""
    @State private var isGenerating = false

    enum GeneratorMode: String, CaseIterable {
        case random = "Random"
        case passphrase = "Passphrase"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Generate Password")
                .font(.headline)

            Picker("Mode", selection: $mode) {
                ForEach(GeneratorMode.allCases, id: \.self) { m in
                    Text(m.rawValue).tag(m)
                }
            }
            .pickerStyle(.segmented)

            if mode == .random {
                HStack {
                    Text("Length: \(Int(length))")
                    Slider(value: $length, in: 8...128, step: 1)
                }

                Toggle("Uppercase (A-Z)", isOn: $uppercase)
                Toggle("Lowercase (a-z)", isOn: $lowercase)
                Toggle("Digits (0-9)", isOn: $digits)
                Toggle("Symbols (!@#…)", isOn: $symbols)
                Toggle("Exclude ambiguous (0O, 1Il)", isOn: $excludeAmbiguous)
            } else {
                HStack {
                    Text("Words: \(Int(words))")
                    Slider(value: $words, in: 4...8, step: 1)
                }

                TextField("Separator", text: $separator)
                    .frame(width: 80)
            }

            Divider()

            // Preview
            if !preview.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text(preview)
                        .font(.system(.body, design: .monospaced))
                        .textSelection(.enabled)
                        .padding(8)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(.quaternary)
                        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))

                    if !strength.isEmpty {
                        Text(strength)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            HStack {
                Button("Generate") {
                    generate()
                }
                .disabled(isGenerating)

                Spacer()

                Button("Use This") {
                    onSelect(preview)
                }
                .disabled(preview.isEmpty)
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(16)
        .frame(width: 320)
        .onAppear { generate() }
    }

    private func generate() {
        isGenerating = true
        Task {
            defer { isGenerating = false }
            do {
                if mode == .random {
                    preview = try await vault.generatePassword(
                        length: Int(length),
                        options: ZPBridge.GeneratePasswordOptions(
                            uppercase: uppercase,
                            lowercase: lowercase,
                            digits: digits,
                            symbols: symbols,
                            excludeAmbiguous: excludeAmbiguous
                        )
                    )
                } else {
                    preview = try await vault.generatePassphrase(
                        words: Int(words),
                        separator: separator
                    )
                }
                // Score the generated password
                let score = try await vault.scorePassword(preview)
                strength = "Strength: \(score.score)/4 — \(score.feedback)"
            } catch {
                preview = "Generation failed"
                strength = error.localizedDescription
            }
        }
    }
}
