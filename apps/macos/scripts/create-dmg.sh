#!/usr/bin/env bash
set -euo pipefail

APP_PATH="${1:-}"
DMG_PATH="${2:-}"

if [[ -z "$APP_PATH" || -z "$DMG_PATH" ]]; then
  echo "usage: $0 /path/to/ZeroPass.app /path/to/ZeroPass.dmg" >&2
  exit 2
fi

if [[ ! -d "$APP_PATH" ]]; then
  echo "app not found: $APP_PATH" >&2
  exit 2
fi

WORK_DIR="$(mktemp -d)"
cleanup() { rm -rf "$WORK_DIR"; }
trap cleanup EXIT

VOL_NAME="ZeroPass"
STAGING="$WORK_DIR/staging"
mkdir -p "$STAGING"

cp -R "$APP_PATH" "$STAGING/"
ln -s /Applications "$STAGING/Applications"

hdiutil create \
  -volname "$VOL_NAME" \
  -srcfolder "$STAGING" \
  -ov \
  -format UDZO \
  "$DMG_PATH" >/dev/null

echo "Created DMG: $DMG_PATH"