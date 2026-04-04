import Foundation

nonisolated final class BookmarkStore {
    private let defaults = UserDefaults.standard
    private let key = "vaultBookmark"

    func saveVaultURL(_ url: URL) throws {
        let data = try url.bookmarkData(
            options: [.withSecurityScope],
            includingResourceValuesForKeys: nil,
            relativeTo: nil
        )
        defaults.set(data, forKey: key)
    }

    func loadVaultURL() -> URL? {
        guard let data = defaults.data(forKey: key) else { return nil }
        var stale = false
        guard let url = try? URL(
            resolvingBookmarkData: data,
            options: [.withSecurityScope],
            relativeTo: nil,
            bookmarkDataIsStale: &stale
        ) else {
            return nil
        }
        return url
    }

    func clear() {
        defaults.removeObject(forKey: key)
    }
}
