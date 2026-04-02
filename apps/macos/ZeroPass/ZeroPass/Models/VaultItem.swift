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
