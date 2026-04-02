import SwiftUI
import AppKit

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var showingCreate = false
    @State private var showingOpen = false

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("ZeroPass")
                .font(.largeTitle)
                .bold()

            Text("Create a new vault or open an existing one.")
                .foregroundStyle(.secondary)

            HStack {
                Button("Create Vault") { showingCreate = true }
                    .keyboardShortcut("n", modifiers: [.command])

                Button("Open Vault") { showingOpen = true }
                    .keyboardShortcut("o", modifiers: [.command])
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }
        }
        .padding(24)
        .frame(minWidth: 520, minHeight: 260)
        .sheet(isPresented: $showingCreate) {
            CreateVaultView()
                .environmentObject(vault)
        }
        .sheet(isPresented: $showingOpen) {
            OpenVaultView()
                .environmentObject(vault)
        }
    }
}

private struct OpenVaultView: View {
    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State private var isBusy = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Open Vault")
                .font(.title2)
                .bold()

            Text("Choose your vault folder.")
                .foregroundStyle(.secondary)

            HStack {
                Button("Choose Folder…") { chooseFolder() }
                    .disabled(isBusy)

                Spacer()

                Button("Cancel") { dismiss() }
                    .disabled(isBusy)
            }
        }
        .padding(20)
        .frame(width: 520)
    }

    private func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = false

        if panel.runModal() == .OK, let url = panel.url {
            isBusy = true
            Task {
                defer { isBusy = false }
                do {
                    try await vault.openVault(url)
                    dismiss()
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }
}
