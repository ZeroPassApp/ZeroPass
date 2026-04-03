import Foundation

enum VaultItemType: String, Codable, CaseIterable, Identifiable {
    case login
    case apikey
    case sshkey
    case note
    case creditcard
    case identity
    case passkey
    case custom

    var id: String { rawValue }

    var displayName: String {
        switch self {
        case .login: return "Login"
        case .apikey: return "API Key"
        case .sshkey: return "SSH Key"
        case .note: return "Secure Note"
        case .creditcard: return "Credit Card"
        case .identity: return "Identity"
        case .passkey: return "Passkey"
        case .custom: return "Custom"
        }
    }
}
