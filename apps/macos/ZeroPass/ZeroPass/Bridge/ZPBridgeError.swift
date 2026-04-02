import Foundation

enum ZPErrorCode: Int32 {
    case ok = 0
    case locked = 1
    case notFound = 2
    case authFailed = 3
    case internalError = 4
    case busy = 5
}

struct ZPBridgeError: LocalizedError {
    let code: ZPErrorCode
    let message: String

    var errorDescription: String? {
        switch code {
        case .ok:
            return nil
        case .locked:
            return message.isEmpty ? "Vault is locked" : message
        case .notFound:
            return message.isEmpty ? "Not found" : message
        case .authFailed:
            return message.isEmpty ? "Authentication failed" : message
        case .busy:
            return message.isEmpty ? "Vault is busy" : message
        case .internalError:
            return message.isEmpty ? "Internal error" : message
        }
    }
}
