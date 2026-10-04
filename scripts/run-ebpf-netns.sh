#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# Always create our own namespaces, including when invoked directly as root.
set -eu
if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  echo 'usage: run-ebpf-netns.sh TEST_BINARY TEST_PATTERN [normal|restricted|no-bpf]' >&2
  exit 2
fi
profile=${3:-normal}
case "$profile" in normal|restricted|no-bpf) ;; *) echo 'unknown capability profile' >&2; exit 2;; esac
exec unshare --net --mount sh -eu -c '
  mount --make-rprivate /
  if [ ! -c /dev/net/tun ]; then
    mount -t tmpfs -o mode=755 tmpfs /dev
    mkdir -p /dev/net
    mknod -m 666 /dev/net/tun c 10 200
    mknod -m 666 /dev/null c 1 3
    mknod -m 666 /dev/urandom c 1 9
    mknod -m 666 /dev/random c 1 8
    mknod -m 666 /dev/zero c 1 5
  fi
  ip link set lo up
  if [ "$3" = restricted ]; then
    # Namespace/device preparation needs SYS_ADMIN, the loaded datapath does not.
    exec setpriv --bounding-set=-sys_admin,-perfmon env \
      PACKETRUSHER_EBPF_TEST=1 PACKETRUSHER_EBPF_RESTRICTED_CAPS=1 \
      "$1" -test.run "$2" -test.v -test.timeout=40s
  fi
  if [ "$3" = no-bpf ]; then
    exec setpriv --bounding-set=-bpf,-sys_admin,-perfmon env \
      PACKETRUSHER_FALLBACK_TEST=1 \
      "$1" -test.run "$2" -test.v -test.timeout=40s
  fi
  PACKETRUSHER_EBPF_TEST=1 "$1" -test.run "$2" -test.v -test.timeout=40s
' sh "$1" "$2" "$profile"
