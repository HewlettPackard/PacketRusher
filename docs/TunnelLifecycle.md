# Tunnel ownership and handover

Each PDU session owns its traffic rules and cleanup operation. Network PDU-session release and UE termination both complete that cleanup before a later registration can reuse the N3 address. Removing a kernel GTP interface alone is insufficient: its userspace worker also holds the UDP socket and must stop.

In dedicated tunnel mode, `val<MSIN>` is a stable dummy interface carrying the UE address for the PDU session. The GTP backends alternate between `gtp0<MSIN>` and `gtp1<MSIN>` during handover. PacketRusher prepares the target backend, installs its rules, replaces the default route, then retires the source backend. The UE address interface, policy rule, routing table and optional `vrf<MSIN>` persist.

Bind applications to the UE IP address:

```sh
iperf3 -B 10.1.0.1 -c IPERF_SERVER
```

With VRF mode, run applications inside the UE's VRF:

```sh
sudo ip vrf exec vrf7005551000 iperf3 -c IPERF_SERVER
```

Inspect `val<MSIN>` for the UE address and the `gtp0<MSIN>`/`gtp1<MSIN>` links for the GTP backend. Scripts that assumed `val<MSIN>` itself was a GTP device must account for the new backend names. Binding a socket to the UE IP or using its VRF are the supported traffic workflows; binding directly to the dummy interface does not select the backing GTP route.

In shared mode, UEs use their gNB's GTP device. A handover stages the UE's target address and rules before switching the route. Source cleanup leaves the gNB's device and every other UE's rules intact. Shared mode does not create a dummy interface for every UE.

Failed target setup returns the new resources while preserving the source binding. Target backend MTU is set during staging; the stable UE endpoint changes MTU just before route replacement. A failed commit restores its source MTU and reports any rollback failure. A repeated setup on the same dedicated N3 address updates the existing GTP rules without creating a second bound socket. Failed updates replay the last completed rule set. If that rollback also fails, PacketRusher releases the tunnel and reports the failure. Switching between QFI enabled and disabled requires a new PDU session; this transition is rejected before an in-place update changes any rule.

Kernel forwarding and uninterrupted application traffic across Xn/N2 handovers require validation with the deployed gtp5g version and core. The unit tests validate ownership and operation ordering with mocked kernel operations. They do not establish zero packet loss during the route switch.
