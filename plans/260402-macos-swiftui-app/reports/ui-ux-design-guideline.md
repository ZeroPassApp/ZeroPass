# ZeroPass macOS App — UI/UX Design Guideline

> Comprehensive design guideline based on Apple HIG, Apple Design Award winners (2022-2025), and gold-standard macOS apps (1Password 8, Apple Passwords, Things 3, Raycast, Bear, Fantastical, Transmit).

---

## Table of Contents

1. [Current Plan Evaluation](#1-current-plan-evaluation)
2. [Design Philosophy](#2-design-philosophy)
3. [Window & Layout Standards](#3-window--layout-standards)
4. [Typography System](#4-typography-system)
5. [Color System](#5-color-system)
6. [Iconography](#6-iconography)
7. [Component Library](#7-component-library)
8. [Authentication Flows](#8-authentication-flows)
9. [Main Layout (3-Column)](#9-main-layout-3-column)
10. [Item Detail Views](#10-item-detail-views)
11. [Quick Search (Spotlight-Style)](#11-quick-search-spotlight-style)
12. [Menu Bar Widget](#12-menu-bar-widget)
13. [Settings Window](#13-settings-window)
14. [Animation & Transitions](#14-animation--transitions)
15. [Keyboard Shortcuts](#15-keyboard-shortcuts)
16. [Accessibility](#16-accessibility)
17. [Dark Mode](#17-dark-mode)
18. [Empty States](#18-empty-states)
19. [Security UX Patterns](#19-security-ux-patterns)
20. [Anti-Patterns to Avoid](#20-anti-patterns-to-avoid)
21. [Pre-Delivery Checklist](#21-pre-delivery-checklist)

---

## 1. Current Plan Evaluation

### Strengths (What the plan does right)

| Area | Assessment | Reference |
|------|-----------|-----------|
| 3-column NavigationSplitView | Correct — matches 1Password, Apple Passwords, Finder pattern | Phase 4 |
| SF Symbols for icons | Correct — native macOS standard | Phase 4 |
| `.regularMaterial` for sidebar | Correct — Apple HIG material recommendation | Phase 4 |
| `.searchable()` modifier | Correct — native search integration | Phase 4 |
| NSPanel for Quick Search | Excellent — `.nonactivatingPanel` is the right approach (vs SwiftUI Window) | Phase 6 |
| Keychain + TouchID flow | Correct — `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` + `.biometryCurrentSet` | Phase 7 |
| Settings TabView (⌘,) | Correct macOS pattern | Phase 8 |
| macOS 14+ deployment target | Correct — required for @Observable maturity | Plan |

### Issues & Gaps Found

#### Critical Issues

| # | Issue | Current Plan | Recommendation | Impact |
|---|-------|-------------|----------------|--------|
| 1 | **Missing macOS menu bar integration** | No mention of standard macOS menu bar menus (File, Edit, View, etc.) | Must define full menu bar structure with keyboard shortcuts | HIGH — Feels non-native without proper menus |
| 2 | **No window sizing specs** | No minimum/maximum/default window sizes defined | Define: default 1024×680pt, min 800×500pt, sidebar 200-250pt | HIGH — Poor resize behavior feels broken |
| 3 | **iOS-scale thinking in detail views** | Phase 5 ASCII mockups suggest iOS-scale layout with large padding | Use macOS density: 13pt body, 20-25pt row height, compact spacing | HIGH — Too much whitespace = inefficient on desktop |
| 4 | **Missing focus/keyboard navigation spec** | Only ↑↓ in list mentioned | Define full Tab order, Return/Space actions, ⌘ shortcuts for every action | MEDIUM — Power users will feel constrained |
| 5 | **No drag-and-drop support** | Not mentioned anywhere | Support item drag between categories/tags, drag URLs from browser | MEDIUM — Expected macOS interaction |
| 6 | **Table vs List confusion** | Phase 4 mentions macOS Table but no column spec | Define sortable columns: Name, Type, Username, Updated — with persist widths | MEDIUM |

#### Design Gaps

| # | Gap | Recommendation |
|---|-----|---------------|
| 7 | No onboarding/first-launch experience defined | Add welcome animation, quick tour (3 screens max) after vault creation |
| 8 | No "toast" notification pattern | Define clipboard copy feedback, auto-lock notification, sync status |
| 9 | No loading skeleton spec | Define skeleton screens for vault loading, search results |
| 10 | No spacing/grid system | Adopt 8pt grid, define all padding/margin values |
| 11 | No visual hierarchy for item types | Define color-coding per type (login=blue, ssh=green, note=yellow, etc.) |
| 12 | Quick search (Phase 6) missing "recent items" when empty | Show last 5 accessed items + pinned/favorites |
| 13 | No Liquid Glass / vibrancy spec | macOS 2025 trend — define where to use materials |
| 14 | Version history UX not fully designed | Needs timeline view with diff indicators |
| 15 | No confirmation pattern standardized | Define when to use alert vs inline confirmation |

---

## 2. Design Philosophy

### Core Principles (Inspired by Apple Design Award Winners)

```
┌─────────────────────────────────────────────────────────────┐
│                   ZeroPass Design Pillars                    │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  NATIVE     │  │  SECURE     │  │  DEVELOPER-FIRST    │ │
│  │  FIRST      │  │  BY DESIGN  │  │                     │ │
│  │             │  │             │  │                     │ │
│  │ Follow HIG  │  │ Privacy as  │  │ Keyboard-driven,    │ │
│  │ SF Symbols  │  │ UX pattern  │  │ information-dense,  │ │
│  │ Materials   │  │ not just    │  │ power-user ready    │ │
│  │ System font │  │ encryption  │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  CLARITY    │  │  DEFERENCE  │  │  DEPTH              │ │
│  │             │  │             │  │                     │ │
│  │ Content     │  │ UI serves   │  │ Vibrancy layers     │ │
│  │ hierarchy   │  │ content,    │  │ reveal hierarchy,   │ │
│  │ guides the  │  │ not the     │  │ translucency over   │ │
│  │ eye         │  │ other way   │  │ decoration          │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Design References

| Reference App | What to Learn | Apply To |
|---------------|---------------|----------|
| **1Password 8** | Vault-colored accents, reveal animations, Watchtower security UI | Item detail, health dashboard |
| **Apple Passwords** | Minimal chrome, modal-first password display, system integration | Unlock flow, simplicity benchmark |
| **Things 3** | Perfect sidebar + list + detail, checkbox animations | 3-column layout, item interactions |
| **Raycast** | Spotlight search UX, stagger animations, keyboard-first | Quick search (⌘K) |
| **Bear** | Tag-based organization, inline editing, markdown | Tag system in sidebar |
| **Fantastical** | Information density, natural language input | Search, compact list view |

---

## 3. Window & Layout Standards

### Window Specifications

```
┌──────────────────────────────────────────────────────────┐
│ Default:  1024 × 680 pt                                  │
│ Minimum:   800 × 500 pt                                  │
│ Maximum:  1800 × 1200 pt (or screen size)                │
│                                                          │
│ ┌──────────┬───────────────────┬───────────────────────┐ │
│ │ Sidebar  │   Item List       │   Detail Pane         │ │
│ │          │                   │                       │ │
│ │ 200-     │   280-400 pt      │   Remaining width     │ │
│ │ 250 pt   │   (min 280)       │   (min 320 pt)        │ │
│ │          │                   │                       │ │
│ │ Collapse │   Sortable Table  │   Scroll if needed    │ │
│ │ to 0pt   │   or List         │                       │ │
│ │ (toggle) │                   │                       │ │
│ └──────────┴───────────────────┴───────────────────────┘ │
│                                                          │
│ Window narrows:                                          │
│  < 1000pt → Hide detail (show on selection as overlay)   │
│  <  700pt → Hide sidebar (toggle via button)             │
└──────────────────────────────────────────────────────────┘
```

### Column Behavior

```swift
// SwiftUI implementation pattern
NavigationSplitView(columnVisibility: $columnVisibility) {
    SidebarView()
        .navigationSplitViewColumnWidth(min: 200, ideal: 220, max: 280)
} content: {
    ItemListView()
        .navigationSplitViewColumnWidth(min: 280, ideal: 350, max: 450)
} detail: {
    DetailView()
        .navigationSplitViewColumnWidth(min: 320, ideal: 400, max: .infinity)
}
.navigationSplitViewStyle(.balanced)
```

### Window State Persistence

- Remember window position + size between launches
- Remember sidebar collapsed state
- Remember column widths
- Remember selected category + sort order
- Use `@SceneStorage` for per-window state

---

## 4. Typography System

### macOS Font Scale (NOT iOS scale)

| Style | Size | Weight | Line Height | Use Case |
|-------|------|--------|-------------|----------|
| **Window Title** | 26pt | Regular | 32pt | App name in unlock screen |
| **Section Header** | 15pt | Semibold | 20pt | Sidebar section labels |
| **Body** | 13pt | Regular | 18pt | Item names, field values, main content |
| **Body Emphasized** | 13pt | Semibold | 18pt | Selected item, labels |
| **Secondary** | 11pt | Regular | 16pt | Usernames, URLs, metadata |
| **Caption** | 10pt | Regular | 14pt | Timestamps, version info, hints |
| **Monospace** | 12pt | Regular (SF Mono) | 16pt | Passwords, API keys, SSH keys, recovery words |
| **Monospace Small** | 11pt | Regular (SF Mono) | 15pt | Code, fingerprints, hashes |

### Font Rules

```swift
// DO — Use system font APIs
Text("Item Name")
    .font(.body)              // 13pt on macOS

Text("username@email.com")
    .font(.subheadline)       // 11pt on macOS
    .foregroundStyle(.secondary)

Text("●●●●●●●●")
    .font(.system(.body, design: .monospaced))

// DON'T — Hard-code sizes
Text("Item Name")
    .font(.system(size: 17))  // This is iOS scale!
```

### Line Length

- Body text: 60-75 characters max per line
- Detail pane field values: full width allowed (monospace wraps)
- Notes field: limit to 75ch with horizontal padding

---

## 5. Color System

### Semantic Colors (MANDATORY — Never hard-code RGB)

```swift
// Text
Color.primary            // Main text (#000 light, #e8e8e8 dark)
Color.secondary          // Subtitles, metadata
Color.tertiary           // Hints, placeholders

// Backgrounds
Color(nsColor: .windowBackgroundColor)       // Window background
Color(nsColor: .controlBackgroundColor)      // Input fields, cards
Color(nsColor: .underPageBackgroundColor)    // Sidebar background
Color(nsColor: .separatorColor)              // Dividers

// System accent (user-configurable in System Settings)
Color.accentColor        // Interactive elements, selected states
```

### Item Type Color Coding

| Item Type | Color | SF Symbol | Hex (for reference only, use named colors) |
|-----------|-------|-----------|---------------------------------------------|
| Login | Blue | `key.fill` | `.blue` |
| API Key | Purple | `wrench.and.screwdriver.fill` | `.purple` |
| SSH Key | Green | `terminal.fill` | `.green` |
| Secure Note | Yellow | `note.text` | `.yellow` |
| Credit Card | Orange | `creditcard.fill` | `.orange` |
| Identity | Teal | `person.crop.circle.fill` | `.teal` |
| Passkey | Indigo | `person.badge.key.fill` | `.indigo` |
| Custom | Gray | `square.grid.2x2.fill` | `.gray` |

### Functional Colors

```swift
// Status
Color.green              // Password strength: Strong, Success
Color.yellow             // Password strength: Fair, Warning
Color.orange             // Password strength: Weak
Color.red                // Password strength: Very Weak, Errors, Destructive actions

// Security
Color.red                // Delete, destructive actions
Color.blue               // Copy, informational actions
Color.green              // Success, verified
```

### Material Usage Map

| Surface | Material | When |
|---------|----------|------|
| Sidebar background | `.regularMaterial` | Always |
| Quick search panel | `.ultraThickMaterial` | Floating panel |
| Menu bar popover | `.regularMaterial` | Always |
| Modal sheets | `.thickMaterial` | Overlay on content |
| Toolbar | System default | Always |
| Unlock screen background | `.ultraThinMaterial` + blur | Over dimmed app content |

---

## 6. Iconography

### SF Symbols Rules

```
✓ Use SF Symbols exclusively (no custom icons, no emojis)
✓ Match symbol weight to adjacent text weight
✓ Use hierarchical rendering for multi-color symbols
✓ Standard size: 16pt in toolbar, 14pt in sidebar, 12pt in list rows
✓ Use .symbolRenderingMode(.hierarchical) for colored icons
```

### Symbol Map

```swift
// Navigation
"sidebar.left"                    // Toggle sidebar
"plus"                            // Add new item
"lock.fill"                       // Lock vault
"lock.open.fill"                  // Vault unlocked
"arrow.triangle.2.circlepath"     // Sync
"gearshape"                       // Settings
"magnifyingglass"                 // Search

// Item Types (filled variants for sidebar, regular for list)
"key.fill"                        // Login
"wrench.and.screwdriver.fill"     // API Key
"terminal.fill"                   // SSH Key
"note.text"                       // Secure Note
"creditcard.fill"                 // Credit Card
"person.crop.circle.fill"         // Identity
"person.badge.key.fill"           // Passkey
"square.grid.2x2.fill"           // Custom

// Actions
"doc.on.doc"                      // Copy
"eye"                             // Reveal
"eye.slash"                       // Hide
"star"                            // Unfavorited
"star.fill"                       // Favorited
"trash"                           // Delete
"pencil"                          // Edit
"arrow.up.arrow.down"             // Sort
"clock.arrow.circlepath"          // Version history
"touchid"                         // TouchID
"faceid"                          // FaceID

// Status
"checkmark.circle.fill"           // Success
"exclamationmark.triangle.fill"   // Warning
"xmark.circle.fill"               // Error
"arrow.clockwise"                 // Sync in progress
```

---

## 7. Component Library

### SecureFieldRow (Reusable)

```
┌──────────────────────────────────────────────────────┐
│  Password     ●●●●●●●●●●●●●●     [👁] [📋]         │
│               ▔▔▔▔▔▔▔▔▔▔▔▔▔▔                        │
│               Strength: Strong ████████████░░         │
└──────────────────────────────────────────────────────┘

States:
- Default: Masked (●●●●●●)
- Revealed: Plaintext (auto-re-mask after 30s)
- Copied: "Copied ✓" toast (2s)
- Hover: Subtle background highlight on row
```

**Specifications:**
- Row height: 44pt (includes label + value + strength bar)
- Label: 11pt secondary color, left-aligned
- Value: 13pt monospace, primary color
- Mask character: `●` (bullet, not asterisk)
- Eye icon: 14pt, `.secondary` color, `.blue` on hover
- Copy icon: 14pt, `.secondary` color, `.blue` on hover
- Strength bar: 4pt height, rounded corners, color-coded

### FieldRow (Non-sensitive)

```
┌──────────────────────────────────────────────────────┐
│  Username     octocat                          [📋]  │
└──────────────────────────────────────────────────────┘
```

- Row height: 32pt
- Label: 11pt secondary, 100pt fixed width
- Value: 13pt primary, fill remaining width
- Copy icon appears on hover (not always visible)

### TagChip

```
┌────────────┐  ┌──────────┐  ┌──────────────┐
│ #work  [×] │  │ #dev     │  │ + Add tag    │
└────────────┘  └──────────┘  └──────────────┘
```

- Height: 22pt
- Font: 11pt
- Background: `.quaternary` fill
- Corner radius: 6pt
- Remove button: 10pt `xmark`, visible on hover only (edit mode always)

### Toast Notification

```
┌──────────────────────────────────────┐
│  ✓  Copied to clipboard (30s)       │
└──────────────────────────────────────┘
```

- Position: Bottom-center of active window, 16pt from bottom
- Background: `.thickMaterial` with rounded corners (8pt)
- Animation: Fade in 150ms, hold 2s, fade out 300ms
- Auto-clear countdown: Subtle progress bar under text

### Empty State

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│                    🔑 (SF Symbol)                     │
│                     48pt, .tertiary                   │
│                                                      │
│              No Items Found                          │
│              13pt, .secondary                        │
│                                                      │
│          Add your first credential                   │
│          11pt, .tertiary                             │
│                                                      │
│            [ + Add Item ]                            │
│            accent color button                       │
│                                                      │
└──────────────────────────────────────────────────────┘
```

- Use `.contentUnavailableView` (macOS 14+) for native rendering
- Symbol size: 48pt
- Center in content area (not full window)

---

## 8. Authentication Flows

### First-Time Setup Wizard

```
Step 1: Welcome               Step 2: Master Password      Step 3: Recovery Key
┌─────────────────────┐      ┌─────────────────────┐      ┌─────────────────────┐
│                     │      │                     │      │                     │
│     🔐              │      │  Create Master      │      │  ⚠️ Write This Down │
│                     │      │  Password           │      │                     │
│  Welcome to         │      │                     │      │  ┌─────┬─────┬────┐ │
│  ZeroPass           │      │  [●●●●●●●●●●●●]    │      │  │ word│ word│word│ │
│                     │      │  [●●●●●●●●●●●●]    │      │  │  1  │  2  │  3 │ │
│  Secure credential  │      │                     │      │  ├─────┼─────┼────┤ │
│  management for     │      │  ████████░░ Strong  │      │  │ word│ word│word│ │
│  developers.        │      │                     │      │  │  4  │  5  │  6 │ │
│                     │      │  ✓ 12+ characters   │      │  ├─────┼─────┼────┤ │
│  [Get Started →]    │      │  ✓ Not common       │      │  │ word│ word│word│ │
│                     │      │  ✓ Good entropy     │      │  │  7  │  8  │  9 │ │
│                     │      │                     │      │  ├─────┼─────┼────┤ │
│                     │      │  [Continue →]       │      │  │ word│ word│word│ │
│                     │      │                     │      │  │ 10  │ 11  │ 12 │ │
└─────────────────────┘      └─────────────────────┘      │  └─────┴─────┴────┘ │
                                                          │                     │
                                                          │  [Copy] [Continue→] │
                                                          └─────────────────────┘

Step 4: Verify Recovery        Step 5: Done
┌─────────────────────┐       ┌─────────────────────┐
│                     │       │                     │
│  Verify Recovery    │       │     ✅              │
│  Key                │       │                     │
│                     │       │  You're all set!    │
│  Enter word #3:     │       │                     │
│  [___________]      │       │  Your vault is      │
│                     │       │  ready to use.      │
│  Enter word #7:     │       │                     │
│  [___________]      │       │  [Open ZeroPass]    │
│                     │       │                     │
│  Enter word #11:    │       │                     │
│  [___________]      │       │                     │
│                     │       │                     │
│  [Verify →]         │       │                     │
│                     │       │                     │
└─────────────────────┘       └─────────────────────┘
```

**Wizard Specs:**
- Window size: Fixed 480×520pt (centered, non-resizable)
- Step indicator: Horizontal dots at top (5 dots)
- Navigation: Back arrow in top-left (except step 1), Continue in bottom-right
- Animations: Cross-fade between steps (200ms)
- Recovery key grid: 4×3, monospace 14pt, each word in bordered cell
- **CRITICAL**: Disable window screenshots during recovery key step (`NSWindow.sharingType = .none`)

### Unlock Screen

```
┌──────────────────────────────────────────────┐
│                                              │
│                                              │
│               🔒  ZeroPass                   │
│                   26pt, semibold             │
│                                              │
│         ┌─────────────────────────────┐      │
│         │ Master Password             │ 🔑   │
│         └─────────────────────────────┘      │
│                                              │
│              [Unlock]    [TouchID 👆]        │
│                                              │
│         Use Recovery Key                     │
│         11pt, link style                     │
│                                              │
│              Error: Incorrect password       │
│              (shake animation + red text)    │
│                                              │
└──────────────────────────────────────────────┘
```

**Unlock Specs:**
- Full-window overlay with `.ultraThinMaterial` background
- App icon: 64pt, centered
- Password field: `SecureField`, 300pt width, centered
- Enter key triggers unlock (no need to click button)
- TouchID: Attempt automatically on appear if enabled
- Error state: Field shakes (3x, 4pt amplitude, 300ms), red error text fades in
- Loading: Replace "Unlock" button with `ProgressView` during Argon2id (200-500ms)

---

## 9. Main Layout (3-Column)

### Sidebar Design

```
┌────────────────────────┐
│  ▼ Categories          │  ← 10pt caption, .tertiary
│                        │
│  🔑  All Items    (24) │  ← Selected: accent bg + .primary text
│  ⭐  Favorites     (3) │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
│  🔐  Logins       (15) │
│  🔧  API Keys      (4) │
│  💻  SSH Keys      (2) │
│  📝  Notes         (1) │
│  💳  Credit Cards   (1) │
│  👤  Identities    (0) │
│  🗝  Passkeys      (1) │
│                        │
│  ▼ Tags                │
│                        │
│  #  work           (8) │
│  #  personal       (6) │
│  #  dev            (4) │
│  #  aws            (2) │
│                        │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
│  📊 24 items           │  ← 10pt caption, .quaternary
│  🔄 Synced 2m ago      │
└────────────────────────┘
```

**Sidebar Specs:**
- Width: 200-250pt (collapsible to 0pt)
- Background: `.regularMaterial` (translucent)
- Row height: 28pt
- Icon size: 14pt SF Symbols
- Text: 13pt body
- Count badge: 11pt, right-aligned, `.secondary` color
- Selected row: System accent color background, `.primary` text
- Section header: 10pt, `.tertiary`, uppercase
- Separator: `Color(nsColor: .separatorColor)`, 0.5pt
- Bottom status: 10pt, `.quaternary`
- Tags section: Dynamic, sorted by count descending

### Item List (Table View)

```
┌──────────────────────────────────────────────────────┐
│  Name ↑          Type      Username       Updated    │  ← Column headers
│  ───────────────────────────────────────────────────  │
│  🔑 GitHub       Login     octocat        2h ago     │  ← Selected row
│  🔑 AWS Console  Login     admin@co..     1d ago     │
│  🔧 Stripe API   API Key   sk_live_..     3d ago     │
│  💻 deploy-key   SSH Key   —              1w ago     │
│  📝 Server Notes Note      —              2w ago     │
│  🔑 Gmail        Login     user@gma..     1mo ago    │
│                                                      │
│           (scrollable, lazy loading)                 │
└──────────────────────────────────────────────────────┘
```

**Table Specs:**
- Use SwiftUI `Table` (macOS native sortable columns)
- Row height: 24pt (compact macOS density)
- Column headers: 11pt, `.secondary`, clickable for sort
- Sort indicator: ↑/↓ arrow next to active sort column
- Name column: Icon (12pt) + text (13pt), min width 150pt
- Type column: Text only (11pt), width 80pt
- Username column: Truncated with `...`, 13pt, width 120pt
- Updated column: Relative date (11pt), right-aligned, width 80pt
- Selected row: System selection color
- Alternating rows: Optional (respect system setting)
- Context menu: Right-click per row (see Section 15)
- Multi-select: ⌘+click for multiple, ⇧+click for range

### Detail Pane

```
┌──────────────────────────────────────────────────────┐
│  ┌────┐                                              │
│  │ 🔑 │  GitHub                         ⭐  ✏️  🗑  │  ← 15pt semibold
│  └────┘  Login  •  #work, #dev                       │  ← 11pt secondary
│  ────────────────────────────────────────────────     │
│                                                      │
│  Username     octocat                          [📋]  │
│  Password     ●●●●●●●●●●●●●       [👁]  [📋]       │
│               ████████████░░ Strong                  │
│  URL          github.com           [🔗]  [📋]       │
│  TOTP         482 193              ⏱ 23s             │
│                                                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│                                                      │
│  Notes                                               │
│  Main GitHub account for work projects               │
│                                                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│                                                      │
│  Custom Fields                                       │
│  2FA Backup   ●●●●●●●●             [👁]  [📋]      │
│                                                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│                                                      │
│  Tags                                                │
│  [#work ×] [#dev ×] [+ Add tag]                     │
│                                                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│                                                      │
│  ▶ Version History (3 versions)                      │
│                                                      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│  Created: Mar 15, 2026  •  Updated: 2 hours ago     │  ← 10pt caption
│  Version: 3                                          │
└──────────────────────────────────────────────────────┘
```

**Detail Pane Specs:**
- Padding: 16pt all sides
- Header: Type icon (32pt) + Name (15pt semibold) + Actions (top-right)
- Actions spacing: 8pt between icons
- Field rows: 32pt height, 8pt vertical spacing
- Label width: 100pt fixed, 11pt `.secondary`, right-aligned
- Value: 13pt, fill remaining width
- Sections: Separated by `Divider()` with 12pt vertical padding
- Version history: `DisclosureGroup`, collapsed by default
- Metadata footer: 10pt `.quaternary`
- Scroll: Vertical scroll if content exceeds pane height

---

## 10. Item Detail Views

### Per-Type Field Layout

#### Login
| Field | Component | Actions |
|-------|-----------|---------|
| Username | FieldRow | Copy |
| Password | SecureFieldRow + Strength | Reveal, Copy |
| URL | FieldRow + link | Open in browser, Copy |
| TOTP | TOTPRow (countdown timer) | Copy code |
| Notes | TextEditor (expandable) | — |
| Custom Fields | Dynamic SecureFieldRow/FieldRow | Reveal, Copy |

#### API Key
| Field | Component | Actions |
|-------|-----------|---------|
| Key ID | FieldRow | Copy |
| API Secret | SecureFieldRow | Reveal, Copy |
| Endpoint URL | FieldRow + link | Open, Copy |
| Environment | Badge (dev/staging/prod) | — |
| Notes | TextEditor | — |

#### SSH Key
| Field | Component | Actions |
|-------|-----------|---------|
| Public Key | MonospaceFieldRow (multi-line) | Copy |
| Private Key | SecureFieldRow (multi-line) | Reveal, Copy |
| Passphrase | SecureFieldRow | Reveal, Copy |
| Fingerprint | MonospaceFieldRow | Copy |
| Notes | TextEditor | — |

#### Secure Note
| Field | Component | Actions |
|-------|-----------|---------|
| Content | Large TextEditor (fill pane) | — |

#### Credit Card
| Field | Component | Actions |
|-------|-----------|---------|
| Card Number | SecureFieldRow (formatted: •••• •••• •••• 4242) | Reveal, Copy |
| Expiry | FieldRow (MM/YY) | Copy |
| CVV | SecureFieldRow | Reveal, Copy |
| Cardholder | FieldRow | Copy |
| Notes | TextEditor | — |

### TOTP Component

```
┌──────────────────────────────────────────┐
│  TOTP    482 193           ⏱ 23s  [📋]  │
│          ▔▔▔ ▔▔▔                         │
│          30pt monospace                  │
│          grouped in 3s                   │
│                                          │
│   [████████████████████░░░░░░░░]  timer  │
│   Green when >10s, Yellow <10s, Red <5s  │
└──────────────────────────────────────────┘
```

- Code font: SF Mono 18pt (larger for readability)
- Grouped: 3-3 with space separator
- Timer: Circular or linear progress, color changes at thresholds
- Auto-refresh: Code updates every 30s with fade animation
- Copy: Copies without spaces ("482193")

### Version History (Expanded DisclosureGroup)

```
┌──────────────────────────────────────────┐
│  ▼ Version History (3 versions)          │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │ v3  Current  •  2 hours ago       │  │
│  │     Changed: password, notes      │  │
│  ├────────────────────────────────────┤  │
│  │ v2  •  3 days ago      [Restore]  │  │
│  │     Changed: username, url        │  │
│  ├────────────────────────────────────┤  │
│  │ v1  •  Mar 15, 2026    [Restore]  │  │
│  │     Initial version               │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

- Row height: 40pt
- Version label: 13pt semibold
- Timestamp: 11pt secondary
- Changed fields: 10pt caption
- Restore button: Text button, `.blue`, with confirmation alert

---

## 11. Quick Search (Spotlight-Style)

### Panel Design (NSPanel)

```
┌──────────────────────────────────────────────────────┐
│  🔍  Search ZeroPass...                          ⌘K  │  ← 15pt, auto-focus
│  ─────────────────────────────────────────────────── │
│                                                      │
│  RECENT                                    10pt cap  │
│  🔑  GitHub          octocat            2h ago       │  ← Hover highlight
│  🔑  AWS Console     admin@company      1d ago       │
│  🔧  Stripe API      sk_live_...       3d ago       │
│                                                      │
│  ─────────────────────────────────────────────────── │
│  ⏎ Copy password  ⌘⏎ Open  ⇧⏎ Open URL  ⌘C Copy   │  ← 10pt hints
└──────────────────────────────────────────────────────┘

With search results:
┌──────────────────────────────────────────────────────┐
│  🔍  git|                                            │
│  ─────────────────────────────────────────────────── │
│                                                      │
│  ▸ 🔑  GitHub           octocat         Login        │  ← Selected (accent bg)
│    🔑  GitLab           user@gitlab     Login        │
│    💻  git-deploy-key                   SSH Key      │
│                                                      │
│  ─────────────────────────────────────────────────── │
│  ⏎ Copy password  ⌘⏎ Open  ⇧⏎ Open URL             │
└──────────────────────────────────────────────────────┘
```

**Panel Specs:**
- Window: `NSPanel`, `.nonactivatingPanel`, `.floating`
- Size: 580×400pt (max 10-12 results visible)
- Position: Center-top of screen (like Spotlight), 200pt from top
- Background: `.ultraThickMaterial` (opaque-ish frosted glass)
- Corner radius: 12pt
- Shadow: `.shadow(color: .black.opacity(0.3), radius: 20)`
- Search field: 15pt, no border, full width
- Results: Stagger fade-in (each +30ms, like Raycast)
- Result row height: 36pt
- Result icon: 16pt, type-colored
- Result title: 13pt semibold
- Result subtitle: 11pt secondary (username/URL)
- Result type badge: 10pt, right-aligned, `.quaternary`
- Selected: System accent color background
- Keyboard hints footer: 10pt `.quaternary`, separator above
- Dismiss: Escape or click outside

### Search Behavior
- Debounce: 150ms keystrokes
- Fuzzy matching: "gh" matches "GitHub"
- Target: < 50ms result display (per PRD)
- Empty query: Show last 5 accessed items
- No results: "No items matching 'query'" + "Create new item" link

---

## 12. Menu Bar Widget

```
Menu Bar: [🔒] or [🔓] icon

Popover (when unlocked):                Popover (when locked):
┌──────────────────────────┐           ┌──────────────────────────┐
│  🔍 Search...            │           │                          │
│  ─────────────────────── │           │    🔒 Vault Locked       │
│                          │           │                          │
│  RECENT                  │           │    [Unlock with TouchID] │
│  🔑 GitHub    [📋]       │           │    [Enter Password]      │
│  🔑 AWS       [📋]       │           │                          │
│  🔧 Stripe    [📋]       │           │  ─────────────────────── │
│  🔑 Gmail     [📋]       │           │  Open ZeroPass           │
│  💻 deploy    [📋]       │           │  Quit                    │
│                          │           └──────────────────────────┘
│  ─────────────────────── │
│  Open ZeroPass     ⌘O    │
│  Lock Vault        ⌘L    │
│  ─────────────────────── │
│  Quit              ⌘Q    │
└──────────────────────────┘
```

**Menu Bar Specs:**
- Style: `.menuBarExtraStyle(.window)` for popover panel
- Icon: `lock.fill` (locked) / `lock.open.fill` (unlocked), 18×18pt
- Popover width: 280pt
- Search field: 13pt, compact
- Recent items: Max 5, with quick-copy button
- Copy button: Appears on hover per row
- Sections: Divider between items and actions

---

## 13. Settings Window

### Tab Structure

```
┌─────────────────────────────────────────────────────────┐
│     [⚙️ General]  [🔒 Security]  [🔄 Sync]  [ℹ️ About] │
├─────────────────────────────────────────────────────────┤

General Tab:                            Security Tab:
┌───────────────────────────────┐      ┌───────────────────────────────┐
│                               │      │                               │
│  Vault Location               │      │  Auto-Lock                    │
│  ~/.zeropass/vaults/default   │      │  Lock after:  [5 minutes ▾]   │
│  [Change...]                  │      │  Lock on sleep:     [✓]       │
│                               │      │  Lock on screensaver: [✓]     │
│  Startup                      │      │                               │
│  Launch at login:   [✓]      │      │  Clipboard                    │
│  Show in menu bar:  [✓]      │      │  Auto-clear after: [30s ▾]    │
│                               │      │                               │
│  Quick Search                 │      │  Biometric                    │
│  Hotkey: [⌘⇧P] [Record...]  │      │  Unlock with TouchID: [✓]     │
│                               │      │                               │
│  Appearance                   │      │  Master Password              │
│  ○ System  ○ Light  ○ Dark   │      │  [Change Password...]         │
│                               │      │                               │
└───────────────────────────────┘      │  Recovery Key                 │
                                       │  [Regenerate Recovery Key...] │
Sync Tab:                              │                               │
┌───────────────────────────────┐      │  Data                         │
│                               │      │  [Export Vault...]            │
│  Sync Server                  │      │  [Import...]                  │
│  Enable sync:    [✓]         │      │                               │
│  Server URL:                  │      └───────────────────────────────┘
│  [https://sync.example.com]  │
│  API Key:                     │
│  [●●●●●●●●●●●●]             │
│  Device Name:                 │
│  [Justin's MacBook Pro]      │
│                               │
│  [Test Connection]            │
│  Last synced: 2 minutes ago   │
│  [Sync Now]                   │
│                               │
└───────────────────────────────┘
```

**Settings Specs:**
- Window size: 550×450pt (fixed, centered)
- Tab style: System TabView (icons + text at top)
- Section headers: 13pt semibold
- Form fields: Standard SwiftUI Form with grouped style
- Spacing: 12pt between sections, 8pt between fields
- Buttons: Standard `.bordered` style, destructive in red

---

## 14. Animation & Transitions

### Animation Timing (macOS is FASTER than iOS)

| Action | Duration | Easing | Notes |
|--------|----------|--------|-------|
| Sidebar toggle | 250ms | `.easeInOut` | Width animation |
| Navigation (list → detail) | 200ms | `.easeInOut` | Cross-dissolve |
| Sheet present | 250ms | `.spring(response: 0.3)` | Slide up + fade |
| Sheet dismiss | 200ms | `.easeIn` | Slide down + fade |
| Alert appear | 150ms | `.easeOut` | Scale 0.95→1.0 + fade |
| Quick search open | 200ms | `.spring(response: 0.25, dampingFraction: 0.8)` | Scale 0.95→1.0 |
| Quick search close | 150ms | `.easeIn` | Fade out |
| Search result stagger | 30ms each | `.easeOut` | Each result 30ms after previous |
| Copy feedback toast | Fade in 150ms, hold 2s, fade out 300ms | `.easeInOut` | Bottom-center |
| Password reveal | 150ms | `.easeInOut` | Cross-fade masked→plain |
| Error shake | 300ms total | `.default` | 3 oscillations, 4pt amplitude |
| Unlock loading | Indeterminate | — | ProgressView replaces button |
| Checkbox/toggle | 200ms | `.spring(dampingFraction: 0.7)` | Bouncy toggle |
| TOTP code change | 300ms | `.easeInOut` | Fade old → fade new |
| Item type icon | Instant | — | No animation on type icons |

### Animation Rules

```swift
// DO — Use withAnimation for state changes
withAnimation(.easeInOut(duration: 0.2)) {
    selectedItem = item
}

// DO — Respect reduced motion
@Environment(\.accessibilityReduceMotion) var reduceMotion

withAnimation(reduceMotion ? .none : .easeInOut(duration: 0.2)) {
    isExpanded.toggle()
}

// DON'T — Animate everything (causes visual noise)
// DON'T — Use animations longer than 400ms
// DON'T — Use bouncy springs for navigation transitions
```

---

## 15. Keyboard Shortcuts

### Global Shortcuts

| Action | Shortcut | Context |
|--------|----------|---------|
| Quick Search | `⌘⇧P` (configurable) | Global hotkey — works from any app |
| New Item | `⌘N` | Main window |
| Lock Vault | `⌘L` | Anywhere in app |
| Settings | `⌘,` | Standard macOS |
| Find/Filter | `⌘F` | Focus search in main window |
| Close Window | `⌘W` | Standard macOS |
| Quit | `⌘Q` | Standard macOS |
| Hide | `⌘H` | Standard macOS |
| Minimize | `⌘M` | Standard macOS |
| Full Screen | `⌃⌘F` | Standard macOS |

### Item Actions

| Action | Shortcut | Context |
|--------|----------|---------|
| Copy Password | `⌘⇧C` | Item selected |
| Copy Username | `⌘⌃C` | Item selected |
| Open URL | `⌘⇧O` | Login item selected |
| Edit Item | `⌘E` | Item selected |
| Delete Item | `⌘⌫` | Item selected (with confirmation) |
| Toggle Favorite | `⌘⇧F` | Item selected |
| Show Version History | `⌘⌃H` | Item selected |

### Navigation

| Action | Shortcut | Context |
|--------|----------|---------|
| Toggle Sidebar | `⌘⌃S` | Main window |
| Next Item | `↓` | Item list focused |
| Previous Item | `↑` | Item list focused |
| Open Item | `⏎` | Item list focused |
| Back to List | `⌘[` | Detail view |
| Tab Between Columns | `Tab` | Main window |

### macOS Menu Bar Structure

```
ZeroPass  File  Edit  View  Item  Window  Help
────────  ────  ────  ────  ────  ──────  ────

ZeroPass:
  About ZeroPass
  ─────────
  Settings...         ⌘,
  ─────────
  Hide ZeroPass       ⌘H
  Hide Others         ⌥⌘H
  Show All
  ─────────
  Quit ZeroPass       ⌘Q

File:
  New Item            ⌘N
  ─────────
  Import...           ⌘I
  Export...           ⌘⇧E
  ─────────
  Close Window        ⌘W

Edit:
  Undo               ⌘Z
  Redo               ⌘⇧Z
  ─────────
  Cut                ⌘X
  Copy               ⌘C
  Paste              ⌘V
  Select All         ⌘A
  ─────────
  Find               ⌘F

View:
  Toggle Sidebar     ⌘⌃S
  ─────────
  Sort by Name
  Sort by Type
  Sort by Date Modified
  ─────────
  Show Toolbar
  Customize Toolbar...
  ─────────
  Enter Full Screen  ⌃⌘F

Item:
  Edit               ⌘E
  Duplicate          ⌘D
  ─────────
  Copy Password      ⌘⇧C
  Copy Username      ⌘⌃C
  Open URL           ⌘⇧O
  ─────────
  Toggle Favorite    ⌘⇧F
  ─────────
  Delete             ⌘⌫

Window:
  Minimize           ⌘M
  Zoom
  ─────────
  Bring All to Front

Help:
  ZeroPass Help
  ─────────
  Report a Problem...
  GitHub Repository
```

---

## 16. Accessibility

### VoiceOver Requirements

```swift
// Every interactive element must have an accessibility label
Button(action: { copyPassword() }) {
    Image(systemName: "doc.on.doc")
}
.accessibilityLabel("Copy password")
.accessibilityHint("Copies the password to clipboard. Auto-clears in 30 seconds.")

// Passwords — describe as hidden, not read aloud
Text(maskedPassword)
    .accessibilityLabel("Password, hidden")
    .accessibilityHint("Double-tap to copy. Activate eye button to reveal.")

// Password strength
ProgressView(value: strengthScore, total: 4)
    .accessibilityLabel("Password strength: \(strengthLabel)")
    .accessibilityValue("\(Int(strengthScore)) out of 4")

// TOTP countdown
Text(totpCode)
    .accessibilityLabel("Verification code: \(totpCode)")
    .accessibilityValue("Expires in \(secondsRemaining) seconds")
```

### Minimum Requirements

| Requirement | Standard | Test Method |
|-------------|----------|-------------|
| Color contrast (body text) | 4.5:1 minimum (target 7:1) | Xcode Accessibility Inspector |
| Color contrast (large text) | 3:1 minimum | Xcode Accessibility Inspector |
| Focus indicators | Visible ring on all interactive elements | Tab through all controls |
| Touch/click targets | 24pt minimum on Mac | Measure in Xcode |
| VoiceOver navigation | All elements reachable and labeled | Enable VoiceOver (⌘⌥Z) |
| Reduced motion | Respect `.accessibilityReduceMotion` | System Settings > Accessibility |
| Increase contrast | Semantic colors adapt automatically | System Settings > Accessibility |
| Dynamic Type | Support larger text sizes | System Settings > Accessibility |
| Color-blind safe | Never use color alone for meaning | Always add text/icon alongside |

### Accessibility Audit Checklist

- [ ] All buttons have `.accessibilityLabel`
- [ ] All images have `.accessibilityLabel` or are decorative (`.accessibilityHidden(true)`)
- [ ] Password fields never read content in VoiceOver
- [ ] Tab order matches visual layout (sidebar → list → detail)
- [ ] Error messages announced via `.accessibilityAnnouncement`
- [ ] Loading states announced ("Unlocking vault...")
- [ ] TOTP countdown accessible with remaining seconds
- [ ] Context menus accessible via keyboard
- [ ] Sheets and alerts trap focus correctly
- [ ] Empty states are accessible and describe next action

---

## 17. Dark Mode

### Implementation Rules

```swift
// DO — Use semantic colors (auto-adapt)
Text("Item Name")
    .foregroundStyle(.primary)

// DO — Use adaptive materials
.background(.regularMaterial)

// DO — Use Color assets with light/dark variants
// (Only if semantic colors don't suffice)

// DON'T — Hard-code colors
.foregroundColor(Color(red: 0, green: 0, blue: 0)) // WRONG

// DON'T — Check colorScheme manually for basic colors
@Environment(\.colorScheme) var scheme
Text("Label")
    .foregroundColor(scheme == .dark ? .white : .black) // WRONG — use .primary
```

### Color Behavior Across Modes

| Element | Light Mode | Dark Mode | Implementation |
|---------|-----------|-----------|----------------|
| Body text | Near-black (#1D1D1F) | Near-white (#E8E8E8) | `.primary` |
| Secondary text | Gray (#666) | Light gray (#888) | `.secondary` |
| Window background | White (#FFFFFF) | Dark gray (#1D1D1D) | `.windowBackgroundColor` |
| Sidebar | Translucent light | Translucent dark | `.regularMaterial` |
| Dividers | Light gray (#D1D1D6) | Dark gray (#38383A) | `.separatorColor` |
| Accent colors | Standard saturation | +10-20% brightness | `.accentColor` (automatic) |
| Item type colors | Standard | Slightly brighter | System named colors |
| Destructive red | #FF3B30 | #FF453A | `.red` |
| Success green | #34C759 | #30D158 | `.green` |

### Testing Checklist

- [ ] All text readable in both modes
- [ ] Materials look correct (not solid black/white)
- [ ] Item type colors distinguishable in both modes
- [ ] Dividers visible in both modes
- [ ] Selected row highlight visible in both modes
- [ ] Password strength colors readable in both modes
- [ ] Quick search panel looks correct in both modes
- [ ] Menu bar icon adapts (dark icon on light menu bar, vice versa)
- [ ] Increase Contrast setting doesn't break layout

---

## 18. Empty States

### State Definitions

| State | Context | Illustration | Message | Action |
|-------|---------|-------------|---------|--------|
| New vault | After setup, no items | `tray.fill` 48pt | "Your vault is empty" | "Add First Item" / "Import from browser" |
| No search results | Search query → 0 results | `magnifyingglass` 48pt | "No items matching '...'" | "Clear search" |
| No favorites | Favorites filter, 0 favorited | `star` 48pt | "No favorites yet" | "Star items to add them here" |
| Category empty | Filter by type, 0 items | Type icon 48pt | "No [type] items" | "Add [Type]" |
| No tags | Tags section, 0 tags | `tag` 48pt | "No tags" | "Tags are added when creating items" |
| Sync not configured | Sync settings, not setup | `arrow.triangle.2.circlepath` 48pt | "Sync not configured" | "Set up sync server" |
| Vault locked | Detail pane when locked | `lock.fill` 48pt | "Vault is locked" | "Unlock to view items" |
| Select an item | Detail pane, nothing selected | `sidebar.right` 48pt | "Select an item" | — (no action, just hint) |

### Implementation

```swift
// Use ContentUnavailableView (macOS 14+)
ContentUnavailableView {
    Label("Your vault is empty", systemImage: "tray.fill")
} description: {
    Text("Add your first credential or import from a browser.")
} actions: {
    Button("Add Item") { showNewItemSheet = true }
        .buttonStyle(.borderedProminent)
    Button("Import...") { showImport = true }
        .buttonStyle(.bordered)
}
```

---

## 19. Security UX Patterns

### Principle: "Security should be invisible until necessary"

| Pattern | Implementation |
|---------|----------------|
| **Passwords always masked** | Default `●●●●●●`, never show in plaintext unless explicitly requested |
| **Auto-reveal timeout** | Re-mask password after 30s of reveal |
| **Copy > Reveal** | Copy button is more prominent than reveal (encourage copy over read) |
| **Clipboard auto-clear** | Clear after 30s (configurable), show toast |
| **Lock on idle** | Default 5 minutes, configurable |
| **Lock on sleep** | Enabled by default |
| **Lock on screensaver** | Enabled by default |
| **Screenshot protection** | `NSWindow.sharingType = .none` during recovery key display |
| **Memory zeroing** | Zero sensitive data (passwords, vault keys) immediately after use |
| **No password logging** | Never log passwords, keys, or mnemonics |
| **Error messages** | Generic "Incorrect password" — never reveal if vault exists |
| **Biometric fallback** | Always offer password entry if biometric fails |
| **Master password rules** | Length ≥ 12, zxcvbn score ≥ 3 — enforced in setup, suggested on change |

### Destructive Action Confirmation

```
Level 1 — Single item delete:
  Alert: "Delete 'GitHub'? This cannot be undone."
  Buttons: [Cancel] [Delete (red)]

Level 2 — Bulk delete / Export:
  Alert with destructive framing + type-to-confirm:
  "Export creates an UNENCRYPTED file containing all credentials."
  "Type 'export' to confirm: [_________]"
  Buttons: [Cancel] [Export]

Level 3 — Vault destructive (change master password, regenerate recovery):
  Require current password entry:
  "Enter your current master password to continue."
  [●●●●●●●●●●●●]
  Buttons: [Cancel] [Continue]
```

---

## 20. Anti-Patterns to Avoid

### Layout Anti-Patterns

| Don't | Why | Do Instead |
|-------|-----|-----------|
| Use iOS 17pt body text | Too large for Mac density | Use 13pt (macOS standard) |
| Hard-code window size | Breaks with different displays | Set min/max, remember position |
| Create custom window chrome | Non-native, confusing | Use system title bar + toolbar |
| Hide menu bar items | macOS users expect all menus | Disable (gray out) instead |
| Use hamburger menu | Not a macOS pattern | Use sidebar with toggle |
| Make toolbar the only way to access actions | Breaks accessibility | Mirror all toolbar actions in menu bar |
| Use iOS-style tab bar at bottom | Not macOS pattern | Use sidebar or top TabView |
| Full-width cards with huge padding | Wastes screen on Mac | Use compact rows, 8pt spacing |

### Interaction Anti-Patterns

| Don't | Why | Do Instead |
|-------|-----|-----------|
| Skip context menu support | Expected on macOS | Right-click on every list row |
| Skip keyboard shortcuts | Power users rely on them | Define shortcut for every action |
| Use only ⌘+click for multi-select | Unintuitive | Support ⇧+click for range select too |
| Make drag-and-drop the only way | Not discoverable | Always provide menu/button alternative |
| Use iOS-style swipe actions | macOS uses right-click | Use context menus |
| Animate everything | Visual noise | Only animate state changes, max 300ms |

### Security Anti-Patterns

| Don't | Why | Do Instead |
|-------|-----|-----------|
| Show passwords by default | Security risk | Always mask, require explicit reveal |
| Store master password in UserDefaults | Insecure | Use Keychain only |
| Keep vault key in memory indefinitely | Attack surface | Zero on lock |
| Use plaintext logging for errors | May leak sensitive data | Sanitize all log output |
| Auto-fill without user confirmation | Security risk | Require click/shortcut |
| Skip clipboard clear | Password lingers in clipboard | Default 30s auto-clear |

---

## 21. Pre-Delivery Checklist

### Before Every Build

- [ ] **Typography**: All text uses system font APIs (`.body`, `.subheadline`, `.caption`), no hard-coded sizes
- [ ] **Colors**: All colors are semantic (`.primary`, `.secondary`, `.accentColor`), no hard-coded RGB
- [ ] **Icons**: All icons are SF Symbols, no emojis in UI, no custom icon files
- [ ] **Materials**: Sidebar uses `.regularMaterial`, panels use appropriate material
- [ ] **Dark Mode**: Tested in both light and dark mode
- [ ] **Keyboard**: All actions have keyboard shortcuts, Tab navigation works
- [ ] **Menu Bar**: All actions available in macOS menu bar menus
- [ ] **Context Menus**: Right-click works on all list items
- [ ] **Window Sizing**: Min/max sizes set, remembers position, resizes gracefully
- [ ] **Focus States**: All interactive elements show focus ring
- [ ] **VoiceOver**: All elements labeled, passwords hidden from screen readers
- [ ] **Reduced Motion**: Animations respect `.accessibilityReduceMotion`
- [ ] **Empty States**: All empty collections show helpful message + action
- [ ] **Loading States**: All async operations show progress indicator
- [ ] **Error States**: All errors show user-friendly message near the problem
- [ ] **Sensitive Data**: Passwords masked, clipboard auto-clears, memory zeroed
- [ ] **Confirmation**: Destructive actions require explicit confirmation
- [ ] **Toast/Feedback**: Copy action shows brief confirmation
- [ ] **Spacing**: 8pt grid system, consistent padding throughout

### Before Release

- [ ] Run with Accessibility Inspector (Xcode)
- [ ] Test VoiceOver end-to-end flow
- [ ] Test with Increase Contrast enabled
- [ ] Test with Reduce Motion enabled
- [ ] Test on smallest supported window size (800×500)
- [ ] Test on external 4K display (HiDPI)
- [ ] Test in Full Screen mode
- [ ] Test Stage Manager compatibility
- [ ] Verify all keyboard shortcuts work
- [ ] Verify clipboard auto-clear works
- [ ] Verify auto-lock on sleep/screensaver works
- [ ] Verify TouchID flow (first unlock → biometric)
- [ ] Profile for memory leaks in sensitive data handling

---

## Appendix A: Code Patterns

### Recommended SwiftUI Patterns

```swift
// 1. NavigationSplitView with proper sizing
struct MainView: View {
    @State private var columnVisibility: NavigationSplitViewVisibility = .all
    @State private var selectedCategory: SidebarCategory? = .all
    @State private var selectedItemID: String?
    
    var body: some View {
        NavigationSplitView(columnVisibility: $columnVisibility) {
            SidebarView(selection: $selectedCategory)
                .navigationSplitViewColumnWidth(min: 200, ideal: 220, max: 280)
        } content: {
            ItemListView(
                category: selectedCategory,
                selection: $selectedItemID
            )
            .navigationSplitViewColumnWidth(min: 280, ideal: 350, max: 450)
        } detail: {
            if let id = selectedItemID {
                ItemDetailView(itemID: id)
            } else {
                ContentUnavailableView(
                    "Select an Item",
                    systemImage: "sidebar.right",
                    description: Text("Choose an item from the list to view details.")
                )
            }
        }
        .navigationSplitViewStyle(.balanced)
    }
}

// 2. Proper semantic styling
struct ItemRowView: View {
    let item: VaultItem
    
    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: item.type.sfSymbol)
                .foregroundStyle(item.type.color)
                .frame(width: 16)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(item.name)
                    .font(.body)
                    .lineLimit(1)
                
                if let subtitle = item.subtitle {
                    Text(subtitle)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }
            
            Spacer()
            
            Text(item.updatedAt.relativeFormatted)
                .font(.caption)
                .foregroundStyle(.tertiary)
        }
        .padding(.vertical, 2)
    }
}

// 3. Respect reduced motion
struct AnimatedView: View {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var isVisible = false
    
    var body: some View {
        content
            .opacity(isVisible ? 1 : 0)
            .offset(y: isVisible ? 0 : (reduceMotion ? 0 : 10))
            .onAppear {
                withAnimation(reduceMotion ? .none : .easeOut(duration: 0.2)) {
                    isVisible = true
                }
            }
    }
}
```

---

## Appendix B: Design Token Summary

```
SPACING
  xs:   4pt
  sm:   8pt
  md:  12pt
  lg:  16pt
  xl:  24pt
  2xl: 32pt

CORNER RADIUS
  xs:   4pt  (buttons, badges)
  sm:   6pt  (tags, chips)
  md:   8pt  (cards, toasts)
  lg:  12pt  (panels, sheets)

ROW HEIGHTS
  compact:  24pt  (table rows)
  standard: 28pt  (sidebar items)
  spacious: 36pt  (search results)
  field:    32pt  (form field rows)
  header:   44pt  (detail pane header)

ICON SIZES
  xs:   12pt  (inline with text)
  sm:   14pt  (sidebar, list rows)
  md:   16pt  (toolbar, search results)
  lg:   24pt  (detail header type icon)
  xl:   32pt  (detail header)
  2xl:  48pt  (empty states)
  3xl:  64pt  (unlock screen app icon)

Z-INDEX (for overlays)
  base:       0
  sidebar:   10
  toolbar:   20
  toast:     30
  panel:     40  (quick search)
  modal:     50  (sheets, alerts)
  menuBar:   60

ANIMATION
  instant:    0ms  (selection highlight)
  fast:     150ms  (micro-interactions, reveals)
  normal:   200ms  (navigation, transitions)
  slow:     300ms  (panel open/close, fade)
  emphasis: 400ms  (first-launch, onboarding)
```

---

*Last updated: 2026-04-02*
*Based on: Apple HIG (2025), Apple Design Awards 2022-2025, 1Password 8, Apple Passwords, Things 3, Raycast, Bear, Fantastical, Transmit, CleanShot X*
