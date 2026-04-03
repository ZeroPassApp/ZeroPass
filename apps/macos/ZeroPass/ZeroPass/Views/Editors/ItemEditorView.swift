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
    @State private var showPasswordGenerator = false
    @State private var passwordStrength: String = ""
    @State private var generatorTargetKey: String?

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

    // MARK: - Body

    var body: some View {
        VStack(spacing: 0) {
            // Title
            HStack {
                Image(systemName: item.type.symbolName)
                    .font(.title3)
                    .foregroundStyle(item.type.color)

                Text(item.id.isEmpty ? "New Item" : "Edit Item")
                    .font(.title2)
                    .bold()

                Spacer()
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 8)

            Form {
                Section {
                    Picker("Type", selection: $item.type) {
                        ForEach(VaultItemType.allCases) { t in
                            Label(t.displayName, systemImage: t.symbolName).tag(t)
                        }
                    }

                    TextField("Name", text: $item.name)

                    Toggle("Favorite", isOn: $item.favorite)
                }

                Section("Fields") {
                    ForEach(orderedFieldKeys, id: \.self) { key in
                        HStack {
                            if isSecretField(key) {
                                SecureField(fieldLabel(key), text: Binding(
                                    get: { item.fields[key] ?? "" },
                                    set: { item.fields[key] = $0 }
                                ))
                            } else {
                                TextField(fieldLabel(key), text: Binding(
                                    get: { item.fields[key] ?? "" },
                                    set: { item.fields[key] = $0 }
                                ))
                            }

                            if isSecretField(key) {
                                Button {
                                    generatorTargetKey = key
                                    showPasswordGenerator = true
                                } label: {
                                    Image(systemName: "wand.and.stars")
                                }
                                .buttonStyle(.borderless)
                                .help("Generate password")
                            }

                            Button(role: .destructive) {
                                item.fields.removeValue(forKey: key)
                            } label: {
                                Image(systemName: "minus.circle")
                            }
                            .buttonStyle(.borderless)
                            .help("Remove field")
                        }
                    }

                    // Add new field
                    HStack {
                        TextField("Key", text: $newFieldKey)
                            .frame(maxWidth: 120)
                        TextField("Value", text: $newFieldValue)
                        Button {
                            guard !newFieldKey.isEmpty else { return }
                            item.fields[newFieldKey] = newFieldValue
                            newFieldKey = ""
                            newFieldValue = ""
                        } label: {
                            Image(systemName: "plus.circle.fill")
                        }
                        .buttonStyle(.borderless)
                        .disabled(newFieldKey.isEmpty)
                        .help("Add field")
                    }
                }

                if !passwordStrength.isEmpty {
                    Section {
                        Text(passwordStrength)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Section("Notes") {
                    TextEditor(text: $item.notes)
                        .font(.body)
                        .frame(minHeight: 60, maxHeight: 120)
                        .scrollContentBackground(.hidden)
                }
            }
            .formStyle(.grouped)

            // Error + Actions
            VStack(spacing: 8) {
                if let err = vault.lastError {
                    Text(err)
                        .foregroundStyle(.red)
                        .font(.callout)
                        .textSelection(.enabled)
                }

                HStack {
                    Button("Cancel") { dismiss() }
                        .keyboardShortcut(.cancelAction)
                        .disabled(isBusy)

                    Spacer()

                    if isBusy {
                        ProgressView()
                            .controlSize(.small)
                    }

                    Button("Save") { save() }
                        .disabled(isBusy || item.name.isEmpty)
                        .keyboardShortcut(.defaultAction)
                }
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 16)
        }
        .frame(minWidth: 500, idealWidth: 600, minHeight: 480, idealHeight: 560)
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

    // MARK: - Actions

    private func updatePasswordStrength(_ password: String) {
        Task {
            let summary = await vault.scorePasswordSummary(password)
            passwordStrength = summary
        }
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
