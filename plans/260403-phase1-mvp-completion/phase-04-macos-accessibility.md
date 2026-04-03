# Phase 4: macOS Accessibility — VoiceOver & Keyboard Navigation

## Context

- **Priority:** HIGH | **Effort:** 3h | **Risk:** Low
- PRD §20 Phase 1: "Full keyboard navigation, Screen reader support (VoiceOver), Focus indicators"
- Current state: Only 3 `accessibilityLabel` in MainShellView + 1 `@FocusState` in QuickSearchView
- 6 of 9 target views have ZERO accessibility annotations

## Related Files

| Action | File | Changes |
|--------|------|---------|
| MODIFY | `apps/macos/.../Views/Auth/UnlockVaultView.swift` | `@FocusState` on password field, auto-focus `.onAppear` |
| MODIFY | `apps/macos/.../Views/Auth/CreateVaultView.swift` | `@FocusState` on password field, auto-focus `.onAppear` |
| MODIFY | `apps/macos/.../Views/Main/SidebarView.swift` | `accessibilityLabel` on category labels with item count |
| MODIFY | `apps/macos/.../Views/Main/ItemListView.swift` | `accessibilityLabel` per row, `accessibilityHint` |
| MODIFY | `apps/macos/.../Views/Main/ItemDetailView.swift` | `accessibilityLabel` on buttons, `accessibilityValue` on masked fields |
| MODIFY | `apps/macos/.../Views/Editors/ItemEditorView.swift` | `accessibilityLabel` on buttons, `@FocusState` on Name |
| MODIFY | `apps/macos/.../Views/Main/MainShellView.swift` | `accessibilityElement(children:)` grouping |
| MODIFY | `apps/macos/.../Views/MenuBar/MenuBarView.swift` | `accessibilityLabel` on all menu actions |
| MODIFY | `apps/macos/.../Views/Search/QuickSearchView.swift` | `accessibilityLabel` on result rows |

Base path: `apps/macos/ZeroPass/ZeroPass/`

## Accessibility Rules

1. Every icon-only `Button` → `.accessibilityLabel("descriptive action")`
2. Every `SecureField` → descriptive label association
3. Decorative `Image(systemName:)` → `.accessibilityHidden(true)`
4. Interactive `Image(systemName:)` → `.accessibilityLabel()` + `.accessibilityAddTraits(.isButton)`
5. Auth views → `@FocusState` with `.onAppear { focused = true }`

## Implementation Steps

### 1. UnlockVaultView.swift

```swift
@FocusState private var passwordFocused: Bool

SecureField("Master password", text: $password)
    .focused($passwordFocused)
    .accessibilityLabel("Master password")
    .accessibilityHint("Enter your vault master password to unlock")
    .onAppear { passwordFocused = true }
```

### 2. CreateVaultView.swift

```swift
@FocusState private var passwordFocused: Bool

SecureField("Create master password", text: $password)
    .focused($passwordFocused)
    .accessibilityLabel("Create master password")
    .accessibilityHint("Choose a strong master password for your vault")
    .onAppear { passwordFocused = true }

SecureField("Confirm password", text: $confirmPassword)
    .accessibilityLabel("Confirm master password")
```

### 3. SidebarView.swift

Each category label needs count context:

```swift
Label("Logins", systemImage: "key")
    .accessibilityLabel("Logins, \(loginCount) items")

Label("Secure Notes", systemImage: "note.text")
    .accessibilityLabel("Secure Notes, \(noteCount) items")
```

### 4. ItemListView.swift

```swift
ForEach(items) { item in
    ItemRow(item: item)
        .accessibilityLabel("\(item.name), \(item.type.displayName)")
        .accessibilityHint("Double-click to view details")
}
```

### 5. ItemDetailView.swift

```swift
// Copy buttons
Button { copyPassword() } label: {
    Image(systemName: "doc.on.doc")
}
.accessibilityLabel("Copy password")

// Reveal/hide toggle
Button { toggleReveal() } label: {
    Image(systemName: revealed ? "eye.slash" : "eye")
}
.accessibilityLabel(revealed ? "Hide password" : "Show password")

// Masked fields
Text(masked ? "••••••••" : password)
    .accessibilityValue(revealed ? "Password visible" : "Password hidden")
```

### 6. ItemEditorView.swift

```swift
@FocusState private var nameFocused: Bool

TextField("Item name", text: $name)
    .focused($nameFocused)
    .accessibilityLabel("Item name")
    .onAppear { nameFocused = true }

Button { addField() } label: {
    Image(systemName: "plus")
}
.accessibilityLabel("Add field")

Button { removeField(at: index) } label: {
    Image(systemName: "minus.circle")
}
.accessibilityLabel("Remove field")

Button { generatePassword() } label: {
    Image(systemName: "wand.and.stars")
}
.accessibilityLabel("Generate password")
```

### 7. MainShellView.swift

Already has 3 toolbar labels. Add region grouping:

```swift
NavigationSplitView {
    SidebarView()
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Sidebar navigation")
} detail: {
    DetailView()
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Item details")
}
```

### 8. MenuBarView.swift

```swift
Button("Quick Search") { ... }
    .accessibilityLabel("Open quick search")

Button("Lock Vault") { ... }
    .accessibilityLabel("Lock vault")

Button("Copy Password") { ... }
    .accessibilityLabel("Copy password to clipboard")
```

### 9. QuickSearchView.swift

Already has `@FocusState`. Add labels on result rows:

```swift
ForEach(results) { result in
    SearchResultRow(item: result)
        .accessibilityLabel("\(result.name), \(result.type.displayName)")
        .accessibilityHint("Press Enter to copy password")
}
```

## Todo List

- [x] UnlockVaultView: @FocusState + auto-focus + labels
- [x] CreateVaultView: @FocusState + auto-focus + labels
- [x] SidebarView: accessibilityLabel with item counts
- [x] ItemListView: accessibilityLabel + hint per row
- [x] ItemDetailView: labels on all buttons, value on masked fields
- [x] ItemEditorView: @FocusState + labels on action buttons
- [x] MainShellView: accessibilityElement grouping
- [x] MenuBarView: accessibilityLabel on all actions
- [x] QuickSearchView: accessibilityLabel on result rows

## Success Criteria

- VoiceOver announces all interactive elements correctly
- Tab key navigates through all focusable elements
- Auth views auto-focus password field on appear
- All icon-only buttons have descriptive labels
- Masked fields report visibility state
- Xcode build succeeds with no accessibility warnings
