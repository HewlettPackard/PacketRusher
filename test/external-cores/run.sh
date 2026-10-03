#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
fixture_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_dir=$(cd -- "$fixture_dir/../.." && pwd)
core=${1:?usage: run.sh free5gc|open5gs /absolute/new/artifact-directory [userspace|ebpf]}
state=${2:?provide an absolute fresh artifact directory}
backend=${3:-userspace}
[[ "$core" == free5gc || "$core" == open5gs ]] || { echo 'unsupported core' >&2; exit 2; }
[[ "$backend" == userspace || "$backend" == ebpf ]] || { echo 'unsupported backend' >&2; exit 2; }
[[ "$state" == /* ]] || { echo 'artifact path must be absolute' >&2; exit 2; }
docker info >/dev/null
docker compose version
project="packetrusher-${core}-${backend}-$$"
mkdir -p "$state"
python3 "$fixture_dir/compose.py" --core "$core" --state "$state" --backend "$backend"
compose=(docker compose --project-name "$project" --file "$state/compose.json")
cleanup() {
  status=$?
  trap - EXIT
  "${compose[@]}" logs --no-color > "$state/containers.log" 2>&1 || true
  "${compose[@]}" ps --all --format json > "$state/containers.json" 2>&1 || true
  cleanup_status=0
  timeout 30s "${compose[@]}" down --volumes --remove-orphans --timeout 10 > "$state/cleanup.log" 2>&1 || cleanup_status=$?
  if [[ "$status" == 0 && "$cleanup_status" != 0 ]]; then status=$cleanup_status; fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 143' TERM
trap 'exit 130' INT
"${compose[@]}" config --quiet
mkdir -p "$fixture_dir/.runtime"
(cd "$repo_dir"; CGO_ENABLED=0 scripts/build.sh "$fixture_dir/.runtime/client")
"${compose[@]}" pull --ignore-buildable
"${compose[@]}" build
"${compose[@]}" up --detach --wait --wait-timeout 90 db
"${compose[@]}" exec --interactive=false --no-TTY db mongosh --quiet "$core" /artifacts/subscriber.js > "$state/provision.log"
if [[ "$core" == free5gc ]]; then
  "${compose[@]}" up --detach nrf udr udm ausf nssf pcf amf
else
  [[ -c /dev/net/tun ]] || { echo 'real Open5GS job requires /dev/net/tun' >&2; exit 1; }
  "${compose[@]}" up --detach core
fi
timeout --signal=TERM --kill-after=15s 180s "${compose[@]}" run --rm --no-deps --interactive=false --no-TTY ran
python3 - "$state/result.json" <<'PY'
import json,sys
result=json.load(open(sys.argv[1]))
if result.get('success') is not True:
    raise SystemExit('missing real-core success assertions')
PY
