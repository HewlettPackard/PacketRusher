# Run PacketRusher in Docker on Linux

Run commands below from the repository root. Edit the YAML configuration before
starting: its N2/N3 addresses must exist in PacketRusher's network namespace,
and its AMF, PLMN, slice, DNN and SIM credentials must match the core. Register
both consecutive MSINs when using the example `-n 2` command.

The container uses the host kernel. SCTP must be available on the host; load it
with `sudo modprobe sctp` if necessary. User-plane tunnels additionally require
installing the bundled gtp5g module on the host and loading it with
`sudo modprobe gtp5g`. The image does not install or load kernel modules.
Docker Desktop is not covered by this Linux workflow.

## Build and inspect the CLI

```bash
docker build -f docker/Dockerfile --target packetrusher -t packetrusher:local .
docker run --rm packetrusher:local --help
docker run --rm packetrusher:local multi-ue --help
```

The image starts the PacketRusher executable directly and passes through CLI
arguments. It prints help without arguments. The default configuration lives at
`/packetrusher/config/config.yml`; mount your own file there or pass `--config`.
Configuration mounts can be read-only: there is no environment-variable rendering
or in-place editing of credentials.

## Host networking

Host networking is the default Compose example. The example configuration uses
`127.0.0.1` for N2, N3 and the AMF; change these when the core runs elsewhere.
This mode also works with a core whose containers expose N2 and N3 on the host.
Confirm that the AMF advertises a UPF N3 address reachable from PacketRusher.

```bash
docker compose -f docker/docker-compose.yml config
docker compose -f docker/docker-compose.yml up --build
```

To mount another configuration, set `PACKETRUSHER_CONFIG` to an absolute path:

```bash
PACKETRUSHER_CONFIG="$PWD/config/config.yml" \
  docker compose -f docker/docker-compose.yml up --build
```

The default command launches two UEs without user-plane interfaces; it needs no
added container capability. To create tunnels, enable the overlay after loading
gtp5g on the host:

```bash
docker compose -f docker/docker-compose.yml \
  -f docker/docker-compose.tunnel.yml up --build
```

The overlay adds `NET_ADMIN`, which permits PacketRusher to create interfaces,
addresses, routes and policy rules. With host networking these changes occur in
the host network namespace. Run this on a dedicated test host. Neither workflow
uses `privileged: true`, and gtp5g devices do not need `/dev/net/tun`.

## An existing core bridge network

Use this alternative when PacketRusher and the core share a Docker bridge. The
example expects an existing `docker_privnet` network with subnet
`10.100.200.0/24`, reserves `10.100.200.200` for PacketRusher, and uses
`10.100.200.201:38412` for the AMF. Change both the Compose address and YAML
endpoints for your network. The UPF's advertised N3 address must be reachable
on this network; published N2 ports alone are insufficient.

```bash
docker network inspect docker_privnet
docker compose -f docker/docker-compose.bridge.yml config
docker compose -f docker/docker-compose.bridge.yml up --build
```

Set `PACKETRUSHER_NETWORK` to use a differently named existing network. The same
`tunnel` overlay enables user-plane devices in the container network namespace:

```bash
docker compose -f docker/docker-compose.bridge.yml \
  -f docker/docker-compose.tunnel.yml up --build
```

No tunnel-interface health check is included: registration-only runs intentionally
have no tunnel, and interface existence does not prove registration or traffic
success. Check PacketRusher and core logs, then send traffic from the assigned
UE address, for example `ping -I <UE-IP> <data-network-IP>` inside the container.

## Multiple gNBs and handover

PacketRusher uses consecutive N2 and N3 addresses for multiple gNBs. With
`--dedicatedGnb`, there is one gNB per UE. A handover option automatically creates
at least two gNBs, even for one UE. Add all required addresses before starting
the process. The N2 and N3 addresses can use the same interface, or separate
interfaces when the core uses separate networks.

For host networking, provision aliases on the host interfaces. For example,
when your configuration starts at N2 `192.168.2.10` and N3 `192.168.3.10`, a
second gNB requires:

```bash
sudo ip address add 192.168.2.11/24 dev <N2-interface>
sudo ip address add 192.168.3.11/24 dev <N3-interface>
```

For a bridge container, Compose's static address gives only the first address
on each attached network. Add further aliases inside the container before
executing PacketRusher. This example assumes the existing network above and
an AMF at `.201`; first copy `docker/config.bridge.yml` to
`docker/config.handover.yml` and change **both** gNB addresses to `.202`.
Reserve `.202` and `.203` so no other container uses them. Replace `eth0` if
your container uses another interface name.

```bash
docker run --rm --name packetrusher-handover \
  --stop-signal SIGINT --stop-timeout 70 \
  --cap-add NET_ADMIN --network docker_privnet --ip 10.100.200.202 \
  --mount "type=bind,src=$PWD/docker/config.handover.yml,dst=/packetrusher/config/config.yml,readonly" \
  --entrypoint /bin/sh packetrusher:local -ec '
    ip address add 10.100.200.203/24 dev eth0
    exec /packetrusher/packetrusher --config /packetrusher/config/config.yml \
      multi-ue -n 1 --timeBeforeNgapHandover 5000
  '
```

Add `--tunnel` to that PacketRusher command for user-plane traffic. Without
`--dedicatedGnb`, UEs share one GTP-U device per gNB, so additional UEs alone do
not require additional N2/N3 addresses. Dedicated mode needs addresses per gNB;
`--tunnel-vrf=false` selects policy routing rather than a VRF in that mode.
Boolean values use `=`: `--tunnel-vrf=false`, not `--tunnel-vrf false`.

The local handover simulator keeps its gNB contexts in one process. Running one
gNB in each of two separate containers does not provide that shared context.

Compose and the image use `SIGINT` on stop because the current multi-UE runner
handles that signal for cleanup. They allow up to 70 seconds for tunnel release.
