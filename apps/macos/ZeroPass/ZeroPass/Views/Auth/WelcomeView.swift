import AppKit
import SwiftUI

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient
    @State private var activeAuthModal: VaultClient.AuthModal?

    var body: some View {
        ZStack {
            ZPTheme.authSceneBackground
                .ignoresSafeArea()

            backgroundMotifs

            VStack(spacing: 0) {
                Spacer(minLength: ZPTheme.spacing24)

                VStack(alignment: .center, spacing: ZPTheme.spacing24) {
                    brandHeader

                    if let errorText = vault.authFlowError, !errorText.isEmpty {
                        AuthMessageView(
                            text: errorText,
                            systemImage: "exclamationmark.triangle.fill",
                            tone: .error
                        )
                    }

                    actionButtons

                    if let recentVaultURL {
                        recentVaultSection(for: recentVaultURL)
                    }

                    assuranceRow
                }
                .frame(maxWidth: 460)
                .frame(maxWidth: .infinity)

                Spacer(minLength: ZPTheme.spacing24)

                footer
            }
            .padding(.horizontal, ZPTheme.spacing32)
            .padding(.vertical, ZPTheme.spacing32)
        }
        .background(WelcomeWindowCommandObserver(activeAuthModal: $activeAuthModal))
        .focusedSceneValue(\.welcomeAuthModal, $activeAuthModal)
        .sheet(item: $activeAuthModal) { modal in
            switch modal {
            case .createVault:
                CreateVaultView()
                    .environmentObject(vault)
            case .openVault:
                OpenVaultSheet(mode: .openExisting)
                    .environmentObject(vault)
            }
        }
        .onChange(of: vault.state) { _, newState in
            guard activeAuthModal != nil else { return }
            if case .noVault = newState {
                return
            }
            activeAuthModal = nil
        }
    }

    private var versionString: String {
        let v = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? ""
        let b = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? ""
        if v.isEmpty && b.isEmpty { return "" }
        if b.isEmpty { return v }
        if v.isEmpty { return b }
        return "\(v) (\(b))"
    }

    private var brandHeader: some View {
        VStack(alignment: .center, spacing: ZPTheme.spacing16) {
            Image(nsImage: NSApp.applicationIconImage)
                .resizable()
                .interpolation(.high)
                .frame(width: 74, height: 74)
                .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
                .shadow(color: ZPTheme.accentGlow, radius: 18, y: 10)
                .accessibilityHidden(true)

            VStack(alignment: .center, spacing: ZPTheme.spacing8) {
                Text("ZeroPass")
                    .font(.system(size: 44, weight: .bold, design: .rounded))
                    .foregroundStyle(ZPTheme.textPrimary)

                Text("Local-first. Zero-knowledge. Total privacy.")
                    .font(.title3.weight(.medium))
                    .foregroundStyle(ZPTheme.textSecondary)

                Text("Create or open a vault to get started.")
                    .font(.callout)
                    .foregroundStyle(ZPTheme.textTertiary)
            }
        }
    }

    private var actionButtons: some View {
        VStack(spacing: ZPTheme.spacing12) {
            Button {
                vault.authFlowError = nil
                activeAuthModal = .createVault
            } label: {
                Text("Create New Vault")
                    .frame(maxWidth: .infinity)
            }
            .controlSize(.large)
            .buttonStyle(.borderedProminent)
            .keyboardShortcut(.defaultAction)
            .accessibilityIdentifier("welcome.createVaultButton")

            Button {
                vault.authFlowError = nil
                activeAuthModal = .openVault
            } label: {
                Text("Open Existing Vault")
                    .frame(maxWidth: .infinity)
            }
            .controlSize(.large)
            .buttonStyle(.bordered)
            .accessibilityIdentifier("welcome.openVaultButton")
        }
        .frame(maxWidth: 360)
    }

    @ViewBuilder
    private func recentVaultSection(for url: URL) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing10) {
            Text("Recent Vaults")
                .font(.system(size: 11, weight: .semibold))
                .tracking(1.8)
                .foregroundStyle(ZPTheme.textTertiary)
                .textCase(.uppercase)

            Button {
                openRecentVault(url)
            } label: {
                HStack(spacing: ZPTheme.spacing12) {
                    ZStack {
                        RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                            .fill(ZPTheme.chipBackground)

                        Image(systemName: "externaldrive.fill")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(ZPTheme.textSecondary)
                    }
                    .frame(width: 40, height: 40)

                    VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                        HStack(spacing: ZPTheme.spacing8) {
                            Text(url.lastPathComponent)
                                .font(.subheadline.weight(.semibold))
                                .foregroundStyle(ZPTheme.textPrimary)
                                .lineLimit(1)

                            Text("RECENT")
                                .font(.system(size: 9, weight: .bold))
                                .foregroundStyle(ZPTheme.accent)
                                .padding(.horizontal, ZPTheme.spacing6)
                                .padding(.vertical, 3)
                                .background(ZPTheme.pillBackground, in: Capsule())
                                .overlay(
                                    Capsule()
                                        .stroke(ZPTheme.pillBorder, lineWidth: 1)
                                )
                        }

                        Text(recentVaultSubtitle(for: url))
                            .font(.caption)
                            .foregroundStyle(ZPTheme.textSecondary)
                            .lineLimit(1)
                    }

                    Spacer()

                    Image(systemName: "chevron.right")
                        .font(.caption.weight(.bold))
                        .foregroundStyle(ZPTheme.textTertiary)
                }
                .padding(.horizontal, ZPTheme.spacing16)
                .padding(.vertical, ZPTheme.spacing14)
                .frame(maxWidth: .infinity, alignment: .leading)
                .zpSurface(.elevated)
            }
            .buttonStyle(.plain)
        }
        .frame(maxWidth: 360, alignment: .leading)
    }

    private var assuranceRow: some View {
        HStack(spacing: ZPTheme.spacing10) {
            welcomePill("Encrypted locally", systemImage: "lock.shield")
            welcomePill("Offline-first", systemImage: "externaldrive")
            welcomePill("Zero-knowledge", systemImage: "eye.slash")
        }
        .fixedSize(horizontal: false, vertical: true)
    }

    private var footer: some View {
        HStack(spacing: ZPTheme.spacing12) {
            footerItem(versionString.isEmpty ? "ZeroPass" : "ZeroPass \(versionString)")
            footerBullet
            footerItem("Secured offline")
            footerBullet
            footerItem("Documentation")
            footerBullet
            footerItem("Support")
            footerBullet
            footerItem("Release Notes")
        }
        .frame(maxWidth: .infinity, alignment: .center)
        .lineLimit(1)
        .minimumScaleFactor(0.8)
    }

    private var recentVaultURL: URL? {
        BookmarkStore().loadVaultURL()
    }

    @ViewBuilder
    private var backgroundMotifs: some View {
        ZStack {
            Image(systemName: "lock.fill")
                .font(.system(size: 94, weight: .regular))
                .foregroundStyle(ZPTheme.textMuted.opacity(0.10))
                .offset(x: -300, y: 170)
                .accessibilityHidden(true)

            Image(systemName: "shield.fill")
                .font(.system(size: 72, weight: .regular))
                .foregroundStyle(ZPTheme.textMuted.opacity(0.08))
                .offset(x: 310, y: -185)
                .accessibilityHidden(true)

            RoundedRectangle(cornerRadius: 220, style: .continuous)
                .fill(ZPTheme.accentGlow)
                .frame(width: 480, height: 320)
                .blur(radius: 90)
                .offset(y: -180)
                .accessibilityHidden(true)
        }
    }

    @ViewBuilder
    private func welcomePill(_ title: String, systemImage: String) -> some View {
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

    @ViewBuilder
    private func footerItem(_ title: String) -> some View {
        Text(title.uppercased())
            .font(.system(size: 10, weight: .medium))
            .tracking(1.2)
            .foregroundStyle(ZPTheme.textTertiary)
    }

    private var footerBullet: some View {
        Text("•")
            .font(.caption)
            .foregroundStyle(ZPTheme.textMuted)
    }

    private func recentVaultSubtitle(for url: URL) -> String {
        var components: [String] = []

        if let values = try? url.resourceValues(forKeys: [.contentModificationDateKey]),
           let modifiedAt = values.contentModificationDate {
            components.append("Updated \(modifiedAt.formatted(date: .abbreviated, time: .omitted))")
        }

        if let values = try? url.resourceValues(forKeys: [.fileSizeKey]),
           let fileSize = values.fileSize,
           fileSize > 0 {
            components.append(ByteCountFormatter.string(fromByteCount: Int64(fileSize), countStyle: .file))
        }

        if components.isEmpty {
            return "Ready on this Mac"
        }

        return components.joined(separator: " • ")
    }

    private func openRecentVault(_ url: URL) {
        vault.authFlowError = nil

        Task {
            do {
                try await vault.openVault(url)
            } catch {
                vault.authFlowError = error.localizedDescription
            }
        }
    }
}

private struct WelcomeWindowCommandObserver: NSViewRepresentable {
    @Binding var activeAuthModal: VaultClient.AuthModal?

    func makeCoordinator() -> Coordinator {
        Coordinator(activeAuthModal: $activeAuthModal)
    }

    func makeNSView(context: Context) -> NSView {
        NSView()
    }

    func updateNSView(_ nsView: NSView, context: Context) {
        DispatchQueue.main.async {
            guard let window = nsView.window else {
                return
            }

            context.coordinator.attach(to: window)
        }
    }

    final class Coordinator {
        private let activeAuthModal: Binding<VaultClient.AuthModal?>
        private weak var observedWindow: NSWindow?
        private var modalObserver: NSObjectProtocol?

        init(activeAuthModal: Binding<VaultClient.AuthModal?>) {
            self.activeAuthModal = activeAuthModal
        }

        deinit {
            if let modalObserver {
                NotificationCenter.default.removeObserver(modalObserver)
            }
        }

        func attach(to window: NSWindow) {
            guard observedWindow !== window else {
                return
            }

            if let modalObserver {
                NotificationCenter.default.removeObserver(modalObserver)
            }

            observedWindow = window
            modalObserver = NotificationCenter.default.addObserver(
                forName: .welcomeAuthModalRequest,
                object: window,
                queue: .main
            ) { [weak self] notification in
                guard
                    let rawValue = notification.userInfo?[WelcomeAuthModalRequest.notificationUserInfoKey] as? String,
                    let modal = VaultClient.AuthModal(rawValue: rawValue)
                else {
                    return
                }

                self?.activeAuthModal.wrappedValue = modal
            }
        }
    }

}

private struct WelcomeAuthModalFocusedKey: FocusedValueKey {
    typealias Value = Binding<VaultClient.AuthModal?>
}

enum WelcomeAuthModalRequest {
    static let notificationUserInfoKey = "modal"
}

extension Notification.Name {
    static let welcomeAuthModalRequest = Notification.Name("ZeroPass.WelcomeAuthModalRequest")
}

extension FocusedValues {
    var welcomeAuthModal: Binding<VaultClient.AuthModal?>? {
        get { self[WelcomeAuthModalFocusedKey.self] }
        set { self[WelcomeAuthModalFocusedKey.self] = newValue }
    }
}
