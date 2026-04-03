# Phase 5: macOS Accessibility — High Contrast Mode

## Context

- **Priority:** MEDIUM | **Effort:** 1h | **Risk:** Low
- PRD §20: "High contrast mode option" — not implemented
- GeneralSettingsView.swift currently only has System/Light/Dark theme picker
- Builds on Phase 4 accessibility work

## Related Files

| Action | File |
|--------|------|
| MODIFY | `apps/macos/ZeroPass/ZeroPass/Views/Settings/GeneralSettingsView.swift` |
| MODIFY | `apps/macos/ZeroPass/ZeroPass/ZeroPassApp.swift` |

## Implementation Steps

### 1. Add toggle in GeneralSettingsView.swift

In the Appearance section, add high contrast toggle:

```swift
@AppStorage("highContrastMode") private var highContrastMode = false

Section("Appearance") {
    Picker("Appearance", selection: $appearanceMode) {
        Text("System").tag("system")
        Text("Light").tag("light")
        Text("Dark").tag("dark")
    }
    .pickerStyle(.segmented)

    Toggle("High contrast", isOn: $highContrastMode)
    Text("Increases contrast for better readability")
        .font(.caption)
        .foregroundStyle(.secondary)
}
```

### 2. Apply environment in ZeroPassApp.swift

Read `@AppStorage("highContrastMode")` and apply on root view:

```swift
@AppStorage("highContrastMode") private var highContrastMode = false

var body: some Scene {
    WindowGroup {
        ContentView()
            .environmentObject(vault)
            .environmentObject(quickSearch)
            .preferredColorScheme(preferredColorScheme)
            .environment(\.colorSchemeContrast, highContrastMode ? .increased : .standard)
    }
}
```

> **Note:** SwiftUI's `.environment(\.colorSchemeContrast, ...)` is read-only. Instead, use a custom environment key and apply visual changes conditionally:

```swift
// Custom environment key
private struct HighContrastKey: EnvironmentKey {
    static let defaultValue = false
}

extension EnvironmentValues {
    var highContrast: Bool {
        get { self[HighContrastKey.self] }
        set { self[HighContrastKey.self] = newValue }
    }
}

// In ZeroPassApp:
ContentView()
    .environment(\.highContrast, highContrastMode)
```

### 3. Apply visual changes when high contrast enabled

Create a `HighContrastModifier` ViewModifier:

- Border widths: 1pt → 2pt
- Text colors: `.secondary` → `.primary`
- Minimum font size: 13px → 14px
- Button borders: visible when high contrast

Apply via `.modifier(HighContrastModifier())` or use the environment value in individual views.

## Todo List

- [x] Add @AppStorage toggle in GeneralSettingsView
- [x] Define HighContrastKey environment key
- [x] Apply .environment(\.highContrast) in ZeroPassApp
- [x] Create HighContrastModifier or apply per-view changes
- [x] Test visual changes with toggle on/off
- [x] Xcode build succeeds

## Success Criteria

- Settings → High contrast toggle visible and persists across launches
- When enabled: borders thicker, text more visible, fonts at least 14px
- When disabled: normal appearance restored
- Works with all 3 theme modes (system/light/dark)
