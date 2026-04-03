# Phase 03 — P2 production readiness and release gates

## Overview

Priority: P2  
Status: completed  
Goal: make the optional sync server honestly deployable and put release gates around the full repo so shipping is repeatable.

## Outcome

- Added a real in-repo Dockerfile for `services/syncserver`.
- Hardened the container to run as a dedicated non-root user with a writable data directory.
- Added repeatable smoke-test coverage and a `services/syncserver/smoke-test.sh` script for binary and Docker operator paths.
- Updated release-facing docs with smoke-test instructions, smoke-tested binary/Docker operator guidance, a preview-grade `systemd` example, and release gates.

## Concrete tasks

1. **Add real deployment artifact(s)**
   - Create an in-repo `Dockerfile` for `services/syncserver` that matches current build/runtime flags.
   - Optionally add a minimal compose example for local operator smoke tests.
   - Ensure container runtime covers:
     - persistent DB volume
     - API key configuration
     - exposed port
     - non-root execution if practical

2. **Production-readiness pass on sync server**
   - Verify startup and shutdown behavior from `services/syncserver/main.go` in container and bare binary modes.
   - Confirm supported storage modes and operator expectations:
     - sqlite works end-to-end
     - postgres/blob options are clearly labeled experimental or supported
   - Add missing smoke/integration scenarios for:
     - auth on/off
     - restart persistence
     - conflict response behavior
     - malformed requests

3. **Document supported operating modes**
   - Deployment guide should separate:
     - verified local dev setup
     - self-hosted preview deployment
     - not-yet-hardened/unsupported production claims
   - README should describe sync as optional and current maturity level.

4. **Define ship gates and packaging matrix**
   - Release candidate must include:
     - green Go tests
     - green macOS tests
     - buildable CLI artifact
     - buildable sync server artifact
     - working container build for sync server
     - updated docs matching shipped scope
   - If sync is included in the release narrative, require at least one documented smoke-test procedure.

5. **Decide release positioning**
   - Recommended ship stance:
     - **Local vault/CLI/macOS app:** ship-ready
     - **Sync server:** self-hosted preview / experimental until deployment + test gaps close
   - Avoid marketing sync as complete until operator story and validation are real.

## Docs vs code fixes in this phase

| Mismatch | Fix docs | Fix code |
|---|---|---|
| Deployment guide shows a Dockerfile snippet, but repo has no Dockerfile | Keep docs only after artifact exists; otherwise label as example-only and not yet shipped | Add actual `Dockerfile` and validate build |
| Sync server implied as production-ready | Reposition as preview until smoke tests + operator docs exist | Add packaging/tests needed for that claim |
| Broad production storage claims (postgres/blob) exceed validated path | Narrow docs to validated path first (likely sqlite) | Only keep advanced backends if they are smoke-tested and documented |

## Definition of done

- Repo contains a real, buildable `Dockerfile` for the sync server.
- Deployment docs describe only validated operator paths.
- Sync server has a repeatable smoke-test recipe and passes it.
- Release checklist exists and distinguishes GA features from preview features.
- README/release notes position sync accurately.

## Quick wins (< 0.5 day)

- Add the minimal `Dockerfile` and verify it builds.
- Add a short release checklist section to docs/README.
- Label postgres/blob storage paths as unverified if they are not exercised yet.

## Risks

- Containerizing may uncover hidden assumptions about file paths, DB creation, or environment-variable handling.
- Attempting to harden every storage/backend path before ship can balloon scope; prioritize the single validated deployment path first.
