# ZeroPass: Project Overview & Product Development Requirements

## Vision

**ZeroPass** is a developer-first, zero-knowledge credential manager purpose-built for the modern DevOps and security-conscious developer workflow. It combines military-grade encryption, offline-first architecture, and seamless CLI integration to eliminate the friction between security and productivity.

---

## Problem Statement

Developers face a critical gap in credential management:

1. **Existing PMs are bloated** — Bitwarden/1Password designed for general users; over-engineered for dev-only use cases
2. **Cloud-first architecture conflicts with offline work** — Limited functionality without connectivity
3. **Metadata exposure** — Most tools encrypt passwords but leak service names, URLs, metadata (potential fingerprinting)
4. **Poor dev integration** — No native `.env` injection, clunky CLI, browser-centric UX
5. **Vendor lock-in** — Proprietary formats, limited portability, unclear data ownership

---

## Target Users

### DevOps Dan
- Manages 50+ staging/prod API keys, SSH keys, database credentials
- Needs instant secret injection into CI/CD pipelines and dev environments
- Values speed, scriptability, offline capability
- Won't compromise on security; expects military-grade encryption out-of-the-box

### Security-First Sarah
- CISO/security engineer evaluating credential solutions
- Requires zero-knowledge architecture, auditable source code, no metadata leakage
- Values per-item encryption, recovery options, and breach detection
- Wants to self-host sync server; needs compliance documentation

### Indie Ian
- Solo developer/indie hacker managing side projects
- Needs affordable, no-subscription, privacy-respecting solution
- Values simplicity, portability, offline-first capability
- Willing to self-host infrastructure if it's straightforward

---

## Unique Value Proposition

| Feature | ZeroPass | Bitwarden | 1Password | KeePass |
|---------|----------|-----------|-----------|---------|
| **CLI-native** | ✅ | ⚠️ (secondary) | ✅ (limited) | ✅ |
| **Offline-first** | ✅ | ⚠️ (cloud sync required) | ⚠️ (cloud-first) | ✅ |
| **Zero-knowledge** | ✅ | ✅ | ✅ | ✅ |
| **Metadata encrypted** | ✅ | ❌ (encrypted, but indexable) | ❌ | ✅ |
| **Per-item keys** | ✅ | ❌ (account-wide key) | ❌ | ❌ |
| **BIP-39 recovery** | ✅ | ❌ | ❌ | ❌ |
| **Self-hosted sync** | ✅ | ✅ (Enterprise only) | ❌ | ❌ |
| **Dev secrets native** | ✅ | ❌ | ❌ | ❌ |
| **Source-available** | ✅ (MIT) | ✅ (AGPL) | ❌ | ✅ (GPL) |
| **Zero dependencies** | ✅ (pure Go) | ❌ (NodeJS) | ❌ | ✅ |

**Key differentiator:** ZeroPass is the only offline-first, CLI-native, fully zero-knowledge PM with metadata encryption, per-item keys, and BIP-39 recovery — optimized for developers and self-hosting.

---

## Scope: MVP (Phase 1)

### In MVP
- ✅ Local vault (encrypted JSON + SQLite FTS5 index)
- ✅ Full encryption engine (Argon2id KDF, AES-256-GCM, HKDF per-item keys, BIP-39 recovery)
- ✅ 7 credential types (login, API key, SSH key, secure note, credit card, identity, custom)
- ✅ CLI: init, unlock/lock, add, get, edit, delete, search, list, generate, health, recovery
- ✅ Runtime secret injection via `zp run` using `zp://item-name/field-name` references
- ✅ Import from 9 sources: Chrome, Firefox, Safari, 1Password (CSV), 1PUX, Bitwarden, LastPass, KeePass, generic CSV
- ✅ Export (JSON, CSV, encrypted) with `--force` flag and plaintext warning
- ✅ Password health analyzer (weak, reused, old)
- ✅ HIBP integration (k-anonymity model)
- ✅ Memory safety (secure wiping, unsafe.Pointer barriers)

### Out of Scope (Phase 2+)
- Sync server and delta sync (Phase 2)
- Browser extension (Phase 3)
- Web UI (Phase 3)
- Passkeys/WebAuthn (Phase 3)
- Mobile apps (Phase 4)
- Shared vaults (Phase 3)
- Enterprise admin dashboard (Phase 4)
- Interactive TUI mode (deferred)
- Recovery PDF export (deferred)

---

## Success Metrics

### Phase 1 (MVP)
- **Functionality:** All 14 CLI commands working end-to-end with 90%+ test coverage
- **Security:** No leaked metadata, proper memory wiping, all crypto reviewed
- **Performance:** Vault with 1000 items loads in <500ms
- **UX:** Time-to-value <2 minutes (init → add → get)
- **Quality:** Zero critical security issues in first release

### Phase 2 (Sync)
- **Functionality:** Delta sync working across 3+ devices with conflict resolution
- **Performance:** Sync <5 seconds for typical changes (10 items)
- **Reliability:** Zero data loss across sync conflicts
- **Security:** Server cannot read encrypted items

### Phase 3 (Extensions)
- **Adoption:** 10k+ GitHub stars, 1k+ active users
- **Feature parity:** Browser extension feature-complete with CLI
- **Web UX:** 90%+ satisfaction in usability testing

### Phase 4 (Mobile/Enterprise)
- **Availability:** iOS and Android apps in app stores
- **Enterprise:** Standardized for 100+ teams

---

## Competitive Positioning

**Market Landscape:**
- Bitwarden: Strong ecosystem, but bloated CLI, metadata indexed
- 1Password: Premium UX, but closed-source, cloud-first
- KeePass: Lightweight, but dated UX, no built-in sync
- pass: Minimal, but no GUI/mobile, limited import/export

**ZeroPass Position:** "KeePass security + Bitwarden ecosystem + CLI-first UX"

**Go-to-Market:**
1. Launch MVP on GitHub (open-source)
2. Target DevOps communities (HN, Reddit, Dev.to)
3. Offer self-hosted sync server for enterprises
4. Build ecosystem around import/export integrations
5. Phase 3: Browser extension + web UI for broader appeal

---

## Development Phases

### Phase 1: MVP (Q1-Q2 2026) — **COMPLETE**
**Goal:** Shippable local vault with crypto engine and CLI

- Local encryption engine (Argon2id, AES-256-GCM, BIP-39)
- 7 credential types, CRUD operations, FTS5 search
- 14 CLI commands implemented
- Import/export (9 sources), password health, HIBP
- Recovery key auto-rotation after recovery unlock (one-time use per PRD §8.5)
- `ValidateRecovery()` for non-destructive mnemonic validation
- macOS SwiftUI app with 9-source import, CSV export, accessibility (VoiceOver, keyboard nav, high contrast)
- Comprehensive test suite, security review

**Milestones:**
- Week 1-2: Core crypto engine (KDF, cipher, key management)
- Week 3-4: Vault engine (item storage, indexing, versioning)
- Week 5-6: CLI commands (add, get, search, edit, delete)
- Week 7-8: Import/export, health check, recovery
- Week 9-10: Testing, security review, documentation

---

### Phase 2: Sync Server (Q2-Q3 2026) — **TARGET: ~35% complete (preview)**
**Goal:** Self-hosted multi-device sync with predictable conflict resolution

- REST API sync server (HTTP handlers for pull/push)
- SQLite storage for sync state (WAL mode)
- Delta sync protocol (timestamps, integer versions)
- Conflict resolution (current server path is timestamp-first; versions remain in protocol payloads for reconciliation)
- Integration testing across devices

**Dependencies:** Phase 1 complete

---

### Phase 3: Extensions (Q3-Q4 2026)
**Goal:** Browser-based access, passkeys support

- Browser extension (Chrome, Firefox, Safari)
- Web UI (Next.js + shadcn/ui)
- WebAuthn/passkeys support
- Shared vault collaboration (encrypted shares)

**Dependencies:** Phase 2 complete

---

### Phase 4: Mobile & Enterprise (2027+)
**Goal:** Complete ecosystem coverage

- iOS app (SwiftUI)
- Android app (Jetpack Compose)
- Enterprise admin dashboard
- SAML/OIDC integration, team management

**Dependencies:** Phase 3 complete

---

## Dependencies & Constraints

### Technical
- Go 1.26+ (stable, pure Go crypto)
- SQLite (no external DB required)
- No external identity provider (self-contained)

### Market
- Open-source community for adoption
- Self-hosting capability (no cloud lock-in)
- Backward compatibility for import sources

### Timeline
- MVP hard deadline: 90 days
- Sync server: 60 days after MVP
- Browser extension: must follow sync server (requires cloud sync)

---

## Go-to-Market Strategy

1. **Soft Launch (MVP ready):** GitHub release, HN post, Reddit communities (r/devops, r/golang)
2. **Community Building:** Discord, discussions, GitHub issues feedback loop
3. **Ecosystem:** Import/export plugins, sync server Docker image
4. **Enterprise:** Self-hosted sync, audit logs, team management (Phase 3)
5. **Monetization (optional Phase 4):** Managed sync, premium support, enterprise licensing

---

## Success Definition

**MVP Success:** 
- Functional local vault with zero data loss
- 1,000 installs within 3 months
- 5+ positive mentions on HN/Reddit
- Zero security vulnerabilities discovered

**Series Success:**
- 10k+ GitHub stars by Phase 3
- 1k+ active users running sync server
- Browser extension achieving feature parity with CLI

---

## Appendix: Phased Roadmap Summary

| Phase | Status | ETA | Focus | Deliverables |
|-------|--------|-----|-------|--------------|
| 1 (MVP) | ✅ Complete | Done | Local vault + CLI + macOS app | Vault, crypto, 14 cmds, 9-source import/export, macOS SwiftUI app |
| 2 (Sync) | ~35% | Week 20 | Preview self-hosted sync | Server, timestamped delta sync, conflict resolution |
| 3 (Extensions) | 0% | Week 35 | Browser + passkeys | Extension, web UI, shared vaults |
| 4 (Mobile/Ent.) | 0% | 2027+ | Complete coverage | Mobile apps, admin dashboard |

---

**Document Version:** 1.1  
**Last Updated:** June 2025  
**Owner:** Product & Engineering
