import SwiftUI

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient

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
                        vault.presentCreateVaultSheet()
                    } label: {
                        Text("Create Vault…")
                            .frame(maxWidth: .infinity)
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityIdentifier("welcome.createVaultButton")

                    Button {
                        vault.presentOpenVaultSheet()
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
        .sheet(item: activeAuthModalBinding) { modal in
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
            guard vault.activeAuthModal != nil else { return }
            if case .noVault = newState {
                return
            }
            vault.dismissAuthModal()
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

    private var activeAuthModalBinding: Binding<VaultClient.AuthModal?> {
        Binding(
            get: { vault.activeAuthModal },
            set: { vault.activeAuthModal = $0 }
        )
    }
}
