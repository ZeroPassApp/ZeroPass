import XCTest
@testable import ZeroPass

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
}
