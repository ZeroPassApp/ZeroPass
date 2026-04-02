Agent is idle (waiting for messages). agent_id: swiftui-macos-research, agent_type: researcher, status: idle, description: Research SwiftUI macOS app, elapsed: 569s, total_turns: 1

[Turn 0]
I now have comprehensive research across all requested topics. Here is the complete report:

---

# ZeroPass macOS Desktop App — Comprehensive Research Report

**Date:** 2026-04-02  
**Target:** Native macOS SwiftUI app for ZeroPass password manager  
**Environment:** macOS 26.3, Swift 6.3, Go 1.26.1, Apple Silicon (arm64)

---

## 1. SwiftUI macOS App Architecture (2024-2026 Best Practices)

### 1.1 App Lifecycle

Modern SwiftUI uses `@main` App protocol — no AppDelegate needed for most cases:

```swift
@main
struct ZeroPassApp: App {
    @State private var vaultManager = VaultManager()
    
    var body: some Scene {
        // Main vault browser window
        WindowGroup {
            ContentView()
                .environment(vaultManager)
        }
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("New Item") { vaultManager.showNewItem = true }
                    .keyboardShortcut("n", modifiers: .command)
            }
            CommandMenu("Vault") {
                Button("Lock Vault") { vaultManager.lock() }
                    .keyboardShortcut("l", modifiers: [.command, .shift])
                Button("Quick Search") { vaultManager.showQuickSearch = true }
                    .keyboardShortcut("k", modifiers: .command)
            }
        }
        
        // Settings window (⌘,)
        Settings {
            SettingsView()
                .environment(vaultManager)
        }
        
        // Menu bar extra
        MenuBarExtra("ZeroPass", systemImage: "lock.shield.fill") {
            MenuBarView()
                .environment(vaultManager)
        }
        .menuBarExtraStyle(.window) // Popover-style, not just menu
        
        // Quick search floating window
        Window("Quick Search", id: "quick-search") {
            QuickSearchView()
                .environment(vaultManager)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 600, height: 400)
    }
}
```

### 1.2 NavigationSplitView (3-column layout) — IDEAL for password managers

```swift
struct ContentView: View {
    @Environment(VaultManager.self) var vault
    @State private var selectedCategory: SidebarCategory? = .allItems
    @State private var selectedItemID: String?
    @State private var columnVisibility: NavigationSplitViewVisibility = .all
    
    var body: some View {
        NavigationSplitView(columnVisibility: $columnVisibility) {
            // Column 1: Sidebar (categories/tags)
            SidebarView(selection: $selectedCategory)
        } content: {
            // Column 2: Item list
            ItemListView(
                category: selectedCategory,
                selectedItemID: $selectedItemID
            )
        } detail: {
            // Column 3: Item detail/editor
            if let id = selectedItemID {
                ItemDetailView(itemID: id)
            } else {
                ContentUnavailableView("Select an Item", 
                    systemImage: "key.fill",
                    description: Text("Choose an item from the list"))
            }
        }
        .searchable(text: $vault.searchText, prompt: "Search vault...")
        .navigationTitle("ZeroPass")
        .toolbar { MainToolbar() }
    }
}
```

### 1.3 @Observable macro (Swift 5.9+) — THE modern approach

**Use `@Observable` instead of `ObservableObject`**. No `@Published` needed. Automatic fine-grained tracking:

```swift
@Observable
class VaultManager {
    // State
    var isLocked: Bool = true
    var items: [VaultItem] = []
    var searchText: String = ""
    var selectedCategory: SidebarCategory? = .allItems
    var showNewItem: Bool = false
    var showQuickSearch: Bool = false
    var error: AppError?
    
    // Computed
    var filteredItems: [VaultItem] {
        items.filter { item in
            if searchText.isEmpty { return true }
            return item.name.localizedCaseInsensitiveContains(searchText)
        }
    }
    
    // Bridge to Go
    private let bridge: GoBridge
    
    init(bridge: GoBridge = .shared) {
        self.bridge = bridge
    }
    
    func unlock(password: String) async throws {
        try bridge.unlock(password: password)
        isLocked = false
        items = try bridge.listItems()
    }
    
    func lock() {
        bridge.lock()
        isLocked = true
        items = []
    }
}
```

**Usage in views** — cleaner than `@ObservedObject`:
```swift
@Environment(VaultManager.self) var vault  // Inject via .environment()
// No @ObservedObject, @StateObject, @EnvironmentObject needed
```

### 1.4 macOS-specific Views

**Table** (macOS native, sortable):
```swift
Table(vault.filteredItems, selection: $selectedItemID, sortOrder: $sortOrder) {
    TableColumn("Name", value: \.name) { item in
        Label(item.name, systemImage: item.type.iconName)
    }
    TableColumn("Type", value: \.type.rawValue) { item in
        Text(item.type.displayName)
    }
    TableColumn("Username") { item in
        Text(item.fields["username"] ?? "")
    }
    TableColumn("Updated", value: \.updatedAt) { item in
        Text(item.updatedAt, style: .relative)
    }
}
.contextMenu(forSelectionType: String.self) { ids in
    Button("Copy Password") { copyPassword(ids.first) }
    Button("Delete", role: .destructive) { deleteItems(ids) }
}
```

**Form** (Settings):
```swift
Form {
    Section("Security") {
        Picker("Auto-lock after", selection: $settings.autoLockTimeout) {
            Text("1 minute").tag(60)
            Text("5 minutes").tag(300)
            Text("15 minutes").tag(900)
            Text("Never").tag(0)
        }
        Stepper("Clipboard clear: \(settings.clipboardClearSec)s", 
                value: $settings.clipboardClearSec, in: 5...120)
        Toggle("Lock on sleep", isOn: $settings.lockOnSleep)
        Toggle("Unlock with TouchID", isOn: $settings.useBiometrics)
    }
    Section("Appearance") {
        Picker("Theme", selection: $settings.theme) { ... }
    }
}
.formStyle(.grouped)
```

### 1.5 SwiftData vs Core Data

**Neither needed.** ZeroPass already has its own encrypted storage (Go vault). The macOS app is a **view layer only** — all data lives in the Go vault (`~/.zeropass/`). No local caching DB needed.

If caching becomes necessary (e.g., decrypted item metadata for fast search), use in-memory only — never write decrypted data to disk from the Swift side.

### 1.6 macOS-only App Structure

```
ZeroPass/
├── Package.swift              # SPM manifest
├── Sources/
│   └── ZeroPass/
│       ├── ZeroPassApp.swift  # @main entry
│       ├── Bridge/
│       │   ├── GoBridge.swift        # Swift wrapper around Go C API
│       │   ├── GoTypes.swift         # Codable structs matching Go types
│       │   └── libzeropass.h         # Generated C header
│       ├── Models/
│       │   ├── VaultManager.swift    # @Observable main state
│       │   ├── VaultItem.swift       # Swift model (decoded from Go JSON)
│       │   └── Settings.swift        # App settings
│       ├── Views/
│       │   ├── ContentView.swift     # NavigationSplitView root
│       │   ├── Sidebar/
│       │   │   └── SidebarView.swift
│       │   ├── ItemList/
│       │   │   ├── ItemListView.swift
│       │   │   └── ItemRowView.swift
│       │   ├── ItemDetail/
│       │   │   ├── ItemDetailView.swift
│       │   │   ├── LoginDetailView.swift
│       │   │   ├── APIKeyDetailView.swift
│       │   │   └── FieldRow.swift
│       │   ├── Editors/
│       │   │   ├── NewItemSheet.swift
│       │   │   └── EditItemView.swift
│       │   ├── Auth/
│       │   │   ├── UnlockView.swift
│       │   │   └── SetupView.swift
│       │   ├── Settings/
│       │   │   └── SettingsView.swift
│       │   ├── MenuBar/
│       │   │   └── MenuBarView.swift
│       │   └── QuickSearch/
│       │       └── QuickSearchView.swift
│       ├── Services/
│       │   ├── KeychainService.swift  # Keychain API wrapper
│       │   ├── BiometricService.swift # TouchID/FaceID
│       │   ├── ClipboardService.swift # Copy + auto-clear
│       │   └── HotkeyService.swift   # Global keyboard shortcut
│       └── Utilities/
│           ├── Extensions.swift
│           └── Constants.swift
├── Resources/
│   └── Assets.xcassets
└── Frameworks/
    └── libzeropass.dylib          # Go shared library
```

### 1.7 Window Management

```swift
// Open specific windows programmatically
@Environment(\.openWindow) var openWindow

Button("Quick Search") {
    openWindow(id: "quick-search")
}

// Multiple window types via Window scenes
Window("Password Generator", id: "generator") {
    PasswordGeneratorView()
}
.defaultSize(width: 400, height: 300)
.windowResizability(.contentSize)
```

### 1.8 Keyboard Shortcuts & Accessibility

```swift
// In-app shortcuts
.keyboardShortcut("c", modifiers: [.command, .shift])  // Copy password
.keyboardShortcut("e", modifiers: .command)             // Edit item
.keyboardShortcut(.delete, modifiers: .command)         // Delete item

// Menu bar keyboard shortcut  
// Global shortcut requires AppKit: NSEvent.addGlobalMonitorForEvents
// Or: HotKey library (github.com/soffes/HotKey) for user-configurable global shortcut

// Accessibility
.accessibilityLabel("Password field")
.accessibilityHint("Double-tap to copy")
.accessibilityAddTraits(.isSecureTextField)
```

---

## 2. Go ↔ Swift Interoperability — Full Analysis

### Comparison Matrix

| Approach | Complexity | Latency | Security | Persistent State | Sandbox Compatible |
|----------|-----------|---------|----------|-----------------|-------------------|
| **A: CLI subprocess** | ★☆☆ Low | 50-100ms/call | ★★★ Process isolation | ✗ No (each call fresh) | ✗ No |
| **B: Go .dylib** | ★★☆ Medium | <1ms/call | ★★☆ In-process | ✓ Yes (Go globals) | ✓ Yes |
| **C: Go daemon (Unix socket)** | ★★★ High | 1-5ms/call | ★★★ Process isolation | ✓ Yes | ✗ No |
| **D: XPC Service** | ★★★★ Very High | 1-5ms/call | ★★★★ Best | ✓ Yes | ✓ Yes |
| **E: Swift rewrite** | ★★★★★ Extreme | <0.1ms | ★★★★ Native | ✓ Yes | ✓ Yes |

### ⭐ RECOMMENDATION: Option B (Go .dylib) — Best balance

**Proven working** (verified in this research). Here's why:

1. **Near-zero latency**: Function calls are direct C ABI calls (<1ms)
2. **Persistent state**: Go runtime stays loaded; vault key stays in memory
3. **Existing code reuse**: All Go crypto/vault/index code works as-is
4. **Single process**: No daemon management, no socket cleanup
5. **Reasonable complexity**: ~25 `//export` functions + Swift wrapper
6. **Proven pattern**: Rust+Swift (1Password), C+Swift (KeePassXC) use same concept

**Verified working bridge pattern:**

```go
// go/bridge/bridge.go
package main

/*
#include <stdlib.h>
*/
import "C"
import (
    "encoding/json"
    "sync"
    "unsafe"
    
    "github.com/zeropass/zeropass/core/vault/store"
    "github.com/zeropass/zeropass/core/vault/item"
    "github.com/zeropass/zeropass/core/vault/index"
)

var (
    mu      sync.Mutex
    vault   *store.Vault
    mgr     *item.Manager
    idx     *index.Index
)

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func jsonResponse(data interface{}, err error) *C.char {
    r := Response{Success: err == nil, Data: data}
    if err != nil { r.Error = err.Error() }
    b, _ := json.Marshal(r)
    return C.CString(string(b))
}

//export ZPFree
func ZPFree(s *C.char) { C.free(unsafe.Pointer(s)) }

//export ZPOpen
func ZPOpen(path *C.char) *C.char {
    mu.Lock()
    defer mu.Unlock()
    v, err := store.Open(C.GoString(path))
    if err != nil { return jsonResponse(nil, err) }
    vault = v
    return jsonResponse(map[string]bool{"opened": true}, nil)
}

//export ZPUnlock
func ZPUnlock(password *C.char) *C.char {
    mu.Lock()
    defer mu.Unlock()
    err := vault.Unlock(C.GoString(password))
    if err != nil { return jsonResponse(nil, err) }
    // Initialize item manager + index
    var idxErr error
    idx, idxErr = index.Open(vault.IndexPath())
    if idxErr != nil { return jsonResponse(nil, idxErr) }
    mgr = item.NewManager(vault.ItemsPath(), vault.VaultKey, idx)
    return jsonResponse(map[string]bool{"unlocked": true}, nil)
}

//export ZPListItems
func ZPListItems(filterJSON *C.char) *C.char {
    mu.Lock()
    defer mu.Unlock()
    // Parse filter, call mgr.ListItems(), return JSON
    items, err := mgr.AllItems()
    return jsonResponse(items, err)
}

//export ZPGetItem
func ZPGetItem(id *C.char) *C.char {
    mu.Lock()
    defer mu.Unlock()
    itm, err := mgr.GetItem(C.GoString(id))
    return jsonResponse(itm, err)
}

// ... ~20 more exported functions following same pattern

func main() {} // Required for c-shared
```

**Swift wrapper:**

```swift
// Bridge/GoBridge.swift
final class GoBridge {
    static let shared = GoBridge()
    
    struct BridgeResponse<T: Decodable>: Decodable {
        let success: Bool
        let data: T?
        let error: String?
    }
    
    private func call<T: Decodable>(_ cString: UnsafeMutablePointer<CChar>?) throws -> T {
        guard let cStr = cString else { throw BridgeError.nullResponse }
        defer { ZPFree(cStr) }
        let json = String(cString: cStr)
        let response = try JSONDecoder().decode(BridgeResponse<T>.self, 
                                                 from: Data(json.utf8))
        guard response.success, let data = response.data else {
            throw BridgeError.goError(response.error ?? "Unknown error")
        }
        return data
    }
    
    func unlock(password: String) throws {
        let _: [String: Bool] = try call(ZPUnlock(password.cString(using: .utf8)))
    }
    
    func listItems() throws -> [VaultItem] {
        return try call(ZPListItems(nil))
    }
    
    func getItem(id: String) throws -> VaultItem {
        return try call(ZPGetItem(id.cString(using: .utf8)))
    }
}
```

**Build command:**
```bash
cd go/bridge && go build -buildmode=c-shared -o ../../Frameworks/libzeropass.dylib
# Sign: codesign --sign "Developer ID Application: ..." libzeropass.dylib
```

### Why NOT the other options:

**Option A (CLI subprocess):** Can't maintain unlock state. Each `zeropass get X` requires re-entering password. Would need session management (new feature) or daemon mode (becomes Option C). Adds 50-100ms latency per operation.

**Option C (Go daemon):** Works but requires lifecycle management (start daemon on app launch, stop on quit, handle crashes, socket cleanup). More moving parts. No advantage over dylib for single-app use.

**Option D (XPC):** Go can't implement NSXPCListenerDelegate. Would need Obj-C/Swift wrapper around Go process — triple the complexity for marginal benefit.

**Option E (Swift rewrite):** Duplicates all crypto, vault, index code. Must maintain two implementations. Defeats purpose of shared Go core. Only makes sense if abandoning Go entirely.

---

## 3. macOS Security APIs & Keychain

### 3.1 Keychain Storage for Vault Key

**Verified working** — store the vault key (or a derived unlock token) in Keychain for biometric unlock:

```swift
final class KeychainService {
    private static let service = "dev.zeropass.vault"
    
    /// Store vault key after master password unlock (enables biometric re-unlock)
    static func storeVaultKey(_ key: Data, forVault vaultID: String) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "vault-key-\(vaultID)",
            kSecValueData as String: key,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly
            // ↑ Requires device passcode, never backed up, never migrated
        ]
        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.storeFailed(status)
        }
    }
    
    /// Store vault key with biometric protection (TouchID required to retrieve)
    static func storeVaultKeyBiometric(_ key: Data, forVault vaultID: String) throws {
        var error: Unmanaged<CFError>?
        guard let access = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly,
            [.biometryCurrentSet],  // Invalidated if fingerprints change
            &error
        ) else {
            throw KeychainError.accessControlFailed
        }
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "vault-key-bio-\(vaultID)",
            kSecValueData as String: key,
            kSecAttrAccessControl as String: access
        ]
        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.storeFailed(status)
        }
    }
    
    /// Retrieve vault key (triggers TouchID if biometric-protected)
    static func retrieveVaultKey(forVault vaultID: String, biometric: Bool) -> Data? {
        let account = biometric ? "vault-key-bio-\(vaultID)" : "vault-key-\(vaultID)"
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        return status == errSecSuccess ? result as? Data : nil
    }
}
```

### 3.2 Keychain Access Policies (which to use)

| Policy | When Available | Backup | Best For |
|--------|---------------|--------|----------|
| `kSecAttrAccessibleWhenUnlocked` | Device unlocked | ✓ iCloud | General data |
| `kSecAttrAccessibleWhenUnlockedThisDeviceOnly` | Device unlocked | ✗ | Sensitive, no migration |
| **`kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly`** | Device unlocked + passcode set | ✗ | **⭐ Vault keys** |
| `kSecAttrAccessibleAfterFirstUnlock` | After first unlock | ✓ | Background services |

**Recommendation:** `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` — data deleted if user removes passcode. Never backed up. Never migrated. Most secure option.

### 3.3 TouchID/Biometric Authentication

```swift
final class BiometricService {
    static func canUseBiometrics() -> Bool {
        let context = LAContext()
        var error: NSError?
        return context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: &error)
    }
    
    static func authenticate() async throws -> Bool {
        let context = LAContext()
        context.localizedReason = "Unlock ZeroPass vault"
        context.localizedCancelTitle = "Use Master Password"
        
        return try await context.evaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics,
            localizedReason: "Unlock your ZeroPass vault"
        )
    }
}
```

### 3.4 Secure Enclave

**Can be used** for protecting vault key with hardware-bound encryption:

```swift
// Create Secure Enclave P256 key pair
// Use it to encrypt/decrypt the vault key
// Key never leaves Secure Enclave
// Requires biometric auth to use
```

**However:** Not needed for MVP. Biometric-protected Keychain is sufficient. Secure Enclave adds complexity (P256 only, can't do AES directly). Consider for Phase 2+.

### 3.5 Hardened Runtime & Notarization

**Hardened Runtime** (required for notarization):
- Enabled in Xcode: Signing & Capabilities → Hardened Runtime
- Go dylib needs `com.apple.security.cs.disable-library-validation` OR sign dylib with same team ID
- Go runtime may need `com.apple.security.cs.allow-unsigned-executable-memory`

**Notarization flow:**
```bash
# 1. Build and sign
xcodebuild -scheme ZeroPass -configuration Release archive
codesign --deep --force --sign "Developer ID Application: ..." ZeroPass.app
codesign --force --sign "Developer ID Application: ..." ZeroPass.app/Contents/Frameworks/libzeropass.dylib

# 2. Create DMG
hdiutil create -volname "ZeroPass" -srcfolder ZeroPass.app -ov ZeroPass.dmg

# 3. Notarize
xcrun notarytool submit ZeroPass.dmg --apple-id "$APPLE_ID" --team-id "$TEAM_ID" --password "$APP_SPECIFIC_PASSWORD" --wait

# 4. Staple
xcrun stapler staple ZeroPass.dmg
```

---

## 4. macOS App Distribution

### Recommended: Direct Distribution + Homebrew Cask

| Channel | Sandbox | Auto-Update | Effort | Best For |
|---------|---------|-------------|--------|----------|
| **Direct (DMG)** | No | Sparkle | Medium | Developer audience ⭐ |
| **Homebrew Cask** | No | brew upgrade | Low | Developer audience ⭐ |
| Mac App Store | Required | Built-in | High | Consumer audience |

**Why NOT Mac App Store for ZeroPass:**
1. Sandbox blocks access to `~/.zeropass` vault directory
2. Sandbox blocks spawning CLI subprocess
3. Sandbox blocks loading Go `.dylib` (without significant entitlement workarounds)
4. App Review may reject password managers that don't use iCloud Keychain
5. Developer audience prefers `brew install` anyway

### Sparkle Auto-Update Setup

```swift
// Package.swift dependency
.package(url: "https://github.com/sparkle-project/Sparkle.git", from: "2.7.0")

// In SettingsView
import Sparkle

struct SettingsView: View {
    private let updaterController = SPUStandardUpdaterController(
        startingUpdater: true, updaterDelegate: nil, userDriverDelegate: nil)
    
    var body: some View {
        Form {
            Section("Updates") {
                CheckForUpdatesView(updater: updaterController.updater)
                Toggle("Automatically check for updates", 
                       isOn: Binding(
                           get: { updaterController.updater.automaticallyChecksForUpdates },
                           set: { updaterController.updater.automaticallyChecksForUpdates = $0 }
                       ))
            }
        }
    }
}
```

### Homebrew Cask Formula

```ruby
cask "zeropass" do
  version "1.0.0"
  sha256 "abc123..."
  
  url "https://github.com/zeropass/zeropass/releases/download/v#{version}/ZeroPass-#{version}.dmg"
  name "ZeroPass"
  desc "Developer-first zero-knowledge credential manager"
  homepage "https://github.com/zeropass/zeropass"
  
  app "ZeroPass.app"
  binary "#{appdir}/ZeroPass.app/Contents/Resources/zeropass"  # CLI tool bundled
  
  zap trash: [
    "~/Library/Preferences/dev.zeropass.plist",
    "~/Library/Application Support/ZeroPass",
  ]
end
```

---

## 5. Swift Libraries — Final Recommendations

### Essential (include from day 1)

| Library | Version | Purpose |
|---------|---------|---------|
| **Sparkle** | 2.7+ | Auto-updates for direct distribution |
| **LaunchAtLogin-Modern** | 1.0+ | Start at login (SMAppService) |

### Nice-to-have (add when needed)

| Library | Version | Purpose |
|---------|---------|---------|
| **Defaults** | 9.0+ | Type-safe UserDefaults |
| **HotKey** | 0.2+ | Global keyboard shortcut (⌘⇧P) |

### Skip

| Library | Reason |
|---------|--------|
| KeychainAccess | Only ~10 Keychain calls; raw Security.framework is fine |
| SwiftUI-Introspect | SwiftUI macOS 14+ sufficient for our needs |
| swift-argument-parser | macOS app doesn't need CLI args |
| gRPC-Swift / SwiftProtobuf | Using dylib bridge, not gRPC |
| SwiftData / Core Data | No local DB needed; Go vault IS the database |

### Open Source References to Study

| Project | Why |
|---------|-----|
| **Secretive** (maxgoedjen/secretive) | SwiftUI macOS + Secure Enclave + security |
| **Strongbox** (strongbox-password-safe) | SwiftUI native password manager |

---

## 6. Password Manager UX Patterns

### 6.1 Three-Pane Layout (industry standard)

```
┌──────────────┬───────────────────────┬──────────────────────┐
│  SIDEBAR     │  ITEM LIST            │  ITEM DETAIL         │
│              │                       │                      │
│  ★ Favorites │  🔍 Search...         │  GitHub Login        │
│  🔑 Logins   │                       │                      │
│  🔧 API Keys │  GitHub     login     │  Username: octocat   │
│  🔐 SSH Keys │  AWS API    apikey    │  Password: ●●●●●● 📋 │
│  📝 Notes    │  Stripe     apikey    │  URL: github.com     │
│  💳 Cards    │  Server     sshkey    │  TOTP: 123 456  ⏱   │
│  👤 Identity │  Gmail      login     │                      │
│              │  ...                  │  Notes: Main account │
│  TAGS        │                       │                      │
│  #work       │                       │  Tags: work, dev     │
│  #personal   │                       │  Updated: 2 min ago  │
│  #servers    │                       │                      │
│              │  47 items             │  [Edit] [Delete]     │
└──────────────┴───────────────────────┴──────────────────────┘
```

### 6.2 Sidebar Categories

```swift
enum SidebarCategory: String, CaseIterable {
    case allItems = "All Items"
    case favorites = "Favorites"
    case logins = "Logins"
    case apiKeys = "API Keys"
    case sshKeys = "SSH Keys"
    case secureNotes = "Secure Notes"
    case creditCards = "Credit Cards"
    case identities = "Identities"
    
    var icon: String {
        switch self {
        case .allItems: return "tray.full"
        case .favorites: return "star.fill"
        case .logins: return "key.fill"
        case .apiKeys: return "wrench.and.screwdriver"
        case .sshKeys: return "terminal"
        case .secureNotes: return "note.text"
        case .creditCards: return "creditcard"
        case .identities: return "person.crop.circle"
        }
    }
}
```

### 6.3 Quick Search (⌘K pattern)

Spotlight-style floating panel — all modern tools use this (1Password: ⌘⇧Space, Raycast: ⌥Space):

```swift
struct QuickSearchView: View {
    @Environment(VaultManager.self) var vault
    @State private var query = ""
    @State private var results: [VaultItem] = []
    @FocusState private var isSearchFocused: Bool
    
    var body: some View {
        VStack(spacing: 0) {
            // Search field
            HStack {
                Image(systemName: "magnifyingglass")
                TextField("Search vault...", text: $query)
                    .textFieldStyle(.plain)
                    .font(.title2)
                    .focused($isSearchFocused)
                    .onSubmit { selectFirst() }
            }
            .padding()
            
            Divider()
            
            // Results
            List(results, selection: $selectedID) { item in
                HStack {
                    Image(systemName: item.type.iconName)
                    VStack(alignment: .leading) {
                        Text(item.name).font(.body)
                        Text(item.subtitle).font(.caption).foregroundStyle(.secondary)
                    }
                    Spacer()
                    Text("⏎ Copy password").font(.caption2).foregroundStyle(.tertiary)
                }
            }
        }
        .frame(width: 600, height: 400)
        .onAppear { isSearchFocused = true }
        .onChange(of: query) { _, new in
            results = vault.search(new)
        }
    }
}
```

### 6.4 Clipboard Management

```swift
final class ClipboardService {
    private static var clearTimer: Timer?
    
    static func copy(_ text: String, clearAfter seconds: Int = 30) {
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(text, forType: .string)
        
        // Cancel previous timer
        clearTimer?.invalidate()
        
        // Schedule auto-clear
        if seconds > 0 {
            clearTimer = Timer.scheduledTimer(withTimeInterval: TimeInterval(seconds), 
                                              repeats: false) { _ in
                // Only clear if clipboard still has our content
                if pasteboard.string(forType: .string) == text {
                    pasteboard.clearContents()
                }
            }
        }
    }
}
```

### 6.5 Menu Bar Widget

```swift
struct MenuBarView: View {
    @Environment(VaultManager.self) var vault
    @State private var searchText = ""
    
    var body: some View {
        VStack(spacing: 8) {
            if vault.isLocked {
                Button("Unlock Vault") { vault.showUnlockWindow() }
            } else {
                TextField("Search...", text: $searchText)
                    .textFieldStyle(.roundedBorder)
                
                ForEach(vault.recentItems.prefix(5)) { item in
                    Button {
                        ClipboardService.copy(item.password)
                    } label: {
                        Label(item.name, systemImage: item.type.iconName)
                    }
                }
                
                Divider()
                
                Button("Open ZeroPass") { NSApp.activate(ignoringOtherApps: true) }
                    .keyboardShortcut("o")
                Button("Lock Vault") { vault.lock() }
                    .keyboardShortcut("l")
            }
            
            Divider()
            Button("Quit") { NSApp.terminate(nil) }
                .keyboardShortcut("q")
        }
        .padding(8)
        .frame(width: 280)
    }
}
```

### 6.6 Dark/Light Mode

SwiftUI handles automatically. Use semantic colors:
```swift
Color.primary         // Adapts to dark/light
Color.secondary       // Adapts
.background(.regularMaterial)  // Vibrancy
.tint(.accentColor)   // User's system accent color
```

---

## 7. Recommended Architecture — Final Decision

### Architecture: SwiftUI + Go .dylib Bridge

```
┌─────────────────────────────────────────────┐
│  ZeroPass.app (SwiftUI)                     │
│  ┌───────────────────────────────────────┐  │
│  │  UI Layer (SwiftUI Views)             │  │
│  │  NavigationSplitView, Table, Forms    │  │
│  └────────────────┬──────────────────────┘  │
│                   │                          │
│  ┌────────────────▼──────────────────────┐  │
│  │  VaultManager (@Observable)           │  │
│  │  State management, UI logic           │  │
│  └────────────────┬──────────────────────┘  │
│                   │                          │
│  ┌────────────────▼──────────────────────┐  │
│  │  GoBridge (Swift)                     │  │
│  │  JSON encode/decode, C function calls │  │
│  └────────────────┬──────────────────────┘  │
│                   │ C ABI                    │
│  ┌────────────────▼──────────────────────┐  │
│  │  libzeropass.dylib (Go)               │  │
│  │  All crypto, vault, search, import    │  │
│  │  Existing Go core — zero rewrite      │  │
│  └───────────────────────────────────────┘  │
│                                              │
│  ┌───────────────────────────────────────┐  │
│  │  Native macOS Services (Swift)        │  │
│  │  Keychain, TouchID, Clipboard,        │  │
│  │  Notifications, MenuBarExtra          │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
         │
         ▼
  ~/.zeropass/vaults/default/   (shared with CLI)
```

### Key Design Decisions

1. **Go .dylib for core** — reuse 100% of existing Go crypto/vault/index code
2. **Swift for platform integration** — Keychain, TouchID, clipboard, MenuBar, global shortcuts
3. **JSON bridge** — all data crosses Go↔Swift boundary as JSON (simple, debuggable, type-safe on both sides via Codable)
4. **No sandbox** — direct distribution via DMG + Homebrew, enables full filesystem access to shared vault
5. **Shared vault** — CLI and GUI operate on same `~/.zeropass` directory
6. **@Observable + NavigationSplitView** — modern SwiftUI, no legacy patterns

### Build Pipeline

```bash
# 1. Build Go dylib
cd go/bridge && go build -buildmode=c-shared -o ../../macos/Frameworks/libzeropass.dylib

# 2. Build Swift app
cd macos && xcodebuild -scheme ZeroPass -configuration Release

# 3. Sign everything
codesign --sign "Developer ID..." ZeroPass.app/Contents/Frameworks/libzeropass.dylib
codesign --deep --sign "Developer ID..." ZeroPass.app

# 4. Create DMG
create-dmg ZeroPass.app

# 5. Notarize
xcrun notarytool submit ZeroPass.dmg ...
xcrun stapler staple ZeroPass.dmg
```

### Estimated Effort

| Component | Effort | Priority |
|-----------|--------|----------|
| Go bridge layer (~25 exported functions) | 2-3 days | P0 |
| Swift GoBridge wrapper + types | 1-2 days | P0 |
| Unlock/Setup views | 1-2 days | P0 |
| NavigationSplitView (3-pane) | 2-3 days | P0 |
| Item list + detail views (8 types) | 3-4 days | P0 |
| New/Edit item sheets | 2-3 days | P0 |
| Search (⌘K) | 1 day | P0 |
| Settings view | 1 day | P1 |
| Keychain + TouchID integration | 1-2 days | P1 |
| MenuBarExtra | 1 day | P1 |
| Clipboard auto-clear | 0.5 day | P1 |
| Global keyboard shortcut | 1 day | P2 |
| Sparkle auto-update | 0.5 day | P2 |
| DMG + notarization pipeline | 1 day | P2 |
| **Total** | **~18-24 days** | — |

---

## Unresolved Questions

1. **Vault locking concurrency**: If CLI and GUI both access `~/.zeropass`, file locking needed? Currently no `flock()` mechanism exists in Go vault code.
2. **Go dylib + Hardened Runtime**: Need to verify exact entitlements needed for Go runtime (GC uses `mmap` with executable pages). May need `allow-unsigned-executable-memory`.
3. **Universal binary**: Should the Go dylib be built as Universal (arm64+x86_64)? Adds ~2x size but supports Intel Macs. Use `lipo` to combine two builds.
4. **Password passing to Go**: When Swift calls `ZPUnlock(password)`, the password crosses the C boundary as `*char`. Should zero the C string after use. Go side should zero `[]byte` (already does via `crypto.ZeroBytes`). Swift side: `password.withCString { ptr in ZPUnlock(UnsafeMutablePointer(mutating: ptr)) }` — but Swift strings are immutable and may be cached. Consider using `Data` + `UnsafeMutableBufferPointer` for password handling.
5. **App bundle structure**: Where to place `libzeropass.dylib` — `Contents/Frameworks/` is standard. Need `@rpath` set correctly in Xcode build settings.
6. **Go runtime initialization**: `c-shared` Go libraries auto-initialize the Go runtime when loaded. First call may have ~50ms overhead. Acceptable for app startup.