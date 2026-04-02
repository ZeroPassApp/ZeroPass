import SwiftUI

extension VaultItemType {
    var symbolName: String {
        switch self {
        case .login: return "key.fill"
        case .apikey: return "wrench.and.screwdriver.fill"
        case .sshkey: return "terminal.fill"
        case .note: return "note.text"
        case .creditcard: return "creditcard.fill"
        case .identity: return "person.crop.circle.fill"
        case .passkey: return "person.badge.key.fill"
        case .custom: return "square.grid.2x2.fill"
        }
    }

    var color: Color {
        switch self {
        case .login: return .blue
        case .apikey: return .purple
        case .sshkey: return .green
        case .note: return .yellow
        case .creditcard: return .orange
        case .identity: return .teal
        case .passkey: return .indigo
        case .custom: return .gray
        }
    }
}
