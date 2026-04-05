# ZeroPass: Project Changelog

All significant changes, features, and fixes are documented here with versioning context and impact notes.

---

## [Unreleased]

### UI/UX Polish & Testing Infrastructure
- **macOS Surrounding Auth Surfaces Stitch Alignment (2026-04-05)**
  - Restyled `CreateVaultView.swift` and `OpenVaultSheet.swift` to match the newer unlock/auth panel language with tighter hierarchy, clearer guidance, and more consistent desktop-native spacing
  - Tightened `RecoveryPhraseView.swift`, `RecoveryPhraseCardView.swift`, and the settings recovery regeneration sheet so create → recovery → unlock now reads like one continuous auth journey instead of separate UI families
  - Added a required recovery confirmation toggle before onboarding can continue, aligning the first-run recovery step with the same seriousness already expected in settings regeneration
  - Sanitized create/open user-facing bridge errors into friendlier copy and hid decorative auth icons from accessibility where they were purely visual
  - **Validation:** targeted macOS UI tests passed for `testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell`, `testOpenExistingVaultFlowUnlocksToMainShell`, and `testRelaunchExistingVaultShowsLockedStateAndUnlocks`
  - **Impact:** macOS auth/recovery UI polish, onboarding safety, and UI-test robustness only; no changes to vault cryptography, bridge contracts, CLI behavior, or sync protocols

- **macOS Welcome & Unlock Flow Stitch Alignment (2026-04-05)**
  - Refined `WelcomeView.swift` to match the approved Stitch direction more closely: centered brand composition, side-by-side primary auth actions, compact chrome actions for documentation/settings, and clickable footer destinations for documentation, support, and release notes
  - Hardened the recent-vault path in the welcome screen by hiding stale bookmarks and clearing failed recent-vault entries instead of repeatedly surfacing a broken CTA
  - Updated `UnlockVaultView.swift` to better mirror the Stitch unlock hierarchy by hiding the extra segmented-control label, renaming the primary method to `Master Password`, tightening the selected-vault chip/panel composition, and matching the top-right overflow affordance more closely
  - Preserved the stable auth UI automation identifiers and revalidated the critical create-vault onboarding flow after reverting a risky auth-window size tweak
  - **Validation:** targeted macOS UI test `testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell` passed after the final welcome/layout adjustments
  - **Impact:** macOS auth flow polish and small support affordances only; no changes to unlock/create logic, vault cryptography, bridge APIs, CLI behavior, or sync contracts

- **macOS Auth/Vault Lifecycle UI-Test Coverage (2026-04-05)**
  - Extended `ZeroPassUITests.swift` from auth smoke coverage to deterministic end-to-end vault lifecycle coverage:
    - `testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell`
    - `testRelaunchExistingVaultShowsLockedStateAndUnlocks`
    - `testOpenExistingVaultFlowUnlocksToMainShell`
  - macOS UI test mode now injects deterministic folder selection through `UITEST_PICK_DIRECTORY_PATH` in `VaultFolderPicker`, avoiding `NSOpenPanel` for create/open flows during automation
  - Added stable auth/main-shell accessibility identifiers used by the full-flow assertions, including `unlockVault.passwordField`, `unlockVault.submitButton`, `mainShell.root`, and `mainShell.newItemButton`
  - Fresh validation passed on macOS arm64 with `xcodebuild test` for the targeted lifecycle flows
  - Outcome: reviewer approved with nits; tester marked the work ready to keep
  - **Impact:** macOS UI automation coverage only; no changes to vault crypto, CLI behavior, bridge APIs, or sync contracts

- **macOS Item Editor Modal Redesign Polish (2026-04-05)**
  - Redesigned `ItemEditorView.swift` to align with Stitch modal visual direction, improving hierarchy, grouping, and spacing
  - Fixed compile regression: made `FlowLayout` shared within module in `ItemDetailView.swift` and made frame modifier explicit in `ItemEditorView.swift`
  - Automated validation passed: `xcodebuild test` verified all existing tests remain passing
  - **Status:** Implementation complete; manual visual QA of editor modal is pending
  - **Impact:** macOS app UI polish only; no changes to item data model, save behavior, validation, or field defaults

- **macOS Unlock Screen Stitch Fidelity Pass (2026-04-05)**
  - Rebuilt the locked-state unlock composition to match the Stitch hierarchy more literally: compact selected-vault chip, nested unlock panel, lighter footer badges/helper text, icon-only vault options affordance, and tighter password/recovery chrome
  - Reduced locked/recovery window sizing and changed the locked window title to `Unlock Vault`
  - **Validation:** Full macOS `xcodebuild test` suite passed after the final unlock-screen changes
  - **Status:** Implementation and automated validation complete; manual unlock-specific visual and accessibility regression is still pending
  - **Impact:** macOS app UI polish only; no changes to unlock logic, vault behavior, CLI flows, bridge APIs, or configuration

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
