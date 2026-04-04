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
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var vault = VaultClient()
    private let quickSearch = QuickSearchPanelController.shared

    @AppStorage("quickSearchKeyCode") private var quickSearchKeyCode: Int = Int(HotkeyService.Hotkey.quickSearchDefault.keyCode)
    @AppStorage("quickSearchModifiers") private var quickSearchModifiers: Int = Int(HotkeyService.Hotkey.quickSearchDefault.modifiers)
    @AppStorage("showMenuBar") private var showMenuBar: Bool = true
    @AppStorage("appearanceMode") private var appearanceMode: String = "system"

    private var isUITesting: Bool {
        ProcessInfo.processInfo.arguments.contains("UITEST_MODE")
    }

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
                    guard !isUITesting else {
                        return
                    }

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
                Button(vault.hasVault ? "Choose Different Vault…" : "Open Vault…") {
                    if vault.hasVault {
                        chooseVaultFolder(replacingCurrent: true)
                    } else {
                        vault.presentOpenVaultSheet()
                        appDelegate.revealMainWindowIfNeeded()
                    }
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

@MainActor
private final class AppDelegate: NSObject, NSApplicationDelegate {
    func revealMainWindowIfNeeded() {
        Task { @MainActor in
            await ensureMainWindowVisible()
        }
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        revealMainWindowIfNeeded()
    }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        if bringMainWindowToFrontIfAvailable() {
            return true
        }

        openNewWindowFromMenuIfAvailable()
        return true
    }

    private func ensureMainWindowVisible() async {
        NSApp.activate(ignoringOtherApps: true)

        if bringMainWindowToFrontIfAvailable() {
            return
        }

        for _ in 0..<10 {
            try? await Task.sleep(nanoseconds: 100_000_000)
            if bringMainWindowToFrontIfAvailable() {
                return
            }
        }

        openNewWindowFromMenuIfAvailable()

        for _ in 0..<10 {
            try? await Task.sleep(nanoseconds: 100_000_000)
            if bringMainWindowToFrontIfAvailable() {
                return
            }
        }
    }

    private func bringMainWindowToFrontIfAvailable() -> Bool {
        guard let window = mainWindow else {
            return false
        }

        NSApp.activate(ignoringOtherApps: true)
        if window.isMiniaturized {
            window.deminiaturize(nil)
        }
        window.makeKeyAndOrderFront(nil)
        return true
    }

    private var mainWindow: NSWindow? {
        NSApp.windows.first { window in
            guard window.canBecomeKey else {
                return false
            }

            if let identifier = window.identifier?.rawValue, identifier.contains("AppWindow") {
                return true
            }

            let appName = Bundle.main.object(forInfoDictionaryKey: "CFBundleDisplayName") as? String
                ?? Bundle.main.object(forInfoDictionaryKey: "CFBundleName") as? String
                ?? "ZeroPass"
            return window.title == appName
        }
    }

    private func openNewWindowFromMenuIfAvailable() {
        guard
            let newWindowItem = newWindowMenuItem(in: NSApp.mainMenu),
            let action = newWindowItem.action
        else {
            return
        }

        _ = NSApp.sendAction(action, to: newWindowItem.target, from: newWindowItem)
    }

    private func newWindowMenuItem(in menu: NSMenu?) -> NSMenuItem? {
        guard let menu else {
            return nil
        }

        for item in menu.items {
            if item.submenu == nil,
               item.keyEquivalent.lowercased() == "n",
               item.keyEquivalentModifierMask.contains(.command) {
                return item
            }

            if let nestedMatch = newWindowMenuItem(in: item.submenu) {
                return nestedMatch
            }
        }

        return nil
    }
}
