# Real-core CI

`Real core interoperability` is separate from the local AIO5GC fixtures. It uses real upstream binaries, a fresh subscriber database, SCTP/NAS registration and observable AMF deregistration. The workflow runs on the standard `ubuntu-24.04` hosted VM with Docker, SCTP and Linux TUN; it never installs `gtp5g`.

| Core | Pinned release | Required assertions |
| --- | --- | --- |
| free5GC | v4.3.0 | Authenticated registration, positive zero-session readiness, automatic deregistration, registration success=1 and PDU attempts=0 |
| Open5GS | v2.8.0 | Authenticated registration, one accepted IPv4 PDU session, persistent userspace UE tunnel, UDP echo through the real UPF in both directions, manual deregistration |

free5GC's native UPF requires its own [gtp5g module](https://github.com/free5gc/free5gc-compose/tree/v4.3.0). PacketRusher's userspace RAN tunnel does not remove that independent core dependency. This job runs the real NRF, AUSF, UDM, UDR, NSSF, PCF and AMF, with no SMF or UPF. It does not claim free5GC PDU/session or user-plane coverage. Open5GS's [Linux TUN setup](https://open5gs.org/open5gs/docs/guide/02-building-open5gs-from-sources/) supports its real UPF without gtp5g.

`multi-ue --numPduSessions 0` explicitly requests registration only. The default remains one session. Zero sessions cannot be combined with `--tunnel`. Readiness requires a live, registered UE with its assigned AMF identity and a completed NG Setup, and does not fabricate a PDU setup response. The free5GC test requires deregistration before the unresolved-readiness timeout fallback and rejects any PDU attempt, failure, cancellation or pending metric.

## Docker runner

From the repository root, with Docker/Compose, Go 1.26.2, Python 3, SCTP and `/dev/net/tun` available:

```sh
go mod download
test/external-cores/run.sh free5gc "$(mktemp -d /tmp/packetrusher-free5gc.XXXXXX)"
test/external-cores/run.sh open5gs "$(mktemp -d /tmp/packetrusher-open5gs.XXXXXX)"
```

Each run creates its own Compose project, bridge and MongoDB volume. No service publishes a host port. The Open5GS core and RAN receive `NET_ADMIN` and the TUN device; the registration-only RAN needs neither. All MongoDB processes disable Unix sockets and use a bounded 256 MiB WiredTiger cache. Provisioning targets only that run's database and uses public synthetic test credentials.

Startup is bounded and precedes the single UE attempt. free5GC must expose its NF listeners and registered NF profiles through real NRF discovery. Open5GS must expose every NF listener, register every SBI NF, and associate SMF/UPF. A loopback PCAP must contain an accepted PFCP Association Setup Response between the configured peers, while both peers remain associated. Startup logs alone cannot prove the test passed. Failed NAS/PDU attempts are never retried by the harness.

The Open5GS DN echo socket binds only the UPF-side `10.45.0.1` endpoint. The probe binds the allocated UE address `10.45.0.2`, checks its route selects PacketRusher's persistent TUN, and verifies three distinct random payloads return from the DN. The N3 PCAP must contain all three distinct payloads in uplink and downlink GTP-U packets with nonzero TEIDs and the allocated UE/DN addresses and UDP port. Real AMF state must move from no registered UEs to one and back to none. Completed PacketRusher reports must contain the exact expected registration/PDU success counts and zero failed, pending or cancelled attempts.

The workflow has a 25-minute job deadline; startup, control operations, traffic, deregistration and shutdown have shorter bounds. Cleanup addresses only the owned Compose project and volumes. Logs, generated configs, control responses, reports, result JSON and both PCAPs are uploaded even on failure. A missing peer, rejected session, incomplete report or failed cleanup fails the job.

## Existing native binaries

Native validation uses the same config generator, real subscriber provisioning and acceptance probe. It requires existing built core binaries, MongoDB `mongod`/`mongosh`, `ip`, `unshare`, `nsenter`, `tcpdump`, Python 3, SCTP and passwordless sudo. It does not build, install or modify host core services.

```sh
CGO_ENABLED=0 go build -o /tmp/packetrusher-core-ci ./cmd
test/external-cores/native.sh \
  --core open5gs --prefix /path/to/open5gs/install \
  --packetrusher /tmp/packetrusher-core-ci \
  --state "$(mktemp -d /tmp/packetrusher-open5gs-native.XXXXXX)"
# free5GC prefix must contain bin/{nrf,udr,udm,ausf,pcf,nssf,amf}
```

`native.sh` makes mount propagation private before creating an isolated network namespace and, when needed, a private `/dev` TUN node. A child RAN namespace is addressed by its owned PID; it never creates host `/run/netns` entries. It uses a fresh MongoDB TCP listener/dbpath in the parent namespace, generated NF configs and a veth between the core and RAN. Shutdown joins the owned processes and retires the veth. Timeout cleanup kills only process groups created by this runner, including descendants that still own a namespace after their leader exits. Record the native binary/source and MongoDB versions alongside results: native mode does not infer a version from a directory name.

## Pins and scope

[versions.json](../test/external-cores/versions.json) records release/source pins and public image digests. free5GC uses verified official v4.3.0 NF images; its config templates come from [free5gc-compose v4.3.0](https://github.com/free5gc/free5gc-compose/tree/887b9a139643715190cc481df0654e84256ecbe7). Subscriber authentication keys follow that release's [WebUI schema](https://github.com/free5gc/webconsole/blob/c40b94b6896109e90e46e59d57ebc5f7fa4c32a2/backend/WebUI/utils.go). Open5GS builds [v2.8.0 source](https://github.com/open5gs/open5gs/tree/157f611a530e292e40ec50f9d23f0ef5d4fcd6a6) with both otherwise moving freeDiameter/prometheus subprojects pinned. Ubuntu/MongoDB images and Actions are pinned by digest/SHA. Distro package repositories and hosted runner images can change, so this does not promise identical build bytes.

To update a core, verify its official release, source/config/subscriber schema and image digests, update the explicit pins, then repeat positive native/CI checks and the negative traffic/report guards. Do not substitute a simulated UPF or interpret an expected PDU failure as registration-only success. These jobs cover one UE, IPv4, registration and deregistration. They do not establish external-core handover, IPv6, scale or free5GC UPF interoperability.
