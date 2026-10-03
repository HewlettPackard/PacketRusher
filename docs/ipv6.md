# IPv6 PDU sessions

Set the requested PDU session family in the existing configuration:

```yaml
ue:
  pdusessiontype: IPv6   # IPv4 (default), IPv6, or IPv4v6
  tunnelbackend: userspace
  tunnelmtu: 1400       # 0 derives the MTU from the assigned N3 interface
```

```sh
sudo ./packetrusher --config config/config.yml ue
sudo ./packetrusher --config config/config.yml multi-ue -n 2 --tunnel
```

The core's subscriber and DNN configuration must allow the requested family.
An IPv4v6 request can be accepted as IPv4, IPv6, or both; a single-family request
must receive that family. All PDU session requests of a configured UE use this
selection. User traffic currently uses PDU session 1. Tunnels disabled with
`ue --disableTunnel`, or without `multi-ue --tunnel`, still exercise IPv6 NAS
negotiation without creating host networking objects.

An IPv6 NAS PDU address carries an eight-byte interface identifier, rather than
a global IPv6 address. PacketRusher uses that identifier for its link-local
address and sends Router Solicitations through the negotiated GTP-U session.
The SMF's Router Advertisement travels through the UPF; its autonomous /64
prefix provides the upper half of the UE's address. This follows
[TS 23.501 §5.8.2.2.3](https://www.etsi.org/deliver/etsi_ts/123500_123599/123501/18.06.00_60/ts_123501v180600p.pdf)
and [TS 24.501 §9.11.4.10](https://www.etsi.org/deliver/etsi_ts/124500_124599/124501/17.14.00_60/ts_124501v171400p.pdf).

Prefix discovery validates the ICMPv6 checksum, hop limit 255, link-local source,
destination, complete option lengths and autonomous /64 prefix lifetimes under
[RFC 4861](https://www.rfc-editor.org/rfc/rfc4861.html) and
[RFC 4862](https://www.rfc-editor.org/rfc/rfc4862.html). Initial setup waits for
up to three solicitations, four seconds apart. Missing or invalid advertisements
produce a setup error and remove staged routing, addresses and the TUN.
Router Advertisements are consumed by the session rather than delivered to host
SLAAC, preventing unintended global addresses and main-table default routes.
Fragmented or extended-header advertisements are rejected.

The persistent `val<MSIN>` TUN owns the negotiated IPv4 address, IPv6 address,
or both. IPv6 source policy uses the same reserved session table as IPv4;
VRF mode installs the route in the UE's VRF table. Applications bind to the
logged negotiated address, for example:

```sh
ping -6 -I 2001:db8:1234::7 2001:db8:ffff::9
```

These are documentation addresses; substitute the addresses allocated and
reachable through your core. The IPv6 packet validator requires the allocated
prefix and interface identifier. Prefix renewal refreshes address lifetimes;
renumbering stages the new address and policy before replacing the route and
retiring the old address. Expired or withdrawn allocations stop global traffic,
remove their routing, and solicit a replacement. The SMF's globally unique
per-session prefix permits skipping Duplicate Address Detection, as specified
in TS 23.501. Existing application connections survive an N3/TEID handover that
retains the IPv6 prefix.

IPv6 user traffic requires the userspace backend and a TUN MTU of at least 1280.
The CLI rejects incompatible backend or explicit MTU selections before starting
telecom sockets. Automatic MTU validation also happens before modifying a live
TUN, so a small target N3 interface cannot disable IPv6 during failed handover.
The existing gtp5g datapath remains available for IPv4. N2/N3 configuration
continues to use an IPv4 underlay; IPv6 and IPv4v6 describe the inner PDU traffic.
DHCPv6 prefix delegation and multiple simultaneous IPv6 prefixes are outside
this initial implementation.

The aio5gc mock encodes IPv6/IPv4v6 NAS addresses and matching NGAP session
types, with deterministic documentation prefixes and non-wrapping allocation.
It has no UPF. Isolated tests use a separate UDP fake UPF to exchange Router
Solicitation/Advertisement messages and actual application traffic through
Linux TUN, policy routing and VRFs. External free5GC/Open5GS interoperability
and IPv6 throughput have not been measured. Procedure metrics continue to
describe control-plane establishment, rather than user-plane prefix discovery
or throughput.
