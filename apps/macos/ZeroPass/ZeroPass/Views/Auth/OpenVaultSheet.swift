import AppKit
import SwiftUI

struct OpenVaultSheet: View {
    enum Mode {
        case openExisting
        case replaceCurrent

        var title: String {
            switch self {
            case .openExisting:
                return "Open Existing Vault"
            case .replaceCurrent:
                return "Choose Different Vault"
            }
        }

        var actionTitle: String {
            switch self {
            case .openExisting:
                return "Choose Folder…"
            case .replaceCurrent:
                return "Choose Different Vault…"
            }
        }
    }

    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    let mode: Mode

    @State private var isBusy = false
    @State private var localError: String?

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            Text(mode.title)
                .font(.title2)
                .fontWeight(.semibold)

            Text(descriptionText)
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)

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

                Spacer()

                Button {
                    chooseFolder()
                } label: {
                    HStack(spacing: ZPTheme.spacing8) {
                        if isBusy {
                            ProgressView()
                                .controlSize(.small)
                        }
                        Label(mode.actionTitle, systemImage: "folder")
                    }
                }
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(isBusy)
            }
        }
        .padding(ZPTheme.spacing24)
        .frame(width: ZPTheme.authSheetWidth)
    }

    private var descriptionText: String {
        switch mode {
        case .openExisting:
            return "Select the folder containing your ZeroPass vault to continue."
        case .replaceCurrent:
            return "Select another vault folder. ZeroPass will keep your current vault open unless the new vault opens successfully."
        }
    }

    private func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = false

        guard panel.runModal() == .OK, let url = panel.url else {
            return
        }

        localError = nil
        isBusy = true

        Task {
            defer { isBusy = false }

            do {
                switch mode {
                case .openExisting:
                    try await vault.openVault(url)
                case .replaceCurrent:
                    try await vault.replaceVault(with: url)
                }
                dismiss()
            } catch {
                localError = error.localizedDescription
            }
        }
    }
}
