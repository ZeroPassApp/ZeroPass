import Foundation

enum ZPBridge {
    typealias Handle = Int

    struct CreateVaultResponse: Codable {
        let handle: Handle
        let mnemonic: String
    }

    struct OpenVaultResponse: Codable {
        let handle: Handle
    }

    struct VaultKeyResponse: Codable {
        let vaultKeyBase64: String
    }

    struct UnlockWithRecoveryResponse: Codable {
        let newMnemonic: String
    }

    struct RegenerateRecoveryResponse: Codable {
        let mnemonic: String
    }

    struct ImportResponse: Codable {
        let imported: Int
    }

    struct SyncConfig: Codable {
        let serverURL: String
        let deviceID: String
        let apiKey: String?
        let lastSyncTime: Int64?

        enum CodingKeys: String, CodingKey {
            case serverURL = "server_url"
            case deviceID = "device_id"
            case apiKey = "api_key"
            case lastSyncTime = "last_sync_time"
        }
    }

    struct SyncResult: Codable {
        let pulled: Int
        let pushed: Int
        let conflicts: Int
    }

    struct SyncItem: Codable {
        let itemID: String
        let version: Int
        let deviceID: String
        let payload: String
        let timestamp: Int64
        let checksum: String
        let deleted: Bool

        enum CodingKeys: String, CodingKey {
            case itemID = "item_id"
            case version
            case deviceID = "device_id"
            case payload
            case timestamp
            case checksum
            case deleted
        }
    }

    struct SyncConflictRecord: Codable {
        let itemID: String
        let winner: SyncItem
        let loser: SyncItem
        let reason: String

        enum CodingKeys: String, CodingKey {
            case itemID = "ItemID"
            case winner = "Winner"
            case loser = "Loser"
            case reason = "Reason"
        }
    }

    struct SyncFullResponse: Codable {
        let result: SyncResult
        let conflicts: [SyncConflictRecord]
        let syncedAt: Date

        enum CodingKeys: String, CodingKey {
            case result
            case conflicts
            case syncedAt = "synced_at"
        }
    }

    struct PasswordScore: Codable {
        let score: Int
        let feedback: String
        let crackTimeSecs: Double

        enum CodingKeys: String, CodingKey {
            case score = "Score"
            case feedback = "Feedback"
            case crackTimeSecs = "CrackTimeSecs"
        }
    }

    struct GeneratePasswordOptions: Codable {
        var uppercase: Bool = true
        var lowercase: Bool = true
        var digits: Bool = true
        var symbols: Bool = true
        var excludeAmbiguous: Bool = false

        enum CodingKeys: String, CodingKey {
            case uppercase = "Uppercase"
            case lowercase = "Lowercase"
            case digits = "Digits"
            case symbols = "Symbols"
            case excludeAmbiguous = "ExcludeAmbiguous"
        }
    }

    struct GeneratedPassword: Codable {
        let password: String
    }

    struct GeneratedPassphrase: Codable {
        let passphrase: String
    }

    private nonisolated static func dataFromResult(_ result: ZPResult) throws -> Data {
        let r = result
        defer { ZPFreeResult(r) }

        let code = ZPErrorCode(rawValue: r.code) ?? .internalError
        if code != .ok {
            let msg = r.error.map { String(cString: $0) } ?? ""
            throw ZPBridgeError(code: code, message: msg)
        }

        guard let p = r.data else {
            return Data()
        }

        let s = String(cString: p)
        return Data(s.utf8)
    }

    private nonisolated static func decode<T: Decodable>(_ type: T.Type, from result: ZPResult) throws -> T {
        let data = try dataFromResult(result)
        return try ZPJSON.decoder.decode(T.self, from: data)
    }

    private nonisolated static func withMutableCString<T>(_ s: String, _ body: (UnsafeMutablePointer<CChar>) -> T) -> T {
        s.withCString { c in
            body(UnsafeMutablePointer(mutating: c))
        }
    }

    nonisolated static func createVault(path: String, masterPassword: String) throws -> CreateVaultResponse {
        let res: ZPResult = withMutableCString(path) { p in
            withMutableCString(masterPassword) { pw in
                ZPCreateVault(p, pw)
            }
        }
        return try decode(CreateVaultResponse.self, from: res)
    }

    nonisolated static func openVault(path: String) throws -> Handle {
        let res: ZPResult = withMutableCString(path) { p in
            ZPOpenVault(p)
        }
        return try decode(OpenVaultResponse.self, from: res).handle
    }

    nonisolated static func unlock(handle: Handle, masterPassword: String) throws {
        let res: ZPResult = withMutableCString(masterPassword) { pw in
            ZPUnlock(handle, pw)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func unlockWithRecovery(handle: Handle, mnemonic: String) throws -> String {
        let res: ZPResult = withMutableCString(mnemonic) { m in
            ZPUnlockWithRecovery(handle, m)
        }
        return try decode(UnlockWithRecoveryResponse.self, from: res).newMnemonic
    }

    nonisolated static func unlockWithKey(handle: Handle, vaultKeyBase64: String) throws {
        let res: ZPResult = withMutableCString(vaultKeyBase64) { k in
            ZPUnlockWithKey(handle, k)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func getVaultKeyBase64(handle: Handle) throws -> String {
        let res: ZPResult = ZPGetVaultKey(handle)
        return try decode(VaultKeyResponse.self, from: res).vaultKeyBase64
    }

    nonisolated static func lock(handle: Handle) throws {
        let res: ZPResult = ZPLock(handle)
        _ = try dataFromResult(res)
    }

    nonisolated static func close(handle: Handle) throws {
        let res: ZPResult = ZPCloseVault(handle)
        _ = try dataFromResult(res)
    }

    nonisolated static func isLocked(handle: Handle) -> Bool {
        ZPIsLocked(handle) != 0
    }

    nonisolated static func listItems(handle: Handle, filter: ItemFilter? = nil) throws -> [VaultItem] {
        let filterJSON: String
        if let filter {
            let d = try ZPJSON.encoder.encode(filter)
            filterJSON = String(decoding: d, as: UTF8.self)
        } else {
            filterJSON = ""
        }

        let res: ZPResult = withMutableCString(filterJSON) { f in
            ZPListItems(handle, f)
        }

        let data = try dataFromResult(res)
        let trimmed = String(decoding: data, as: UTF8.self)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed == "null" {
            return []
        }
        if trimmed.isEmpty {
            throw ZPBridgeError(code: .internalError, message: "bridge protocol error: expected JSON array, got empty payload")
        }

        return try ZPJSON.decoder.decode([VaultItem].self, from: data)
    }

    nonisolated static func getItem(handle: Handle, id: String) throws -> VaultItem {
        let res: ZPResult = withMutableCString(id) { c in
            ZPGetItem(handle, c)
        }
        return try decode(VaultItem.self, from: res)
    }

    nonisolated static func createItem(handle: Handle, item: VaultItem) throws -> VaultItem {
        let d = try ZPJSON.encoder.encode(item)
        let json = String(decoding: d, as: UTF8.self)

        let res: ZPResult = withMutableCString(json) { c in
            ZPCreateItem(handle, c)
        }
        return try decode(VaultItem.self, from: res)
    }

    nonisolated static func updateItem(handle: Handle, id: String, item: VaultItem) throws -> VaultItem {
        let d = try ZPJSON.encoder.encode(item)
        let json = String(decoding: d, as: UTF8.self)

        let res: ZPResult = withMutableCString(id) { iid in
            withMutableCString(json) { payload in
                ZPUpdateItem(handle, iid, payload)
            }
        }
        return try decode(VaultItem.self, from: res)
    }

    nonisolated static func deleteItem(handle: Handle, id: String) throws {
        let res: ZPResult = withMutableCString(id) { iid in
            ZPDeleteItem(handle, iid)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func searchIDs(handle: Handle, query: String) throws -> [String] {
        let res: ZPResult = withMutableCString(query) { q in
            ZPSearch(handle, q)
        }
        return try decode([String].self, from: res)
    }

    nonisolated static func versionHistory(handle: Handle, id: String) throws -> [ItemVersion] {
        let res: ZPResult = withMutableCString(id) { iid in
            ZPGetVersionHistory(handle, iid)
        }
        return try decode([ItemVersion].self, from: res)
    }

    nonisolated static func restoreVersion(handle: Handle, id: String, version: Int32) throws -> VaultItem {
        let res: ZPResult = withMutableCString(id) { iid in
            ZPRestoreVersion(handle, iid, version)
        }
        return try decode(VaultItem.self, from: res)
    }

    nonisolated static func changeMasterPassword(handle: Handle, old: String, new: String) throws {
        let res: ZPResult = withMutableCString(old) { o in
            withMutableCString(new) { n in
                ZPChangeMasterPassword(handle, o, n)
            }
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func regenerateRecovery(handle: Handle) throws -> String {
        let res: ZPResult = ZPRegenerateRecovery(handle)
        return try decode(RegenerateRecoveryResponse.self, from: res).mnemonic
    }

    nonisolated static func exportJSON(handle: Handle, path: String) throws {
        let res: ZPResult = withMutableCString(path) { p in
            ZPExportJSON(handle, p)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func exportEncrypted(handle: Handle, path: String) throws {
        let res: ZPResult = withMutableCString(path) { p in
            ZPExportEncrypted(handle, p)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func exportCSV(handle: Handle, path: String) throws {
        let res: ZPResult = withMutableCString(path) { p in
            ZPExportCSV(handle, p)
        }
        _ = try dataFromResult(res)
    }

    @discardableResult
    nonisolated static func importCSV(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportCSV(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importChrome(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportChrome(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importFirefox(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportFirefox(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importSafari(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportSafari(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func import1Password(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImport1Password(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func import1PUX(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImport1PUX(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importBitwarden(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportBitwarden(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importLastPass(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportLastPass(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    @discardableResult
    nonisolated static func importKeePass(handle: Handle, path: String) throws -> Int {
        let res: ZPResult = withMutableCString(path) { p in
            ZPImportKeePass(handle, p)
        }
        return try decode(ImportResponse.self, from: res).imported
    }

    nonisolated static func syncSetup(handle: Handle, config: SyncConfig) throws {
        let d = try ZPJSON.encoder.encode(config)
        let json = String(decoding: d, as: UTF8.self)

        let res: ZPResult = withMutableCString(json) { payload in
            ZPSyncSetup(handle, payload)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func syncRegister(handle: Handle, deviceName: String) throws {
        let res: ZPResult = withMutableCString(deviceName) { name in
            ZPSyncRegister(handle, name)
        }
        _ = try dataFromResult(res)
    }

    nonisolated static func syncFull(handle: Handle) throws -> SyncFullResponse {
        let res: ZPResult = ZPSyncFull(handle)
        return try decode(SyncFullResponse.self, from: res)
    }

    nonisolated static func scorePassword(_ password: String) throws -> PasswordScore {
        let res: ZPResult = withMutableCString(password) { pw in
            ZPScorePassword(pw)
        }
        return try decode(PasswordScore.self, from: res)
    }

    nonisolated static func generatePassword(length: Int, options: GeneratePasswordOptions? = nil) throws -> String {
        let optJSON: String
        if let options {
            let d = try ZPJSON.encoder.encode(options)
            optJSON = String(decoding: d, as: UTF8.self)
        } else {
            optJSON = ""
        }

        if optJSON.isEmpty {
            let res = ZPGeneratePassword(Int32(length), nil)
            return try decode(GeneratedPassword.self, from: res).password
        }

        let res: ZPResult = withMutableCString(optJSON) { opts in
            ZPGeneratePassword(Int32(length), opts)
        }
        return try decode(GeneratedPassword.self, from: res).password
    }

    nonisolated static func generatePassphrase(words: Int, separator: String = "-") throws -> String {
        let res: ZPResult = withMutableCString(separator) { sep in
            ZPGeneratePassphrase(Int32(words), sep)
        }
        return try decode(GeneratedPassphrase.self, from: res).passphrase
    }
}
