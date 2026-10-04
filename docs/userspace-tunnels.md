# Userspace GTP-U tunnels

PacketRusher supports `gtp5g`, `userspace` and its own [eBPF backend](ebpf-backend.md).
`gtp5g` remains the default. The
`userspace` backend uses Linux TUN and UDP and requires no gtp5g module, build,
DKMS installation, or module access. Both support the existing tunnel options,
source routing, per-UE VRFs and handover.

```yaml
ue:
  tunnelbackend: userspace
  tunnelmtu: 1400
```

An explicit global flag overrides the YAML selection:

```sh
sudo ./packetrusher --tunnel-backend userspace ue
sudo ./packetrusher --tunnel-backend userspace multi-ue -n 100 --tunnel
sudo ./packetrusher --tunnel-backend userspace multi-ue -n 2 --tunnel -d --tunnel-vrf
```

The backend selection does not enable tunnels: `multi-ue` still requires
`--tunnel`, while `ue` enables them unless `--disableTunnel` is supplied. The
userspace backend supports a tunnel for PDU session 1, matching the kernel
backend's current restriction. Its packet queues are bounded; a slow UE drops
excess downlink traffic without blocking another UE on the same N3 socket.

Each UE gets a persistent `val<MSIN>` TUN for the PDU session. Applications bind
to the negotiated UE IPv4 address, or run with `ip vrf exec vrf<MSIN>`. Sessions
share one UDP port 2152 socket per gNB N3 address and are isolated by downlink
TEID, peer endpoint and inner destination address. Uplink encapsulation uses the
UPF TEID and QFI in an uplink PDU Session Container. The receiver validates GTP
lengths, optional/extension header chains and IP lengths; it also answers GTP-U
Echo Requests. Unrecognized TEIDs, non-IP T-PDUs and packets from unexpected UDP
peers are dropped.

Handover stages the target N3 socket and TEID before changing the uplink binding.
The TUN, UE address, policy rule and routing table remain stable. A target bind or
TEID conflict leaves the source working. Cleanup interrupts the TUN workers,
removes this session's routing objects and releases the shared N3 socket after
its final session exits. Existing interface names are never adopted or deleted.

Linux, `/dev/net/tun`, `CAP_NET_ADMIN` and an assigned IPv4 N3 address are still
required. For containers, pass `--device /dev/net/tun` and `--cap-add NET_ADMIN`
and use the documented N2/N3 network configuration. `gtp5g` and `userspace` cannot
both bind the same N3 address and port. MTU selection reserves the same complete
GTP/QFI overhead as the kernel backend; set `ue.tunnelmtu` lower for a constrained
path. The gtp5g-specific 500 ms dedicated-tunnel creation floor does not apply to
userspace tunnels.

The userspace backend carries IPv4, IPv6 and dual-stack PDU traffic over an
IPv4 N3 underlay. See [IPv6 PDU sessions](ipv6.md) for address-family selection
and UPF prefix discovery. Performance against a
real UPF can be compared using the [three-backend iperf3 benchmark](../test/external-cores/benchmarks/README.md),
which matches routing, MTU and CPU settings and records receiver throughput,
loss and CPU use. Existing metrics describe control-plane procedures, not
user-plane throughput.

For an isolated live TUN/UDP check, build the userspace and service test binaries
with `go test -race -c`, enter a fresh `unshare --net` namespace, enable loopback,
and set `PACKETRUSHER_TUN_TEST=1`. The explicit tests exchange application UDP
through the TUN and GTP socket, retain the application connection across handover,
and verify both routing and socket cleanup.
