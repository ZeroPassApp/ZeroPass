import SwiftUI

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var showingCreate = false
    @State private var showingOpen = false

    var body: some View {
        AuthSceneScaffold(
            title: "ZeroPass",
            subtitle: "Local-first, zero-knowledge password manager",
            detail: "Create a new vault or open an existing one to continue.",
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

                VStack(spacing: ZPTheme.spacing12) {
                    Button {
                        vault.authFlowError = nil
                        showingCreate = true
                    } label: {
                        Label("Create New Vault", systemImage: "plus.circle.fill")
                            .frame(maxWidth: .infinity)
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut("n", modifiers: [.command])

                    Button {
                        vault.authFlowError = nil
                        showingOpen = true
                    } label: {
                        Label("Open Existing Vault", systemImage: "folder")
                            .frame(maxWidth: .infinity)
                    }
                    .controlSize(.large)
                    .buttonStyle(.bordered)
                    .keyboardShortcut("o", modifiers: [.command])
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
        .sheet(isPresented: $showingCreate) {
            CreateVaultView()
                .environmentObject(vault)
        }
        .sheet(isPresented: $showingOpen) {
            OpenVaultSheet(mode: .openExisting)
                .environmentObject(vault)
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
