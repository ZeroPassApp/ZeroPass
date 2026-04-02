import Foundation
import SwiftUI
import UserNotifications

@MainActor
final class NotificationService {
    static let shared = NotificationService()

    @AppStorage("notificationsEnabled") var notificationsEnabled: Bool = true

    private init() {}

    func requestAuthorizationIfNeeded() {
        guard notificationsEnabled else { return }
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound]) { _, _ in }
    }

    func notify(title: String, body: String) {
        guard notificationsEnabled else { return }

        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body

        let req = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(req)
    }
}
