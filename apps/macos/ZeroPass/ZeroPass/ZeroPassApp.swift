//
//  ZeroPassApp.swift
//  ZeroPass
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

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
        .commands {
            CommandGroup(after: .newItem) {
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
}
