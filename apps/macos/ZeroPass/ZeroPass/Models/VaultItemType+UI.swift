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

    var defaultFields: [String: String] {
        switch self {
        case .login: return ["username": "", "password": "", "url": ""]
        case .apikey: return ["api_key": "", "api_secret": "", "endpoint": ""]
        case .sshkey: return ["private_key": "", "public_key": "", "passphrase": ""]
        case .note: return [:]
        case .creditcard: return ["card_number": "", "expiry": "", "cvv": "", "cardholder": ""]
        case .identity: return ["full_name": "", "email": "", "phone": "", "address": ""]
        case .passkey: return ["credential_id": "", "relying_party": "", "user_handle": ""]
        case .custom: return [:]
        }
    }

    /// Field keys that contain sensitive data and should be masked by default
    var sensitiveFieldKeys: Set<String> {
        switch self {
        case .login: return ["password"]
        case .apikey: return ["api_secret", "api_key"]
        case .sshkey: return ["private_key", "passphrase"]
        case .note: return []
        case .creditcard: return ["cvv", "card_number"]
        case .identity: return []
        case .passkey: return ["credential_id"]
        case .custom: return []
        }
    }
}
