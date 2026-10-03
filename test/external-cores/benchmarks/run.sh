#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# Explicit guest-only wrapper: private mounts/network before creating endpoints.
sudo -n unshare --mount --net -- bash -c '
  mount --make-rprivate /
  exec taskset -c 0-3 python3 "$@"
' _ "$script_dir/run.py" "$@"
