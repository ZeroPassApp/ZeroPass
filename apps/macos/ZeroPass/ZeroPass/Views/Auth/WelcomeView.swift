import SwiftUI
import AppKit

struct WelcomeView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var showingCreate = false
    @State private var showingOpen = false
    @State private var iconAppeared = false

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            VStack(spacing: 24) {
                Image(systemName: "lock.shield.fill")
                    .font(.system(size: 64))
                    .foregroundStyle(.tint)
                    .symbolEffect(.appear, isActive: iconAppeared)
                    .accessibilityHidden(true)

                VStack(spacing: 6) {
                    Text("ZeroPass")
                        .font(.largeTitle)
                        .bold()

                    Text("Zero-knowledge. Local-first.")
                        .font(.title3)
                        .foregroundStyle(.secondary)
                }

                VStack(spacing: 10) {
                    Button {
                        showingCreate = true
                    } label: {
                        Label("Create New Vault", systemImage: "plus.circle")
                            .frame(maxWidth: 240)
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut("n", modifiers: [.command])

                    Button {
                        showingOpen = true
                    } label: {
                        Label("Open Existing Vault", systemImage: "folder")
                            .frame(maxWidth: 240)
                    }
                    .controlSize(.large)
                    .buttonStyle(.bordered)
                    .keyboardShortcut("o", modifiers: [.command])
                }
            }

            Spacer()

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .font(.callout)
                    .textSelection(.enabled)
                    .padding(.bottom, 8)
            }

            Text("Version \(versionString)")
                .font(.caption)
                .foregroundStyle(.tertiary)
                .padding(.bottom, 16)
        }
        .frame(minWidth: 560, minHeight: 420)
        .onAppear {
            iconAppeared = true
        }
        .sheet(isPresented: $showingCreate) {
            CreateVaultView()
                .environmentObject(vault)
        }
        .sheet(isPresented: $showingOpen) {
            OpenVaultView()
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

private struct OpenVaultView: View {
    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State private var isBusy = false

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "folder.badge.questionmark")
                .font(.system(size: 36))
                .foregroundStyle(.secondary)

            Text("Open Vault")
                .font(.title2)
                .bold()

            Text("Select the folder containing your ZeroPass vault.")
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            HStack(spacing: 12) {
                Button("Cancel") { dismiss() }
                    .keyboardShortcut(.cancelAction)
                    .disabled(isBusy)

                Button {
                    chooseFolder()
                } label: {
                    Label("Choose Folder…", systemImage: "folder")
                }
                .keyboardShortcut(.defaultAction)
                .disabled(isBusy)
            }
            .padding(.top, 4)
        }
        .padding(24)
        .frame(width: 400)
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
