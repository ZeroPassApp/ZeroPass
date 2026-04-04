import Foundation
import Combine

@MainActor
final class VaultClient: ObservableObject {
    private static let uiTestModeArgument = "UITEST_MODE"
    private static let uiTestResetStateArgument = "UITEST_RESET_STATE"

    enum State: Equatable {
        case noVault
        case locked
        case showingRecovery(mnemonic: String)
        case unlocked
    }

    enum AuthModal: String, Identifiable {
        case createVault
        case openVault

        var id: String { rawValue }
    }

    @Published private(set) var state: State = .noVault
    @Published private(set) var unlockPasswordFocusRequestID = UUID()
    @Published var activeAuthModal: AuthModal?
    @Published var items: [VaultItem] = []
    @Published var selectedItemID: VaultItem.ID?
    @Published var lastError: String?
    @Published var authFlowError: String?

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

    private let keychain: any KeychainStoring
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
            if !syncEnabled && !suppressSyncDisableCleanup {
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

    @Published var syncAPIKey: String

    @Published var syncDeviceName: String {
        didSet { UserDefaults.standard.set(syncDeviceName, forKey: "syncDeviceName") }
    }

    @Published private(set) var syncLastSyncedAt: Date?
    @Published private(set) var syncStatus: String?

    private var syncLastSyncTimeNanos: Int64 = 0
    private var suppressSyncDisableCleanup = false

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

    init(keychain: any KeychainStoring = KeychainService()) {
	    self.keychain = keychain
        Self.resetStateForUITestsIfNeeded()

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

        self.syncAPIKey = ""
        self.syncDeviceName = UserDefaults.standard.string(forKey: "syncDeviceName") ?? Self.defaultDeviceName()

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

    deinit {
        MainActor.assumeIsolated {
            autoLock.stop()
            endSecurityScopedAccess()
        }
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

    static var isRunningUITests: Bool {
        ProcessInfo.processInfo.arguments.contains(uiTestModeArgument)
    }

    private static func resetStateForUITestsIfNeeded() {
        let arguments = ProcessInfo.processInfo.arguments
        guard arguments.contains(uiTestModeArgument), arguments.contains(uiTestResetStateArgument) else {
            return
        }

        if let bundleIdentifier = Bundle.main.bundleIdentifier {
            UserDefaults.standard.removePersistentDomain(forName: bundleIdentifier)
        }
        BookmarkStore().clear()
        UserDefaults.standard.synchronize()
    }

    func requestUnlockPasswordFocus() {
        unlockPasswordFocusRequestID = UUID()
    }

    func presentCreateVaultSheet() {
        authFlowError = nil
        activeAuthModal = .createVault
    }

    func presentOpenVaultSheet() {
        authFlowError = nil
        activeAuthModal = .openVault
    }

    func dismissAuthModal() {
        activeAuthModal = nil
    }

    func restoreLastVaultIfAvailable() async {
        guard state == .noVault else { return }
        guard let url = bookmarks.loadVaultURL() else { return }
        do {
            try await openVault(url)
        } catch {
            authFlowError = error.localizedDescription
        }
    }

    func createVault(_ url: URL, masterPassword: String) async throws {
        authFlowError = nil
        try beginSecurityScopedAccess(url)

        let path = url.path
        let resp = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.createVault(path: path, masterPassword: masterPassword)
        }.value

        vaultURL = url
        handle = resp.handle
        state = .showingRecovery(mnemonic: resp.mnemonic)
        authFlowError = nil

        loadSyncConfigFromDisk()
        try? await rehydrateSyncClientFromStoredSecretIfNeeded()
        try? bookmarks.saveVaultURL(url)
    }

    func openVault(_ url: URL) async throws {
        authFlowError = nil
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
        authFlowError = nil

        loadSyncConfigFromDisk()
        try? await rehydrateSyncClientFromStoredSecretIfNeeded()
        try? bookmarks.saveVaultURL(url)
    }

    func replaceVault(with url: URL) async throws {
        if vaultURL == url, handle != nil {
            authFlowError = nil
            return
        }

        authFlowError = nil

        guard url.startAccessingSecurityScopedResource() else {
            throw ZPBridgeError(code: .internalError, message: "Failed to access vault folder")
        }

        do {
            let path = url.path
            let newHandle = try await Task.detached(priority: .userInitiated) {
                try ZPBridge.openVault(path: path)
            }.value

            if let oldHandle = handle {
                _ = try? await Task.detached(priority: .utility) {
                    try ZPBridge.close(handle: oldHandle)
                }.value
            }

            securityScopedURL?.stopAccessingSecurityScopedResource()
            securityScopedURL = url
            vaultURL = url
            handle = newHandle
            items = []
            selectedItemID = nil
            state = .locked
            autoLock.setVaultUnlocked(false)
            authFlowError = nil

            resetSyncFromVault()
            loadSyncConfigFromDisk()
                try? await rehydrateSyncClientFromStoredSecretIfNeeded()
            try? bookmarks.saveVaultURL(url)
        } catch {
            url.stopAccessingSecurityScopedResource()
            throw error
        }
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

        let newMnemonic = try await Task.detached(priority: .userInitiated) {
            try ZPBridge.unlockWithRecovery(handle: h, mnemonic: mnemonic)
        }.value

        state = .showingRecovery(mnemonic: newMnemonic)
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
        authFlowError = nil

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

    func exportCSV(to url: URL) async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        try await Task.detached(priority: .utility) {
            try ZPBridge.exportCSV(handle: h, path: url.path)
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

    enum ImportSource {
        case chrome, firefox, safari, onePassword, onePasswordPUX, bitwarden, lastPass, keepass, csv
    }

    @discardableResult
    func importFrom(source: ImportSource, url: URL) async throws -> Int {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

        let count = try await Task.detached(priority: .utility) {
            switch source {
            case .chrome: return try ZPBridge.importChrome(handle: h, path: url.path)
            case .firefox: return try ZPBridge.importFirefox(handle: h, path: url.path)
            case .safari: return try ZPBridge.importSafari(handle: h, path: url.path)
            case .onePassword: return try ZPBridge.import1Password(handle: h, path: url.path)
            case .onePasswordPUX: return try ZPBridge.import1PUX(handle: h, path: url.path)
            case .bitwarden: return try ZPBridge.importBitwarden(handle: h, path: url.path)
            case .lastPass: return try ZPBridge.importLastPass(handle: h, path: url.path)
            case .keepass: return try ZPBridge.importKeePass(handle: h, path: url.path)
            case .csv: return try ZPBridge.importCSV(handle: h, path: url.path)
            }
        }.value

        autoLock.recordActivity()
        try await refreshItems()
        return count
    }

    func applySyncConfig() async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
	    guard let url = vaultURL else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

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

        let previousAPIKey = (try keychain.loadSyncAPIKey(for: url) ?? "")
            .trimmingCharacters(in: .whitespacesAndNewlines)

        do {
            if apiKey.isEmpty {
                try keychain.deleteSyncAPIKey(for: url)
            } else {
                try keychain.storeSyncAPIKey(apiKey, for: url)
            }

            try await Task.detached(priority: .userInitiated) {
                try ZPBridge.syncSetup(handle: h, config: cfg)
            }.value
        } catch {
            if previousAPIKey.isEmpty {
                try? keychain.deleteSyncAPIKey(for: url)
            } else {
                try? keychain.storeSyncAPIKey(previousAPIKey, for: url)
            }
            throw error
        }

        loadSyncConfigFromDisk()
        syncStatus = "Saved"
        autoLock.recordActivity()
    }

    func syncRegisterDevice() async throws {
        guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
        try await rehydrateSyncClientFromStoredSecretIfNeeded()
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
	    try await rehydrateSyncClientFromStoredSecretIfNeeded()

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
        syncAPIKey = ""
        syncLastSyncTimeNanos = 0
        syncLastSyncedAt = nil
        syncStatus = nil
    }

    private func disableSyncConfigOnDisk() {
        if let url = vaultURL {
            try? keychain.deleteSyncAPIKey(for: url)
            try? FileManager.default.removeItem(at: syncConfigURL(for: url))
        }
        UserDefaults.standard.removeObject(forKey: "syncAPIKey")
        syncAPIKey = ""
        syncLastSyncTimeNanos = 0
        syncLastSyncedAt = nil
        syncStatus = nil
    }

    private func loadSyncConfigFromDisk() {
        guard let url = vaultURL else { return }
        suppressSyncDisableCleanup = true
        defer { suppressSyncDisableCleanup = false }

        migrateLegacySyncAPIKeyIfNeeded(for: url)

        let fp = syncConfigURL(for: url)
        guard let data = try? Data(contentsOf: fp) else {
            syncServerURL = ""
            syncAPIKey = ""
            syncEnabled = false
            syncLastSyncTimeNanos = 0
            syncLastSyncedAt = nil
            syncStatus = nil
            return
        }

        do {
            let cfg = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: data)
            syncServerURL = cfg.serverURL
            syncDeviceID = cfg.deviceID
            syncAPIKey = (try? keychain.loadSyncAPIKey(for: url)) ?? ""
            syncLastSyncTimeNanos = cfg.lastSyncTime ?? 0

            if syncLastSyncTimeNanos > 0 {
                syncLastSyncedAt = Date(timeIntervalSince1970: TimeInterval(syncLastSyncTimeNanos) / 1_000_000_000)
            } else {
                syncLastSyncedAt = nil
            }
            syncEnabled = !cfg.serverURL.isEmpty && !cfg.deviceID.isEmpty
        } catch {
            syncServerURL = ""
            syncDeviceID = ""
            syncAPIKey = ""
            syncEnabled = false
            syncLastSyncTimeNanos = 0
            syncLastSyncedAt = nil
            syncStatus = "Failed to read sync config"
        }
    }

    private func rehydrateSyncClientFromStoredSecretIfNeeded() async throws {
        guard syncEnabled else { return }
        guard let h = handle else { return }
        guard let url = vaultURL else { return }

        let server = syncServerURL.trimmingCharacters(in: .whitespacesAndNewlines)
        let device = syncDeviceID.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !server.isEmpty, !device.isEmpty else { return }

        let apiKey = (try keychain.loadSyncAPIKey(for: url) ?? "")
            .trimmingCharacters(in: .whitespacesAndNewlines)
        syncAPIKey = apiKey

        let cfg = ZPBridge.SyncConfig(
            serverURL: server,
            deviceID: device,
            apiKey: apiKey.isEmpty ? nil : apiKey,
            lastSyncTime: syncLastSyncTimeNanos > 0 ? syncLastSyncTimeNanos : nil
        )

        try await Task.detached(priority: .utility) {
            try ZPBridge.syncSetup(handle: h, config: cfg)
        }.value
    }

    func migrateLegacySyncAPIKeyIfNeeded(for url: URL) {
        let legacyDefaultsKey = UserDefaults.standard.string(forKey: "syncAPIKey")?
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let configURL = syncConfigURL(for: url)

        let legacyConfig: ZPBridge.SyncConfig?
        if let data = try? Data(contentsOf: configURL) {
            legacyConfig = try? JSONDecoder().decode(ZPBridge.SyncConfig.self, from: data)
        } else {
            legacyConfig = nil
        }

        let storedKeychainKey = (try? keychain.loadSyncAPIKey(for: url))?
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let legacyDiskKey = legacyConfig?.apiKey?
            .trimmingCharacters(in: .whitespacesAndNewlines)

        let legacyDefaultsCandidate = legacyConfig == nil ? nil : legacyDefaultsKey

        let candidate = [storedKeychainKey, legacyDiskKey, legacyDefaultsCandidate]
            .compactMap { value -> String? in
                guard let value, !value.isEmpty else { return nil }
                return value
            }
            .first

        let hasStoredKeychainKey = storedKeychainKey?.isEmpty == false
        var migratedToKeychain = hasStoredKeychainKey

        if !hasStoredKeychainKey, let candidate {
            do {
                try keychain.storeSyncAPIKey(candidate, for: url)
                migratedToKeychain = true
            } catch {
                return
            }
        }

        if migratedToKeychain, let legacyConfig, legacyConfig.apiKey != nil {
            try? persistMetadataOnlySyncConfig(legacyConfig, to: configURL)
        }

        if migratedToKeychain {
            UserDefaults.standard.removeObject(forKey: "syncAPIKey")
        }
    }

    func syncConfigURL(for url: URL) -> URL {
        url.appendingPathComponent("sync.json")
    }

    func persistMetadataOnlySyncConfig(_ cfg: ZPBridge.SyncConfig, to url: URL) throws {
        let metadataOnly = ZPBridge.SyncConfig(
            serverURL: cfg.serverURL,
            deviceID: cfg.deviceID,
            apiKey: nil,
            lastSyncTime: cfg.lastSyncTime
        )
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        let data = try encoder.encode(metadataOnly)
        try data.write(to: url, options: .atomic)
    }

    private static func defaultDeviceName() -> String {
        let hostName = ProcessInfo.processInfo.hostName
            .split(separator: ".", maxSplits: 1, omittingEmptySubsequences: true)
            .first
            .map(String.init)?
            .trimmingCharacters(in: .whitespacesAndNewlines)

        if let hostName, !hostName.isEmpty {
            return hostName
        }
        return "Mac"
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
