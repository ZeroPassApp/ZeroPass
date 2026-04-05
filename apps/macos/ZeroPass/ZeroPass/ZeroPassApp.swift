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
    @FocusedValue(\.welcomeAuthModal) private var focusedWelcomeAuthModal
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
        .defaultSize(width: 900, height: 620)
        .commands {
            CommandGroup(after: .newItem) {
                Button(vault.hasVault ? "Choose Different Vault…" : "Open Vault…") {
                    if vault.hasVault {
                        chooseReplacementVaultFolder()
                    } else {
                        presentWelcomeAuthModal(.openVault)
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

    private func presentWelcomeAuthModal(_ modal: VaultClient.AuthModal) {
        vault.authFlowError = nil

        if presentFocusedWelcomeAuthModal(modal) {
            return
        }

        Task { @MainActor in
            if let window = await appDelegate.waitUntilMainWindowIsVisible() {
                for _ in 0..<10 {
                    if presentFocusedWelcomeAuthModal(modal) {
                        return
                    }

                    NotificationCenter.default.post(
                        name: .welcomeAuthModalRequest,
                        object: window,
                        userInfo: [WelcomeAuthModalRequest.notificationUserInfoKey: modal.rawValue]
                    )

                    try? await Task.sleep(nanoseconds: 100_000_000)
                }
            }
        }
    }

    private func presentFocusedWelcomeAuthModal(_ modal: VaultClient.AuthModal) -> Bool {
        guard let focusedWelcomeAuthModal else {
            return false
        }

        focusedWelcomeAuthModal.wrappedValue = modal
        return true
    }

    private func chooseReplacementVaultFolder() {
        Task { @MainActor in
            guard let url = await VaultFolderPicker.pickDirectory(canCreateDirectories: false) else {
                return
            }

            vault.lastError = nil

            do {
                try await vault.replaceVault(with: url)
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}

@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate {
    func revealMainWindowIfNeeded() {
        Task { @MainActor in
            _ = await waitUntilMainWindowIsVisible()
        }
    }

    func waitUntilMainWindowIsVisible() async -> NSWindow? {
        await ensureMainWindowVisible()
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        revealMainWindowIfNeeded()
    }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        if bringMainWindowToFrontIfAvailable() != nil {
            return true
        }

        openNewWindowFromMenuIfAvailable()
        return true
    }

    private func ensureMainWindowVisible() async -> NSWindow? {
        NSApp.activate(ignoringOtherApps: true)

        if let window = bringMainWindowToFrontIfAvailable() {
            return window
        }

        for _ in 0..<10 {
            try? await Task.sleep(nanoseconds: 100_000_000)
            if let window = bringMainWindowToFrontIfAvailable() {
                return window
            }
        }

        openNewWindowFromMenuIfAvailable()

        for _ in 0..<10 {
            try? await Task.sleep(nanoseconds: 100_000_000)
            if let window = bringMainWindowToFrontIfAvailable() {
                return window
            }
        }

        return nil
    }

    private func bringMainWindowToFrontIfAvailable() -> NSWindow? {
        guard let window = mainWindow else {
            return nil
        }

        NSApp.activate(ignoringOtherApps: true)
        if window.isMiniaturized {
            window.deminiaturize(nil)
        }
        window.makeKeyAndOrderFront(nil)
        return window
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
