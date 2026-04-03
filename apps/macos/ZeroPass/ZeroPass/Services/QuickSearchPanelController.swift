import AppKit
import Combine
import SwiftUI

@MainActor
final class QuickSearchPanelController: ObservableObject {
    static let shared = QuickSearchPanelController()

    private var panel: QuickSearchPanel?

    func toggle(vault: VaultClient) {
        if let panel, panel.isVisible {
            close()
            return
        }
        show(vault: vault)
    }

    func show(vault: VaultClient) {
        let panel = ensurePanel(vault: vault)
        position(panel: panel)

        NSApp.activate(ignoringOtherApps: true)
        panel.makeKeyAndOrderFront(nil)
    }

    func close() {
        panel?.orderOut(nil)
    }

    private func ensurePanel(vault: VaultClient) -> QuickSearchPanel {
        if let panel { return panel }

        let panel = QuickSearchPanel()
        panel.isReleasedWhenClosed = false

        let root = QuickSearchView()
            .environmentObject(vault)
            .environmentObject(self)

        panel.contentView = NSHostingView(rootView: root)
        self.panel = panel
        return panel
    }

    private func position(panel: NSPanel) {
        let screen = NSScreen.main ?? NSScreen.screens.first
        guard let screen else { return }

        let size = NSSize(width: 580, height: 400)
        let frame = screen.visibleFrame

        let x = frame.midX - size.width / 2
        let y = frame.maxY - size.height - 200

        panel.setFrame(NSRect(origin: CGPoint(x: x, y: y), size: size), display: false)
    }
}
