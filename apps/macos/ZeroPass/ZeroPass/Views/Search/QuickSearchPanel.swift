import AppKit

final class QuickSearchPanel: NSPanel {
    init() {
        super.init(
            contentRect: NSRect(x: 0, y: 0, width: 580, height: 400),
            styleMask: [.titled, .fullSizeContentView, .utilityWindow],
            backing: .buffered,
            defer: false
        )

        titleVisibility = .hidden
        titlebarAppearsTransparent = true
        isMovableByWindowBackground = true

        level = .floating
        collectionBehavior = [.transient, .moveToActiveSpace]
        isOpaque = false
        backgroundColor = .clear

        hidesOnDeactivate = true
    }

    override var canBecomeKey: Bool { true }
    override var canBecomeMain: Bool { true }
}
