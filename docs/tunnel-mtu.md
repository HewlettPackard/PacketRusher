# GTP-U tunnel MTU

PacketRusher sets the inner IPv4 MTU for both shared and dedicated GTP-U
interfaces. By default it finds the host interface holding the gNB's N3 source
address and subtracts 44 bytes from that interface's MTU. An Ethernet N3 interface
with MTU 1500 therefore produces a tunnel MTU of 1456. IP aliases are supported.

The 44 bytes cover outer IPv4 (20), UDP (8), the GTP-U base header (8), optional
GTP-U fields (4), and the PDU Session Container (4). The same conservative limit
applies to sessions without QFI marking. IPv4's maximum total length also limits
the result when the N3 address is on a loopback interface.

To match a UPF configured with MTU 1400, or accommodate a lower path MTU:

```yaml
ue:
  # Keep the other UE configuration fields here.
  tunnelmtu: 1400
```

Zero or an omitted field selects automatic calculation. Explicit values must be
at least 68 and no larger than the calculated limit. PacketRusher reports a setup
error instead of silently retaining the kernel module's default when the N3
address is missing, the MTU is invalid, or the MTU update fails.

This calculation uses the local N3 interface MTU; it does not discover a smaller
MTU elsewhere in the path. Configure a smaller value when needed. A shared gNB
has one GTP-U interface, so all its UEs use the same MTU.

To validate on a Linux host with gtp5g, inspect the tunnel using `ip link show`.
With N3 MTU 1500 and an IPv4 UE address, `ping -M do -s 1428 -I UE_IP DN_IP`
produces an inner IPv4 packet of 1456 bytes (20 IPv4 + 8 ICMP + 1428 payload).
Capture on N3 and confirm the encapsulated packet is 1500 bytes without IPv4
fragmentation. Repeat with payload 1429 and verify it is rejected locally.
