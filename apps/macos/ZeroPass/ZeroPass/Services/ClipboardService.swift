import AppKit
import Foundation

@MainActor
final class ClipboardService {
    static let shared = ClipboardService()

    private var clearTask: Task<Void, Never>?

    func copySensitive(_ value: String, clearAfterSeconds: Int) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(value, forType: .string)

        clearTask?.cancel()
        guard clearAfterSeconds > 0 else { return }

        clearTask = Task { [value] in
            try? await Task.sleep(nanoseconds: UInt64(clearAfterSeconds) * 1_000_000_000)
            guard !Task.isCancelled else { return }

            if NSPasteboard.general.string(forType: .string) == value {
                NSPasteboard.general.clearContents()
            }
        }
    }

    func clear() {
        clearTask?.cancel()
        clearTask = nil
        NSPasteboard.general.clearContents()
    }
}
