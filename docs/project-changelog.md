# ZeroPass: Project Changelog

All significant changes, features, and fixes are documented here with versioning context and impact notes.

---

## [Unreleased]

### UI/UX Polish & Testing Infrastructure
- **macOS Auth HIG Compliance & UI Automation (2026-04-04)**
  - Restored standard macOS auth window chrome (title bar, traffic lights, shadow)
  - Replaced custom floating auth overlays with native SwiftUI sheets
  - Simplified auth scaffold layout and keyboard focus restoration
  - Updated auth copy (welcome, unlock, create, recovery) toward HIG conventions
  - Updated button hierarchy and visual spacing for native sheet appearance
  - Added stable accessibility identifiers for auth UI elements (accessibility, testing):
    - Welcome actions: `welcome.createVaultButton`, `welcome.openVaultButton`
    - Create/open sheets: `createVault.title`, `openVault.title`
  - Added 4 macOS UI smoke tests for auth flow (all passing):
    - `testWelcomeScreenShowsPrimaryActionsOnFreshLaunch`
    - `testCreateVaultSheetCanBeOpenedAndCancelled`
    - `testOpenVaultSheetCanBeOpenedAndCancelled`
    - `testLaunchPerformance`
  - Added launch/reopen window recovery to ensure visible main window on fresh launch and menu commands
  - Unified no-vault `Open Vault…` command with shared auth sheet state (eliminates UI path divergence)
  - **Status:** All 15 automated tests passing; fresh-launch smoke tests verified via `/tmp/zeropass-macos-full-tests-5.log`
  - **Impact:** UI/UX and testing infrastructure only; no breaking changes to vault operations or CLI behavior

- **macOS Auth & Window Follow-Up Cleanup (2026-04-04)**
  - Implemented multi-window-safe auth modal routing via focused scene values and `AppDelegate` window tracking
  - Added UI test mode support with automatic test-state reset in `VaultClient` initialization
  - Introduced `VaultClient.AuthModal` enum and focus-request ID pattern for test-driven UI automation
  - Improved `WelcomeView` state management: replaced dual-boolean state with unified `activeAuthModal` binding
  - Enhanced `UnlockRecoverySection` with focus restoration via `focusRequestID` listener patterns
  - Refined `UnlockRecoverySection` styling: removed shadow, improved typography (callout weight), dynamic error highlighting
  - Added recovery phrase privacy marking via `.privacySensitive()` SwiftUI modifier
  - Consolidated command menu logic to direct no-vault `Open Vault…` flow through native auth sheets
  - Added `AppDelegate` lifecycle coordination for reliable main window visibility on fresh launch and menu-driven reopen
  - **Test Status:** Full macOS scheme validated and passed on 2026-04-04 (log: `/tmp/zeropass-macos-full-tests-followups.log`)
  - **Impact:** Auth window state is now window-local instead of process-global; UI automation is now testable via explicit focus request methods; no breaking changes

---

## How to Read This Changelog

- **[Unreleased]** — Changes in `main` branch not yet released
- **[Version]** — Tagged release versions (semantic versioning)
- **Categories:**
  - Features — New functionality or capabilities
  - Improvements — Enhanced existing features
  - UI/UX Polish — Styling, interaction refinement, accessibility
  - Bug Fixes — Resolved issues
  - Security — Security patches or hardening
  - Breaking Changes — API or behavior changes requiring migration
  - Dependencies — Updated external libraries
  - Internal — Refactoring, performance, cleanup

---

## Guidelines for Entries

- Each entry should reference affected components (CLI/Core/Bridge/macOS app)
- Include date of merge for context
- Note breaking changes prominently
- Link to related plan documents or issues when available
- For features: describe user-facing impact, not just code changes
- For security: specify severity and workaround if applicable
