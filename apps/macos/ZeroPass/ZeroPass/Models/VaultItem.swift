import Foundation

struct VaultItem: Codable, Identifiable, Hashable {
    var id: String
    var type: VaultItemType
    var name: String
    var fields: [String: String]
    var notes: String
    var tags: [String]
    var favorite: Bool
    var customFields: [String: String]
    var createdAt: Date
    var updatedAt: Date
    var lastAccessedAt: Date
    var version: Int

    enum CodingKeys: String, CodingKey {
        case id
        case type
        case name
        case fields
        case notes
        case tags
        case favorite
        case customFields = "custom_fields"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case lastAccessedAt = "last_accessed_at"
        case version
    }

    static func new(type: VaultItemType) -> VaultItem {
        let now = Date()
        return VaultItem(
            id: "",
            type: type,
            name: "",
            fields: [:],
            notes: "",
            tags: [],
            favorite: false,
            customFields: [:],
            createdAt: now,
            updatedAt: now,
            lastAccessedAt: now,
            version: 0
        )
    }
}

extension VaultItem {
    static let preferredIdentityFieldKeys = [
        "username", "email", "user", "login", "cardholder", "full_name", "relying_party"
    ]

    static let preferredSecretFieldKeys = [
        "password", "api_key", "api_secret", "secret", "private_key", "cvv", "card_number", "credential_id", "passphrase"
    ]

    static let preferredURLFieldKeys = ["url", "endpoint"]

    var preferredIdentityField: (key: String, value: String)? {
        firstNonEmptyField(in: Self.preferredIdentityFieldKeys)
    }

    var preferredSecretField: (key: String, value: String)? {
        firstNonEmptyField(in: Self.preferredSecretFieldKeys)
    }

    var preferredURLField: (key: String, value: String)? {
        firstNonEmptyField(in: Self.preferredURLFieldKeys)
    }

    func firstNonEmptyField(in keys: [String]) -> (key: String, value: String)? {
        for key in keys {
            let value = fields[key, default: ""].trimmingCharacters(in: .whitespacesAndNewlines)
            if !value.isEmpty {
                return (key, value)
            }
        }

        return nil
    }
}
