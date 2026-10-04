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
    ./scripts/run-ebpf-netns.sh "$1" "$2" "${3:-normal}"
  else
    sudo -n ./scripts/run-ebpf-netns.sh "$1" "$2" "${3:-normal}"
  fi
}
run_namespace "$scratch/backend.test" '^TestActualVerifierLoadDeclaredCapabilities$|^TestActualKernelPacketBoundsChecksumsAndTupleIsolation$|^TestActualKernelOptionalHeadersDelegateOnlyCurrentTuple$|^TestActualKernelFree5UPFCapturedSequenceHeader$|^TestActualKernelUDPChecksumChunkBoundaries$|^TestActualKernelIPv6PrefixIIDAndHiddenNDAdmission$|^TestActualKernelUplinkRejectsTruncatedIPv4Options$|^TestActualKernelChecksumRelayRequiresCurrentOwner$|^TestNativeChecksumStageOwnershipAndEffectiveFeatures$' restricted
run_namespace "$scratch/backend.test" '^TestNativeLoopbackRouterAdvertisementAdmission$|^TestNativeKernelChecksumAndReassemblyBeforeFallback$|^TestNativeJumboChecksumTailsAndQFI$|^TestNativeOptionalHeadersBothBackendsAndRecoveryEcho$' restricted
run_namespace "$scratch/backend.test" '^TestActualVerifierLoad$|^TestNativeManagementEchoAndJoinedClose$|^TestNativeBidirectionalHandoverAndCleanup$'
run_namespace "$scratch/backend.test" '^TestNativeRemotePeerHandoverAndCleanup$'
run_namespace "$scratch/backend.test" '^TestNativeTCPBulkTrafficAndCleanup$'
run_namespace "$scratch/service.test" '^TestNativeEBPFServiceRoutingHandoverRollbackAndRelease$|^TestNativeAutoUsesEBPFAndKeepsItAcrossHandover$'
run_namespace "$scratch/service.test" '^TestNativeDefaultFallbackWithoutBPFCapability$' no-bpf
run_namespace "$scratch/service.test" '^TestNativeEBPFServiceTCPBulkTrafficAndRelease$'
run_namespace "$scratch/service.test" '^TestNativeEBPFServiceInitialOpenFailureRetiresStaging$'
run_namespace "$scratch/service.test" '^TestNativeEBPFRealIPv6DualStackPolicyVRFAndHandover$|^TestUserspaceRealIPv6DualStackPolicyVRFAndHandover$'
run_namespace "$scratch/service.test" '^TestNativeEBPFProductionTCPFamiliesPolicyVRFAndDeviceBinding$|^TestNativeUserspaceProductionTCPFamiliesPolicyVRFAndDeviceBinding$'
run_namespace "$scratch/service.test" '^TestNativeProductionIPv6TCPVirtualOffloadAndDeviceBinding$'
run_namespace "$scratch/service.test" '^TestNativeEBPFVirtualN3ChecksumFallbackAndGatewayMark$'
run_namespace "$scratch/service.test" '^TestNativeEBPFJumboOptionsAndZeroQFI$'
run_namespace "$scratch/service.test" '^TestNativeEBPFFailedAuthorizationDisablesTrafficAndRetainsClaims$'
