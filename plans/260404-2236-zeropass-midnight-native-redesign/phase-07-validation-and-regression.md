# Phase 7: Validation + Regression

## Context Links
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`
- `apps/macos/ZeroPass/ZeroPassTests/RecoveryPhraseSupportTests.swift`
- `.github/copilot-instructions.md`

## Overview
- **Priority:** P1
- **Status:** Pending
- **Description:** Verify the redesign across UI automation, manual QA, and repo-level build/test commands so the visual rewrite lands without collateral damage.

## Key Insights
- The repo already has macOS UI smoke coverage for the welcome/create/open flow, plus unit coverage around recovery helpers and search matcher behavior.
- UI redesigns often fail in focus order, keyboard flow, appearance variants, and selection persistence long before they fail in compiler output.
- Even though the work is macOS-heavy, the repo convention is still to keep Go build/test health green before finishing.

## Requirements
- Preserve and update existing UI smoke tests instead of discarding them.
- Add coverage where the redesign meaningfully changes interaction surfaces: unlock layout, settings access, quick search, and editor entry points.
- Validate system appearance, dark/light overrides, reduced motion, increase contrast, Full Keyboard Access, and VoiceOver-relevant labels.
- Run both macOS app validation and repo-wide Go build/test commands before closing the work.

## Architecture
- Treat validation as a matrix, not one happy path.
- Keep tests targeted: UI tests for cross-screen flows, unit tests for helper logic only where helper behavior changed.
- Use the repo’s existing build/test commands; do not invent a parallel verification workflow.

## Related Code Files

### Files to Edit
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITests.swift`
- `apps/macos/ZeroPass/ZeroPassUITests/ZeroPassUITestsLaunchTests.swift`
- `apps/macos/ZeroPass/ZeroPassTests/ZeroPassTests.swift`
- `apps/macos/ZeroPass/ZeroPassTests/RecoveryPhraseSupportTests.swift`

### Validation Commands
- `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- `go test ./...`
- `go vet ./...`
- `go build -o zp ./packages/cli/`
- `go build -o syncserver ./services/syncserver/`
- `bash services/syncserver/smoke-test.sh both`

## Implementation Steps
1. Update existing UI tests to survive any renamed identifiers or redesigned auth/settings/editor entry points.
2. Add a few focused UI smoke tests for the redesigned unlocked experience where it matters most: quick search invocation, settings window surfacing, and item editor open/cancel flow.
3. Add or adjust unit tests only if search/recovery helper behavior changes during the redesign.
4. Run the macOS scheme test command first, then full Go repo verification, and only then close the redesign work.
5. Finish with a manual QA pass across auth, workspace, settings, search, menu bar, editor, and recovery flows in system/light/dark accessibility variants.

## Todo List
- [ ] Preserve and expand macOS UI smoke coverage
- [ ] Update/add unit tests only where helper behavior changed
- [ ] Run macOS validation command
- [ ] Run repo-wide Go build/test/vet/smoke commands
- [ ] Complete manual QA matrix for appearance and accessibility variants

## Success Criteria
- The redesign passes automated UI tests and repo-level verification.
- Keyboard and accessibility flows still work across redesigned screens.
- The app looks better without breaking Go-side confidence or hidden regression checks.

## Risk Assessment
- **Risk:** UI-only changes quietly break focus order or keyboard shortcuts.  
  **Mitigation:** Add focused smoke coverage and manual keyboard QA for the changed surfaces.
- **Risk:** Test updates become brittle if they overfit visual structure.  
  **Mitigation:** Assert stable affordances and flows, not pixel-level layout.

## Security Considerations
- Ensure tests and QA do not capture or persist real recovery phrases, passwords, or API keys.
- Keep copy-sensitive paths and destructive confirmations in the validation matrix.

## Next Steps
- Synchronize docs so the new direction is recorded and older plans do not point future contributors back to the pre-Midnight UI.