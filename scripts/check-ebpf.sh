#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# No host interfaces, bpffs mounts or services are changed.
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d "${TMPDIR:-/tmp}/packetrusher-ebpf-check.XXXXXX")
trap 'rm -rf -- "$scratch"' EXIT HUP INT TERM
"${GO:-go}" test -race -c -o "$scratch/backend.test" ./internal/control_test_engine/ue/gtp/ebpfgtp
"${GO:-go}" test -race -c -o "$scratch/service.test" ./internal/control_test_engine/ue/gtp/service
run_namespace() {
  if [ "$(id -u)" -eq 0 ]; then
    ./scripts/run-ebpf-netns.sh "$1" "$2"
  else
    sudo -n ./scripts/run-ebpf-netns.sh "$1" "$2"
  fi
}
run_namespace "$scratch/backend.test" '^TestActualVerifierLoad$|^TestActualKernelPacketBoundsChecksumsAndTupleIsolation$|^TestNativeManagementEchoAndJoinedClose$|^TestNativeBidirectionalHandoverAndCleanup$'
run_namespace "$scratch/backend.test" '^TestNativeRemotePeerHandoverAndCleanup$'
run_namespace "$scratch/service.test" '^TestNativeEBPFServiceRoutingHandoverRollbackAndRelease$'
