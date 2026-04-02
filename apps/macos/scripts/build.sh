#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
PROJ="$ROOT_DIR/apps/macos/ZeroPass/ZeroPass.xcodeproj"
SCHEME="ZeroPass"
OUT_DIR="$ROOT_DIR/dist/macos"
DERIVED_DATA="$ROOT_DIR/.derivedData-release"

mkdir -p "$OUT_DIR"

# Archive (Release). For CI/unsigned builds, set CODE_SIGNING_ALLOWED=NO.
CODE_SIGNING_ALLOWED="${CODE_SIGNING_ALLOWED:-YES}"

EXTRA_SIGN_FLAGS=()
if [[ "$CODE_SIGNING_ALLOWED" == "NO" ]]; then
  EXTRA_SIGN_FLAGS+=(CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY="")
fi

xcodebuild \
  archive \
  -project "$PROJ" \
  -scheme "$SCHEME" \
  -configuration Release \
  -destination 'generic/platform=macOS' \
  -derivedDataPath "$DERIVED_DATA" \
  CODE_SIGNING_ALLOWED="$CODE_SIGNING_ALLOWED" \
  "${EXTRA_SIGN_FLAGS[@]}" \
  -archivePath "$OUT_DIR/ZeroPass.xcarchive"

echo "Archived to: $OUT_DIR/ZeroPass.xcarchive"

echo "Tip: create a DMG via apps/macos/scripts/create-dmg.sh \"$OUT_DIR/ZeroPass.xcarchive/Products/Applications/ZeroPass.app\" $OUT_DIR/ZeroPass.dmg"