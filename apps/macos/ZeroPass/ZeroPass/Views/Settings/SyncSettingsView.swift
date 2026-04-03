import SwiftUI

struct SyncSettingsView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var isBusy = false

    var body: some View {
        Form {
            Section("Sync") {
                Toggle("Enable Sync", isOn: $vault.syncEnabled)
                    .disabled(!vault.hasVault)

                TextField("Server URL", text: $vault.syncServerURL)
                    .textContentType(.URL)
                    .disabled(!vault.syncEnabled)

                TextField("Device ID", text: $vault.syncDeviceID)
                    .font(.system(.body, design: .monospaced))
                    .disabled(!vault.syncEnabled)

                SecureField("API Key (optional)", text: $vault.syncAPIKey)
                    .disabled(!vault.syncEnabled)

                if let last = vault.syncLastSyncedAt {
                    Text("Last synced: \(last.formatted(date: .abbreviated, time: .standard))")
                        .foregroundStyle(.secondary)
                } else {
                    Text("Last synced: Never")
                        .foregroundStyle(.secondary)
                }

                if let status = vault.syncStatus {
                    Text(status)
                        .foregroundStyle(.secondary)
                }
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
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }
        }
        .formStyle(.grouped)
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
