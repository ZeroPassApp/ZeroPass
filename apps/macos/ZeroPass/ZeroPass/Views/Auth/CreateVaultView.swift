import SwiftUI
import AppKit

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
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            Text("Create Vault")
                .font(.title2)
                .fontWeight(.semibold)

            Text("Choose a folder on this Mac and set the master password you’ll use to unlock this vault.")
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Vault Location")
                    .font(.caption)
                    .fontWeight(.medium)
                    .foregroundStyle(ZPTheme.textSecondary)

                HStack {
                    Text(folderURL?.path ?? "No folder selected yet")
                        .lineLimit(1)
                        .truncationMode(.middle)
                        .foregroundStyle(folderURL == nil ? ZPTheme.textSecondary : ZPTheme.textPrimary)
                        .accessibilityLabel(folderURL != nil ? "Selected folder: \(folderURL!.lastPathComponent)" : "No folder selected")

                    Spacer()

                    Button("Choose Folder…") {
                        chooseFolder()
                    }
                    .disabled(isBusy)
                    .accessibilityLabel("Choose vault folder")
                }
                .padding(.horizontal, ZPTheme.spacing12)
                .padding(.vertical, ZPTheme.spacing10)
                .background(ZPTheme.authInsetBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
                .overlay(
                    RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                        .stroke(ZPTheme.authInsetBorder, lineWidth: 1)
                )
            }

            VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                Text("Master Password")
                    .font(.caption)
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
                    .font(.caption)
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
                Text("Passwords do not match")
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
                    .disabled(isBusy)

                Spacer()

                Button("Create Vault") {
                    create()
                }
                    .buttonStyle(.borderedProminent)
                    .disabled(isBusy || folderURL == nil || password.isEmpty || password != confirm)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityLabel("Create vault")
            }
        }
        .padding(24)
        .frame(width: 560)
        .onAppear { focusedField = .password }
        .onDisappear { passwordStrengthTask?.cancel() }
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

    private func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = true

        if panel.runModal() == .OK, let url = panel.url {
            folderURL = url
        }
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
