import Foundation

struct ItemVersion: Codable, Identifiable, Hashable {
    var id: Int { version }

    let version: Int
    let item: VaultItem
    let savedAt: Date

    enum CodingKeys: String, CodingKey {
        case version
        case item
        case savedAt = "saved_at"
    }
}
