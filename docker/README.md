# PacketRusher in Docker

The container uses the kernel of the Linux host: SCTP must be available on it (`sudo modprobe sctp`), and so
must the gtp5g module for user-plane tunnels on the gtp5g backend. Run the commands from the repository root.

```bash
docker build -f docker/Dockerfile --target packetrusher -t packetrusher:local .
docker run --rm packetrusher:local --help
```

The image runs `packetrusher` with the arguments of `docker run`. Its configuration is
`/packetrusher/config/config.yml`, a copy of [config/config.yml](../config/config.yml): mount yours there.

## Host network

The N2 and N3 addresses of the configuration are addresses of the host. Compose registers two UEs with
`config/config.yml`, or with the file whose absolute path is in `PACKETRUSHER_CONFIG`:

```bash
docker compose -f docker/docker-compose.yml up --build
```

User-plane tunnels (`--tunnel`) need the `NET_ADMIN` capability, with which PacketRusher creates its
interfaces, routes and rules in the network namespace of the host. Without the gtp5g module on the host they
also need the TUN device, `--device /dev/net/tun`:

```bash
docker run --rm --network host --cap-add NET_ADMIN \
  -v "$PWD/config/config.yml:/packetrusher/config/config.yml:ro" packetrusher:local multi-ue -n 2 --tunnel
```

## Docker network of the core

Give the container the N2/N3 address of its configuration on the network of the 5G core, here
`docker_privnet` (10.100.200.0/24) with the AMF at 10.100.200.201:

```bash
docker run --rm --network docker_privnet --ip 10.100.200.200 --cap-add NET_ADMIN \
  -v "$PWD/my-config.yml:/packetrusher/config/config.yml:ro" packetrusher:local multi-ue -n 2 --tunnel
```

Each additional gNB (`--dedicatedGnb`, handovers) uses the next N2 and N3 addresses, which must be added to
the interface beforehand (`ip address add`).
