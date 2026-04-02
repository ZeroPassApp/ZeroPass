#!/usr/bin/env bash
set -euo pipefail

DMG_PATH="${1:-}"

if [[ -z "$DMG_PATH" ]]; then
  echo "usage: $0 /path/to/ZeroPass.dmg" >&2
  exit 2
fi

: "${APPLE_ID:?set APPLE_ID}"
: "${APPLE_TEAM_ID:?set APPLE_TEAM_ID}"
: "${NOTARY_PASSWORD:?set NOTARY_PASSWORD (app-specific password)}"

xcrun notarytool submit "$DMG_PATH" \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$NOTARY_PASSWORD" \
  --wait

xcrun stapler staple "$DMG_PATH"
xcrun stapler validate "$DMG_PATH"

echo "Notarized + stapled: $DMG_PATH"