//
//  ZeroPassApp.swift
//  ZeroPass
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

import AppKit
import SwiftUI

@main
struct ZeroPassApp: App {
    @StateObject private var vault = VaultClient()
    private let quickSearch = QuickSearchPanelController.shared

    @AppStorage("quickSearchKeyCode") private var quickSearchKeyCode: Int = Int(HotkeyService.Hotkey.quickSearchDefault.keyCode)
    @AppStorage("quickSearchModifiers") private var quickSearchModifiers: Int = Int(HotkeyService.Hotkey.quickSearchDefault.modifiers)
    @AppStorage("showMenuBar") private var showMenuBar: Bool = true
    @AppStorage("appearanceMode") private var appearanceMode: String = "system"


    private var preferredColorScheme: ColorScheme? {
        switch appearanceMode {
        case "light": return .light
        case "dark": return .dark
        default: return nil
        }
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(vault)
                .environmentObject(quickSearch)
                .preferredColorScheme(preferredColorScheme)
                .onAppear {
                    NotificationService.shared.requestAuthorizationIfNeeded()

                    HotkeyService.shared.onHotkey = {
                        quickSearch.toggle(vault: vault)
                    }
                    HotkeyService.shared.register(
                        .init(
                            keyCode: UInt32(quickSearchKeyCode),
                            modifiers: UInt32(quickSearchModifiers)
                        )
                    )
                }
        }
        .defaultSize(width: 620, height: 480)
        .commands {
            CommandGroup(after: .newItem) {
                Button(vault.hasVault ? "Choose Different Vault…" : "Open Existing Vault…") {
                    chooseVaultFolder(replacingCurrent: vault.hasVault)
                }
                .keyboardShortcut("o", modifiers: [.command])

                Button("Close Vault") {
                    Task { await vault.closeVault() }
                }
                .disabled(!vault.hasVault)

                Divider()

                Button("Quick Search") {
                    quickSearch.toggle(vault: vault)
                }
                .keyboardShortcut("k", modifiers: [.command])

                Button("Lock Vault") {
                    Task { await vault.lock() }
                }
                .keyboardShortcut("l", modifiers: [.command, .shift])
                .disabled(!vault.isUnlocked)

                Button("Refresh") {
                    Task { await vault.refresh() }
                }
                .keyboardShortcut("r", modifiers: [.command])
                .disabled(!vault.isUnlocked)
            }
        }

        MenuBarExtra(
            "ZeroPass",
            systemImage: vault.isUnlocked ? "lock.open.fill" : "lock.fill",
            isInserted: $showMenuBar
        ) {
            MenuBarView()
                .environmentObject(vault)
                .environmentObject(quickSearch)
        }
        .menuBarExtraStyle(.window)

        Settings {
            SettingsView()
                .environmentObject(vault)
                .environmentObject(quickSearch)
        }
    }

    private func chooseVaultFolder(replacingCurrent: Bool) {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.canCreateDirectories = false

        guard panel.runModal() == .OK, let url = panel.url else {
            return
        }

        vault.lastError = nil
        if !replacingCurrent {
            vault.authFlowError = nil
        }

        Task {
            do {
                if replacingCurrent {
                    try await vault.replaceVault(with: url)
                } else {
                    try await vault.openVault(url)
                }
            } catch {
                if replacingCurrent {
                    vault.lastError = error.localizedDescription
                } else {
                    vault.authFlowError = error.localizedDescription
                }
            }
        }
    }
}
