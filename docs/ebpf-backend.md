# PacketRusher eBPF tunnel backend

Select `ue.tunnelbackend: ebpf` or `--tunnel-backend ebpf`. PacketRusher loads its
own small GTP-U programs and installs the negotiated peer, UL/DL TEIDs, QFI and
UE address. It does not require gtp5g, eUPF or PFCP.

```sh
sudo ./packetrusher --config ./config/ethernet-n3.yml --tunnel-backend ebpf \
  multi-ue -n 1 --tunnel
# Dedicated gNBs also work with --dedicatedGnb --tunnel-vrf=false.
```

This initial profile supports IPv4 PDU session 1, IPv4 N3 over an ordinary
Ethernet interface or veth, and policy routing. Assign the configured N3 address
to the interface first. N3 must have a single route, no VRF/master, and an MTU
between the required tunnel size and 1500. The UE endpoint's MTU is N3 MTU minus
44, optionally lowered by `ue.tunnelmtu` (68..1456). Linux 6.6 or newer with
TCX/LWT BPF support, CAP_BPF/CAP_NET_ADMIN, and `/dev/net/tun` are required.
Linux little-endian amd64 and arm64 builds include the BPF object; selecting this
backend needs no compiler. Runtime verifier/attachment failures are reported
without changing to another backend.

IPv6/IPv4v6, VRF, loopback N3, multiple routes, VLAN/bond/bridge attachment,
jumbo N3, outer fragmentation, IPv4 header options, GTP sequence/N-PDU options,
and extension chains beyond the single PDU Session Container are unsupported.
Choose `userspace` for IPv6/VRF. The `ue` command uses policy routing with eBPF;
other backends retain its existing VRF behavior. Applications bind the assigned
UE source IPv4 address; socket `SO_BINDTODEVICE` to the UE TUN is unsupported
because LWT outer rerouting retains the socket's bound-interface constraint. This is an experimental supported
profile, not a claim of complete replacement for every gtp5g deployment.

Uplink uses a source policy and per-session LWT route on the stable UE TUN. The
BPF program adds IPv4/UDP/GTP-U and the uplink QFI container; Linux reroutes the
outer packet through N3 and resolves its neighbour. Owned ingress TCX links
validate outer lengths/checksums, local/peer address, DL TEID, QFI and inner UE
destination before removing the encapsulation and redirecting to TUN ingress.
User-plane packets never pass through a Go reader/writer. TCX leaves other host
addresses and UDP ports untouched; the backend exclusively reserves its local
UDP/2152 so it cannot intercept another application's GTP-U socket.

Handover stages the new local port and downlink tuple, then atomically replaces
one canonical session map entry used by both directions. Staged and retired
aliases must also match that entry before delivery. The UE TUN, source rule,
routing table and application sockets stay the same. Failed staging retains
the source. A committed handover tracks unsuccessful retirement for later
cleanup instead of rolling back the newly committed mapping.

A small joined UDP management worker answers strictly validated Echo Requests
from committed peers, preserving their sequence and returning Recovery IE 0.
The backend does not initiate Echo Requests. End Markers and unsupported GTP
control messages are discarded; NAS/NGAP owns session release. Echo handling is
the only GTP work performed in Go.

Release deactivates the canonical mapping before retiring downlink aliases,
local port ownership and TCX references. Removing the TUN removes its address
before the source policy is removed. Failed map/link/socket cleanup remains
owned; unrecoverable deactivation or endpoint close retains a strong reference
to the TUN and quarantines its routing table until process exit. This prevents
interface-index or source-policy reuse after incomplete cleanup. No host qdisc,
foreign BPF link or bpffs mount is adopted or removed.

`Registry.Stats` exposes kernel uplink successes, downlink successes and owned
ingress drops. It excludes uplink rejection, management rejection and packets
dropped by subsequent kernel routing; it is not a total-drop counter.

## Rebuild and native checks

The source is `internal/control_test_engine/ue/gtp/ebpfgtp/bpf/gtpu.c`; the Apache
2.0 object is embedded beside its Go loader. Rebuild with clang 18 and Linux UAPI
headers, then commit the generated object with the source:

```sh
CLANG=clang-18 ./scripts/build-ebpf.sh
# Requires Go's race toolchain, sudo/unshare, iproute2, ethtool and /dev/net/tun.
./scripts/check-ebpf.sh
```

The checks create private network/mount namespaces and a second private fake
UPF namespace. They require real verifier loading and actual bidirectional UDP
payloads, keepalive responses, TEID/N3 handover, wrong-TEID/destination drops,
unrelated UDP passthrough, release and reinstall. Separate encoded fixtures
exercise checksum, type, QFI, length and tuple rejection through the kernel
program. Socket-free tests inject commit/cleanup failures. The
`eBPF native datapath` workflow runs this same command on relevant pull requests, pushes to main, and manual
dispatch. It fails if kernel support or packet proofs fail; ordinary unprivileged tests skip only the
explicitly gated native fixtures. No external-core interoperability claim is
made solely from these fake-UPF checks.

The loader uses [cilium/ebpf v0.22.0](https://github.com/cilium/ebpf/releases/tag/v0.22.0),
not eUPF code. The encapsulation/room-change contracts are documented in the
[Linux BPF UAPI](https://github.com/torvalds/linux/blob/master/include/uapi/linux/bpf.h);
outer rerouting is implemented by
[Linux LWT BPF](https://github.com/torvalds/linux/blob/master/net/core/lwt_bpf.c).
TCX attachment uses [cilium's owned-link API](https://pkg.go.dev/github.com/cilium/ebpf/link#AttachTCX).
