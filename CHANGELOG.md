# Changelog

## v2.0.0

The first release since the `20250225` snapshot.

### Highlights

- **User-plane tunnels without a kernel module.** `--tunnel` works out of the box: the new `ebpf` backend carries
  the traffic of the UEs in the kernel with eBPF programs (Linux 6.6 or later), and the new `userspace` backend
  runs anywhere. gtp5g remains supported, and `--tunnel-backend auto`, the default, takes the first available of
  `ebpf`, `gtp5g` and `userspace`.
- **As fast as the kernel module, or faster.** Through a free5GC UPF on one 16-core host, four UEs carry
  10 Gbit/s uplink and 14 Gbit/s downlink in total with `ebpf`, against 11 and 10.5 with gtp5g. The README has
  the figures.
- **IPv6 and dual-stack PDU sessions**, with `ue.pdusessiontype`.
- **Several UEs carry traffic through one gNB**: `--tunnel` no longer requires `--dedicatedGnb`.
- **UEs driven while a test runs**: `multi-ue --control-socket`, then `control` for one action (idle, reconnect,
  handover, deregister, register…) or `run-scenario` for a JSON list of them.
- **Procedure results** as a JSON report (`--report-json`) and Prometheus metrics (`--metrics-addr`).
- **Tagged releases** come with Linux amd64 and arm64 archives and a `ghcr.io/hewlettpackard/packetrusher`
  image, and `packetrusher --version` tells which one is running.

### Added

- `--tunnel-backend ebpf|gtp5g|userspace|auto`. Only `auto` falls back: a backend requested by name that is not
  available is an error.
- gNB IDs of 22 to 32 bits and a cell ID: `gnodeb.plmnlist.gnbidlength` and `cellid`.
- `multi-ue --number-of-gnbs`, for handover targets; `--numPduSessions 0`, to only register; `--loopCount` and
  `--timeBeforeReregistration`, for registration loops.
- `ue.tunnelmtu`. By default the MTU of a tunnel is the MTU of the N3 interface minus the 44 bytes of GTP-U.
- The gNB re-establishes a lost AMF association, with a new NG Setup, after releasing the UEs it served.
- A gNB whose N2 or N3 address is unavailable tries the next ones.
- Each UE reports its own IMEISV, which strict cores can now decode.
- CI registers a UE against Open5GS, with traffic through its UPF, and against free5GC.

### Changed

- gtp5g is optional, and its bundled version is 0.10.2, which builds on Linux up to 7.0.
- A handover keeps the address and the connections of the UE: the address is on `val<MSIN>`, and the tunnel
  alternates between `gtp0<MSIN>` and `gtp1<MSIN>`.
- A run removes the devices and routing rules that a killed one left behind.
- A gNB that cannot reach its AMF or complete NG Setup ends the command with an error, and Ctrl-C during startup
  is a clean exit.
- The first UE uses the configured gNB, and each following UE the next one.
- Logs are written by zap: a line is now `<time> <level> <message>`.
- Go 1.26.2 is required.

### Fixed

- NGAP messages of a UE are handled in order (#102), and one UE's failed procedure no longer stops the simulator.
- Races during fast registration and deregistration cycles (#187).
- Duplicate gNB addresses after NG Setup retries (#138).
- Uplink GTP-U packets leave from the N3 address of their gNB.
- The bundled gtp5g on recent kernels: a handover could leave a CPU dropping every packet until reboot (Linux
  7.0), and the device was no longer lockless (Linux 6.12 and later), which capped its throughput.
- The SQN is formatted on 6 bytes, and a PDU Session Resource Setup Request without NAS-PDU is accepted.
