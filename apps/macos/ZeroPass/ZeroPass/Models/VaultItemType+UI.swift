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
        case .login: return ZPTheme.dynamicColor(light: "#2D6DE6", dark: "#67B0FF")
        case .apikey: return ZPTheme.dynamicColor(light: "#7A52D9", dark: "#B391FF")
        case .sshkey: return ZPTheme.dynamicColor(light: "#0E9D73", dark: "#4FD6A3")
        case .note: return ZPTheme.dynamicColor(light: "#A98100", dark: "#E1BF5F")
        case .creditcard: return ZPTheme.dynamicColor(light: "#C46A24", dark: "#F1A565")
        case .identity: return ZPTheme.dynamicColor(light: "#16838E", dark: "#5DD0D8")
        case .passkey: return ZPTheme.dynamicColor(light: "#5664DB", dark: "#8EA1FF")
        case .custom: return ZPTheme.dynamicColor(light: "#74808C", dark: "#8E9AA5")
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
