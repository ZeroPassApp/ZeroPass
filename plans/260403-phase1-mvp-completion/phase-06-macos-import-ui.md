# Phase 6: macOS App — Import UI Expansion (Full Sources)

## Context

- **Priority:** MEDIUM | **Effort:** 1h | **Risk:** Low
- SecuritySettingsView.swift has only "Import CSV…" button
- PRD requires import from Chrome/Firefox/Safari/1Password/Bitwarden/LastPass/KeePass via GUI
- Bridge layer already exports all 9 import sources (`ZPImportChrome`, `ZPImportSafari`, etc.)
- VaultClient.swift only wraps `importCSV()` — other 8 sources not exposed

## Related Files

| Action | File |
|--------|------|
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift` |
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` |

## Implementation Steps

### 1. Replace single button with dropdown menu in SecuritySettingsView

Replace:
```swift
Button("Import CSV…") {
    importCSV()
}
.disabled(!vault.isUnlocked)
```

With:
```swift
Menu("Import from…") {
    Button("Chrome CSV…") { importFrom(.chrome) }
    Button("Firefox CSV…") { importFrom(.firefox) }
    Button("Safari CSV…") { importFrom(.safari) }
    Divider()
    Button("1Password CSV…") { importFrom(.onePassword) }
    Button("1Password 1PUX…") { importFrom(.onePasswordPUX) }
    Button("Bitwarden CSV…") { importFrom(.bitwarden) }
    Button("LastPass CSV…") { importFrom(.lastPass) }
    Button("KeePass CSV…") { importFrom(.keepass) }
    Divider()
    Button("Generic CSV…") { importFrom(.csv) }
}
.disabled(!vault.isUnlocked)
```

### 2. Add ImportSource enum

```swift
enum ImportSource: String, CaseIterable {
    case chrome, firefox, safari
    case onePassword, onePasswordPUX, bitwarden, lastPass, keepass
    case csv

    var fileTypes: [UTType] {
        switch self {
        case .onePasswordPUX: return [UTType(filenameExtension: "1pux")].compactMap { $0 }
        default: return [.commaSeparatedText]
        }
    }

    var label: String {
        switch self {
        case .chrome: return "Chrome"
        case .firefox: return "Firefox"
        case .safari: return "Safari"
        case .onePassword: return "1Password"
        case .onePasswordPUX: return "1Password 1PUX"
        case .bitwarden: return "Bitwarden"
        case .lastPass: return "LastPass"
        case .keepass: return "KeePass"
        case .csv: return "Generic CSV"
        }
    }
}
```

### 3. Add `importFrom(_:)` handler

```swift
private func importFrom(_ source: ImportSource) {
    let panel = NSOpenPanel()
    panel.allowedContentTypes = source.fileTypes
    panel.allowsMultipleSelection = false
    panel.message = "Select \(source.label) export file"

    guard panel.runModal() == .OK, let url = panel.url else { return }

    Task {
        do {
            let count = try await vault.importFile(from: url, source: source)
            // Show success notification
        } catch {
            // Show error alert
        }
    }
}
```

### 4. Add VaultClient import methods

Add `importFile(from:source:)` dispatcher in VaultClient:

```swift
@discardableResult
func importFile(from url: URL, source: ImportSource) async throws -> Int {
    guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }
    let count = try await Task.detached(priority: .utility) {
        switch source {
        case .chrome:          return try ZPBridge.importChrome(handle: h, path: url.path)
        case .firefox:         return try ZPBridge.importFirefox(handle: h, path: url.path)
        case .safari:          return try ZPBridge.importSafari(handle: h, path: url.path)
        case .onePassword:     return try ZPBridge.import1Password(handle: h, path: url.path)
        case .onePasswordPUX:  return try ZPBridge.import1PUX(handle: h, path: url.path)
        case .bitwarden:       return try ZPBridge.importBitwarden(handle: h, path: url.path)
        case .lastPass:        return try ZPBridge.importLastPass(handle: h, path: url.path)
        case .keepass:         return try ZPBridge.importKeePass(handle: h, path: url.path)
        case .csv:             return try ZPBridge.importCSV(handle: h, path: url.path)
        }
    }.value
    autoLock.recordActivity()
    try await refreshItems()
    return count
}
```

## Todo List

- [x] Add ImportSource enum in SecuritySettingsView or shared types
- [x] Replace "Import CSV…" button with Menu dropdown
- [x] Add importFrom(_:) handler with NSOpenPanel
- [x] Add VaultClient.importFile(from:source:) dispatcher
- [x] Verify ZPBridge Swift wrappers exist for all 9 import sources
- [x] Xcode build succeeds

## Success Criteria

- Settings → Import from… shows dropdown with all 9 sources grouped logically
- Each source opens file picker with correct file type filter
- Import succeeds and items appear in vault
- Disabled when vault is locked
