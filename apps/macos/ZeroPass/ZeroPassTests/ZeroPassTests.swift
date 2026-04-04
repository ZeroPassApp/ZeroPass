import XCTest
import LocalAuthentication
@testable import ZeroPass

private final class FakeKeychainStore: KeychainStoring {
    var vaultKeys: [String: String] = [:]
    var syncKeys: [String: String] = [:]
    var syncStoreError: Error?

    func storeVaultKeyBase64(_ vaultKeyBase64: String, for vaultURL: URL, requireUserPresence: Bool) throws {
        vaultKeys[vaultURL.path] = vaultKeyBase64
    }

    func loadVaultKeyBase64(for vaultURL: URL, context: LAContext) throws -> String? {
        vaultKeys[vaultURL.path]
    }

    func deleteVaultKey(for vaultURL: URL) throws {
        vaultKeys.removeValue(forKey: vaultURL.path)
    }

    func storeSyncAPIKey(_ apiKey: String, for vaultURL: URL) throws {
        if let syncStoreError {
            throw syncStoreError
        }
        syncKeys[vaultURL.path] = apiKey
    }

    func loadSyncAPIKey(for vaultURL: URL) throws -> String? {
        syncKeys[vaultURL.path]
    }

    func deleteSyncAPIKey(for vaultURL: URL) throws {
        syncKeys.removeValue(forKey: vaultURL.path)
    }
}

@MainActor
final class ZeroPassTests: XCTestCase {
    private func makeTempVaultDir() throws -> URL {
        let fm = FileManager.default
        let dir = fm.temporaryDirectory.appendingPathComponent("zeropass-test-\(UUID().uuidString)")
        try fm.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir
    }

    func testUnlockWithVaultKey() throws {
        let fm = FileManager.default
        let dir = try makeTempVaultDir()
        defer { try? fm.removeItem(at: dir) }

        let pw = "test-master-password"
        let created = try ZPBridge.createVault(path: dir.path, masterPassword: pw)

        // Be explicit in case create doesn't leave the vault unlocked.
        try ZPBridge.unlock(handle: created.handle, masterPassword: pw)

        let key = try ZPBridge.getVaultKeyBase64(handle: created.handle)
        XCTAssertFalse(key.isEmpty)

        try ZPBridge.lock(handle: created.handle)
        try ZPBridge.unlockWithKey(handle: created.handle, vaultKeyBase64: key)

        let items = try ZPBridge.listItems(handle: created.handle, filter: nil)
        XCTAssertEqual(items.count, 0)

        try ZPBridge.close(handle: created.handle)
    }

    func testChangeMasterPasswordRoundTrip() throws {
        let fm = FileManager.default
        let dir = try makeTempVaultDir()
        defer { try? fm.removeItem(at: dir) }

        let oldPw = "old-master-password"
        let newPw = "new-master-password"

        let created = try ZPBridge.createVault(path: dir.path, masterPassword: oldPw)
        try ZPBridge.unlock(handle: created.handle, masterPassword: oldPw)

        try ZPBridge.changeMasterPassword(handle: created.handle, old: oldPw, new: newPw)

        try ZPBridge.lock(handle: created.handle)

        XCTAssertThrowsError(try ZPBridge.unlock(handle: created.handle, masterPassword: oldPw))
        try ZPBridge.unlock(handle: created.handle, masterPassword: newPw)

        try ZPBridge.close(handle: created.handle)
    }

    func testVersionHistoryAndRestore() throws {
        let fm = FileManager.default
        let dir = try makeTempVaultDir()
        defer { try? fm.removeItem(at: dir) }

        let pw = "master-password"
        let created = try ZPBridge.createVault(path: dir.path, masterPassword: pw)
        try ZPBridge.unlock(handle: created.handle, masterPassword: pw)

        var item = VaultItem.new(type: .login)
        item.name = "GitHub"
        item.fields = ["username": "dev", "password": "secret"]

        let createdItem = try ZPBridge.createItem(handle: created.handle, item: item)
        var updated = createdItem
        updated.name = "GitHub Updated"
        _ = try ZPBridge.updateItem(handle: created.handle, id: createdItem.id, item: updated)

        let history = try ZPBridge.versionHistory(handle: created.handle, id: createdItem.id)
        XCTAssertGreaterThanOrEqual(history.count, 1)

        if let first = history.first {
            let restored = try ZPBridge.restoreVersion(handle: created.handle, id: createdItem.id, version: Int32(first.version))
            XCTAssertEqual(restored.id, createdItem.id)
        }

        try ZPBridge.close(handle: created.handle)
    }

    func testExportJSONAndImportCSV() throws {
        let fm = FileManager.default
        let dir = try makeTempVaultDir()
        defer { try? fm.removeItem(at: dir) }

        let pw = "master-password"
        let created = try ZPBridge.createVault(path: dir.path, masterPassword: pw)
        try ZPBridge.unlock(handle: created.handle, masterPassword: pw)

        var item = VaultItem.new(type: .login)
        item.name = "Example"
        item.fields = ["url": "https://example.com", "username": "u", "password": "p"]
        _ = try ZPBridge.createItem(handle: created.handle, item: item)

        let exportPath = dir.appendingPathComponent("export.json").path
        try ZPBridge.exportJSON(handle: created.handle, path: exportPath)
        let exported = try String(contentsOfFile: exportPath, encoding: .utf8)
        XCTAssertFalse(exported.isEmpty)

        let csv = "Name,Type,URL,Username,Password,Notes\nImported,login,https://imported.example,iu,ip,hello\n"
        let csvPath = dir.appendingPathComponent("import.csv")
        try csv.data(using: .utf8)?.write(to: csvPath)

        let imported = try ZPBridge.importCSV(handle: created.handle, path: csvPath.path)
        XCTAssertEqual(imported, 1)

        let items = try ZPBridge.listItems(handle: created.handle, filter: nil)
        XCTAssertTrue(items.contains(where: { $0.name == "Imported" }))

        try ZPBridge.close(handle: created.handle)
    }

    func testSyncConfigDecodesWithoutPersistedAPIKey() throws {
        let json = #"{"server_url":"https://sync.example.com","device_id":"device-1","last_sync_time":123456}"#
        let data = try XCTUnwrap(json.data(using: .utf8))
        let cfg = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: data)

        XCTAssertEqual(cfg.serverURL, "https://sync.example.com")
        XCTAssertEqual(cfg.deviceID, "device-1")
        XCTAssertNil(cfg.apiKey)
        XCTAssertEqual(cfg.lastSyncTime, 123456)
    }

    func testMigrateLegacySyncAPIKeyMovesSecretOutOfDiskAndDefaults() throws {
        let keychain = FakeKeychainStore()
        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }
        defer {
            UserDefaults.standard.removeObject(forKey: "syncAPIKey")
        }

        let client = VaultClient(keychain: keychain)
        let configURL = client.syncConfigURL(for: dir)
        let legacyConfig = ZPBridge.SyncConfig(
            serverURL: "https://sync.example.com",
            deviceID: "device-1",
            apiKey: "disk-secret",
            lastSyncTime: 42
        )
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        try encoder.encode(legacyConfig).write(to: configURL, options: .atomic)
        UserDefaults.standard.set("defaults-secret", forKey: "syncAPIKey")

        client.migrateLegacySyncAPIKeyIfNeeded(for: dir)

        XCTAssertEqual(try keychain.loadSyncAPIKey(for: dir), "disk-secret")
        XCTAssertNil(UserDefaults.standard.string(forKey: "syncAPIKey"))

        let persisted = try Data(contentsOf: configURL)
        let migrated = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: persisted)
        XCTAssertEqual(migrated.serverURL, "https://sync.example.com")
        XCTAssertEqual(migrated.deviceID, "device-1")
        XCTAssertNil(migrated.apiKey)
        XCTAssertEqual(migrated.lastSyncTime, 42)
        XCTAssertFalse(String(decoding: persisted, as: UTF8.self).contains("disk-secret"))
    }

    func testMigrateLegacySyncAPIKeyPrefersExistingStoredSecret() throws {
        let keychain = FakeKeychainStore()
        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }
        defer { UserDefaults.standard.removeObject(forKey: "syncAPIKey") }

        let client = VaultClient(keychain: keychain)
        let configURL = client.syncConfigURL(for: dir)
        let legacyConfig = ZPBridge.SyncConfig(
            serverURL: "https://sync.example.com",
            deviceID: "device-1",
            apiKey: "disk-secret",
            lastSyncTime: 7
        )
        try JSONEncoder().encode(legacyConfig).write(to: configURL, options: .atomic)
        try keychain.storeSyncAPIKey("stored-secret", for: dir)
        UserDefaults.standard.set("defaults-secret", forKey: "syncAPIKey")

        client.migrateLegacySyncAPIKeyIfNeeded(for: dir)

        XCTAssertEqual(try keychain.loadSyncAPIKey(for: dir), "stored-secret")
        let persisted = try Data(contentsOf: configURL)
        XCTAssertFalse(String(decoding: persisted, as: UTF8.self).contains("disk-secret"))
    }

    func testMigrateLegacySyncAPIKeyKeepsLegacyCopiesWhenKeychainStoreFails() throws {
        let keychain = FakeKeychainStore()
        keychain.syncStoreError = KeychainError.unexpectedStatus(errSecAuthFailed)

        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }
        defer { UserDefaults.standard.removeObject(forKey: "syncAPIKey") }

        let client = VaultClient(keychain: keychain)
        let configURL = client.syncConfigURL(for: dir)
        let legacyConfig = ZPBridge.SyncConfig(
            serverURL: "https://sync.example.com",
            deviceID: "device-1",
            apiKey: "disk-secret",
            lastSyncTime: 7
        )
        try JSONEncoder().encode(legacyConfig).write(to: configURL, options: .atomic)
        UserDefaults.standard.set("defaults-secret", forKey: "syncAPIKey")

        client.migrateLegacySyncAPIKeyIfNeeded(for: dir)

        XCTAssertNil(try keychain.loadSyncAPIKey(for: dir))
        XCTAssertEqual(UserDefaults.standard.string(forKey: "syncAPIKey"), "defaults-secret")

        let persisted = try Data(contentsOf: configURL)
        let unmigrated = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: persisted)
        XCTAssertEqual(unmigrated.apiKey, "disk-secret")
    }

    func testMigrateLegacySyncAPIKeyIgnoresBlankStoredSecret() throws {
        let keychain = FakeKeychainStore()
        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }

        let client = VaultClient(keychain: keychain)
        let configURL = client.syncConfigURL(for: dir)
        let legacyConfig = ZPBridge.SyncConfig(
            serverURL: "https://sync.example.com",
            deviceID: "device-1",
            apiKey: "disk-secret",
            lastSyncTime: 9
        )
        try JSONEncoder().encode(legacyConfig).write(to: configURL, options: .atomic)
        try keychain.storeSyncAPIKey("   ", for: dir)

        client.migrateLegacySyncAPIKeyIfNeeded(for: dir)

        XCTAssertEqual(try keychain.loadSyncAPIKey(for: dir), "disk-secret")

        let persisted = try Data(contentsOf: configURL)
        let migrated = try JSONDecoder().decode(ZPBridge.SyncConfig.self, from: persisted)
        XCTAssertNil(migrated.apiKey)
    }

    func testOpenVaultWithoutSyncConfigKeepsStoredSyncAPIKey() async throws {
        let keychain = FakeKeychainStore()
        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }

        let created = try ZPBridge.createVault(path: dir.path, masterPassword: "master-password")
        try ZPBridge.close(handle: created.handle)
        try keychain.storeSyncAPIKey("stored-secret", for: dir)

        let client = VaultClient(keychain: keychain)
        try await client.openVault(dir)

        XCTAssertEqual(try keychain.loadSyncAPIKey(for: dir), "stored-secret")
        try await client.closeVault()
    }

    func testMigrateLegacySyncAPIKeyDoesNotAttachDefaultsSecretWithoutSyncMetadata() throws {
        let keychain = FakeKeychainStore()
        let dir = try makeTempVaultDir()
        defer { try? FileManager.default.removeItem(at: dir) }
        defer { UserDefaults.standard.removeObject(forKey: "syncAPIKey") }

        let client = VaultClient(keychain: keychain)
        UserDefaults.standard.set("defaults-secret", forKey: "syncAPIKey")

        client.migrateLegacySyncAPIKeyIfNeeded(for: dir)

        XCTAssertNil(try keychain.loadSyncAPIKey(for: dir))
        XCTAssertEqual(UserDefaults.standard.string(forKey: "syncAPIKey"), "defaults-secret")
    }
}
