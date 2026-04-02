# Phase 9: Build & Distribution Pipeline

## Context
- Depends on: All previous phases complete

## Overview
- **Priority:** P2
- **Status:** Pending
- **Description:** Set up automated build pipeline, code signing, notarization, DMG creation, Sparkle auto-updates, and Homebrew cask distribution.

## Key Insights
- Direct distribution (DMG) — NOT Mac App Store (sandbox blocks vault access)
- Developer ID certificate required for notarization
- Sparkle 2.7+ for auto-updates (EdDSA signed)
- Homebrew cask for `brew install --cask zeropass`
- GitHub Actions: `macos-latest` runner has Xcode + Go pre-installed

## Critical Build Notes

> **Hardened Runtime Required:** The app MUST enable Hardened Runtime for notarization. Because Go runtime uses JIT-like memory management, you MUST add the entitlement `com.apple.security.cs.allow-unsigned-executable-memory` in the `.entitlements` file. Without this, the app will crash on launch when signed with Hardened Runtime.

> **Universal Binary (arm64 + x86_64):** Go `c-archive` must be built TWICE — once for each architecture — then combined with `lipo`. You CANNOT use `go build` with multiple `-arch` flags. The build script must:
> 1. `CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-archive -o libzeropass-arm64.a`
> 2. `CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-archive -o libzeropass-amd64.a`
> 3. `lipo -create -output libzeropass.a libzeropass-arm64.a libzeropass-amd64.a`
> Only one `.h` file is needed (they're identical).

> **Notarization uses `notarytool` (NOT `altool`):** Apple deprecated `altool` in Xcode 14. Use `xcrun notarytool submit ZeroPass.dmg --apple-id ... --team-id ... --password ... --wait`. The `--wait` flag blocks until notarization completes (typically 2–10 minutes). Then `xcrun stapler staple ZeroPass.dmg`.

## Related Code Files

### Files to CREATE

| File | Description |
|------|-------------|
| `macos/ZeroPass/ZeroPass.entitlements` | Hardened Runtime entitlements (allow-unsigned-executable-memory) |
| `macos/scripts/build.sh` | Full build script: Go lib (universal) → Xcode → sign → DMG |
| `macos/scripts/notarize.sh` | Notarization + staple script |
| `macos/scripts/create-dmg.sh` | DMG creation with background image |
| `.github/workflows/macos-build.yml` | GitHub Actions CI/CD |
| `macos/Sparkle/appcast.xml` | Sparkle appcast feed |
| `Casks/zeropass.rb` | Homebrew cask formula |

## Implementation Steps

### Build Script (`build.sh`)
```bash
# 1. Build Go c-archive for arm64
# 2. Build Go c-archive for amd64
# 3. lipo -create → universal libzeropass.a
# 4. Copy libzeropass.a + libzeropass.h to macos/Libraries/
# 5. xcodebuild archive (ARCHS="arm64 x86_64")
# 6. Export archive with exportOptions.plist
# 7. Code sign app + embedded frameworks with Developer ID
# 8. Create DMG
# 9. Notarize with xcrun notarytool submit --wait
# 10. Staple notarization ticket: xcrun stapler staple
```

### GitHub Actions Workflow
1. Trigger: push to `main` with tag `v*`
2. Runner: `macos-latest` (has Xcode + Go)
3. Secrets required: `DEVELOPER_ID_CERT_BASE64`, `CERT_PASSWORD`, `APPLE_ID`, `APPLE_TEAM_ID`, `NOTARY_PASSWORD` (app-specific password)
4. Steps: checkout → setup-go → build Go lib (arm64 + amd64 + lipo) → import certificate to Keychain → xcodebuild → sign → notarize → staple → DMG → create GitHub release
5. **Certificate import:** Decode base64 cert → `security import` into ephemeral Keychain → `security set-key-partition-list` to allow codesign
6. **Artifact:** Upload DMG + dSYM to release assets

### Sparkle Integration
1. Add Sparkle 2.7 via SPM
2. Configure appcast URL in Info.plist
3. EdDSA key pair for update signing
4. Host appcast.xml on GitHub Pages or releases

### Homebrew Cask
1. Create formula with SHA256 of DMG
2. Submit to homebrew/homebrew-cask or self-hosted tap

## Success Criteria
- [ ] `./build.sh` produces signed universal (arm64 + x86_64) ZeroPass.app
- [ ] Hardened Runtime enabled with `allow-unsigned-executable-memory` entitlement
- [ ] DMG opens with drag-to-Applications layout
- [ ] Notarization passes (`xcrun stapler validate`)
- [ ] App launches on BOTH Apple Silicon and Intel Macs
- [ ] Sparkle checks for updates on launch
- [ ] GitHub Actions builds on push to main with tag
- [ ] `brew install --cask zeropass` works from tap
