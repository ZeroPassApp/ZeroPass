import SwiftUI

struct SyncSettingsView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var isBusy = false

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            settingsIntro(
                title: "Sync",
                message: "Configure the preview sync service, register this Mac, and trigger manual sync operations."
            )

            Form {
                Section("Sync Preview") {
                    Toggle("Enable Sync", isOn: $vault.syncEnabled)
                        .disabled(!vault.hasVault)

                    TextField("Server URL", text: $vault.syncServerURL)
                        .textContentType(.URL)
                        .disabled(!vault.syncEnabled)

                    TextField("Device ID", text: $vault.syncDeviceID)
                        .font(.system(.body, design: .monospaced))
                        .disabled(!vault.syncEnabled)

                    SecureField("API Key (stored in Keychain)", text: $vault.syncAPIKey)
                        .disabled(!vault.syncEnabled)

                    if let last = vault.syncLastSyncedAt {
                        Text("Last synced: \(last.formatted(date: .abbreviated, time: .standard))")
                            .foregroundStyle(ZPTheme.textSecondary)
                    } else {
                        Text("Last synced: Never")
                            .foregroundStyle(ZPTheme.textSecondary)
                    }

                    if let status = vault.syncStatus {
                        Text(status)
                            .foregroundStyle(ZPTheme.textSecondary)
                    }

                    Text("Sync remains a preview. API keys are stored in Keychain; sync.json now keeps metadata only.")
                        .font(.footnote)
                        .foregroundStyle(ZPTheme.textSecondary)
                }

                Section("Actions") {
                    TextField("Device name", text: $vault.syncDeviceName)
                        .disabled(!vault.syncEnabled)

                    HStack {
                        Button("Save") { save() }
                            .disabled(isBusy || !vault.syncEnabled || vault.syncServerURL.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)

                        Button("Register") { register() }
                            .disabled(isBusy || !vault.syncEnabled || vault.syncDeviceName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)

                        Button("Sync Now") { syncNow() }
                            .disabled(isBusy || !vault.syncEnabled)
                    }
                }

                if let err = vault.lastError {
                    Text(err)
                        .foregroundStyle(ZPTheme.destructive)
                        .textSelection(.enabled)
                }
            }
            .formStyle(.grouped)
        }
        .background(ZPTheme.workspaceBackground)
    }

    @ViewBuilder
    private func settingsIntro(title: String, message: String) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing6) {
            Text(title)
                .font(.headline)
                .foregroundStyle(ZPTheme.textPrimary)

            Text(message)
                .font(.callout)
                .foregroundStyle(ZPTheme.textSecondary)
        }
        .padding(ZPTheme.spacing18)
        .zpSurface(.muted)
    }

    private func save() {
        isBusy = true
        vault.lastError = nil
        Task {
            defer { isBusy = false }
            do {
                try await vault.applySyncConfig()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }

    private func register() {
        isBusy = true
        vault.lastError = nil
        Task {
            defer { isBusy = false }
            do {
                try await vault.syncRegisterDevice()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }

    private func syncNow() {
        isBusy = true
        vault.lastError = nil
        Task {
            defer { isBusy = false }
            do {
                try await vault.syncNow()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}
