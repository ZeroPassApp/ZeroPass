import AppKit
import SwiftUI

extension View {
    func authWindowLayout(for state: VaultClient.State) -> some View {
        modifier(AuthWindowLayoutModifier(state: state))
    }
}

struct AuthWindowLayoutModifier: ViewModifier {
    let state: VaultClient.State

    func body(content: Content) -> some View {
        content.background(WindowLayoutObserver(layout: .init(state: state)))
    }
}

private struct WindowLayoutObserver: NSViewRepresentable {
    let layout: AuthWindowLayout

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeNSView(context: Context) -> NSView {
        NSView()
    }

    func updateNSView(_ nsView: NSView, context: Context) {
        DispatchQueue.main.async {
            guard let window = nsView.window else { return }
            context.coordinator.attach(to: window)
            context.coordinator.currentKind = layout.kind

            layout.applyAppearance(to: window)

            guard context.coordinator.lastLayout != layout else { return }
            context.coordinator.lastLayout = layout

            layout.apply(to: window, rememberedUnlockedSize: context.coordinator.lastUnlockedSize)

            if layout.kind == .unlocked {
                context.coordinator.lastUnlockedSize = window.contentRect(forFrameRect: window.frame).size
            }
        }
    }

    final class Coordinator {
        var lastLayout: AuthWindowLayout?
        var currentKind: AuthWindowLayout.Kind = .noVault
        var lastUnlockedSize: CGSize?

        private weak var observedWindow: NSWindow?
        private var resizeObserver: NSObjectProtocol?

        deinit {
            if let resizeObserver {
                NotificationCenter.default.removeObserver(resizeObserver)
            }
        }

        func attach(to window: NSWindow) {
            guard observedWindow !== window else { return }

            if let resizeObserver {
                NotificationCenter.default.removeObserver(resizeObserver)
            }

            observedWindow = window
            resizeObserver = NotificationCenter.default.addObserver(
                forName: NSWindow.didResizeNotification,
                object: window,
                queue: .main
            ) { [weak self, weak window] _ in
                guard let self, let window else { return }
                guard self.currentKind == .unlocked else { return }
                self.lastUnlockedSize = window.contentRect(forFrameRect: window.frame).size
            }
        }
    }
}

private struct AuthWindowLayout: Equatable {
    enum Kind: Equatable {
        case noVault
        case locked
        case showingRecovery
        case unlocked
    }

    private static let authWindowMask: NSWindow.StyleMask = [.borderless]
    private static let unlockedWindowMask: NSWindow.StyleMask = [
        .titled,
        .closable,
        .miniaturizable,
        .resizable,
        .fullSizeContentView
    ]

    let kind: Kind
    let minContentSize: CGSize
    let idealContentSize: CGSize

    private var usesFloatingAuthChrome: Bool {
        switch kind {
        case .noVault, .locked, .showingRecovery:
            return true
        case .unlocked:
            return false
        }
    }

    init(state: VaultClient.State) {
        switch state {
        case .noVault:
            kind = .noVault
            minContentSize = CGSize(width: 560, height: 420)
            idealContentSize = CGSize(width: 560, height: 420)
        case .locked:
            kind = .locked
            minContentSize = CGSize(width: 560, height: 460)
            idealContentSize = CGSize(width: 560, height: 460)
        case .showingRecovery:
            kind = .showingRecovery
            minContentSize = CGSize(width: 640, height: 520)
            idealContentSize = CGSize(width: 640, height: 520)
        case .unlocked:
            kind = .unlocked
            minContentSize = CGSize(width: 800, height: 500)
            idealContentSize = CGSize(width: 1000, height: 680)
        }
    }

    func apply(to window: NSWindow, rememberedUnlockedSize: CGSize?) {
        guard !window.styleMask.contains(.fullScreen) else { return }

        if usesFloatingAuthChrome {
            window.contentMinSize = idealContentSize
            window.contentMaxSize = idealContentSize
        } else {
            window.contentMinSize = minContentSize
            window.contentMaxSize = CGSize(width: 10_000, height: 10_000)
        }

        let currentSize = window.contentRect(forFrameRect: window.frame).size
        let targetSize: CGSize

        switch kind {
        case .unlocked:
            if let rememberedUnlockedSize {
                targetSize = CGSize(
                    width: max(rememberedUnlockedSize.width, minContentSize.width),
                    height: max(rememberedUnlockedSize.height, minContentSize.height)
                )
            } else {
                targetSize = idealContentSize
            }
        case .noVault, .locked, .showingRecovery:
            targetSize = idealContentSize
        }

        guard shouldResize(from: currentSize, to: targetSize) else { return }

        let frameSize = window.frameRect(forContentRect: CGRect(origin: .zero, size: targetSize)).size
        let currentFrame = window.frame
        let targetOrigin = CGPoint(
            x: currentFrame.midX - (frameSize.width / 2),
            y: currentFrame.midY - (frameSize.height / 2)
        )

        window.setFrame(CGRect(origin: targetOrigin, size: frameSize), display: true, animate: false)
    }

    func applyAppearance(to window: NSWindow) {
        if usesFloatingAuthChrome {
            if window.styleMask != Self.authWindowMask {
                window.styleMask = Self.authWindowMask
            }
            window.titleVisibility = .hidden
            window.titlebarAppearsTransparent = false
            window.isMovableByWindowBackground = true
            window.backgroundColor = .clear
            window.isOpaque = false
            window.hasShadow = false
        } else {
            if window.styleMask != Self.unlockedWindowMask {
                window.styleMask = Self.unlockedWindowMask
            }
            window.titleVisibility = .visible
            window.titlebarAppearsTransparent = false
            window.isMovableByWindowBackground = false
            window.backgroundColor = .windowBackgroundColor
            window.isOpaque = true
            window.hasShadow = true
        }

        window.standardWindowButton(.closeButton)?.isHidden = usesFloatingAuthChrome
        window.standardWindowButton(.miniaturizeButton)?.isHidden = usesFloatingAuthChrome
        window.standardWindowButton(.zoomButton)?.isHidden = usesFloatingAuthChrome

        window.invalidateShadow()
    }

    private func shouldResize(from currentSize: CGSize, to targetSize: CGSize) -> Bool {
        abs(currentSize.width - targetSize.width) > 1 || abs(currentSize.height - targetSize.height) > 1
    }
}
