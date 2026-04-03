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

    private enum Field { case password, confirm }
    @FocusState private var focusedField: Field?

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
                    .accessibilityLabel(folderURL != nil ? "Selected folder: \(folderURL!.lastPathComponent)" : "No folder selected")

                Spacer()

                Button("Choose Folder…") { chooseFolder() }
                    .disabled(isBusy)
                    .accessibilityLabel("Choose vault folder")
            }

            SecureField("Master Password", text: $password)
                .focused($focusedField, equals: .password)
                .accessibilityLabel("Master password")
                .onChange(of: password) { _, newValue in
                    Task { strength = await vault.scorePasswordSummary(newValue) }
                }

            SecureField("Confirm Password", text: $confirm)
                .focused($focusedField, equals: .confirm)
                .accessibilityLabel("Confirm password")

            if !strength.isEmpty {
                Text(strength)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityLabel("Password strength: \(strength)")
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
                    .accessibilityLabel("Error: \(err)")
            }

            HStack {
                Button("Cancel") { dismiss() }
                    .disabled(isBusy)

                Spacer()

                Button("Create") { create() }
                    .disabled(isBusy || folderURL == nil || password.isEmpty || password != confirm)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityLabel("Create vault")
            }
        }
        .padding(20)
        .frame(width: 560)
        .onAppear { focusedField = .password }
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
