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

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Create Vault")
                .font(.title2)
                .bold()

            HStack {
                Text(folderURL?.path ?? "No folder selected")
                    .lineLimit(1)
                    .truncationMode(.middle)
                    .foregroundStyle(folderURL == nil ? .secondary : .primary)

                Spacer()

                Button("Choose Folder…") { chooseFolder() }
                    .disabled(isBusy)
            }

            SecureField("Master Password", text: $password)
                .onChange(of: password) { _, newValue in
                    Task { strength = await vault.scorePasswordSummary(newValue) }
                }

            SecureField("Confirm Password", text: $confirm)

            if !strength.isEmpty {
                Text(strength)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }

            HStack {
                Button("Cancel") { dismiss() }
                    .disabled(isBusy)

                Spacer()

                Button("Create") { create() }
                    .disabled(isBusy || folderURL == nil || password.isEmpty || password != confirm)
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding(20)
        .frame(width: 560)
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
        vault.lastError = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.createVault(url, masterPassword: password)
                dismiss()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}
