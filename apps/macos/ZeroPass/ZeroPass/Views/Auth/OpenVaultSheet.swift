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
            header

            if let localError, !localError.isEmpty {
                AuthMessageView(
                    text: localError,
                    systemImage: "exclamationmark.triangle.fill",
                    tone: .error
                )
            }

            guidancePanel

            footerActions
        }
        .padding(ZPTheme.spacing24)
        .background(ZPTheme.workspaceBackground)
        .frame(minWidth: ZPTheme.authSheetWidth, idealWidth: ZPTheme.authSheetWidth)
    }

    private var header: some View {
        HStack(alignment: .center, spacing: ZPTheme.spacing12) {
            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .fill(ZPTheme.accentSoft)

                Image(systemName: mode == .openExisting ? "folder" : "arrow.triangle.2.circlepath")
                    .font(.title3.weight(.semibold))
                    .foregroundStyle(ZPTheme.accent)
            }
            .frame(width: 42, height: 42)
            .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                Text(mode.title)
                    .font(.title2.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)
                    .accessibilityIdentifier(mode == .openExisting ? "openVault.title" : "replaceVault.title")

                Text(descriptionText)
                    .foregroundStyle(ZPTheme.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var guidancePanel: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
            HStack(alignment: .top, spacing: ZPTheme.spacing12) {
                Image(systemName: mode == .openExisting ? "externaldrive.fill" : "arrow.clockwise.circle.fill")
                    .font(.callout.weight(.semibold))
                    .foregroundStyle(ZPTheme.accent)
                    .padding(.top, 2)
                    .accessibilityHidden(true)

                VStack(alignment: .leading, spacing: ZPTheme.spacing6) {
                    Text("Vault folders can be opened directly from any accessible location on this Mac.")
                        .font(.callout)
                        .foregroundStyle(ZPTheme.textSecondary)

                    Text(mode == .openExisting
                         ? "ZeroPass opens the selected vault in a locked state so you can verify it before unlocking."
                         : "If the replacement vault cannot be opened, your current vault stays exactly as it is.")
                        .font(.caption)
                        .foregroundStyle(ZPTheme.textTertiary)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }

            HStack(spacing: ZPTheme.spacing8) {
                authPill(mode == .openExisting ? "Locked First" : "Safe Replace", systemImage: "lock")
                authPill("No Data Leaves This Mac", systemImage: "checkmark.shield")
            }
        }
        .padding(ZPTheme.spacing18)
        .zpSurface(.muted)
    }

    private var footerActions: some View {
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
            .accessibilityIdentifier(mode == .openExisting ? "openVault.chooseFolderButton" : "replaceVault.chooseFolderButton")
        }
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
                localError = openVaultErrorMessage(for: error)
            }
        }
    }

    private func openVaultErrorMessage(for error: Error) -> String {
        if let bridgeError = error as? ZPBridgeError {
            switch bridgeError.code {
            case .notFound:
                return "ZeroPass couldn’t find a vault in that folder. Choose a valid vault folder and try again."
            case .busy:
                return "That vault is busy right now. Wait a moment and try again."
            default:
                return mode == .openExisting
                    ? "ZeroPass couldn’t open that vault. Check the folder and try again."
                    : "ZeroPass couldn’t switch to that vault. Your current vault is still open."
            }
        }

        return mode == .openExisting
            ? "ZeroPass couldn’t open that vault right now. Please try again."
            : "ZeroPass couldn’t switch vaults right now. Please try again."
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

enum VaultFolderPicker {
    private enum UITestDirectoryOverride {
        static let environmentKey = "UITEST_PICK_DIRECTORY_PATH"

        @MainActor
        static func resolvedURL(canCreateDirectories: Bool) -> URL? {
            guard VaultClient.isRunningUITests else { return nil }

            let rawPath = ProcessInfo.processInfo.environment[environmentKey]?
                .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            guard !rawPath.isEmpty else { return nil }

            let url = URL(fileURLWithPath: rawPath, isDirectory: true)
            let fileManager = FileManager.default

            if canCreateDirectories {
                try? fileManager.createDirectory(at: url, withIntermediateDirectories: true)
                return url
            }

            var isDirectory: ObjCBool = false
            guard fileManager.fileExists(atPath: url.path, isDirectory: &isDirectory), isDirectory.boolValue else {
                return nil
            }

            return url
        }
    }

    @MainActor
    static func pickDirectory(canCreateDirectories: Bool) async -> URL? {
        if let overrideURL = UITestDirectoryOverride.resolvedURL(canCreateDirectories: canCreateDirectories) {
            return overrideURL
        }

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
