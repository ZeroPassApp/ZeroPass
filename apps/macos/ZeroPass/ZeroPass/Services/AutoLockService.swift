import AppKit
import Foundation

@MainActor
final class AutoLockService {
    var onLock: (() -> Void)?

    private var isVaultUnlocked = false

    private var timeoutSeconds: Int = 0
    private var lockOnSleep: Bool = true
    private var lockOnScreenSleep: Bool = true

    private var pendingTask: Task<Void, Never>?
    private var observers: [NSObjectProtocol] = []

    func configure(timeoutSeconds: Int, lockOnSleep: Bool, lockOnScreenSleep: Bool) {
        self.timeoutSeconds = max(0, timeoutSeconds)
        self.lockOnSleep = lockOnSleep
        self.lockOnScreenSleep = lockOnScreenSleep
        reschedule()
    }

    func setVaultUnlocked(_ unlocked: Bool) {
        isVaultUnlocked = unlocked
        if unlocked {
            start()
            reschedule()
        } else {
            pendingTask?.cancel()
            pendingTask = nil
        }
    }

    func recordActivity() {
        guard isVaultUnlocked else { return }
        reschedule()
    }

    func stop() {
        pendingTask?.cancel()
        pendingTask = nil

        for o in observers {
            NSWorkspace.shared.notificationCenter.removeObserver(o)
        }
        observers.removeAll()
    }

    private func start() {
        guard observers.isEmpty else { return }

        let nc = NSWorkspace.shared.notificationCenter
        observers.append(
            nc.addObserver(forName: NSWorkspace.willSleepNotification, object: nil, queue: .main) { [weak self] _ in
                guard let self, self.lockOnSleep, self.isVaultUnlocked else { return }
                self.onLock?()
                NotificationService.shared.notify(title: "ZeroPass", body: "Vault locked (sleep)")
            }
        )
        observers.append(
            nc.addObserver(forName: NSWorkspace.screensDidSleepNotification, object: nil, queue: .main) { [weak self] _ in
                guard let self, self.lockOnScreenSleep, self.isVaultUnlocked else { return }
                self.onLock?()
                NotificationService.shared.notify(title: "ZeroPass", body: "Vault locked (screen sleep)")
            }
        )
    }

    private func reschedule() {
        pendingTask?.cancel()
        pendingTask = nil

        guard isVaultUnlocked else { return }
        guard timeoutSeconds > 0 else { return }

        pendingTask = Task { [timeoutSeconds] in
            try? await Task.sleep(nanoseconds: UInt64(timeoutSeconds) * 1_000_000_000)
            guard !Task.isCancelled else { return }
            await MainActor.run {
                self.onLock?()
                NotificationService.shared.notify(title: "ZeroPass", body: "Vault auto-locked")
            }
        }
    }
}
