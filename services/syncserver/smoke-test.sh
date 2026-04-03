#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-both}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE_TAG="${ZEROPASS_SMOKE_IMAGE:-zeropass-syncserver-smoke}"
API_KEY="smoke-secret-key"
ITEM_ID="smoke-item"
ITEM_PAYLOAD="smoke-payload"

TMP_DIRS=()
PIDS=()
CONTAINERS=()
VOLUMES=()

cleanup() {
  for pid in "${PIDS[@]:-}"; do
    [[ -n "${pid}" ]] || continue
    kill "${pid}" >/dev/null 2>&1 || true
    wait "${pid}" >/dev/null 2>&1 || true
  done
  for cid in "${CONTAINERS[@]:-}"; do
    [[ -n "${cid}" ]] || continue
    docker rm -f "${cid}" >/dev/null 2>&1 || true
  done
  for volume in "${VOLUMES[@]:-}"; do
    [[ -n "${volume}" ]] || continue
    docker volume rm -f "${volume}" >/dev/null 2>&1 || true
  done
  for dir in "${TMP_DIRS[@]:-}"; do
    [[ -n "${dir}" ]] || continue
    rm -rf "${dir}"
  done
}
trap cleanup EXIT

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 1; }
}

random_port() {
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
}

wait_for_health() {
  python3 - "$1" <<'PY'
import json, sys, time, urllib.request
url = sys.argv[1]
last = None
for _ in range(50):
    try:
        with urllib.request.urlopen(url, timeout=1) as resp:
            body = json.load(resp)
        if body.get("status") == "ok":
            raise SystemExit(0)
    except SystemExit:
        raise
    except Exception as exc:  # pragma: no cover - smoke helper
        last = exc
        time.sleep(0.2)
raise SystemExit(f"health check failed for {url}: {last}")
PY
}

request_json() {
  local method="$1" url="$2" auth="$3" body="${4:-}"
  local args=(-fsS -X "$method")
  [[ -n "$auth" ]] && args+=(-H "Authorization: Bearer $auth")
  [[ -n "$body" ]] && args+=(-H "Content-Type: application/json" --data "$body")
  curl "${args[@]}" "$url"
}

request_status() {
  local method="$1" url="$2" auth="$3" body="${4:-}"
  local args=(-sS -o /dev/null -w "%{http_code}" -X "$method")
  [[ -n "$auth" ]] && args+=(-H "Authorization: Bearer $auth")
  [[ -n "$body" ]] && args+=(-H "Content-Type: application/json" --data "$body")
  curl "${args[@]}" "$url"
}

assert_container_user() {
  local cid="$1" configured_user
  configured_user="$(docker inspect "$cid" --format '{{.Config.User}}')"
  [[ "$configured_user" == "10001:10001" ]] || {
    echo "expected docker container user 10001:10001, got ${configured_user:-<empty>}" >&2
    exit 1
  }
}

stop_container() {
  local cid="$1"
  docker stop "$cid" >/dev/null
  docker rm "$cid" >/dev/null
}

assert_pull_contains() {
  python3 - "$ITEM_ID" "$1" <<'PY'
import json, sys
expected = sys.argv[1]
items = json.loads(sys.argv[2]).get("items", [])
if not any(item.get("item_id") == expected for item in items):
    raise SystemExit(f"expected {expected} in pull response, got {items}")
PY
}

assert_conflict_for() {
  python3 - "$ITEM_ID" "$1" <<'PY'
import json, sys
expected = sys.argv[1]
conflicts = json.loads(sys.argv[2]).get("conflicts", [])
if not any(item.get("item_id") == expected for item in conflicts):
    raise SystemExit(f"expected conflict for {expected}, got {conflicts}")
PY
}

seed_server() {
  local base_url="$1"
  wait_for_health "$base_url/health"
  [[ "$(request_status GET "$base_url/sync/pull?device_id=peer&since=0" "")" == "401" ]] || {
    echo "expected unauthorized pull without bearer token" >&2
    exit 1
  }
  [[ "$(request_status POST "$base_url/sync/push" "$API_KEY" '{bad json')" == "400" ]] || {
    echo "expected malformed JSON push to return 400" >&2
    exit 1
  }

  request_json POST "$base_url/devices/register" "$API_KEY" '{"device_id":"smoke-device-a","device_name":"Smoke Device A"}' >/dev/null
  request_json POST "$base_url/sync/push" "$API_KEY" "{\"device_id\":\"smoke-device-a\",\"items\":[{\"item_id\":\"${ITEM_ID}\",\"version\":1,\"device_id\":\"smoke-device-a\",\"payload\":\"${ITEM_PAYLOAD}\",\"timestamp\":1000,\"checksum\":\"smoke-c1\",\"deleted\":false}]}" >/dev/null
}

verify_restart_state() {
  local base_url="$1"
  wait_for_health "$base_url/health"
  local pull_resp conflict_resp
  pull_resp="$(request_json GET "$base_url/sync/pull?device_id=smoke-device-b&since=0" "$API_KEY")"
  assert_pull_contains "$pull_resp"
  conflict_resp="$(request_json POST "$base_url/sync/push" "$API_KEY" "{\"device_id\":\"smoke-device-b\",\"items\":[{\"item_id\":\"${ITEM_ID}\",\"version\":1,\"device_id\":\"smoke-device-b\",\"payload\":\"older-payload\",\"timestamp\":500,\"checksum\":\"smoke-old\",\"deleted\":false}]}")"
  assert_conflict_for "$conflict_resp"
}

run_binary_smoke() {
  need go; need curl; need python3
  local tmp_dir bin db_path port port2 pid
  tmp_dir="$(mktemp -d)"
  TMP_DIRS+=("$tmp_dir")
  bin="$tmp_dir/syncserver"
  db_path="$tmp_dir/zeropass-sync.db"

  (cd "$ROOT_DIR" && go build -o "$bin" ./services/syncserver/)

  port="$(random_port)"
  "$bin" --port="$port" --db-path="$db_path" --api-key="$API_KEY" >"$tmp_dir/binary-1.log" 2>&1 &
  pid=$!
  PIDS+=("$pid")
  seed_server "http://127.0.0.1:$port"
  kill "$pid" >/dev/null 2>&1 || true
  wait "$pid" >/dev/null 2>&1 || true

  port2="$(random_port)"
  "$bin" --port="$port2" --db-path="$db_path" --api-key="$API_KEY" >"$tmp_dir/binary-2.log" 2>&1 &
  pid=$!
  PIDS+=("$pid")
  verify_restart_state "http://127.0.0.1:$port2"
  kill "$pid" >/dev/null 2>&1 || true
  wait "$pid" >/dev/null 2>&1 || true
  echo "✓ binary smoke test passed"
}

run_docker_smoke() {
  need docker; need curl; need python3
  local volume port port2 cid
  volume="zeropass-smoke-$RANDOM-$RANDOM"
  VOLUMES+=("$volume")
  docker build -t "$IMAGE_TAG" -f "$ROOT_DIR/services/syncserver/Dockerfile" "$ROOT_DIR" >/dev/null
  docker volume create "$volume" >/dev/null

  port="$(random_port)"
  cid="$(docker run -d -p "127.0.0.1:${port}:8443" -v "$volume:/var/lib/zeropass" "$IMAGE_TAG" --port=8443 --db-path=/var/lib/zeropass/zeropass-sync.db --api-key="$API_KEY")"
  CONTAINERS+=("$cid")
  assert_container_user "$cid"
  seed_server "http://127.0.0.1:$port"
  stop_container "$cid"

  port2="$(random_port)"
  cid="$(docker run -d -p "127.0.0.1:${port2}:8443" -v "$volume:/var/lib/zeropass" "$IMAGE_TAG" --port=8443 --db-path=/var/lib/zeropass/zeropass-sync.db --api-key="$API_KEY")"
  CONTAINERS+=("$cid")
  assert_container_user "$cid"
  verify_restart_state "http://127.0.0.1:$port2"
  stop_container "$cid"
  echo "✓ docker smoke test passed"
}

case "$MODE" in
  binary) run_binary_smoke ;;
  docker) run_docker_smoke ;;
  both)
    run_binary_smoke
    run_docker_smoke
    ;;
  *)
    echo "usage: bash services/syncserver/smoke-test.sh [binary|docker|both]" >&2
    exit 1
    ;;
esac