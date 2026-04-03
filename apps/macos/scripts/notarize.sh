#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
usage:
  notarize.sh [--staple <path>]... <artifact> [artifact...]

auth (choose one):
  - NOTARY_KEYCHAIN_PROFILE (recommended; created via `xcrun notarytool store-credentials`)
  - App Store Connect API key: ASC_API_KEY_PATH + ASC_API_KEY_ID + ASC_API_ISSUER_ID
  - Apple ID: APPLE_ID + APPLE_TEAM_ID + NOTARY_PASSWORD (app-specific password)

notes:
  - artifacts can be .dmg, .zip, .pkg, or even a .app directory
  - stapling is optional; pass --staple for the .app and/or .dmg you want stapled
USAGE
}

STAPLE_PATHS=()
ARTIFACTS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --staple)
      shift
      [[ $# -gt 0 ]] || { echo "error: --staple requires a path" >&2; exit 2; }
      STAPLE_PATHS+=("$1")
      shift
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "error: unknown option: $1" >&2
      usage
      exit 2
      ;;
    *)
      ARTIFACTS+=("$1")
      shift
      ;;
  esac
done

for a in "$@"; do
  ARTIFACTS+=("$a")
done

if [[ ${#ARTIFACTS[@]} -eq 0 ]]; then
  usage
  exit 2
fi

NOTARY_FLAGS=()
if [[ -n "${NOTARY_KEYCHAIN_PROFILE:-}" ]]; then
  NOTARY_FLAGS+=(--keychain-profile "$NOTARY_KEYCHAIN_PROFILE")
elif [[ -n "${ASC_API_KEY_PATH:-}" && -n "${ASC_API_KEY_ID:-}" && -n "${ASC_API_ISSUER_ID:-}" ]]; then
  NOTARY_FLAGS+=(--key "$ASC_API_KEY_PATH" --key-id "$ASC_API_KEY_ID" --issuer "$ASC_API_ISSUER_ID")
else
  : "${APPLE_ID:?set APPLE_ID}"
  : "${APPLE_TEAM_ID:?set APPLE_TEAM_ID}"
  : "${NOTARY_PASSWORD:?set NOTARY_PASSWORD (app-specific password)}"
  NOTARY_FLAGS+=(--apple-id "$APPLE_ID" --team-id "$APPLE_TEAM_ID" --password "$NOTARY_PASSWORD")
fi

for artifact in "${ARTIFACTS[@]}"; do
  if [[ ! -e "$artifact" ]]; then
    echo "error: artifact not found: $artifact" >&2
    exit 2
  fi

  echo "Notarizing: $artifact"

  set +e
  out="$(xcrun notarytool submit "$artifact" "${NOTARY_FLAGS[@]}" --wait 2>&1)"
  rc=$?
  set -e

  if [[ $rc -ne 0 ]]; then
    echo "$out" >&2
    req_id="$(printf '%s\n' "$out" | sed -n 's/^id: //p' | head -n 1)"
    if [[ -n "$req_id" ]]; then
      echo "Fetching notarization log for request: $req_id" >&2
      xcrun notarytool log "$req_id" "${NOTARY_FLAGS[@]}" >&2 || true
    fi
    exit $rc
  fi

done

for p in "${STAPLE_PATHS[@]}"; do
  if [[ ! -e "$p" ]]; then
    echo "error: staple target not found: $p" >&2
    exit 2
  fi
  echo "Stapling: $p"
  xcrun stapler staple "$p"
  xcrun stapler validate "$p"
done

echo "Notarization complete."
