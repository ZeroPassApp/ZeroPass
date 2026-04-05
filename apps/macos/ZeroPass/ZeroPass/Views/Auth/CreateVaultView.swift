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
            header

            locationPanel

            passwordPanel

            if let localError, !localError.isEmpty {
                AuthMessageView(
                    text: localError,
                    systemImage: "exclamationmark.triangle.fill",
                    tone: .error
                )
            }

            footerActions
        }
        .padding(ZPTheme.spacing24)
        .background(ZPTheme.workspaceBackground)
        .frame(minWidth: ZPTheme.authSheetWidth, idealWidth: ZPTheme.authSheetWidth)
        .onAppear { focusedField = .password }
        .onDisappear { passwordStrengthTask?.cancel() }
    }

    private var header: some View {
        HStack(alignment: .center, spacing: ZPTheme.spacing12) {
            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .fill(ZPTheme.accentSoft)

                Image(systemName: "lock.shield")
                    .font(.title3.weight(.semibold))
                    .foregroundStyle(ZPTheme.accent)
            }
            .frame(width: 42, height: 42)
            .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                Text("Create Vault")
                    .font(.title2.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)
                    .accessibilityIdentifier("createVault.title")

                Text("Choose a folder and set the master password you’ll use to unlock this vault.")
                    .foregroundStyle(ZPTheme.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var locationPanel: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
            Label("Vault Location", systemImage: "folder")
                .font(.callout.weight(.semibold))
                .foregroundStyle(ZPTheme.textPrimary)

            HStack(alignment: .center, spacing: ZPTheme.spacing12) {
                VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                    Text(folderDisplayText)
                        .font(.callout.weight(folderURL == nil ? .regular : .semibold))
                        .foregroundStyle(folderURL == nil ? ZPTheme.textSecondary : ZPTheme.textPrimary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                        .accessibilityLabel(folderAccessibilityLabel)

                    Text(folderURL == nil ? "Choose a destination folder for the new vault." : "ZeroPass will create and manage the vault files in this folder.")
                        .font(.caption)
                        .foregroundStyle(ZPTheme.textTertiary)
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Button("Choose Folder…") {
                    Task { await chooseFolder() }
                }
                .controlSize(.small)
                .disabled(isBusy)
                .accessibilityLabel("Choose vault folder")
                .accessibilityIdentifier("createVault.chooseFolderButton")
            }
            .padding(.horizontal, ZPTheme.spacing14)
            .padding(.vertical, ZPTheme.spacing14)
            .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)

            HStack(spacing: ZPTheme.spacing8) {
                authPill("AES-256 Encrypted", systemImage: "lock.shield")
                authPill("Recovery Phrase Included", systemImage: "key.horizontal")
            }
        }
        .padding(ZPTheme.spacing18)
        .zpSurface(.muted)
    }

    private var passwordPanel: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Master Password")
                    .font(.callout.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                SecureField("Enter a master password", text: $password)
                    .textFieldStyle(.plain)
                    .focused($focusedField, equals: .password)
                    .accessibilityLabel("Master password")
                    .accessibilityIdentifier("createVault.masterPasswordField")
                    .onChange(of: password) { _, newValue in
                        updatePasswordStrength(for: newValue)
                    }
                    .padding(.horizontal, ZPTheme.spacing14)
                    .padding(.vertical, ZPTheme.spacing12)
                    .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
            }

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Confirm Password")
                    .font(.callout.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                SecureField("Confirm your master password", text: $confirm)
                    .textFieldStyle(.plain)
                    .focused($focusedField, equals: .confirm)
                    .accessibilityLabel("Confirm password")
                    .accessibilityIdentifier("createVault.confirmPasswordField")
                    .padding(.horizontal, ZPTheme.spacing14)
                    .padding(.vertical, ZPTheme.spacing12)
                    .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)
            }

            if !password.isEmpty {
                VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                    PasswordStrengthBar(score: strengthScore)
                        .accessibilityLabel("Password strength: \(strengthLabel)")

                    if !strength.isEmpty {
                        Text(strength)
                            .font(.caption)
                            .foregroundStyle(strengthColor)
                    }
                }
            }

            if !confirm.isEmpty && password != confirm {
                Text("Passwords do not match.")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.destructive)
            }

            Text("You’ll receive a one-time recovery phrase after creation. Save it offline before continuing.")
                .font(.caption)
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(ZPTheme.spacing18)
        .zpSurface(.elevated)
    }

    private var footerActions: some View {
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
            .accessibilityIdentifier("createVault.submitButton")
        }
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
                localError = createVaultErrorMessage(for: error)
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

    private func createVaultErrorMessage(for error: Error) -> String {
        if let bridgeError = error as? ZPBridgeError {
            switch bridgeError.code {
            case .notFound:
                return "The selected folder is no longer available. Choose another location and try again."
            case .busy:
                return "ZeroPass is still working with this vault. Please wait a moment and try again."
            default:
                return "ZeroPass couldn’t create the vault in that location. Check folder access and try again."
            }
        }

        return "ZeroPass couldn’t create the vault right now. Please try again."
    }

    @ViewBuilder
    private func authPill(_ title: String, systemImage: String) -> some View {
        Label(title, systemImage: systemImage)
            .font(.caption.weight(.semibold))
            .foregroundStyle(ZPTheme.textSecondary)
            .padding(.horizontal, ZPTheme.spacing10)
            .padding(.vertical, ZPTheme.spacing8)
            .background(ZPTheme.chipBackground, in: Capsule())
            .overlay(
                Capsule()
                    .stroke(ZPTheme.panelBorder, lineWidth: 1)
            )
    }
}
