import AppKit
import SwiftUI

struct SecuritySettingsView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var showingChangePassword = false

    @State private var showingRecoveryConfirm = false
    @State private var showingRecoverySheet = false
    @State private var recoveryMnemonic: String = ""

    @State private var showingExportConfirm = false
    @State private var exportConfirmText: String = ""

    @State private var showingCSVExportConfirm = false
    @State private var csvExportConfirmText: String = ""

    var body: some View {
        Form {
            Section("Auto-lock") {
                Stepper(value: $vault.autoLockTimeoutSeconds, in: 0...3600, step: 60) {
                    if vault.autoLockTimeoutSeconds == 0 {
                        Text("Auto-lock: Never")
                    } else {
                        Text("Auto-lock after \(vault.autoLockTimeoutSeconds / 60)m")
                    }
                }

                Toggle("Lock on sleep", isOn: $vault.lockOnSleepEnabled)
                Toggle("Lock on screen sleep", isOn: $vault.lockOnScreenSleepEnabled)
            }

            Section("Clipboard") {
                Toggle("Auto-clear clipboard", isOn: $vault.clipboardAutoClearEnabled)

                Stepper(value: $vault.clipboardAutoClearSeconds, in: 5...300, step: 5) {
                    Text("Clear clipboard after \(vault.clipboardAutoClearSeconds)s")
                }
                .disabled(!vault.clipboardAutoClearEnabled)
            }

            Section("Biometrics") {
                Toggle("Enable Touch ID unlock", isOn: $vault.biometricUnlockEnabled)
                    .disabled(!vault.isBiometricAvailable)
            }

            Section("Master Password") {
                Button("Change master password…") {
                    showingChangePassword = true
                }
                .disabled(!vault.hasVault)
            }

            Section("Recovery Key") {
                Button("Regenerate recovery key…") {
                    showingRecoveryConfirm = true
                }
                .disabled(!vault.isUnlocked)

                Text("Recovery key is shown ONLY during creation or regeneration (never stored).")
                    .foregroundStyle(.secondary)
                    .font(.system(size: 12))
            }

            Section("Data") {
                HStack {
                    Button("Export Encrypted…") {
                        exportEncrypted()
                    }
                    .disabled(!vault.isUnlocked)

                    Button("Export JSON…") {
                        showingExportConfirm = true
                    }
                    .disabled(!vault.isUnlocked)

                    Button("Export CSV…") {
                        showingCSVExportConfirm = true
                    }
                    .disabled(!vault.isUnlocked)
                }

                Menu("Import from…") {
                    Button("Chrome CSV…") { importFrom(.chrome) }
                    Button("Firefox CSV…") { importFrom(.firefox) }
                    Button("Safari CSV…") { importFrom(.safari) }
                    Divider()
                    Button("1Password CSV…") { importFrom(.onePassword) }
                    Button("1Password 1PUX…") { importFrom(.onePasswordPUX) }
                    Button("Bitwarden…") { importFrom(.bitwarden) }
                    Button("LastPass CSV…") { importFrom(.lastPass) }
                    Button("KeePass CSV…") { importFrom(.keepass) }
                    Divider()
                    Button("Generic CSV…") { importFrom(.csv) }
                }
                .disabled(!vault.isUnlocked)
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }
        }
        .formStyle(.grouped)
        .sheet(isPresented: $showingChangePassword) {
            ChangePasswordSheet()
                .environmentObject(vault)
                .frame(width: 520)
        }
        .alert("Regenerate Recovery Key", isPresented: $showingRecoveryConfirm) {
            Button("Cancel", role: .cancel) {}
            Button("Regenerate", role: .destructive) {
                Task {
                    do {
                        recoveryMnemonic = try await vault.regenerateRecoveryMnemonic()
                        showingRecoverySheet = true
                    } catch {
                        vault.lastError = error.localizedDescription
                    }
                }
            }
        } message: {
            Text("This will invalidate your current recovery key. Make sure you can record the new one.")
        }
        .sheet(isPresented: $showingRecoverySheet, onDismiss: { recoveryMnemonic = "" }) {
            RecoveryMnemonicSheet(mnemonic: $recoveryMnemonic)
                .frame(width: 520)
        }
        .alert("Export Unencrypted JSON", isPresented: $showingExportConfirm) {
            TextField("Type EXPORT to continue", text: $exportConfirmText)
            Button("Cancel", role: .cancel) { exportConfirmText = "" }
            Button("Export", role: .destructive) {
                guard exportConfirmText.uppercased() == "EXPORT" else { return }
                exportJSON()
                exportConfirmText = ""
            }
        } message: {
            Text("Exported JSON is unencrypted. Anyone with the file can read your secrets.")
        }
        .alert("Export Unencrypted CSV", isPresented: $showingCSVExportConfirm) {
            TextField("Type EXPORT to continue", text: $csvExportConfirmText)
            Button("Cancel", role: .cancel) { csvExportConfirmText = "" }
            Button("Export", role: .destructive) {
                guard csvExportConfirmText.uppercased() == "EXPORT" else { return }
                exportCSV()
                csvExportConfirmText = ""
            }
        } message: {
            Text("Exported CSV is unencrypted. Anyone with the file can read your secrets.")
        }
    }

    private func exportEncrypted() {
        let panel = NSSavePanel()
        panel.nameFieldStringValue = "zeropass-export.zpenc"

        if panel.runModal() == .OK, let url = panel.url {
            Task {
                do {
                    try await vault.exportEncrypted(to: url)
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }

    private func exportJSON() {
        let panel = NSSavePanel()
        panel.nameFieldStringValue = "zeropass-export.json"

        if panel.runModal() == .OK, let url = panel.url {
            Task {
                do {
                    try await vault.exportJSON(to: url)
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }

    private func exportCSV() {
        let panel = NSSavePanel()
        panel.nameFieldStringValue = "zeropass-export.csv"

        if panel.runModal() == .OK, let url = panel.url {
            Task {
                do {
                    try await vault.exportCSV(to: url)
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }

    private func importCSV() {
        importFrom(.csv)
    }

    private func importFrom(_ source: VaultClient.ImportSource) {
        let panel = NSOpenPanel()
        panel.canChooseFiles = true
        panel.canChooseDirectories = false
        panel.allowsMultipleSelection = false

        if panel.runModal() == .OK, let url = panel.url {
            Task {
                do {
                    _ = try await vault.importFrom(source: source, url: url)
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }
}

private struct ChangePasswordSheet: View {
    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State private var oldPassword = ""
    @State private var newPassword = ""
    @State private var confirmPassword = ""
    @State private var strengthSummary = ""

    @State private var isBusy = false
    @State private var errorText: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Change Master Password")
                .font(.title2)
                .bold()

            Form {
                SecureField("Current password", text: $oldPassword)
                SecureField("New password", text: $newPassword)
                SecureField("Confirm new password", text: $confirmPassword)

                if !strengthSummary.isEmpty {
                    Text(strengthSummary)
                        .foregroundStyle(.secondary)
                        .font(.system(size: 12))
                }

                if let errorText {
                    Text(errorText)
                        .foregroundStyle(.red)
                }
            }
            .formStyle(.grouped)

            HStack {
                Spacer()
                Button("Cancel") { dismiss() }
                    .disabled(isBusy)

                Button("Change") {
                    changePassword()
                }
                .disabled(isBusy || oldPassword.isEmpty || newPassword.isEmpty || newPassword != confirmPassword)
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(20)
        .onChange(of: newPassword) { _, _ in
            Task {
                strengthSummary = await vault.scorePasswordSummary(newPassword)
            }
        }
    }

    private func changePassword() {
        isBusy = true
        errorText = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.changeMasterPassword(old: oldPassword, new: newPassword)
                dismiss()
            } catch {
                errorText = error.localizedDescription
            }
        }
    }
}

private struct RecoveryMnemonicSheet: View {
    @Binding var mnemonic: String
    @Environment(\.dismiss) private var dismiss

    @State private var confirmed = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("New Recovery Key")
                .font(.title2)
                .bold()

            Text("Write this down. It will not be shown again.")
                .foregroundStyle(.secondary)

            Text(mnemonic)
                .font(.system(.body, design: .monospaced))
                .textSelection(.enabled)
                .padding(10)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(.regularMaterial)
                .clipShape(RoundedRectangle(cornerRadius: 8))

            Toggle("I have saved this recovery key", isOn: $confirmed)

            HStack {
                Spacer()
                Button("Done") {
                    dismiss()
                }
                .disabled(!confirmed)
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(20)
    }
}
