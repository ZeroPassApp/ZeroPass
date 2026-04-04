import AppKit
import SwiftUI

struct OpenVaultSheet: View {
    enum Mode {
        case openExisting
        case replaceCurrent

        var title: String {
            switch self {
            case .openExisting:
                return "Open Vault"
            case .replaceCurrent:
                return "Choose Different Vault"
            }
        }

        var actionTitle: String {
            "Choose Folder…"
        }
    }

    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    let mode: Mode

    @State private var isBusy = false
    @State private var localError: String?

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing20) {
            Text(mode.title)
                .font(.title2)
                .fontWeight(.semibold)
                .accessibilityIdentifier(mode == .openExisting ? "openVault.title" : "replaceVault.title")

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
                .accessibilityIdentifier(mode == .openExisting ? "openVault.cancelButton" : "replaceVault.cancelButton")

                Spacer()

                Button {
                    Task { await chooseFolder() }
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
        .frame(minWidth: ZPTheme.authSheetWidth, idealWidth: ZPTheme.authSheetWidth)
    }

    private var descriptionText: String {
        switch mode {
        case .openExisting:
            return "Select the folder containing your ZeroPass vault."
        case .replaceCurrent:
            return "Select another vault folder. ZeroPass will keep your current vault open unless the new vault opens successfully."
        }
    }

    @MainActor
    private func chooseFolder() async {
        guard let url = await VaultFolderPicker.pickDirectory(canCreateDirectories: false) else {
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

enum VaultFolderPicker {
    @MainActor
    static func pickDirectory(canCreateDirectories: Bool) async -> URL? {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = canCreateDirectories

        return await withCheckedContinuation { continuation in
            let finish: (NSApplication.ModalResponse) -> Void = { response in
                continuation.resume(returning: response == .OK ? panel.url : nil)
            }

            if let targetWindow = presentationWindow() {
                panel.beginSheetModal(for: targetWindow, completionHandler: finish)
            } else {
                finish(panel.runModal())
            }
        }
    }

    @MainActor
    private static func presentationWindow() -> NSWindow? {
        NSApp.keyWindow ?? NSApp.mainWindow ?? NSApp.windows.first(where: \.isVisible)
    }
}
