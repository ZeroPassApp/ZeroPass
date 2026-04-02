# Phase 5: Item CRUD Views

## Context
- Depends on: [Phase 4 — Main UI Layout](./phase-04-main-ui-layout.md)
- **Design Reference**: [UI/UX Guideline §10 Item Detail Views](./reports/ui-ux-design-guideline.md#10-item-detail-views) + [§7 Component Library](./reports/ui-ux-design-guideline.md#7-component-library)

## Overview
- **Priority:** P0
- **Status:** Pending
- **Description:** Build detail views for all 8 item types, plus new item sheet and edit view. Each type has specific field layouts and actions.

## Key Insights
- 8 item types: login, apikey, sshkey, note, creditcard, identity, passkey, custom
- Each type has predefined fields (from `types.go` constants) + custom fields support
- Password fields: masked by default, toggle visibility, copy button
- Sensitive fields (password, api_secret, private_key, cvv): use `SecureField` or masked `Text`
- TOTP: show countdown timer if TOTP seed present in login fields
  - **Note:** Go core has NO TOTP code generation. Need Swift-side TOTP library (e.g., SwiftOTP) to generate time-based codes from the stored seed/URI
- **Passkey items are view/copy only for MVP** — acting as macOS Credential Provider (autofill into browsers) is Phase 3 scope. Show placeholder UI: "Credential Provider coming soon".
- **Version history** available via `ZPGetVersionHistory` — show in detail view

## Design Specs (from [UI/UX Guideline](./reports/ui-ux-design-guideline.md))

### Detail Pane Layout
- **Padding**: 16pt all sides
- **Header**: Type icon (32pt) + Name (15pt semibold) + Actions (top-right)
- **Field rows**: 32pt height, 8pt vertical spacing
- **Label width**: 100pt fixed, 11pt `.secondary`, right-aligned
- **Value**: 13pt, fill remaining width
- **Sections**: Separated by `Divider()` with 12pt vertical padding
- **Version history**: `DisclosureGroup`, collapsed by default
- **Metadata footer**: 10pt `.quaternary`

### SecureFieldRow Component
- **Row height**: 44pt (includes label + value + strength bar)
- **Mask character**: `●` (bullet, not asterisk)
- **Eye icon**: 14pt, `.secondary` default, `.blue` on hover
- **Copy icon**: 14pt, `.secondary` default, `.blue` on hover
- **Auto-re-mask**: Password hidden again after 30s of reveal
- **Strength bar**: 4pt height, color-coded (red/orange/yellow/green)

### FieldRow Component
- **Row height**: 32pt
- **Copy icon**: Appears on **hover only** (not always visible)

### TOTP Component
- **Code font**: SF Mono 18pt, grouped 3-3 with space separator
- **Timer**: Circular or linear progress
- **Color thresholds**: Green >10s, Yellow <10s, Red <5s
- **Copy**: Copies without spaces ("482193")

### TagChip Component
- **Height**: 22pt, corner radius 6pt
- **Font**: 11pt, `.quaternary` fill background
- **Remove button**: 10pt `xmark`, visible on hover / edit mode

### Toast Notification (Copy Feedback)
- **Position**: Bottom-center of active window, 16pt from bottom
- **Background**: `.thickMaterial` with rounded corners (8pt)
- **Animation**: Fade in 150ms, hold 2s, fade out 300ms

### Version History (DisclosureGroup)
- **Row height**: 40pt per version entry
- **Version label**: 13pt semibold
- **Timestamp**: 11pt `.secondary`
- **Changed fields**: 10pt `.caption`
- **Restore button**: Text button, `.blue`, with confirmation alert

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/Views/ItemDetail/ItemDetailView.swift` | Router: switch on item.type → specific detail view |
| `macos/ZeroPass/Views/ItemDetail/LoginDetailView.swift` | Login (username, password, URL, TOTP) |
| `macos/ZeroPass/Views/ItemDetail/APIKeyDetailView.swift` | API Key (key, secret, endpoint) |
| `macos/ZeroPass/Views/ItemDetail/SSHKeyDetailView.swift` | SSH Key (public, private, passphrase, fingerprint) |
| `macos/ZeroPass/Views/ItemDetail/NoteDetailView.swift` | Secure Note (text editor) |
| `macos/ZeroPass/Views/ItemDetail/CreditCardDetailView.swift` | Card (number, expiry, CVV, holder) |
| `macos/ZeroPass/Views/ItemDetail/IdentityDetailView.swift` | Identity (name, email, phone, address) |
| `macos/ZeroPass/Views/ItemDetail/PasskeyDetailView.swift` | Passkey — **view/copy only** (credential_id, rp_id, public_key) + "Credential Provider coming soon" banner |
| `macos/ZeroPass/Views/ItemDetail/CustomDetailView.swift` | Custom (key-value list) |
| `macos/ZeroPass/Views/ItemDetail/VersionHistoryView.swift` | Version history list + restore button (calls ZPRestoreVersion) |
| `macos/ZeroPass/Views/Components/SecureFieldRow.swift` | Reusable: masked value + reveal toggle + copy |
| `macos/ZeroPass/Views/Components/FieldRow.swift` | Reusable: label + value + copy button |
| `macos/ZeroPass/Views/Components/TagEditor.swift` | Tag chips with add/remove |
| `macos/ZeroPass/Views/Editors/NewItemSheet.swift` | Create new item (type picker → form) |
| `macos/ZeroPass/Views/Editors/EditItemView.swift` | Edit existing item (in-place or sheet) |
| `macos/ZeroPass/Views/Editors/PasswordGeneratorSheet.swift` | Generate password/passphrase inline |

## Implementation Steps

### ItemDetailView (router)
1. Receive `VaultItem` as binding or ID
2. Switch on `item.type` → render specific detail view
3. Toolbar: Edit button, Delete button, Favorite toggle
4. Show metadata: created, updated, last accessed, version
5. "Version History" expandable section → `VersionHistoryView`
   - List of past versions with timestamps
   - "Restore" button per version → calls `ZPRestoreVersion`
   - Confirmation dialog before restore

### Per-Type Detail Views
Each follows this pattern:
```
┌──────────────────────────────────────┐
│  [Icon] Item Name              ★ ✏️  │
│  Type: Login  •  Tags: #work, #dev  │
├──────────────────────────────────────┤
│  Username     octocat         📋     │
│  Password     ●●●●●●●●  👁️  📋     │
│  URL          github.com      🔗     │
│  TOTP         123 456    ⏱ 23s      │
├──────────────────────────────────────┤
│  Notes                               │
│  Main GitHub account for work        │
├──────────────────────────────────────┤
│  Custom Fields                       │
│  2FA Backup   [masked]   👁️  📋     │
├──────────────────────────────────────┤
│  Created: Mar 15, 2026               │
│  Updated: 2 hours ago                │
│  Version: 3                          │
└──────────────────────────────────────┘
```

### SecureFieldRow (reusable)
1. Label on left (100pt fixed width, 11pt `.secondary`, right-aligned)
2. Masked value ("●●●●●●") by default
3. Eye toggle to reveal (auto-re-mask after 30s)
4. Copy button (uses ClipboardService with auto-clear) — shows toast notification
5. For passwords: show strength indicator (4pt color bar)
6. Copy icon appears on hover, not always visible
7. VoiceOver: label as "Password, hidden" — never read content aloud

### NewItemSheet
1. Type picker (grid of 8 types with icons)
2. Dynamic form based on selected type
3. Pre-filled field names based on type constants
4. Password field: "Generate" button → PasswordGeneratorSheet
5. Tag editor
6. Favorite toggle
7. Save → calls `ZeroPassVault.createItem()` → refreshes list

### EditItemView
1. Same form as NewItemSheet, pre-populated
2. Type is read-only (can't change type after creation)
3. "History" section showing version count
4. Save → calls `ZeroPassVault.updateItem()` → refreshes list

### PasswordGeneratorSheet
1. Tab: Random Password | Passphrase
2. Random: length slider (8-128), toggles for uppercase/lowercase/digits/symbols
3. Passphrase: word count (4-8), separator picker
4. Preview: generated result + strength indicator
5. "Regenerate" button
6. "Use This Password" button → fills form field + copies to clipboard

## Success Criteria
- [ ] All 8 item types render correctly with appropriate fields (see [Guideline §10](./reports/ui-ux-design-guideline.md#10-item-detail-views))
- [ ] Detail pane: 16pt padding, 32pt field rows, 100pt label width
- [ ] Sensitive fields masked by default (● bullets), auto-re-mask after 30s
- [ ] Copy button shows toast notification ("Copied ✓", 2s duration)
- [ ] Copy icon appears on hover only (not always visible)
- [ ] New item creation works for all 8 types
- [ ] Edit updates existing items correctly
- [ ] Password generator produces results and fills forms
- [ ] TOTP: SF Mono 18pt, grouped 3-3, countdown with color thresholds
- [ ] Tags: 22pt chips with 6pt corner radius, remove on hover
- [ ] Favorite toggle persists
- [ ] Delete with confirmation alert (not just dialog)
- [ ] Form validation (required fields: name)
- [ ] Version history: `DisclosureGroup`, collapsed by default, restore with confirmation
- [ ] Passkey detail shows "view only" banner
- [ ] All text uses macOS scale (13pt body, NOT 17pt iOS)
- [ ] VoiceOver: passwords labeled as hidden, TOTP announced with remaining seconds
