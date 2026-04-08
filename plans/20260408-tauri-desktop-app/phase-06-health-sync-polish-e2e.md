# Phase 06: Health + Sync + Polish + E2E

## Context Links
- [plan.md](plan.md) | [phase-05](phase-05-settings-platform-features.md)
- [Testing Strategy Report](../reports/researcher-testing-strategy-report.md)

## Overview
- **Priority**: P2
- **Status**: pending
- **Description**: Health dashboard, version history, sync integration, dark mode polish, E2E tests

## Implementation Steps

- [ ] 1. Tauri commands: `commands/health.rs`, `commands/sync.rs`, `commands/version.rs`
- [ ] 2. HealthDashboard — circular score + category cards (weak, reused, old, breached)
- [ ] 3. VersionHistory — item version list with restore action
- [ ] 4. Sync flow connection (pull, push, full sync via SyncSettings)
- [ ] 5. Dark mode CSS polish — verify all components in both themes
- [ ] 6. Accessibility audit — keyboard navigation, focus management, ARIA labels
- [ ] 7. Playwright E2E setup: playwright.config.ts, helpers
- [ ] 8. E2E spec: auth.spec.ts — create vault → recovery → lock → unlock
- [ ] 9. E2E spec: items.spec.ts — CRUD items, all 8 types
- [ ] 10. E2E spec: search.spec.ts — toolbar search + Cmd+K
- [ ] 11. E2E spec: settings.spec.ts — navigate tabs, change settings
- [ ] 12. E2E spec: importexport.spec.ts — import CSV → export JSON
- [ ] 13. Coverage report: Vitest v8 + cargo-tarpaulin

## Success Criteria
- Health dashboard shows password health report
- Version history allows restoring previous versions
- Dark mode renders correctly across all components
- 5 E2E specs pass
- Combined coverage >90%
