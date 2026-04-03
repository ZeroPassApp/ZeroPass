# Phase 7: macOS App — CSV Export

## Context

- **Priority:** MEDIUM | **Effort:** 1h | **Risk:** Low
- SecuritySettingsView has "Export Encrypted…" and "Export JSON…" buttons, NO CSV export
- CLI supports all 3 formats (json, csv, encrypted)
- Bridge layer already exports `ZPExportCSV`
- VaultClient only wraps `exportJSON()` and `exportEncrypted()`
- PRD §8.7 lists CSV as supported export format

## Related Files

| Action | File |
|--------|------|
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Views/Settings/SecuritySettingsView.swift` |
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Services/VaultClient.swift` |

## Implementation Steps

### 1. Add "Export CSV…" button in SecuritySettingsView

Add alongside existing export buttons:

```swift
HStack {
    Button("Export Encrypted…") {
        exportEncrypted()
    }
    .disabled(!vault.isUnlocked)

    Button("Export JSON…") {
        showingExportConfirm = true
    }
    .disabled(!vault.isUnlocked)

    Button("Export CSV…") {
        showingCSVExportConfirm = true
    }
    .disabled(!vault.isUnlocked)
}
```

### 2. Add state variables

```swift
@State private var showingCSVExportConfirm = false
@State private var csvExportConfirmText = ""
```

### 3. Add confirmation alert

Same pattern as existing JSON export — type "EXPORT" to confirm:

```swift
.alert("Export Unencrypted CSV", isPresented: $showingCSVExportConfirm) {
    TextField("Type EXPORT to continue", text: $csvExportConfirmText)
    Button("Cancel", role: .cancel) { csvExportConfirmText = "" }
    Button("Export", role: .destructive) {
        guard csvExportConfirmText.uppercased() == "EXPORT" else { return }
        exportCSV()
        csvExportConfirmText = ""
    }
} message: {
    Text("Exported CSV is unencrypted. Anyone with the file can read your secrets.")
}
```

### 4. Add exportCSV() handler

```swift
private func exportCSV() {
    let panel = NSSavePanel()
    panel.allowedContentTypes = [.commaSeparatedText]
    panel.nameFieldStringValue = "zeropass-export.csv"

    guard panel.runModal() == .OK, let url = panel.url else { return }

    Task {
        do {
            try await vault.exportCSV(to: url)
            // Show success notification
        } catch {
            // Show error alert
        }
    }
}
```

### 5. Add VaultClient.exportCSV()

```swift
func exportCSV(to url: URL) async throws {
    guard let h = handle else { throw ZPBridgeError(code: .notFound, message: "No vault open") }

    try await Task.detached(priority: .utility) {
        try ZPBridge.exportCSV(handle: h, path: url.path)
    }.value

    autoLock.recordActivity()
}
```

## Todo List

- [x] Add state variables for CSV export confirmation
- [x] Add "Export CSV…" button to HStack
- [x] Add confirmation alert with "type EXPORT" pattern
- [x] Add exportCSV() handler with NSSavePanel
- [x] Add VaultClient.exportCSV() method
- [x] Verify ZPBridge.exportCSV Swift wrapper exists
- [x] Xcode build succeeds

## Success Criteria

- Settings → "Export CSV…" button visible alongside Encrypted and JSON
- Clicking shows confirmation dialog requiring "EXPORT" text
- Successful export saves .csv file with correct content
- Disabled when vault is locked
- Warning text clearly communicates plaintext risk
