# Phase 8: Docs Sync + Follow-Through

## Context Links
- `docs/project-changelog.md`
- `docs/project-roadmap.md`
- `plans/260402-macos-swiftui-app/phase-03-authentication-views.md`
- `plans/260402-macos-swiftui-app/phase-04-main-ui-layout.md`
- `plans/260402-macos-swiftui-app/phase-06-search-quick-access.md`
- `plans/260402-macos-swiftui-app/phase-08-settings-preferences.md`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`

## Overview
- **Priority:** P2
- **Status:** Pending
- **Description:** Record the redesign accurately in docs and mark older UI guidance as superseded so future work does not drift back toward the previous look.

## Key Insights
- The repo expects roadmap/changelog maintenance after meaningful UI work, even when the change is “just” presentation.
- Older macOS planning docs still describe earlier auth and layout assumptions; if they remain untouched, they will create future design drift.
- This phase should be small and factual, not a second design exercise.

## Requirements
- Update `docs/project-changelog.md` with the redesign scope, affected macOS areas, and validation outcome.
- Update `docs/project-roadmap.md` if the redesign materially changes macOS polish/status within Phase 1 or ship-readiness work.
- Add superseded/updated notes in older macOS plan docs that still describe the pre-Midnight direction.
- Capture follow-up polish items separately instead of silently extending the redesign scope.

## Architecture
- Keep docs changes limited to source-of-truth files plus the few older plan docs most likely to mislead future contributors.
- Prefer short “superseded by” notes over rewriting historical plan files line by line.

## Related Code Files

### Files to Edit
- `docs/project-changelog.md`
- `docs/project-roadmap.md`
- `plans/260402-macos-swiftui-app/phase-03-authentication-views.md`
- `plans/260402-macos-swiftui-app/phase-04-main-ui-layout.md`
- `plans/260402-macos-swiftui-app/phase-06-search-quick-access.md`
- `plans/260402-macos-swiftui-app/phase-08-settings-preferences.md`
- `plans/260402-macos-swiftui-app/reports/ui-ux-design-guideline.md`

## Implementation Steps
1. Add a changelog entry summarizing the Midnight Native redesign across auth, workspace, settings, quick search, menu bar, editor, and recovery utilities.
2. Update the roadmap if the macOS polish/ship-readiness state changed in a meaningful way.
3. Mark older macOS UI planning docs as updated/superseded where they would otherwise point contributors toward the pre-redesign direction.
4. Write down any deferred polish bugs or stretch ideas as follow-up work instead of quietly enlarging the finished scope.

## Todo List
- [ ] Update changelog with redesign summary and validation notes
- [ ] Update roadmap if progress/milestone framing changed
- [ ] Mark stale macOS UI planning docs as superseded or updated
- [ ] Capture leftover polish as explicit follow-up work

## Success Criteria
- Repo docs point to the new Midnight Native direction rather than the older mixed UI direction.
- Historical plans remain readable but no longer mislead future implementation work.
- The redesign closes with accurate documentation, not just merged code.

## Risk Assessment
- **Risk:** Old docs continue to reintroduce the prior design language.  
  **Mitigation:** Update the roadmap/changelog and add explicit superseded notes in the most relevant earlier UI plan docs.
- **Risk:** Docs work balloons into a rewrite of every historical plan.  
  **Mitigation:** Touch only the docs that still materially affect future implementation decisions.

## Security Considerations
- Do not include screenshots or copied secrets from real vault data in docs updates.
- Keep documentation focused on implementation scope and validation, not live sensitive examples.

## Next Steps
- Hand the plan to implementation with Phase 1 first, then move linearly through the phases so the shared visual system lands before the per-screen polish work.