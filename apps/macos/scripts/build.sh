#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
PROJ="$ROOT_DIR/apps/macos/ZeroPass/ZeroPass.xcodeproj"
SCHEME="ZeroPass"
OUT_DIR="$ROOT_DIR/dist/macos"
DERIVED_DATA="$ROOT_DIR/.derivedData-release"
ARCHIVE_PATH="$OUT_DIR/ZeroPass.xcarchive"

mkdir -p "$OUT_DIR"

# Archive (Release). For CI/unsigned builds, set CODE_SIGNING_ALLOWED=NO.
CODE_SIGNING_ALLOWED="${CODE_SIGNING_ALLOWED:-YES}"
CONFIGURATION="${CONFIGURATION:-Release}"
NOTARIZE="${NOTARIZE:-NO}"

XCB=(xcodebuild archive
  -project "$PROJ"
  -scheme "$SCHEME"
  -configuration "$CONFIGURATION"
  -destination 'generic/platform=macOS'
  -derivedDataPath "$DERIVED_DATA"
  CODE_SIGNING_ALLOWED="$CODE_SIGNING_ALLOWED"
)

if [[ "$CODE_SIGNING_ALLOWED" == "NO" ]]; then
  XCB+=(CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY="")
fi

if [[ -n "${CODE_SIGN_IDENTITY:-}" ]]; then
  # When overriding identity from the CLI, force manual signing to avoid
  # "conflicting provisioning settings" with Automatic signing.
  XCB+=(CODE_SIGN_STYLE=Manual)
  XCB+=(CODE_SIGN_IDENTITY="$CODE_SIGN_IDENTITY")
  # Developer ID builds do not use provisioning profiles.
  XCB+=(PROVISIONING_PROFILE_SPECIFIER="" PROVISIONING_PROFILE="")
fi

if [[ -n "${DEVELOPMENT_TEAM:-}" ]]; then
  XCB+=(DEVELOPMENT_TEAM="$DEVELOPMENT_TEAM")
fi

XCB+=(-archivePath "$ARCHIVE_PATH")

"${XCB[@]}"

APP_PATH="$ARCHIVE_PATH/Products/Applications/ZeroPass.app"
if [[ ! -d "$APP_PATH" ]]; then
  echo "error: app not found in archive: $APP_PATH" >&2
  exit 1
fi

ZIP_PATH="$OUT_DIR/ZeroPass.zip"
DMG_PATH="$OUT_DIR/ZeroPass.dmg"

DO_NOTARIZE=0
if [[ "$NOTARIZE" == "YES" || "$NOTARIZE" == "yes" || "$NOTARIZE" == "1" ]]; then
  DO_NOTARIZE=1
fi

# Create ZIP (.app)
rm -f "$ZIP_PATH"
ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"

# Notarize ZIP, then staple the .app (so the ticket is embedded).
if [[ $DO_NOTARIZE -eq 1 ]]; then
  "$(dirname "$0")/notarize.sh" --staple "$APP_PATH" "$ZIP_PATH"

  # Recreate ZIP from stapled app so offline installs include the ticket.
  rm -f "$ZIP_PATH"
  ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"
fi

# Create DMG from (potentially stapled) app
"$(dirname "$0")/create-dmg.sh" "$APP_PATH" "$DMG_PATH"

# Notarize + staple DMG
if [[ $DO_NOTARIZE -eq 1 ]]; then
  "$(dirname "$0")/notarize.sh" --staple "$DMG_PATH" "$DMG_PATH"
fi

echo "Artifacts:"
echo "  Archive: $ARCHIVE_PATH"
echo "  App:     $APP_PATH"
echo "  ZIP:     $ZIP_PATH"
echo "  DMG:     $DMG_PATH"
