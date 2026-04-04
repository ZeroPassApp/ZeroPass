import AppKit
import SwiftUI

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient
    @State private var activeAuthModal: VaultClient.AuthModal?

    var body: some View {
        AuthSceneScaffold(
            title: "ZeroPass",
            subtitle: "Create or open a vault to get started.",
            symbolName: "lock.shield.fill"
        ) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
                if let errorText = vault.authFlowError, !errorText.isEmpty {
                    AuthMessageView(
                        text: errorText,
                        systemImage: "exclamationmark.triangle.fill",
                        tone: .error
                    )
                }

                VStack(spacing: ZPTheme.spacing16) {
                    Button {
                        vault.authFlowError = nil
                        activeAuthModal = .createVault
                    } label: {
                        Text("Create Vault…")
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
                        Text("Open Vault…")
                            .frame(maxWidth: .infinity)
                    }
                    .controlSize(.large)
                    .buttonStyle(.bordered)
                    .accessibilityIdentifier("welcome.openVaultButton")
                }
            }
        } footer: {
            if !versionString.isEmpty {
                Text("Version \(versionString)")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.textSecondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
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
