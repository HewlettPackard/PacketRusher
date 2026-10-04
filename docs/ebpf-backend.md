# PacketRusher eBPF tunnel backend

Select `ue.tunnelbackend: ebpf` or `--tunnel-backend ebpf`. PacketRusher loads its
own GTP-U programs and installs the negotiated peer, UL/DL TEIDs, QFI and UE
allocation. It needs neither gtp5g nor an eUPF process. The remote core still
needs its own real UPF; this is the UE-side tunnel backend.

```sh
sudo ./packetrusher --config ./config/config.yml --tunnel-backend ebpf \
  multi-ue -n 1 --tunnel
# --tunnel-vrf selects a per-UE VRF; policy routing is also supported.
```

IPv4, IPv6 and IPv4v6 PDU session 1 use the same stable owned TUN, source routing
or per-UE VRF as `userspace`. Applications can bind its allocated address and
use `SO_BINDTODEVICE` on the UE TUN, including inside its VRF. Both portable
backends create the tunnel for PDU session 1; eBPF rejects `--numPduSessions > 1`
up front. This does not increase either backend's session scope.

N3 uses IPv4 UDP/2152. Assign its configured local address before starting.
Supported attachment types are loopback and an Ethernet interface without a
master, including veth. A single source-aware direct or IPv4-gateway route must
use that interface. Multipath routes, N3 VRF/slave devices and non-Ethernet
routed interfaces are rejected. Ordinary Ethernet/veth and loopback are tested;
this is not a claim that every bridge, bond or VLAN topology was validated.
N3 MTU may exceed 1500. The inner MTU is N3 MTU minus 44 bytes, bounded by IPv4's
65535-byte outer length; `ue.tunnelmtu` can lower it. IPv6 and dual-stack need
an inner MTU of at least 1280. Too-small target MTUs fail before changing a live
IPv6 endpoint, and failed handover restores the old MTU and attachment.

Linux 6.6 or newer with TCX, CAP_BPF/CAP_NET_ADMIN and `/dev/net/tun` is required.
Little-endian supported builds embed the object; runtime use needs no compiler.
Verifier, device, socket or attachment errors fail setup without selecting
another backend. The checked restricted-capability stage loads the real object
without CAP_SYS_ADMIN or CAP_PERFMON. Packet inspection uses bounded helper
reads and initialized stack chunks; nonzero outer UDP checksums cover the full
payload, including odd tails.

## Packet and allocation ownership

Uplink first checks the original TUN's canonical allocation, inner length and
MTU, then adds IPv4/UDP/GTP-U. QFI zero emits the eight-byte base GTP header;
other QFIs emit the uplink PDU Session Container. An owned, unaddressed veth hop
completes the inner kernel checksum before its relay redirects to N3. Only that
staging device has checksum/GSO features disabled. Underlay offload settings are
unchanged. The relay rechecks the current canonical owner, outer tuple, UL TEID,
QFI and inner source before forwarding. An explicit validated gateway avoids
inheriting the application's inner VRF or packet mark; local N3 uses a fresh
outer IPv4 route.

The UE TUN also limits TCP GSO to one segment before applications get its route.
The staging hop is necessary because TCX redirect runs before
[`validate_xmit_skb`](https://github.com/torvalds/linux/blob/v6.8/net/core/dev.c),
and TCP control packets may retain CHECKSUM_PARTIAL even when the original TUN
has checksum offload disabled. A UPF receives packet bytes, not that inner skb
metadata. The native offload-enabled IPv6 TCP proof checks complete wire
checksums and exact bidirectional application data.

Downlink TCX validates local and peer N3 ownership, UDP port, lengths, DL TEID,
QFI, direction and allocated inner destination. IPv4 options and inner
fragments retain their normal IP representation. Sequence fields with or
without the single PDU Session Container, including free5UPF's E+S format, are
supported. N-PDU flags and extension chains beyond that container are rejected;
no GTP sequence reordering is performed.

IPv6 uses the core's allocated interface ID and a validated autonomous /64 from
Router Advertisements. Solicitations, Router Advertisements and lease timers
use the shared IPv6 lifecycle with `userspace`. Direct and extension-hidden ND
is kept out of host SLAAC; fragmented ND and malformed/checksum-invalid
advertisements are rejected. The owner callback commits kernel prefix admission
before publishing the address. Expiry, withdrawal and callback failure revoke
it. Router-only loss keeps a still-valid prefix behind the routing table's
fallback barrier. Handover cannot overwrite a newer callback-owned prefix.

## Bounded fallbacks and cleanup

A joined, exclusively owned UDP socket answers validated Echo Requests and
handles IPv6 control. It also admits outer fragments after kernel reassembly,
and checksum-uncertain virtual/loopback packets after normal kernel UDP
checksum validation. TC cannot distinguish a legitimate RX PARTIAL checksum
from corrupt completed bytes by examining the wire sum alone. It delegates
that narrow case to the socket instead of bypassing checksum admission. Go
then applies the same current peer, TEID, QFI, MTU and allocation checks before
one nonblocking write to the owned TUN. Actual corrupt CHECKSUM_NONE packets
are tested to produce no injection.

Room-changing BPF helpers cap non-GSO growth below the maximum IPv4 length.
Valid inner packets larger than 8000 bytes stay on the owned TUN for normal
kernel completion and a joined UDP uplink sender. That sender checks the current
allocation and handover mapping at dispatch. Native maximum-size jumbo traffic
is validated through this path. Normal MTU-sized uplink traffic and completed
downlink packets stay in the kernel; there is no claim that every packet uses
TCX exclusively. See the helper bounds in
[Linux 6.8 filter.c](https://github.com/torvalds/linux/blob/v6.8/net/core/filter.c).

Handover stages the target socket and DL tuple, then atomically replaces the
canonical mapping. Staged and retired aliases must match that entry. The TUN,
VRF, source policies and application sockets stay stable. Failed staging keeps
the source; unsuccessful retirement after commit remains owned for cleanup.
Release cancels and joins readers and IPv6 updates, revokes the canonical map,
then retires aliases, staging links, sockets, TUN and routing objects. Invalid
prefix notification is idempotent after fully successful retirement. If map
retirement fails, the owned TUN is disabled and its descriptor/table claims are
quarantined rather than reused. Disable failures preserve both original errors.
No foreign qdisc, BPF link or bpffs mount is adopted or removed.

`Registry.Stats` returns kernel uplink/downlink redirect attempts and owned
TC ingress drops. Attempts increment before the redirect completes and are not
delivery counts. The registry's `counters` array has uint32 keys/uint64 values:
0 final uplink relay attempts; 1 downlink redirect attempts; 2 owned ingress
drops; 3 UDP checksum/reassembly delegation attempts; 4 oversized uplink
fallback attempts. Go rejection, subsequent routing drops and delivered fallback
packets are separate; these counters cannot prove that all traffic stayed in
BPF. Benchmark snapshots are taken outside the measured intervals.

## Build and checks

```sh
CLANG=clang-18 ./scripts/build-ebpf.sh
# Go race toolchain, sudo/unshare, iproute2, ethtool, VRF module, /dev/net/tun.
sudo modprobe vrf
./scripts/check-ebpf.sh
```

The runner builds the actual embedded object and production service tests,
then creates private network/mount namespaces and a private TUN node when
needed. Its checked matrix includes both backends' identical IPv4/IPv6 TCP
270336-byte exchanges, TUN/VRF/device binding, unchanged virtual N3 offloads,
IPv6/dual-stack lease/handover behavior, gateway-mark routing, jumbo options and
QFI zero, real checksum corruption/reassembly, prefix/retirement failure traffic
blocking, Echo, target refusal, release/reinstall and full resource cleanup.
The first stages drop SYS_ADMIN/PERFMON after namespace creation. The workflow
runs this command and fails if required kernel/module support or a packet proof
fails. Ordinary unprivileged tests skip explicitly gated native fixtures.
External-core interoperability and performance claims require the separate
real-core profiles; the in-process peer is an encoded transport fixture.

The loader uses [cilium/ebpf](https://github.com/cilium/ebpf), and the source and
embedded object are Apache 2.0. See the
[Linux BPF UAPI](https://github.com/torvalds/linux/blob/v6.8/include/uapi/linux/bpf.h)
for encapsulation, neighbour redirect and TCX contracts.
