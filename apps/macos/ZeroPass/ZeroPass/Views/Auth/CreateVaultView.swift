import SwiftUI

struct CreateVaultView: View {
    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State private var folderURL: URL?
    @State private var password: String = ""
    @State private var confirm: String = ""

    @State private var isBusy = false
    @State private var strength: String = ""
    @State private var strengthScore: Int = 0
    @State private var localError: String?
    @State private var passwordStrengthTask: Task<Void, Never>?

    private enum Field { case password, confirm }
    @FocusState private var focusedField: Field?

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing20) {
            Text("Create Vault")
                .font(.title2)
                .fontWeight(.semibold)
                .accessibilityIdentifier("createVault.title")

            Text("Choose a folder and set the master password you’ll use to unlock this vault.")
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Vault Location")
                    .font(.callout)
                    .fontWeight(.medium)
                    .foregroundStyle(ZPTheme.textSecondary)

                HStack(alignment: .firstTextBaseline, spacing: ZPTheme.spacing12) {
                    Text(folderDisplayText)
                        .lineLimit(1)
                        .truncationMode(.middle)
                        .foregroundStyle(folderURL == nil ? ZPTheme.textSecondary : ZPTheme.textPrimary)
                        .accessibilityLabel(folderAccessibilityLabel)

                    Spacer()

                    Button("Choose Folder…") {
                        Task { await chooseFolder() }
                    }
                    .disabled(isBusy)
                    .accessibilityLabel("Choose vault folder")
                }
            }

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Master Password")
                    .font(.callout)
                    .fontWeight(.medium)
                    .foregroundStyle(ZPTheme.textSecondary)

                SecureField("Enter a master password", text: $password)
                    .textFieldStyle(.roundedBorder)
                    .focused($focusedField, equals: .password)
                    .accessibilityLabel("Master password")
                    .onChange(of: password) { _, newValue in
                        updatePasswordStrength(for: newValue)
                    }
            }

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Confirm Password")
                    .font(.callout)
                    .fontWeight(.medium)
                    .foregroundStyle(ZPTheme.textSecondary)

                SecureField("Confirm your master password", text: $confirm)
                    .textFieldStyle(.roundedBorder)
                    .focused($focusedField, equals: .confirm)
                    .accessibilityLabel("Confirm password")
            }

            if !password.isEmpty {
                PasswordStrengthBar(score: strengthScore)
                    .accessibilityLabel("Password strength: \(strengthLabel)")

                if !strength.isEmpty {
                    Text(strength)
                        .font(.caption)
                        .foregroundStyle(strengthColor)
                }
            }

            if !confirm.isEmpty && password != confirm {
                Text("Passwords do not match.")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.destructive)
            }

            if let localError, !localError.isEmpty {
                AuthMessageView(
                    text: localError,
                    systemImage: "exclamationmark.triangle.fill",
                    tone: .error
                )
            }

            HStack {
                Button("Cancel") {
                    dismiss()
                }
                .keyboardShortcut(.cancelAction)
                .disabled(isBusy)
                .accessibilityIdentifier("createVault.cancelButton")

                Spacer()

                Button {
                    create()
                } label: {
                    HStack(spacing: ZPTheme.spacing8) {
                        if isBusy {
                            ProgressView()
                                .controlSize(.small)
                        }
                        Text("Create Vault")
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(isBusy || folderURL == nil || password.isEmpty || password != confirm)
                .keyboardShortcut(.defaultAction)
                .accessibilityLabel("Create vault")
            }
        }
        .padding(ZPTheme.spacing24)
        .frame(minWidth: ZPTheme.authSheetWidth, idealWidth: ZPTheme.authSheetWidth)
        .onAppear { focusedField = .password }
        .onDisappear { passwordStrengthTask?.cancel() }
    }

    private var folderDisplayText: String {
        folderURL?.path(percentEncoded: false) ?? "No folder chosen"
    }

    private var folderAccessibilityLabel: String {
        if let folderURL {
            return "Selected folder: \(folderURL.lastPathComponent)"
        }
        return "No folder selected"
    }

    private var strengthLabel: String {
        switch strengthScore {
        case 0: return "Very weak"
        case 1: return "Weak"
        case 2: return "Fair"
        case 3: return "Strong"
        default: return "Very strong"
        }
    }

    private var strengthColor: Color {
        switch strengthScore {
        case 0: return .red
        case 1: return .orange
        case 2: return .yellow
        case 3: return .green
        default: return .teal
        }
    }

    @MainActor
    private func chooseFolder() async {
        guard let url = await VaultFolderPicker.pickDirectory(canCreateDirectories: true) else {
            return
        }

        folderURL = url
        focusedField = .password
    }

    private func create() {
        guard let url = folderURL else { return }
        isBusy = true
        localError = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.createVault(url, masterPassword: password)
                dismiss()
            } catch {
                localError = error.localizedDescription
            }
        }
    }

    private func updatePasswordStrength(for newValue: String) {
        passwordStrengthTask?.cancel()

        guard !newValue.isEmpty else {
            strength = ""
            strengthScore = 0
            return
        }

        let candidate = newValue
        passwordStrengthTask = Task {
            do {
                let score = try await vault.scorePassword(candidate)
                guard !Task.isCancelled, password == candidate else { return }
                strength = "Score \(score.score)/4 — \(score.feedback)"
                strengthScore = score.score
            } catch {
                guard !Task.isCancelled, password == candidate else { return }
                strength = ""
                strengthScore = 0
            }
        }
    }
}
