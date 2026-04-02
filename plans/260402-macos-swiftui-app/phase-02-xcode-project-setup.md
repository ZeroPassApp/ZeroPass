# Phase 2: Xcode Project + Swift Bridge Wrapper

## Context
- Depends on: [Phase 1 — Go Bridge Layer](./phase-01-go-bridge-layer.md)
- [SwiftUI macOS Research](./research/swiftui-macos-research.md)
- **Design Reference**: [UI/UX Design Guideline](./reports/ui-ux-design-guideline.md) — ALL UI phases reference this document

## Overview
- **Priority:** P0
- **Status:** In Progress
- **Description:** Create the Xcode project, integrate `libzeropass.a` static library, and build the Swift bridge wrapper that provides clean, type-safe, async Swift API over the C exports.

## Design Foundation (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

This phase establishes the app shell that ALL UI phases build upon. Key setup items:

### Window Configuration
- Default: 1024×680pt, Min: 800×500pt (set in `WindowGroup`)
- Use `@SceneStorage` for window state persistence

### macOS Menu Bar Structure (REQUIRED from Phase 2)
The app entry point (`ZeroPassApp.swift`) must define the full macOS menu bar:
```
ZeroPass  File  Edit  View  Item  Window  Help
```
Full menu structure in [Guideline §15](./reports/ui-ux-design-guideline.md#15-keyboard-shortcuts). Use `.commands { }` modifier.

### App Scenes
```swift
@main
struct ZeroPassApp: App {
    var body: some Scene {
        WindowGroup { ContentView() }
            .defaultSize(width: 1024, height: 680)
            .commands { AppCommands() }     // Full menu bar
        
        Settings { SettingsView() }         // ⌘, (Phase 8)
        
        MenuBarExtra("ZeroPass", systemImage: "lock.fill") {
            MenuBarView()                   // Phase 7
        }
        .menuBarExtraStyle(.window)
    }
}
```

### Item Type Enum (with design tokens)
```swift
enum ItemType: String, Codable, CaseIterable {
    case login, apikey, sshkey, note, creditcard, identity, passkey, custom
    
    var sfSymbol: String { /* from Guideline §6 */ }
    var color: Color { /* from Guideline §5 — blue, purple, green, etc. */ }
    var displayName: String { /* ... */ }
}
```

### Design Token Constants
Define shared constants matching [Guideline Appendix B](./reports/ui-ux-design-guideline.md#appendix-b-design-token-summary):
```swift
enum DesignTokens {
    enum Spacing { static let xs: CGFloat = 4; static let sm: CGFloat = 8; ... }
    enum CornerRadius { static let xs: CGFloat = 4; static let sm: CGFloat = 6; ... }
    enum RowHeight { static let compact: CGFloat = 24; static let standard: CGFloat = 28; ... }
    enum IconSize { static let sm: CGFloat = 14; static let md: CGFloat = 16; ... }
    enum Animation { static let fast: Double = 0.15; static let normal: Double = 0.2; ... }
}
```

## Key Insights
- Use SPM (Swift Package Manager) for dependency management, not CocoaPods
- Module map approach (like WireGuard) is cleaner than bridging header for SPM
- `BridgeResult` pattern: auto-free C memory in Swift initializer, decode JSON
- All heavy operations (KDF, search) dispatched to `Task.detached` — never block MainActor
- Swift `Codable` structs mirror Go types for automatic JSON decode

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass.xcodeproj/` | Xcode project |
| `macos/ZeroPass/ZeroPassApp.swift` | @main entry point with WindowGroup, Settings, MenuBarExtra |
| `macos/ZeroPass/Bridge/ZeroPassBridge.swift` | Core bridge: `BridgeResult`, `ZeroPassVault` class |
| `macos/ZeroPass/Bridge/BridgeTypes.swift` | All Codable structs (VaultItem, ItemFilter, HealthReport, etc.) |
| `macos/ZeroPass/Bridge/BridgeErrors.swift` | `ZeroPassError` enum (locked, notFound, authFailed, internal) |
| `macos/ZeroPass/Bridge/module.modulemap` | Module map for libzeropass |
| `macos/ZeroPass/Models/VaultManager.swift` | @Observable main state manager |
| `macos/ZeroPass/Models/AppSettings.swift` | UserDefaults-backed settings |
| `macos/ZeroPass/ContentView.swift` | Root view (locked → unlock, unlocked → main) |
| `macos/ZeroPass/Assets.xcassets` | App icon, accent color |
| `macos/ZeroPass/ZeroPass.entitlements` | Hardened Runtime entitlements |
| `macos/Package.swift` | SPM manifest (Sparkle, LaunchAtLogin) |

## Implementation Steps

1. Create Xcode project: macOS App, SwiftUI, Swift, deployment target macOS 14.0
2. Configure project structure matching `macos/ZeroPass/` layout
3. Add Xcode Build Phase script to invoke `make -C ../bridge build-universal`
4. Link `libzeropass.a` via "Link Binary With Libraries"
5. Create `module.modulemap` for `ZeroPassCore` module importing `libzeropass.h`
6. Also link system frameworks: `Security.framework`, `LocalAuthentication.framework`
7. Create `BridgeErrors.swift` — `ZeroPassError` enum mapping error codes
8. Create `BridgeResult` struct — handles C memory lifecycle + JSON decoding:
   ```swift
   struct BridgeResult {
       let data: Data?
       let error: ZeroPassError?
       init(from result: ZPResult) { /* frees C memory, parses JSON */ }
       func decode<T: Decodable>(as type: T.Type) throws -> T
   }
   ```
9. Create `ZeroPassVault` class — handle-based, wraps all C functions:
   ```swift
   class ZeroPassVault {
       let handle: Int
       func unlock(password: String) async throws
       func unlockWithKey(_ vaultKey: Data) async throws  // TouchID path — bypasses KDF
       func lock() throws
       func changeMasterPassword(old: String, new: String) async throws
       func listItems(filter: ItemFilter?) async throws -> [VaultItem]
       func getVersionHistory(itemID: String) async throws -> [ItemVersion]
       func restoreVersion(itemID: String, version: Int) async throws
       // ... all operations with async variants
   }
   ```
10. Create `BridgeTypes.swift` — Codable types:
    - `VaultItem` (mirrors `types.Item`)
    - `ItemFilter` (mirrors `types.ItemFilter`)
    - `ItemType` enum (login, apikey, sshkey, note, creditcard, identity, passkey, custom)
    - `HealthReport`, `BreachResult`, `SyncResult`, etc.
11. Create `VaultManager.swift` — @Observable:
    - `isLocked`, `items`, `searchText`, `selectedCategory`, `error`
    - Vault lifecycle methods using `ZeroPassVault`
    - `filteredItems` computed property
12. Create `AppSettings.swift` — UserDefaults wrapper for preferences
13. Create `ContentView.swift` — routes between UnlockView and MainView based on lock state
14. Create `ZeroPassApp.swift` — @main with WindowGroup, Settings scene, MenuBarExtra
15. Configure entitlements: `com.apple.security.cs.allow-unsigned-executable-memory` (for Go runtime)
16. **Critical: Test Hardened Runtime NOW** — build, sign, verify notarization passes with this entitlement.
    Do NOT defer this test to Phase 9. If notarization fails, the entire architecture must be revisited.
17. Add SPM dependencies: Sparkle 2.7+, LaunchAtLogin-Modern
18. Verify: app builds and launches, shows placeholder content
19. Configure `JSONDecoder.dateDecodingStrategy = .iso8601` — Go `time.Time` marshals as RFC3339

## Swift Type Definitions (key types)

```swift
enum ItemType: String, Codable, CaseIterable {
    case login, apikey, sshkey, note, creditcard, identity, passkey, custom
    var displayName: String { /* ... */ }
    var iconName: String { /* SF Symbol name */ }
}

struct VaultItem: Identifiable, Codable {
    let id: String
    var type: ItemType
    var name: String
    var fields: [String: String]
    var notes: String
    var tags: [String]
    var favorite: Bool
    var customFields: [String: String]
    let createdAt: Date
    var updatedAt: Date
    var lastAccessedAt: Date?
    var version: Int
}

struct ItemFilter: Codable {
    var type: ItemType?
    var tags: [String]?
    var favorite: Bool?
    var searchQuery: String?
    var sortBy: String?
    var sortOrder: String?
}
```

## Success Criteria
- [ ] Xcode project builds without errors
- [ ] Go library build phase triggers automatically on build
- [ ] `ZeroPassVault.create()` and `.open()` work from Swift
- [ ] JSON round-trip: Go Item → JSON → Swift VaultItem decodes correctly
- [ ] Memory: no leaks in BridgeResult lifecycle (Instruments check)
- [ ] App launches and shows ContentView

## Risk Assessment
- **Hardened Runtime + Go:** Need `allow-unsigned-executable-memory` — **test notarization in THIS phase, not Phase 9**
- **Build phase ordering:** Go build must complete before Swift compile links
- **JSON date format:** Go `time.Time` marshals as RFC3339; configure Swift `JSONDecoder.dateDecodingStrategy = .iso8601`
- **BridgeResult mnemonic handling:** When decoding `ZPCreateVault` result, mnemonic string must be zeroed after display. Use `Data` + `UnsafeMutableBufferPointer`, not plain Swift `String`
