import Foundation
import Combine

@MainActor
final class VaultClient: ObservableObject {
    enum State: Equatable {
        case noVault
        case locked
        case showingRecovery(mnemonic: String)
        case unlocked
    }

    @Published private(set) var state: State = .noVault
    @Published var items: [VaultItem] = []
    @Published var selectedItemID: VaultItem.ID?
    @Published var lastError: String?

    @Published var biometricUnlockEnabled: Bool {
        didSet {
            UserDefaults.standard.set(biometricUnlockEnabled, forKey: "biometricUnlockEnabled")
            if !biometricUnlockEnabled, let url = vaultURL {
                try? keychain.deleteVaultKey(for: url)
            }
        }
    }

    @Published var clipboardAutoClearEnabled: Bool {
        didSet { UserDefaults.standard.set(clipboardAutoClearEnabled, forKey: "clipboardAutoClearEnabled") }
    }

    @Published var clipboardAutoClearSeconds: Int {
        didSet { UserDefaults.standard.set(clipboardAutoClearSeconds, forKey: "clipboardAutoClearSeconds") }
    }

    private let keychain = KeychainService()
    private let biometrics = BiometricService()
    private let autoLock = AutoLockService()

    private var handle: Int?
    private var vaultURL: URL?

    private var securityScopedURL: URL?
    private let bookmarks = BookmarkStore()

    @Published var autoLockTimeoutSeconds: Int {
        didSet {
            UserDefaults.standard.set(autoLockTimeoutSeconds, forKey: "autoLockTimeoutSeconds")
            configureAutoLock()
        }
    }

    @Published var syncEnabled: Bool {
        didSet {
            UserDefaults.standard.set(syncEnabled, forKey: "syncEnabled")
            if !syncEnabled {
                disableSyncConfigOnDisk()
            }
        }
    }

    @Published var syncServerURL: String {
        didSet { UserDefaults.standard.set(syncServerURL, forKey: "syncServerURL") }
    }

    @Published var syncDeviceID: String {
        didSet { UserDefaults.standard.set(syncDeviceID, forKey: "syncDeviceID") }
    }

    @Published var syncAPIKey: String {
        didSet { UserDefaults.standard.set(syncAPIKey, forKey: "syncAPIKey") }
    }

    @Published var syncDeviceName: String {
        didSet { UserDefaults.standard.set(syncDeviceName, forKey: "syncDeviceName") }
    }

    @Published private(set) var syncLastSyncedAt: Date?
    @Published private(set) var syncStatus: String?

    private var syncLastSyncTimeNanos: Int64 = 0

    @Published var lockOnSleepEnabled: Bool {
        didSet {
            UserDefaults.standard.set(lockOnSleepEnabled, forKey: "lockOnSleepEnabled")
            configureAutoLock()
        }
    }

    @Published var lockOnScreenSleepEnabled: Bool {
        didSet {
            UserDefaults.standard.set(lockOnScreenSleepEnabled, forKey: "lockOnScreenSleepEnabled")
            configureAutoLock()
        }
    }

    init() {
        if UserDefaults.standard.object(forKey: "biometricUnlockEnabled") == nil {
            self.biometricUnlockEnabled = false
        } else {
            self.biometricUnlockEnabled = UserDefaults.standard.bool(forKey: "biometricUnlockEnabled")
        }

        self.syncEnabled = UserDefaults.standard.bool(forKey: "syncEnabled")
        self.syncServerURL = UserDefaults.standard.string(forKey: "syncServerURL") ?? ""

        if let id = UserDefaults.standard.string(forKey: "syncDeviceID"), !id.isEmpty {
            self.syncDeviceID = id
        } else {
            let id = UUID().uuidString
            self.syncDeviceID = id
            UserDefaults.standard.set(id, forKey: "syncDeviceID")
        }

        self.syncAPIKey = UserDefaults.standard.string(forKey: "syncAPIKey") ?? ""
        self.syncDeviceName = UserDefaults.standard.string(forKey: "syncDeviceName") ?? Host.current().localizedName ?? "Mac"

        self.syncLastSyncedAt = nil
        self.syncStatus = nil

        if UserDefaults.standard.object(forKey: "clipboardAutoClearEnabled") == nil {
            self.clipboardAutoClearEnabled = true
        } else {
            self.clipboardAutoClearEnabled = UserDefaults.standard.bool(forKey: "clipboardAutoClearEnabled")
        }

        let secs = UserDefaults.standard.integer(forKey: "clipboardAutoClearSeconds")
        self.clipboardAutoClearSeconds = secs > 0 ? secs : 30

        let autoSecs = UserDefaults.standard.integer(forKey: "autoLockTimeoutSeconds")
        self.autoLockTimeoutSeconds = autoSecs >= 0 ? autoSecs : 300

        if UserDefaults.standard.object(forKey: "lockOnSleepEnabled") == nil {
            self.lockOnSleepEnabled = true
        } else {
            self.lockOnSleepEnabled = UserDefaults.standard.bool(forKey: "lockOnSleepEnabled")
        }

        if UserDefaults.standard.object(forKey: "lockOnScreenSleepEnabled") == nil {
            self.lockOnScreenSleepEnabled = true
        } else {
            self.lockOnScreenSleepEnabled = UserDefaults.standard.bool(forKey: "lockOnScreenSleepEnabled")
        }

        autoLock.onLock = { [weak self] in
            Task { await self?.lock() }
        }
        configureAutoLock()
    }

    var vaultPathDisplay: String {
        vaultURL?.path ?? "—"
    }

    var hasVault: Bool { handle != nil }
    var isUnlocked: Bool {
        if case .unlocked = state { return true }
        return false
    }

    var isBiometricAvailable: Bool { biometrics.isAvailable() }

    func restoreLastVaultIfAvailable() async {
        guard state == .noVault else { return }
        guard let url = bookmarks.loadVaultURL() else { return }
        do {
            try await openVault(url)
        } catch {
            lastError = error.localizedDescription
        }
    }

    func createVault(_ url: URL, masterPassword: String) async throws {
        try beginSecurityScopedAccess(url)

        let path = url.path
        let resp = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.createVault(path: path, masterPassword: masterPassword)
        }.value

        vaultURL = url
        handle = resp.handle
        state = .showingRecovery(mnemonic: resp.mnemonic)

        loadSyncConfigFromDisk()
        try? bookmarks.saveVaultURL(url)
    }

    func openVault(_ url: URL) async throws {
        try beginSecurityScopedAccess(url)

        let path = url.path
        let h = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.openVault(path: path)
        }.value

        vaultURL = url
        handle = h
        items = []
        selectedItemID = nil
        state = .locked

        loadSyncConfigFromDisk()
        try? bookmarks.saveVaultURL(url)
    }

    func unlock(masterPassword: String) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.unlock(handle: h, masterPassword: masterPassword)
        }.value

        state = .unlocked
        autoLock.setVaultUnlocked(true)
        autoLock.recordActivity()
        try await refreshItems()
        await persistVaultKeyIfNeeded()
    }

    func unlockWithRecovery(mnemonic: String) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.unlockWithRecovery(handle: h, mnemonic: mnemonic)
        }.value

        state = .unlocked
        autoLock.setVaultUnlocked(true)
        autoLock.recordActivity()
        try await refreshItems()
        await persistVaultKeyIfNeeded()
    }

    func acceptRecoveryPhrase() {
        if case .showingRecovery = state {
            state = .unlocked
            autoLock.setVaultUnlocked(true)
            autoLock.recordActivity()
            Task { await persistVaultKeyIfNeeded() }
        }
    }

    func lock() async {
        guard let h = handle else { return }
        do {
            try await Task.detached(priority: .userInitiated) {
                try ZPBridge.lock(handle: h)
            }.value
            items = []
            selectedItemID = nil
            state = .locked
            autoLock.setVaultUnlocked(false)
        } catch {
            lastError = error.localizedDescription
        }
    }

    func closeVault() async {
        if let h = handle {
            _ = try? await Task.detached(priority: .userInitiated) {
                try ZPBridge.close(handle: h)
            }.value
        }

        handle = nil
        vaultURL = nil
        items = []
        selectedItemID = nil
        state = .noVault
        autoLock.setVaultUnlocked(false)

        resetSyncFromVault()
        endSecurityScopedAccess()
    }

    func unlockWithBiometrics() async throws {
        guard biometricUnlockEnabled else {
            throw ZPBridgeError(code: .internalError, message: "Biometric unlock is disabled")
        }
        guard biometrics.isAvailable() else {
            throw BiometricError.unavailable
        }
        guard let url = vaultURL else {
            throw ZPBridgeError(code: .notFound, message: "No vault selected")
        }
        guard let h = handle else {
            throw ZPBridgeError(code: .notFound, message: "No vault open")
        }

        let ctx = try await biometrics.authenticateContext(reason: "Unlock ZeroPass")
        let key = try keychain.loadVaultKeyBase64(for: url, context: ctx)
        guard let key else {
            throw ZPBridgeError(code: .notFound, message: "No vault key stored in Keychain")
        }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.unlockWithKey(handle: h, vaultKeyBase64: key)
        }.value

        state = .unlocked
        autoLock.setVaultUnlocked(true)
        autoLock.recordActivity()
        try await refreshItems()
    }

    private func persistVaultKeyIfNeeded() async {
        guard biometricUnlockEnabled else { return }
        guard biometrics.isAvailable() else { return }
        guard let url = vaultURL else { return }
        guard let h = handle else { return }

        do {
            let key = try await Task.detached(priority: .utility) {
                try ZPBridge.getVaultKeyBase64(handle: h)
            }.value

            try keychain.storeVaultKeyBase64(key, for: url, requireUserPresence: true)
        } catch {
            lastError = error.localizedDescription
        }
    }

    func refresh() async {
        do {
            try await refreshItems()
        } catch {
            lastError = error.localizedDescription
        }
    }

    func refreshItems() async throws {
        guard let h = handle else { return }

        let newItems = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.listItems(handle: h, filter: nil)
        }.value

        items = newItems
        if selectedItemID == nil {
            selectedItemID = newItems.first?.id
        }

        autoLock.recordActivity()
    }

    func recentItems(limit: Int) -> [VaultItem] {
        Array(items.sorted(by: { $0.lastAccessedAt > $1.lastAccessedAt }).prefix(limit))
    }

    func searchItems(query: String) async throws -> [VaultItem] {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
        let ids = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.searchIDs(handle: h, query: query)
        }.value

        let byID = Dictionary(uniqueKeysWithValues: items.map { ($0.id, $0) })
        autoLock.recordActivity()
        return ids.compactMap { byID[$0] }
    }

    // MARK: - Filtering & Counts

    var favoriteCount: Int {
        items.filter(\.favorite).count
    }

    var allTags: [String] {
        Array(Set(items.flatMap(\.tags))).sorted()
    }

    func count(for type: VaultItemType) -> Int {
        items.filter { $0.type == type }.count
    }

    func filteredItems(type: VaultItemType? = nil, tag: String? = nil, favoritesOnly: Bool = false, sortBy: String = "name", sortOrder: String = "asc") -> [VaultItem] {
        var result = items

        if let type {
            result = result.filter { $0.type == type }
        }
        if favoritesOnly {
            result = result.filter(\.favorite)
        }
        if let tag {
            result = result.filter { $0.tags.contains(tag) }
        }

        switch sortBy {
        case "name":
            result.sort { sortOrder == "asc" ? $0.name.localizedCompare($1.name) == .orderedAscending : $0.name.localizedCompare($1.name) == .orderedDescending }
        case "updated_at":
            result.sort { sortOrder == "asc" ? $0.updatedAt < $1.updatedAt : $0.updatedAt > $1.updatedAt }
        case "created_at":
            result.sort { sortOrder == "asc" ? $0.createdAt < $1.createdAt : $0.createdAt > $1.createdAt }
        case "type":
            result.sort { sortOrder == "asc" ? $0.type.rawValue < $1.type.rawValue : $0.type.rawValue > $1.type.rawValue }
        default:
            break
        }

        return result
    }

    // MARK: - Password Generation

    func generatePassword(length: Int = 20, options: ZPBridge.GeneratePasswordOptions? = nil) async throws -> String {
        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.generatePassword(length: length, options: options)
        }.value
    }

    func generatePassphrase(words: Int = 5, separator: String = "-") async throws -> String {
        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.generatePassphrase(words: words, separator: separator)
        }.value
    }

    func scorePassword(_ password: String) async throws -> ZPBridge.PasswordScore {
        try await Task.detached(priority: .utility) {
            try ZPBridge.scorePassword(password)
        }.value
    }

    // MARK: - Vault Info

    var vaultName: String {
        vaultURL?.lastPathComponent ?? "ZeroPass"
    }

    func toggleFavorite(item: VaultItem) async throws {
        var updated = item
        updated.favorite = !item.favorite
        try await saveItem(updated)
    }

    func saveItem(_ item: VaultItem) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let saved: VaultItem = try await Task.detached(priority: .userInitiated) {
            if item.id.isEmpty {
                return try ZPBridge.createItem(handle: h, item: item)
            }
            return try ZPBridge.updateItem(handle: h, id: item.id, item: item)
        }.value

        if let idx = items.firstIndex(where: { $0.id == saved.id }) {
            items[idx] = saved
        } else {
            items.insert(saved, at: 0)
        }
        selectedItemID = saved.id

        autoLock.recordActivity()
    }

    func deleteItem(id: String) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.deleteItem(handle: h, id: id)
        }.value

        items.removeAll { $0.id == id }
        if selectedItemID == id {
            selectedItemID = items.first?.id
        }

        autoLock.recordActivity()
    }

    func versionHistory(itemID: String) async throws -> [ItemVersion] {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let v = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.versionHistory(handle: h, id: itemID)
        }.value

        autoLock.recordActivity()
        return v
    }

    func restoreVersion(itemID: String, version: Int) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        _ = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.restoreVersion(handle: h, id: itemID, version: Int32(version))
        }.value

        autoLock.recordActivity()
        try await refreshItems()
    }

    func changeMasterPassword(old: String, new: String) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.changeMasterPassword(handle: h, old: old, new: new)
        }.value

        autoLock.recordActivity()
        await persistVaultKeyIfNeeded()
    }

    func scorePasswordSummary(_ password: String) async -> String {
        if password.isEmpty { return "" }
        do {
            let score = try await Task.detached(priority: .utility) {
                try ZPBridge.scorePassword(password)
            }.value
            return "Score \(score.score)/4 — \(score.feedback)"
        } catch {
            return ""
        }
    }

    func regenerateRecoveryMnemonic() async throws -> String {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let mnemonic = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.regenerateRecovery(handle: h)
        }.value

        autoLock.recordActivity()
        return mnemonic
    }

    func exportJSON(to url: URL) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .utility) {
            try ZPBridge.exportJSON(handle: h, path: url.path)
        }.value

        autoLock.recordActivity()
    }

    func exportEncrypted(to url: URL) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .utility) {
            try ZPBridge.exportEncrypted(handle: h, path: url.path)
        }.value

        autoLock.recordActivity()
    }

    @discardableResult
    func importCSV(from url: URL) async throws -> Int {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let count = try await Task.detached(priority: .utility) {
            try ZPBridge.importCSV(handle: h, path: url.path)
        }.value

        autoLock.recordActivity()
        return count
    }

    func applySyncConfig() async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        if !syncEnabled {
            disableSyncConfigOnDisk()
            syncStatus = "Sync disabled"
            return
        }

        let server = syncServerURL.trimmingCharacters(in: .whitespacesAndNewlines)
        let device = syncDeviceID.trimmingCharacters(in: .whitespacesAndNewlines)
        if server.isEmpty || device.isEmpty {
            throw ZPBridgeError(code: .internalError, message: "Server URL and Device ID are required")
        }

        let apiKey = syncAPIKey.trimmingCharacters(in: .whitespacesAndNewlines)
        let cfg = ZPBridge.SyncConfig(
            serverURL: server,
            deviceID: device,
            apiKey: apiKey.isEmpty ? nil : apiKey,
            lastSyncTime: syncLastSyncTimeNanos > 0 ? syncLastSyncTimeNanos : nil
        )

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.syncSetup(handle: h, config: cfg)
        }.value

        loadSyncConfigFromDisk()
        syncStatus = "Saved"
        autoLock.recordActivity()
    }

    func syncRegisterDevice() async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
        let name = syncDeviceName.trimmingCharacters(in: .whitespacesAndNewlines)
        if name.isEmpty {
            throw ZPBridgeError(code: .internalError, message: "Device name must not be empty")
        }

        try await Task.detached(priority: .userInitiated) {
            try ZPBridge.syncRegister(handle: h, deviceName: name)
        }.value

        syncStatus = "Registered"
        NotificationService.shared.notify(title: "ZeroPass", body: "Device registered")
        autoLock.recordActivity()
    }

    func syncNow() async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let resp = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.syncFull(handle: h)
        }.value

        syncLastSyncedAt = resp.syncedAt
        syncStatus = "Pulled \(resp.result.pulled) • Pushed \(resp.result.pushed) • Conflicts \(resp.result.conflicts)"

        // Sync updates config on disk (last_sync_time). Refresh our view and data.
        loadSyncConfigFromDisk()
        try await refreshItems()

        NotificationService.shared.notify(title: "ZeroPass", body: "Sync complete")
        autoLock.recordActivity()
    }

    private func configureAutoLock() {
        autoLock.configure(
            timeoutSeconds: autoLockTimeoutSeconds,
            lockOnSleep: lockOnSleepEnabled,
            lockOnScreenSleep: lockOnScreenSleepEnabled
        )
        autoLock.setVaultUnlocked(isUnlocked)
    }

    private func resetSyncFromVault() {
        syncLastSyncTimeNanos = 0
        syncLastSyncedAt = nil
        syncStatus = nil
    }

    private func disableSyncConfigOnDisk() {
        guard let url = vaultURL else { return }
        let fp = url.appendingPathComponent("sync.json")
        try? FileManager.default.removeItem(at: fp)
        syncLastSyncTimeNanos = 0
        syncLastSyncedAt = nil
        syncStatus = nil
    }

    private func loadSyncConfigFromDisk() {
        guard let url = vaultURL else { return }

        let fp = url.appendingPathComponent("sync.json")
        guard let data = try? Data(contentsOf: fp) else {
            // Keep current values (UserDefaults-backed). Sync is considered disabled.
            syncLastSyncTimeNanos = 0
            syncLastSyncedAt = nil
            syncStatus = nil
            return
        }

        do {
            let cfg = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: data)
            syncServerURL = cfg.serverURL
            syncDeviceID = cfg.deviceID
            syncAPIKey = cfg.apiKey ?? ""
            syncLastSyncTimeNanos = cfg.lastSyncTime ?? 0

            if syncLastSyncTimeNanos > 0 {
                syncLastSyncedAt = Date(timeIntervalSince1970: TimeInterval(syncLastSyncTimeNanos) / 1_000_000_000)
            } else {
                syncLastSyncedAt = nil
            }
            syncEnabled = !cfg.serverURL.isEmpty && !cfg.deviceID.isEmpty
        } catch {
            syncStatus = "Failed to read sync config"
        }
    }

    // MARK: - Security scoped access

    private func beginSecurityScopedAccess(_ url: URL) throws {
        endSecurityScopedAccess()

        guard url.startAccessingSecurityScopedResource() else {
            throw ZPBridgeError(code: .internalError, message: "Failed to access vault folder")
        }
        securityScopedURL = url
    }

    private func endSecurityScopedAccess() {
        securityScopedURL?.stopAccessingSecurityScopedResource()
        securityScopedURL = nil
    }
}
