import AppKit
import SwiftUI

struct GeneralSettingsView: View {
    @EnvironmentObject var vault: VaultClient

    @ObservedObject private var launchAtLogin = LaunchAtLoginService.shared

    @AppStorage("quickSearchKeyCode") private var quickSearchKeyCode: Int = Int(HotkeyService.Hotkey.quickSearchDefault.keyCode)
    @AppStorage("quickSearchModifiers") private var quickSearchModifiers: Int = Int(HotkeyService.Hotkey.quickSearchDefault.modifiers)

    @AppStorage("showMenuBar") private var showMenuBar: Bool = true
    @AppStorage("appearanceMode") private var appearanceMode: String = "system"
    @AppStorage("notificationsEnabled") private var notificationsEnabled: Bool = true

    @State private var isRecordingHotkey = false

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            settingsIntro(
                title: "General",
                message: "Manage vault location, startup behavior, quick access, and appearance defaults."
            )

            Form {
                Section("Vault") {
                    HStack(alignment: .firstTextBaseline) {
                        Text("Vault location")
                        Spacer()
                        Text(vault.vaultPathDisplay)
                            .font(.system(.body, design: .monospaced))
                            .foregroundStyle(ZPTheme.textSecondary)
                            .lineLimit(1)
                            .truncationMode(.middle)

                        Button("Change…") { chooseVaultFolder() }
                            .disabled(vault.isUnlocked)
                            .help(vault.isUnlocked ? "Lock the vault before switching." : "")
                    }
                }

                Section("Startup") {
                    Toggle("Launch at login", isOn: Binding(
                        get: { launchAtLogin.isEnabled },
                        set: { launchAtLogin.setEnabled($0) }
                    ))

                    if let err = launchAtLogin.lastError {
                        Text(err)
                            .foregroundStyle(ZPTheme.destructive)
                            .textSelection(.enabled)
                    }
                }

                Section("Quick Access") {
                    Toggle("Show in menu bar", isOn: $showMenuBar)

                    HStack {
                        Text("Quick Search hotkey")
                        Spacer()
                        Text(HotkeyRecorderView.displayString(keyCode: quickSearchKeyCode, modifiers: quickSearchModifiers))
                            .font(.system(.body, design: .monospaced))
                            .foregroundStyle(ZPTheme.textSecondary)
                        Button("Record…") { isRecordingHotkey = true }
                    }
                }

                Section("Appearance") {
                    Picker("Appearance", selection: $appearanceMode) {
                        Text("System").tag("system")
                        Text("Light").tag("light")
                        Text("Dark").tag("dark")
                    }
                    .pickerStyle(.segmented)
                }

                Section("Notifications") {
                    Toggle("Enable notifications", isOn: $notificationsEnabled)
                }
            }
            .formStyle(.grouped)
        }
        .background(ZPTheme.workspaceBackground)
        .sheet(isPresented: $isRecordingHotkey) {
            HotkeyRecorderView(keyCode: $quickSearchKeyCode, modifiers: $quickSearchModifiers)
                .frame(width: 420, height: 180)
        }
        .onChange(of: quickSearchKeyCode) { _, _ in
            HotkeyService.shared.register(.init(keyCode: UInt32(quickSearchKeyCode), modifiers: UInt32(quickSearchModifiers)))
        }
        .onChange(of: quickSearchModifiers) { _, _ in
            HotkeyService.shared.register(.init(keyCode: UInt32(quickSearchKeyCode), modifiers: UInt32(quickSearchModifiers)))
        }
    }

    @ViewBuilder
    private func settingsIntro(title: String, message: String) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing6) {
            Text(title)
                .font(.headline)
                .foregroundStyle(ZPTheme.textPrimary)

            Text(message)
                .font(.callout)
                .foregroundStyle(ZPTheme.textSecondary)
        }
        .padding(ZPTheme.spacing18)
        .zpSurface(.muted)
    }

    private func chooseVaultFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = false

        if panel.runModal() == .OK, let url = panel.url {
            Task {
                do {
                    if vault.hasVault {
                        try await vault.replaceVault(with: url)
                    } else {
                        try await vault.openVault(url)
                    }
                } catch {
                    vault.lastError = error.localizedDescription
                }
            }
        }
    }
}
