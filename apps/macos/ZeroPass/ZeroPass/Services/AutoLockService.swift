import AppKit
import Foundation

nonisolated final class AutoLockService {
    var onLock: (() -> Void)?

    private var isVaultUnlocked = false

    private var timeoutSeconds: Int = 0
    private var lockOnSleep: Bool = true
    private var lockOnScreenSleep: Bool = true

    private var pendingWorkItem: DispatchWorkItem?
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
            pendingWorkItem?.cancel()
            pendingWorkItem = nil
        }
    }

    func recordActivity() {
        guard isVaultUnlocked else { return }
        reschedule()
    }

    func stop() {
        pendingWorkItem?.cancel()
        pendingWorkItem = nil

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
                self.notify(body: "Vault locked (sleep)")
            }
        )
        observers.append(
            nc.addObserver(forName: NSWorkspace.screensDidSleepNotification, object: nil, queue: .main) { [weak self] _ in
                guard let self, self.lockOnScreenSleep, self.isVaultUnlocked else { return }
                self.onLock?()
                self.notify(body: "Vault locked (screen sleep)")
            }
        )
    }

    private func reschedule() {
        pendingWorkItem?.cancel()
        pendingWorkItem = nil

        guard isVaultUnlocked else { return }
        guard timeoutSeconds > 0 else { return }

        let workItem = DispatchWorkItem { [weak self] in
            guard let self, self.isVaultUnlocked else { return }
            self.onLock?()
            self.notify(body: "Vault auto-locked")
        }
        pendingWorkItem = workItem
        DispatchQueue.main.asyncAfter(deadline: .now() + .seconds(timeoutSeconds), execute: workItem)
    }

    private func notify(body: String) {
        Task { @MainActor in
            NotificationService.shared.notify(title: "ZeroPass", body: body)
        }
    }
}
