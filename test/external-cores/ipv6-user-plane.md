# Optional native Open5GS IPv6 and dual-stack acceptance

These profiles extend the pinned Open5GS v2.8.0 fixture, using the same real AMF/SMF/UPF, authentication, accepted PFCP packet, AMF0→1→0, completed NAS report and manual PDU-retirement gates. Ordinary IPv4 and hosted Docker profiles retain their defaults. Run only in the disposable guest after building a composed CLI with IPv6 support for the selected backend.

```sh
test/external-cores/native.sh --core open5gs \
  --prefix /home/tester/packetrusher-real-cores/open5gs-install \
  --packetrusher /absolute/path/to/exact/packetrusher \
  --state /absolute/fresh/open5gs-userspace-ipv6 \
  --sessions 1 --backend userspace --pdu-session-type IPv6
```

Repeat with `--backend ebpf` and with `--pdu-session-type IPv4v6`, using a fresh state directory each time. Neither portable PacketRusher backend nor the Open5GS UPF requires gtp5g. The eBPF profile disables offloads only on its owned veth pair. N3 remains IPv4 Ethernet; the PDU payload is IPv6 or dual-stack, with source policy routing and no VRF.

The profile provisions the [Open5GS IPv6 session pool](https://github.com/open5gs/open5gs/blob/157f611a530e292e40ec50f9d23f0ef5d4fcd6a6/configs/open5gs/smf.yaml.in#L37) and subscriber type2 or3. UE `2001:db8:cafe:1::2` and DN `2001:db8:cafe::1` occupy separate /64s within the /48 pool. NAS supplies only the UE IID. Acceptance requires the actual [Open5GS SMF advertisement](https://github.com/open5gs/open5gs/blob/157f611a530e292e40ec50f9d23f0ef5d4fcd6a6/src/smf/gtp-path.c#L603) in the N3 capture: link-local source, matching IID destination, hop255, valid ICMPv6 checksum, autonomous /64 PIO, positive router/prefix lifetimes and the expected learned prefix. The backend must own the resulting global /128 and source route.

Each requested family sends three distinct source-bound nonce/sequence datagrams to its real DN socket. Typed GTP-U evidence verifies that family's full UE/DN addresses, UDP lengths, sequence IDs and both directions; IPv6 additionally requires a valid nonzero UDP checksum. Dual-stack cannot pass on IPv4 alone. After both owned DN sockets close, each family must remain registered and ready while a captured uplink produces no UDP echo. ICMP errors quoting the nonce do not count. Final retirement must remove the global IPv6 address and its source policy. `result.json` records independent family results and RA evidence; captures, routes, addresses, policies and lifecycle/report artifacts remain available on failure.

Genuine free5GC v4.3.0 remains IPv4 here: the exact release SMF [initialization](https://github.com/free5gc/smf/blob/1c9d7662fb51de39df0591bad9c609ac85fec64d/internal/context/context.go#L241) hardcodes that supported type, and its [IPv6-request test](https://github.com/free5gc/smf/blob/1c9d7662fb51de39df0591bad9c609ac85fec64d/internal/sbi/processor/pdu_session_test.go#L489) expects an IPv4-only rejection. An enum or pool alone would not establish real release support.

Configuration/packet guards are unit evidence. A runtime result must identify the exact composed CLI, backend object, core binaries, harness and guest kernel; this document does not claim an unexecuted live pass.
