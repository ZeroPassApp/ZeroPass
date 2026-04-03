import Foundation

struct ItemFilter: Codable {
    var type: VaultItemType?
    var tags: [String]?
    var favorite: Bool?
    var searchQuery: String?
    var sortBy: String?
    var sortOrder: String?

    enum CodingKeys: String, CodingKey {
        case type
        case tags
        case favorite
        case searchQuery = "search_query"
        case sortBy = "sort_by"
        case sortOrder = "sort_order"
    }
}
