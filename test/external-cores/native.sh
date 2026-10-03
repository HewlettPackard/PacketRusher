#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
fixture_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# Both mount propagation and network ownership are private before any change.
sudo -n unshare --mount --net -- bash -c '
  mount --make-rprivate /
  if [[ ! -c /dev/net/tun ]]; then
    mount -t tmpfs -o mode=755 tmpfs /dev
    mkdir -p /dev/net
    mknod -m 666 /dev/net/tun c 10 200
    mknod -m 666 /dev/null c 1 3
    mknod -m 666 /dev/urandom c 1 9
    mknod -m 666 /dev/random c 1 8
    mknod -m 666 /dev/zero c 1 5
  fi
  exec python3 "$@"
' _ "$fixture_dir/native.py" "$@"
