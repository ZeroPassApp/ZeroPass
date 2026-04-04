import Foundation
import LocalAuthentication
import Security

enum KeychainError: LocalizedError {
    case unexpectedStatus(OSStatus)
    case invalidData

    var errorDescription: String? {
        switch self {
        case .unexpectedStatus(let status):
            return "Keychain error (status: \(status))"
        case .invalidData:
            return "Keychain returned invalid data"
        }
    }
}

protocol KeychainStoring: AnyObject {
    func storeVaultKeyBase64(_ vaultKeyBase64: String, for vaultURL: URL, requireUserPresence: Bool) throws
    func loadVaultKeyBase64(for vaultURL: URL, context: LAContext) throws -> String?
    func deleteVaultKey(for vaultURL: URL) throws
    func storeSyncAPIKey(_ apiKey: String, for vaultURL: URL) throws
    func loadSyncAPIKey(for vaultURL: URL) throws -> String?
    func deleteSyncAPIKey(for vaultURL: URL) throws
}

nonisolated final class KeychainService: KeychainStoring {
    private let vaultKeyService = "com.zeropass.vaultkey"
    private let syncAPIKeyService = "com.zeropass.sync-api-key"

    private func account(for vaultURL: URL) -> String {
        vaultURL.path
    }

    func storeVaultKeyBase64(_ vaultKeyBase64: String, for vaultURL: URL, requireUserPresence: Bool) throws {
        let acct = account(for: vaultURL)
        let data = Data(vaultKeyBase64.utf8)

        let baseQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: vaultKeyService,
            kSecAttrAccount: acct,
        ]

        let attrs: [CFString: Any] = [kSecValueData: data]
        let updStatus = SecItemUpdate(baseQuery as CFDictionary, attrs as CFDictionary)
        if updStatus == errSecSuccess {
            return
        }
        if updStatus != errSecItemNotFound {
            throw KeychainError.unexpectedStatus(updStatus)
        }

        var addQuery = baseQuery
        if requireUserPresence {
            var err: Unmanaged<CFError>?
            if let ac = SecAccessControlCreateWithFlags(nil, kSecAttrAccessibleWhenUnlockedThisDeviceOnly, [.userPresence], &err) {
                addQuery[kSecAttrAccessControl] = ac
            }
        } else {
            addQuery[kSecAttrAccessible] = kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        }

        addQuery[kSecValueData] = data
        let addStatus = SecItemAdd(addQuery as CFDictionary, nil)
        if addStatus != errSecSuccess {
            throw KeychainError.unexpectedStatus(addStatus)
        }
    }

    func loadVaultKeyBase64(for vaultURL: URL, context: LAContext) throws -> String? {
        let acct = account(for: vaultURL)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: vaultKeyService,
            kSecAttrAccount: acct,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne,
            kSecUseAuthenticationContext: context,
        ]

        var out: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &out)
        if status == errSecItemNotFound {
            return nil
        }
        if status != errSecSuccess {
            throw KeychainError.unexpectedStatus(status)
        }
        guard let data = out as? Data else {
            throw KeychainError.invalidData
        }
        return String(data: data, encoding: .utf8)
    }

    func deleteVaultKey(for vaultURL: URL) throws {
        let acct = account(for: vaultURL)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: vaultKeyService,
            kSecAttrAccount: acct,
        ]

        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            throw KeychainError.unexpectedStatus(status)
        }
    }

    func storeSyncAPIKey(_ apiKey: String, for vaultURL: URL) throws {
        let acct = account(for: vaultURL)
        let data = Data(apiKey.utf8)

        let baseQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: syncAPIKeyService,
            kSecAttrAccount: acct,
        ]
        let attrs: [CFString: Any] = [kSecValueData: data]

        let updStatus = SecItemUpdate(baseQuery as CFDictionary, attrs as CFDictionary)
        if updStatus == errSecSuccess {
            return
        }
        if updStatus != errSecItemNotFound {
            throw KeychainError.unexpectedStatus(updStatus)
        }

        let addQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: syncAPIKeyService,
            kSecAttrAccount: acct,
            kSecAttrAccessible: kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            kSecValueData: data,
        ]
        let addStatus = SecItemAdd(addQuery as CFDictionary, nil)
        if addStatus != errSecSuccess {
            throw KeychainError.unexpectedStatus(addStatus)
        }
    }

    func loadSyncAPIKey(for vaultURL: URL) throws -> String? {
        let acct = account(for: vaultURL)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: syncAPIKeyService,
            kSecAttrAccount: acct,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne,
        ]

        var out: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &out)
        if status == errSecItemNotFound {
            return nil
        }
        if status != errSecSuccess {
            throw KeychainError.unexpectedStatus(status)
        }
        guard let data = out as? Data else {
            throw KeychainError.invalidData
        }
        return String(data: data, encoding: .utf8)
    }

    func deleteSyncAPIKey(for vaultURL: URL) throws {
        let acct = account(for: vaultURL)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: syncAPIKeyService,
            kSecAttrAccount: acct,
        ]

        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            throw KeychainError.unexpectedStatus(status)
        }
    }
}
