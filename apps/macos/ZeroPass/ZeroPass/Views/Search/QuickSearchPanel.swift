import AppKit

final class QuickSearchPanel: NSPanel {
    init() {
        super.init(
            contentRect: NSRect(x: 0, y: 0, width: 640, height: 440),
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
        hasShadow = true

        hidesOnDeactivate = true
    }

    override var canBecomeKey: Bool { true }
    override var canBecomeMain: Bool { true }
}
