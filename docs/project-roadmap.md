# ZeroPass: Development Roadmap

## Overview

ZeroPass is organized into 4 development phases, each building on the previous. This document tracks progress, milestones, and dependencies.

---

## Phase 1: MVP (Local Vault & CLI) — **~90% Complete**

**Goal:** Shippable local credential vault with full encryption engine and functional CLI.

**Status:** **IN PROGRESS** — Core crypto and vault engine complete; CLI commands 85% done; testing 80% coverage.

**Timeline:** Target completion: Week 10 (April 15, 2026)

### Milestones

| Milestone | Status | ETA | Owner |
|-----------|--------|-----|-------|
| 1.1 Crypto engine (Argon2, AES-GCM, HKDF, BIP-39) | ✅ Complete | ✓ | Crypto Team |
| 1.2 Vault engine (CRUD, indexing, versioning) | ✅ Complete | ✓ | Vault Team |
| 1.3 14 CLI commands (init, add, get, search, etc.) | ⏳ 85% | Apr 8 | CLI Team |
| 1.4 Import/export (4 sources: Chrome, Firefox, 1PWD, Bitwarden) | ⏳ 80% | Apr 10 | CLI Team |
| 1.5 Password health & HIBP integration | ⏳ 75% | Apr 12 | Health Team |
| 1.6 Recovery mnemonic (validate, test, regenerate) | ✅ Complete | ✓ | Crypto Team |
| 1.7 Test coverage (90%+ core, 80%+ CLI) | ⏳ 80% | Apr 13 | QA Team |
| 1.8 Security review & documentation | ⏳ 70% | Apr 15 | Security Team |

### Deliverables

- ✅ Encryption engine (Argon2id KDF, AES-256-GCM, HKDF per-item keys, BIP-39 recovery)
- ✅ 7 credential types (login, API key, SSH key, secure note, credit card, identity, custom)
- ✅ 14 CLI commands fully implemented
- ⏳ Import from 4 major password managers (Chrome, Firefox, 1Password, Bitwarden, CSV)
- ⏳ Export to JSON, CSV, encrypted PGP archive
- ⏳ Password health analyzer (weak, reused, old passwords)
- ⏳ HIBP integration (k-anonymity model, optional)
- ✅ BIP-39 recovery mnemonic (12 words, printable)
- ⏳ Comprehensive unit & integration tests (90%+ core modules)
- ⏳ Security documentation and threat model

### In Scope

- Local-only vault (no network required)
- Memory-safe encryption and key management
- Full-text search (SQLite FTS5)
- CLI-first UX
- Import/export for data portability
- Password health and breach detection

### Out of Scope

- **Phase 2+:** Multi-device sync
- **Phase 3+:** Browser extension, web UI
- **Phase 4+:** Mobile apps, enterprise admin

### Dependencies

None — Phase 1 is foundational and independent.

### Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Crypto implementation bugs | 🔴 Critical | Security audit, extensive test coverage, crypto team review |
| Password import failures | 🟠 High | Test with real export files, support most common formats first |
| Performance regression (>500ms load) | 🟡 Medium | Benchmark during development, optimize FTS5 queries |

### Success Criteria

- ✅ All 14 CLI commands working end-to-end
- ✅ Zero data loss across vault operations
- ✅ Vault with 1000 items loads <500ms
- ✅ 90%+ test coverage for core modules
- ✅ No security vulnerabilities discovered (peer review)
- ✅ User can go from init→add→search→export in <5 minutes

---

## Phase 2: Sync Server & Multi-Device — **~40% Complete**

**Goal:** Enable seamless credential sync across 3+ devices with conflict resolution.

**Status:** **IN PROGRESS** — Sync protocol designed; client ~50% done; server ~30% done; integration testing not started.

**Timeline:** Target start: Week 10, Duration: 6 weeks (by Week 20 = May 15, 2026)

### Milestones

| Milestone | Status | ETA | Owner |
|-----------|--------|-----|-------|
| 2.1 Sync protocol design & review | ✅ Complete | ✓ | Architecture |
| 2.2 Sync client implementation (pull/push) | ⏳ 50% | May 1 | Sync Team |
| 2.3 Sync server REST API & handlers | ⏳ 30% | May 3 | Backend Team |
| 2.4 Conflict resolution algorithm | ⏳ 60% | Apr 25 | Sync Team |
| 2.5 SQLite sync storage (WAL mode) | ⏳ 40% | May 3 | Backend Team |
| 2.6 Integration testing (multi-device) | ⏳ 0% | May 10 | QA Team |
| 2.7 Deployment guide & Docker image | ⏳ 10% | May 12 | DevOps |
| 2.8 Documentation & examples | ⏳ 5% | May 15 | Docs Team |

### Deliverables

- Sync client for multi-device coordination
- REST API sync server
- Delta sync protocol (timestamps + version vectors)
- Conflict resolution (last-write-wins → tiebreak → remote authority)
- SQLite storage with WAL mode for reliability
- Docker image for easy deployment
- Integration tests across 3+ devices
- Sync troubleshooting guide

### In Scope

- 2-way sync (pull and push)
- Automatic conflict detection & resolution
- Device registration & tracking
- Zero data loss during sync conflicts
- Optional bearer token authentication
- WAL mode for crash recovery

### Out of Scope

- Real-time sync (polling-based initially)
- End-to-end device-to-device encryption (Phase 3)
- Web UI for sync management (Phase 3)

### Dependencies

- **Requires:** Phase 1 complete (vault engine)
- **Blocks:** Phase 3 (browser extension, web sync)

### Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Conflicts cause data loss | 🔴 Critical | Extensive conflict resolution testing, version vectors, audit trail |
| Clock skew breaks logic | 🟠 High | Server-side timestamps (authority), fallback to logical clocks |
| Network flakiness loses sync state | 🟠 High | Idempotent push/pull, WAL mode for crash recovery |

### Success Criteria

- ✅ Zero data loss across sync operations
- ✅ Syncs <5 seconds for typical changes
- ✅ Conflicts automatically resolved with minimal user intervention
- ✅ Server handles 1000+ devices concurrently
- ✅ Integration tests pass on 3+ devices

---

## Phase 3: Extensions & Web App — **0% Complete**

**Goal:** Expand beyond CLI with browser access and shared vault capability.

**Status:** **NOT STARTED** — Planning phase; requires Phase 2 complete.

**Timeline:** Target start: Week 20, Duration: 8 weeks (by Week 28 = July 10, 2026)

### Milestones

| Milestone | Status | ETA | Owner |
|-----------|--------|-----|-------|
| 3.1 Browser extension architecture | ⏳ 0% | Jun 1 | Architecture |
| 3.2 Chrome extension implementation | ⏳ 0% | Jun 20 | Frontend Team |
| 3.3 Firefox extension implementation | ⏳ 0% | Jun 25 | Frontend Team |
| 3.4 WebAuthn/Passkey support | ⏳ 0% | Jun 15 | Security Team |
| 3.5 Web UI (Next.js + shadcn/ui) | ⏳ 0% | Jul 1 | Frontend Team |
| 3.6 Shared vault collaboration | ⏳ 0% | Jul 5 | Backend Team |
| 3.7 Browser extension distribution | ⏳ 0% | Jul 10 | DevOps |
| 3.8 Browser + web app tests & launch | ⏳ 0% | Jul 15 | QA Team |

### Deliverables

- Chrome extension (feature-parity with CLI)
- Firefox extension (feature-parity with CLI)
- Web UI for credential management
- Passkey/WebAuthn support
- Shared vault encryption & permissions
- Distribution on Chrome Store, Firefox Add-ons
- Integration tests (extension ↔ server)

### In Scope

- Browser password autofill
- Passkey registration & authentication
- Shared vault with granular permissions
- Web UI for account & sync settings

### Out of Scope

- Mobile browser extensions (Phase 4)
- SAML/OIDC integration (Phase 4)
- Team management dashboard (Phase 4)

### Dependencies

- **Requires:** Phase 2 complete (sync server)
- **Blocks:** Phase 4 (mobile apps, enterprise)

---

## Phase 4: Mobile & Enterprise — **0% Complete**

**Goal:** Complete ecosystem with mobile apps and enterprise features.

**Status:** **NOT STARTED** — Post-launch, based on user feedback.

**Timeline:** Target start: 2027 Q1, Duration: Ongoing

### Planned Features

- iOS app (SwiftUI, App Store distribution)
- Android app (Jetpack Compose, Play Store distribution)
- Enterprise admin dashboard
- SAML/OIDC SSO integration
- Team management & audit logging
- Advanced sharing & delegation
- On-premises deployment

### Success Metrics

- 1000+ installs on both App Store and Play Store
- 4.5+ star ratings
- 100+ enterprise customers deploying self-hosted sync server

---

## Cross-Phase Requirements

### Continuous

- **Security:** Monthly penetration testing, dependency updates, CVE monitoring
- **Performance:** Benchmark suite maintained, SLA targets (vault load <500ms, sync <5s)
- **Documentation:** Updated with each phase release
- **Community:** GitHub discussions, Discord support, public roadmap

### Architecture Decisions

| Decision | Rationale | Phases |
|----------|-----------|--------|
| Pure Go implementation | No native dependencies, single binary, cross-platform | All |
| SQLite locally, selfhostable server | Decentralization, no cloud lock-in, privacy-first | 1+ |
| All-encryption-client-side | Zero-knowledge guarantee, even server can't read data | All |
| Interface-driven DI | Testability, swappable implementations, future extensibility | All |

### Technology Debt

| Item | Severity | Phase | Notes |
|------|----------|-------|-------|
| Add integration tests (Phase 2) | 🟠 High | 2 | Critical for sync correctness |
| Benchmark & optimize FTS5 (Phase 1) | 🟡 Medium | 1 | Future-proof for 10k+ items |
| Refactor CLI middleware (Phase 1) | 🟡 Medium | 1 | Code duplication in handlers |
| Add concurrency tests (Phase 2) | 🟠 High | 2 | Sync may have race conditions |

---

## Community & Adoption Timeline

| Milestone | Target | Activity |
|-----------|--------|----------|
| MVP Launch (Phase 1) | 100 installs | GitHub release, HN post, Reddit |
| 1K GitHub Stars | Week 16 | Community blog posts, featured in newsletters |
| 1K Users | Month 4 | Sync server adoption, self-hosting tutorials |
| Browser Extension (Phase 3) | 10K users | Extension store launch, mainstream coverage |
| 10K GitHub Stars | Month 9 | Enterprise inquiries, funding conversations |

---

## Known Constraints

1. **Single Go developer** — Sequential phases, not parallel
2. **No funding** — Open-source volunteer effort
3. **No commercial infrastructure** — Self-hosted only (no managed service yet)
4. **Security audits** — Deferred until Phase 2 sync is complete

---

## How to Read This Roadmap

- **Status indicators:** ✅ = Complete, ⏳ = In Progress, ❌ = Blocked, ⚪ = Not Started
- **Phase completion % = (Completed milestones / Total milestones) × 100**
- **Target dates are estimates** subject to change based on community feedback and blockers
- **All dates are best-effort** — Security issues and blockers may shift timeline

---

## Getting Help

- **GitHub Issues:** Report bugs, request features
- **GitHub Discussions:** Ask questions, share use cases
- **Discord:** Community chat and support
- **Email:** contact@zeropass.dev (optional, future)

---

**Document Version:** 1.0  
**Last Updated:** April 1, 2026  
**Next Review:** April 8, 2026  
**Owner:** Product & Engineering
